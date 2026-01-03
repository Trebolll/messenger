package handler

import (
	"messenger/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
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
	Name      string   `json:"name" binding:"required"`
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
