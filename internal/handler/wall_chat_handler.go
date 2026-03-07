package handler

import (
	"encoding/json"
	"messenger/internal/model"
	"messenger/internal/repository"
	ws_hub "messenger/internal/service/websocket"
	"messenger/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WallChatHandler struct {
	wallHub     *ws_hub.WallHub
	messageRepo *repository.MessageRepository
	jwtSecret   string
}

func NewWallChatHandler(wallHub *ws_hub.WallHub, messageRepo *repository.MessageRepository, jwtSecret string) *WallChatHandler {
	return &WallChatHandler{
		wallHub:     wallHub,
		messageRepo: messageRepo,
		jwtSecret:   jwtSecret,
	}
}

// GET /api/wall/chat/:chat_id/comments — загрузить дерево комментариев
func (h *WallChatHandler) GetComments(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("chat_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный chat_id"})
		return
	}
	comments, err := h.messageRepo.GetCommentsByChatID(chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comments)
}

// POST /api/wall/chat/:chat_id/comments — отправить комментарий (требует авторизации)
func (h *WallChatHandler) PostComment(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("chat_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный chat_id"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неавторизован"})
		return
	}
	userID := val.(uuid.UUID)

	var req struct {
		Content  string     `json:"content"`
		ParentID *uuid.UUID `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "пустой контент"})
		return
	}

	msg := &model.Message{
		ChatID:   chatID,
		SenderID: userID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}
	if err := h.messageRepo.SendComment(msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Рассылаем всем в комнате чата
	h.wallHub.BroadcastToRoom(chatID, map[string]interface{}{
		"type":    "new_comment",
		"comment": msg,
	})

	// Обновляем счетчик на стене в реальном времени
	if repo, ok := h.messageRepo.GetWallRepo(); ok {
		if ownerID, err := repo.GetPostOwnerByChatID(chatID); err == nil {
			var count int
			h.messageRepo.GetDB().QueryRow(`SELECT COUNT(*) FROM messages WHERE chat_id = $1`, chatID).Scan(&count)
			h.wallHub.BroadcastToRoom(ownerID, map[string]interface{}{
				"type":           "update_post_comment_count",
				"chat_id":        chatID,
				"comments_count": count,
			})
		}
	}

	c.JSON(http.StatusCreated, msg)
}

// GET /ws/wall/:chat_id — WebSocket подключение к комнате поста
func (h *WallChatHandler) HandleWallWS(c *gin.Context) {
	chatID, err := uuid.Parse(c.Param("chat_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный chat_id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Авторизация опциональна — анонимы могут читать
	var userID *uuid.UUID
	if token := c.Query("token"); token != "" {
		if claims, err := utils.VerifyJWT(token, h.jwtSecret); err == nil {
			uid := claims.UserID
			userID = &uid
		}
	}

	client := &ws_hub.WallClient{
		ID:     uuid.New(),
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		ChatID: chatID,
	}

	h.wallHub.Register() <- client

	go client.WritePump()
	go client.ReadPump(h.wallHub, func(wc *ws_hub.WallClient, data []byte) {
		// Только авторизованные могут писать через WS
		if wc.UserID == nil {
			return
		}
		var req struct {
			Content  string     `json:"content"`
			ParentID *uuid.UUID `json:"parent_id"`
		}
		if err := json.Unmarshal(data, &req); err != nil || req.Content == "" {
			return
		}
		msg := &model.Message{
			ChatID:   wc.ChatID,
			SenderID: *wc.UserID,
			Content:  req.Content,
			ParentID: req.ParentID,
		}
		if err := h.messageRepo.SendComment(msg); err != nil {
			return
		}
		h.wallHub.BroadcastToRoom(wc.ChatID, map[string]interface{}{
			"type":    "new_comment",
			"comment": msg,
		})
	})
}

// GET /ws/wall-posts/:user_id — WebSocket подключение к стене пользователя
func (h *WallChatHandler) HandleWallPostsWS(c *gin.Context) {
	targetUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный user_id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws_hub.WallClient{
		ID:     uuid.New(),
		Conn:   conn,
		Send:   make(chan []byte, 256),
		ChatID: targetUserID, // Используем UserID как ID комнаты в WallHub
	}

	h.wallHub.Register() <- client

	go client.WritePump()
	client.ReadPump(h.wallHub, nil) // Только чтение для стены
}
