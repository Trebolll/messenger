package handler

import (
	"messenger/internal/model"
	"messenger/internal/service"
	"messenger/internal/service/websocket"
	"messenger/internal/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService *service.UserService
	hub         *websocket.Hub
}

func NewUserHandler(userService *service.UserService, hub *websocket.Hub) *UserHandler {
	return &UserHandler{userService: userService, hub: hub}
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
		BirthDate *string `json:"birth_date"`
		Location  *string `json:"location"`
		Status    *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.userService.UpdateProfile(userID, req.Phone, req.FullName, req.BirthDate, req.Location, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
