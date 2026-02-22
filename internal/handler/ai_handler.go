package handler

import (
	"messenger/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	aiService *service.AIService
}

func NewAIHandler(aiService *service.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

func (h *AIHandler) Suggest(c *gin.Context) {
	var req service.AISuggestRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите text и action (improve | shorten | tone | continue)"})
		return
	}

	if len(req.Text) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Текст не может быть пустым"})
		return
	}

	if len(req.Text) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Текст слишком длинный (максимум 2000 символов)"})
		return
	}

	result, err := h.aiService.Suggest(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить ответ от AI"})
		return
	}

	c.JSON(http.StatusOK, result)
}
