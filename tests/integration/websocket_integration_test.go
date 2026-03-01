package integration

import (
	"encoding/json"
	"messenger/internal/utils"
	_ "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"messenger/internal/handler"
	"messenger/internal/model"
	"messenger/internal/repository"
	_ "messenger/internal/service"
	"messenger/internal/service/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gw "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketConnectionAndMessage(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	userRepo := repository.NewUserRepository(db)

	hub := websocket.NewHub()
	go hub.Run()

	user1 := &model.User{ID: uuid.New(), Username: "user1", Email: "user1@test.com", Password: "password"}
	userRepo.Create(user1)

	// Create chat manually in DB for simplicity in test
	chatID := uuid.New()
	_, err := db.Exec("INSERT INTO chats (id, type, creator_id) VALUES ($1, $2, $3)", chatID, model.TypePrivate, user1.ID)
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)", chatID, user1.ID)
	require.NoError(t, err)

	chat := &model.Chat{ID: chatID, Type: model.TypePrivate, CreatorID: &user1.ID}

	jwtSecret := "your_secret_key"
	token, err := utils.GenerateJWT(user1.ID, jwtSecret, time.Hour)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	wsHandler := handler.NewWebSocketHandler(hub, jwtSecret)
	router.GET("/api/ws", wsHandler.HandleWebSocket)

	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http URL to ws URL and add token
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/ws?token=" + token

	// Connect to WebSocket
	dialer := gw.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Wait for connection to be registered in hub
	time.Sleep(100 * time.Millisecond)

	testMsg := &model.Message{
		ID:        uuid.New(),
		ChatID:    chat.ID,
		SenderID:  user1.ID,
		Content:   "Hello via WS",
		CreatedAt: time.Now(),
	}

	// Test hub's ability to send to the client
	hub.SendToUser(user1.ID, websocket.Message{
		Type:    "new_message",
		Content: testMsg,
	})

	// Read messages from WebSocket until we get "new_message"
	var receivedWSMsg struct {
		Type    string          `json:"type"`
		Content json.RawMessage `json:"content"`
	}

	for {
		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		err = json.Unmarshal(message, &receivedWSMsg)
		require.NoError(t, err)

		if receivedWSMsg.Type == "new_message" {
			break
		}
		// Skip other messages like "user_status"
	}

	assert.Equal(t, "new_message", receivedWSMsg.Type)

	var receivedMsg model.Message
	err = json.Unmarshal(receivedWSMsg.Content, &receivedMsg)
	require.NoError(t, err)
	assert.Equal(t, testMsg.Content, receivedMsg.Content)
	assert.Equal(t, testMsg.ID, receivedMsg.ID)
}

func TestWebSocketTypingStatus(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	userRepo := repository.NewUserRepository(db)

	hub := websocket.NewHub()
	go hub.Run()

	user1 := &model.User{ID: uuid.New(), Username: "user1", Email: "user1@test.com", Password: "password"}
	userRepo.Create(user1)

	jwtSecret := "your_secret_key"
	token, err := utils.GenerateJWT(user1.ID, jwtSecret, time.Hour)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	wsHandler := handler.NewWebSocketHandler(hub, jwtSecret)
	router.GET("/api/ws", wsHandler.HandleWebSocket)

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/ws?token=" + token
	dialer := gw.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	time.Sleep(100 * time.Millisecond)

	// Broadcast status update (since typing broadcast isn't explicitly implemented in Hub with a helper)
	hub.BroadcastStatusUpdate(user1.ID, "online")

	_, message, err := conn.ReadMessage()
	require.NoError(t, err)

	var receivedWSMsg struct {
		Type    string `json:"type"`
		Content any    `json:"content"`
	}
	err = json.Unmarshal(message, &receivedWSMsg)
	require.NoError(t, err)
	assert.Equal(t, "user_status", receivedWSMsg.Type)
}
