package handler

import (
	"messenger/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttachmentHandler struct {
	attachmentService *service.AttachmentService
}

func NewAttachmentHandler(attachmentService *service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{attachmentService: attachmentService}
}

func (h *AttachmentHandler) Upload(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := uuid.Parse(chatIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID чата"})
		return
	}
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "неавторизован"})
		return
	}
	senderID := val.(uuid.UUID)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл не найден в запросе"})
		return
	}

	if fileHeader.Size > 50<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл слишком большой, максимум 50MB"})
		return
	}

	allowed := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true,
		"audio/webm": true, "audio/ogg": true, "audio/mpeg": true,
		"video/webm": true, "video/mp4": true,
	}
	mimeType := fileHeader.Header.Get("Content-Type")
	if !allowed[mimeType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "тип файла не поддерживается"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось прочитать файл"})
		return
	}
	defer file.Close()

	var messageID *uuid.UUID
	if msgIDStr := c.PostForm("message_id"); msgIDStr != "" {
		if parsed, err := uuid.Parse(msgIDStr); err == nil {
			messageID = &parsed
		}
	}

	attachment, err := h.attachmentService.Upload(
		c.Request.Context(),
		senderID,
		chatID,
		messageID,
		file,
		fileHeader.Filename,
		fileHeader.Size,
		fileHeader.Header.Get("Content-Type"),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

func (h *AttachmentHandler) Download(c *gin.Context) {
	attachmentIDStr := c.Param("id")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID вложения"})
		return
	}

	content, size, contentType, filename, err := h.attachmentService.Download(c.Request.Context(), attachmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer content.Close()

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.DataFromReader(http.StatusOK, size, contentType, content, nil)
}

func (h *AttachmentHandler) Delete(c *gin.Context) {
	attachmentIDStr := c.Param("id")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID вложения"})
		return
	}

	if err := h.attachmentService.Delete(c.Request.Context(), attachmentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "вложение удалено"})
}
