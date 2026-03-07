package handler

import (
	"net/http"

	"messenger/internal/model"
	"messenger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FeedHandler struct {
	feedService *service.FeedService
}

func NewFeedHandler(feedService *service.FeedService) *FeedHandler {
	return &FeedHandler{feedService: feedService}
}

// GET /api/feed?limit=30&offset=0
func (h *FeedHandler) GetFeed(c *gin.Context) {
	val, _ := c.Get("userID")
	userID, _ := val.(uuid.UUID)

	var req model.FeedRequest
	_ = c.ShouldBindQuery(&req)
	if req.Limit == 0 {
		req.Limit = 30
	}

	posts, err := h.feedService.GetFeed(userID, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// POST /api/feed/track
// Body: { post_id, event_type, watch_seconds }
// event_type: view | like | comment | video_complete | skip
func (h *FeedHandler) TrackEvent(c *gin.Context) {
	val, _ := c.Get("userID")
	userID, _ := val.(uuid.UUID)

	var req model.TrackEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// mime_type опционально из query — нужен для обновления предпочтений
	mimeType := c.Query("mime")

	// Запускаем асинхронно — клиент не ждёт
	go h.feedService.TrackEvent(userID, &req, mimeType)

	c.JSON(http.StatusAccepted, gin.H{"ok": true})
}
