package handler

import (
	"messenger/internal/model"
	"messenger/internal/service"
	"messenger/internal/service/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WallHandler struct {
	wallService    *service.WallService
	storageService *service.StorageService
	wallHub        *websocket.WallHub
}

func NewWallHandler(wallService *service.WallService, storageService *service.StorageService, wallHub *websocket.WallHub) *WallHandler {
	return &WallHandler{
		wallService:    wallService,
		storageService: storageService,
		wallHub:        wallHub,
	}
}

func (h *WallHandler) UpdateSettings(c *gin.Context) {
	var req struct {
		Bio string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неавторизован"})
		return
	}
	userID := val.(uuid.UUID)

	if err := h.wallService.UpdateSettings(userID, req.Bio); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.wallHub.BroadcastToRoom(userID, map[string]interface{}{
		"type": "update_wall_info",
		"bio":  req.Bio,
	})

	c.JSON(http.StatusOK, gin.H{"message": "настройки обновлены"})
}

func (h *WallHandler) CreatePost(c *gin.Context) {
	var p model.WallPost

	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат запроса"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неавторизован"})
		return
	}
	p.UserID = val.(uuid.UUID)

	if err := h.wallService.CreatePost(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.wallHub.BroadcastToRoom(p.UserID, map[string]interface{}{
		"type": "new_post",
		"post": p,
	})

	c.JSON(http.StatusCreated, p)
}

func (h *WallHandler) GetWall(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID пользователя"})
		return
	}

	val, _ := c.Get("userID")
	viewerID, _ := val.(uuid.UUID)

	wallRes, err := h.wallService.GetWall(userID, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallRes)
}

func (h *WallHandler) ToggleLike(c *gin.Context) {
	postIDStr := c.Param("post_id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID поста"})
		return
	}
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неавторизован"})
		return
	}
	userID := val.(uuid.UUID)

	liked, count, err := h.wallService.ToggleLike(postID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ownerID, err := h.wallService.GetPostOwner(postID)
	if err == nil {
		h.wallHub.BroadcastToRoom(ownerID, map[string]interface{}{
			"type":        "update_post_like",
			"post_id":     postID,
			"likes_count": count,
		})
	}

	c.JSON(http.StatusOK, gin.H{"liked": liked, "likes_count": count})
}

func (h *WallHandler) GetPostChat(c *gin.Context) {
	postIDStr := c.Param("post_id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID поста"})
		return
	}
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неавторизован"})
		return
	}

	chatID, err := h.wallService.GetPostChat(postID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "чат не найден"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"chat_id": chatID})
}

func (h *WallHandler) DeletePost(c *gin.Context) {
	postIDStr := c.Param("post_id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID поста"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неавторизован"})
		return
	}
	userID := val.(uuid.UUID)

	ownerID, err := h.wallService.DeletePost(postID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.wallHub.BroadcastToRoom(ownerID, map[string]interface{}{
		"type":    "delete_post",
		"post_id": postID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "пост удалён"})
}

func (h *WallHandler) DeleteAttachment(c *gin.Context) {
	attIDStr := c.Param("attachment_id")
	attID, err := uuid.Parse(attIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID вложения"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неавторизован"})
		return
	}
	userID := val.(uuid.UUID)

	if err := h.wallService.DeleteAttachment(attID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "медиа удалено"})
}

func (h *WallHandler) UploadAttachment(c *gin.Context) {
	postIDStr := c.Param("post_id")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID поста"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл не найден"})
		return
	}

	// Open file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть файл"})
		return
	}
	defer f.Close()

	// Upload to storage
	objectName := uuid.New().String()
	url, err := h.storageService.Upload(c.Request.Context(), objectName, f, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка загрузки файла"})
		return
	}

	att := model.WallAttachment{
		PostID:    postID,
		Url:       url,
		Filename:  file.Filename,
		MimeType:  file.Header.Get("Content-Type"),
		SizeBytes: file.Size,
	}

	if err := h.wallService.AddAttachment(&att); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast update to wall
	ownerID, err := h.wallService.GetPostOwner(postID)
	if err == nil {
		h.wallHub.BroadcastToRoom(ownerID, map[string]interface{}{
			"type":       "update_post_attachment",
			"post_id":    postID,
			"attachment": att,
		})
	}

	c.JSON(http.StatusOK, att)
}

func (h *WallHandler) GetGlobalMediaFeed(c *gin.Context) {
	val, _ := c.Get("userID")
	viewerID, _ := val.(uuid.UUID)

	posts, err := h.wallService.GetGlobalMediaFeed(viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, posts)
}
