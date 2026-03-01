package handler

import (
	"messenger/internal/model"
	"messenger/internal/service"
	"messenger/internal/service/websocket"
	"messenger/internal/utils"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService    *service.UserService
	hub            *websocket.Hub
	storageService service.Storage
}

func NewUserHandler(userService *service.UserService, hub *websocket.Hub, storageService service.Storage) *UserHandler {
	return &UserHandler{userService: userService, hub: hub, storageService: storageService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var u model.User

	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ошибка": "недействительный текст запроса"})
		return
	}

	if err := h.userService.CreateUser(&u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ошибка": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"сообщение": "пользователь успешно создан"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.userService.LoginUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := utils.GenerateJWT(user.ID, "your_secret_key", 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		query = c.Query("username")
	}

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query parameter 'q' or 'username' is required"})
		return
	}

	users, err := h.userService.SearchUsers(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Phone     *string `json:"phone"`
		FullName  *string `json:"full_name"`
		Username  *string `json:"username"`
		BirthDate *string `json:"birth_date"`
		Location  *string `json:"location"`
		Status    *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.userService.UpdateProfile(userID, req.Phone, req.FullName, req.Username, req.BirthDate, req.Location, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Рассылаем обновление профиля всем онлайн
	if id, ok := userID.(uuid.UUID); ok {
		h.hub.BroadcastProfileUpdate(id, user.AvatarUrl, user.Username, user.FullName, user.Status)
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.userService.UpdateStatus(userID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast статус через WebSocket всем подключённым
	if id, ok := userID.(uuid.UUID); ok {
		h.hub.BroadcastStatusUpdate(id, req.Status)
	} else if idStr, ok := userID.(string); ok {
		if parsed, err := uuid.Parse(idStr); err == nil {
			h.hub.BroadcastStatusUpdate(parsed, req.Status)
		}
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID, exists := c.Get("userID")
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось открыть файл"})
		return
	}
	defer file.Close()

	// Генерируем уникальное имя
	path := filepath.Ext(fileHeader.Filename)
	objectName := "avatars/" + uuid.New().String() + path

	url, err := h.storageService.Upload(c.Request.Context(), objectName, file, fileHeader.Size, mimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось загрузить файл"})
		return
	}

	var uid uuid.UUID
	switch v := userID.(type) {
	case uuid.UUID:
		uid = v
	case string:
		uid, err = uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный ID"})
			return
		}
	}

	if err := h.userService.UpdateAvatarUrl(uid, url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Получаем актуальные данные пользователя для рассылки
	if user, err := h.userService.GetUserByID(uid); err == nil {
		h.hub.BroadcastProfileUpdate(uid, url, user.Username, user.FullName, user.Status)
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": url})
}
