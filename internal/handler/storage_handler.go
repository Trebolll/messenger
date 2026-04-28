package handler

import (
	"fmt"
	"messenger/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StorageHandler struct {
	storageService *service.StorageService
}

func NewStorageHandler(storageService *service.StorageService) *StorageHandler {
	return &StorageHandler{storageService: storageService}
}

func (h *StorageHandler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "файл не найден"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть файл"})
		return
	}
	defer file.Close()

	objectName := fileHeader.Filename

	url, err := h.storageService.Upload(
		c.Request.Context(),
		objectName,
		file,
		fileHeader.Size,
		fileHeader.Header.Get("Content-Type"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка загрузки"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":         url,
		"object_name": objectName,
		"filename":    fileHeader.Filename,
		"size":        fileHeader.Size,
	})
}

func (h *StorageHandler) Download(c *gin.Context) {
	objectName := c.Param("object_name")
	if objectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "имя объекта не указано"})
		return
	}

	object, size, contentType, err := h.storageService.Download(c.Request.Context(), objectName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка скачивания"})
		return
	}
	defer object.Close()

	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", size))
	// Кавычки вокруг filename решают проблему пробелов и обрезания имени
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", objectName))
	c.DataFromReader(http.StatusOK, size, contentType, object, nil)
}

func (h *StorageHandler) Delete(c *gin.Context) {
	objectName := c.Param("object_name")
	if objectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "имя объекта не указано"})
		return
	}

	if err := h.storageService.Delete(c.Request.Context(), objectName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка удаления"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "объект удален"})
}
