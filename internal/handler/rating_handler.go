package handler

import (
	"messenger/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RatingHandler struct {
	ratingService *service.RatingService
}

func NewRatingHandler(ratingService *service.RatingService) *RatingHandler {
	return &RatingHandler{ratingService: ratingService}
}

// POST /api/messages/:message_id/vote  body: {"vote": 1} или {"vote": -1}
func (h *RatingHandler) Vote(c *gin.Context) {
	messageIDStr := c.Param("message_id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID сообщения"})
		return
	}

	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неавторизован"})
		return
	}
	voterID := val.(uuid.UUID)

	var req struct {
		Vote int `json:"vote" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Vote != 1 && req.Vote != -1) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vote должен быть 1 или -1"})
		return
	}

	if err := h.ratingService.Vote(messageID, voterID, req.Vote); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GET /api/users/:user_id/rating
func (h *RatingHandler) GetUserRating(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID пользователя"})
		return
	}

	rating, err := h.ratingService.GetUserRating(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rating": rating})
}
