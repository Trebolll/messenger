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

// GET /api/ai/agents — список доступных агентов
func (h *AIHandler) Agents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"agents": service.Agents})
}

// POST /api/ai/suggest
func (h *AIHandler) Suggest(c *gin.Context) {
	var req service.AISuggestRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите text и action"})
		return
	}

	if len(req.Text) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Текст не может быть пустым"})
		return
	}

	maxLen := 2000
	if req.Action == "reply" {
		maxLen = 8000
	}
	if len(req.Text) > maxLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Текст слишком длинный"})
		return
	}

	result, err := h.aiService.Suggest(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить ответ: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
