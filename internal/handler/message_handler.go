package handler

import (
	"messenger/internal/model"
	"messenger/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	var m model.Message

	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ошибка": "недействительный текст запроса"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"ошибка": "неавторизован"})
		return
	}
	m.SenderID = val.(uuid.UUID)

	if err := h.messageService.SendMessage(&m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ошибка": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, m)
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ошибка": "неверный идентификатор чата"})
		return
	}

	val, _ := c.Get("userID")
	userID := val.(uuid.UUID)

	// 1. Помечаем как прочитанные
	// Мы делаем это ПЕРЕД получением, чтобы в ответе эти сообщения могли уже иметь статус прочитанных (по желанию)
	_ = h.messageService.MarkChatAsRead(chatID, userID)

	// 2. Получаем историю сообщений
	messages, err := h.messageService.GetMessagesByChatID(chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ошибка": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	chatID, _ := uuid.Parse(c.Param("chat_id"))
	val, _ := c.Get("userID")
	userID := val.(uuid.UUID)

	if err := h.messageService.MarkChatAsRead(chatID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
