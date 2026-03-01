package handler

import (
	"messenger/internal/handler"
	"messenger/internal/service/websocket"
	"messenger/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHandleWebSocket_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := websocket.NewHub()
	h := handler.NewWebSocketHandler(hub, "secret")

	r := gin.New()
	r.GET("/ws", h.HandleWebSocket)

	// No token
	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Invalid token
	req = httptest.NewRequest("GET", "/ws?token=invalid", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleWebSocket_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := websocket.NewHub()
	secret := "secret"
	h := handler.NewWebSocketHandler(hub, secret)

	r := gin.New()
	r.GET("/ws", h.HandleWebSocket)

	userID := uuid.New()
	token, _ := utils.GenerateJWT(userID, secret, time.Hour)

	// Valid token but Upgrade will fail because it's not a real WS request
	// However, we can check that it doesn't return 401
	req := httptest.NewRequest("GET", "/ws?token="+token, nil)
	w := httptest.NewRecorder()

	// We expect a failure in Upgrade, but not 401 before that
	r.ServeHTTP(w, req)

	// Since it's not a websocket request, Upgrade returns error and handler returns
	// Usually gin returns 200 if nothing else is set, but since upgrade failed
	// without explicit JSON error in the current implementation (line 44):
	/*
		if err != nil {
			return
		}
	*/
	// It will stay 200 (default) but without the upgrade headers.
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}
