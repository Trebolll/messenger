package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"messenger/internal/handler"
	"messenger/internal/model"
	"messenger/internal/repository"
	"messenger/internal/service"
	"messenger/internal/service/websocket"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTestRouter(t *testing.T, db *sql.DB) (*gin.Engine, *websocket.Hub) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	hub := websocket.NewHub()

	userRepo := repository.NewUserRepository(db)
	wallRepo := repository.NewWallRepository(db)
	wallService := service.NewWallService(wallRepo)
	userService := service.NewUserService(userRepo, wallService)
	wallHub := websocket.NewWallHub()
	go wallHub.Run()
	userHandler := handler.NewUserHandler(userService, wallService, hub, wallHub, nil, "your_secret_key")

	authMiddleware := func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userID, _ := uuid.Parse(userIDStr)
		c.Set("userID", userID)
	}

	router.PUT("/users/profile", authMiddleware, userHandler.UpdateProfile)
	router.PUT("/users/status", authMiddleware, userHandler.UpdateStatus)

	return router, hub
}

func TestUpdateProfileSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupUserTestRouter(t, db)

	userID := createTestUser(t, db, "testuser", "test@example.com", "pass")

	body := []byte(`{"full_name":"New Full Name","status":"Thinking"}`)
	req := httptest.NewRequest("PUT", "/users/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response model.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "New Full Name", response.FullName)
	assert.Equal(t, "Thinking", response.Status)

	var fullName, status string
	db.QueryRow("SELECT full_name, status FROM users WHERE id = $1", userID).Scan(&fullName, &status)
	assert.Equal(t, "New Full Name", fullName)
	assert.Equal(t, "Thinking", status)
}

func TestUpdateStatusSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupUserTestRouter(t, db)

	userID := createTestUser(t, db, "testuser", "test@example.com", "pass")

	body := []byte(`{"status":"Away"}`)
	req := httptest.NewRequest("PUT", "/users/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var status string
	db.QueryRow("SELECT status FROM users WHERE id = $1", userID).Scan(&status)
	assert.Equal(t, "Away", status)
}
