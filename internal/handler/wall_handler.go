package handler

import (
	"messenger/internal/model"
	"messenger/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WallHandler struct {
	wallService    *service.WallService
	storageService *service.StorageService
}

func NewWallHandler(wallService *service.WallService, storageService *service.StorageService) *WallHandler {
	return &WallHandler{
		wallService:    wallService,
		storageService: storageService,
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

	c.JSON(http.StatusCreated, p)
}

func (h *WallHandler) GetWall(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID пользователя"})
		return
	}

	wallRes, err := h.wallService.GetWall(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallRes)
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

	c.JSON(http.StatusOK, att)
}
