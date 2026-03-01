package middleware

import (
	"messenger/internal/middleware"
	"messenger/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test_secret"
	userID := uuid.New()
	token, _ := utils.GenerateJWT(userID, secret, time.Hour)

	r := gin.New()
	r.Use(middleware.AuthMiddleware(secret))
	r.GET("/test", func(c *gin.Context) {
		val, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, userID, val)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test_secret"

	r := gin.New()
	r.Use(middleware.AuthMiddleware(secret))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test_secret"

	r := gin.New()
	r.Use(middleware.AuthMiddleware(secret))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_TokenInQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test_secret"
	userID := uuid.New()
	token, _ := utils.GenerateJWT(userID, secret, time.Hour)

	r := gin.New()
	r.Use(middleware.AuthMiddleware(secret))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test?token="+token, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
