package handler

import (
	"messenger/internal/service"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chatService    *service.ChatService
	storageService *service.StorageService
}

func NewChatHandler(chatService *service.ChatService, storageService *service.StorageService) *ChatHandler {
	return &ChatHandler{chatService: chatService, storageService: storageService}
}

type CreatePrivateChatRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

func (h *ChatHandler) CreatePrivateChat(c *gin.Context) {
	var req CreatePrivateChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Получаем текущего пользователя из токена
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	currentUserID := val.(uuid.UUID)

	chat, err := h.chatService.CreatePrivateChat(currentUserID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chat)
}

type CreateGroupChatRequest struct {
	Name      string   `json:"name"`
	Usernames []string `json:"usernames" binding:"required"`
}

func (h *ChatHandler) CreateGroupChat(c *gin.Context) {
	var req CreateGroupChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Получаем создателя группы из токена
	val, _ := c.Get("userID")
	creatorID := val.(uuid.UUID)

	chat, err := h.chatService.CreateGroupChatByUsernames(req.Name, req.Usernames, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chat)
}

func (h *ChatHandler) GetUserChats(c *gin.Context) {
	// Получаем userID из контекста (который установил JWT middleware)
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := val.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	chats, _ := h.chatService.GetUserChats(userID)

	c.JSON(http.StatusOK, chats)
}

func (h *ChatHandler) UpdateGroupAvatar(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID чата"})
		return
	}

	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл не найден"})
		return
	}

	// Валидация размера (макс 5MB)
	if fileHeader.Size > 5<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл слишком большой, максимум 5MB"})
		return
	}

	// Валидация типа
	mimeType := fileHeader.Header.Get("Content-Type")
	allowed := map[string]bool{"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true}
	if !allowed[mimeType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "разрешены только изображения (jpg, png, gif, webp)"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось прочитать файл"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	objectName := "group-avatars/" + uuid.New().String() + ext

	url, err := h.storageService.Upload(c.Request.Context(), objectName, file, fileHeader.Size, mimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить файл"})
		return
	}

	if err := h.chatService.UpdateGroupAvatarUrl(chatID, url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": url})
}
