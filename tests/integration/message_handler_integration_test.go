package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
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

type MockHubMessage struct {
	messages map[uuid.UUID][]websocket.Message
}

func NewMockHubMessage() *MockHubMessage {
	return &MockHubMessage{messages: make(map[uuid.UUID][]websocket.Message)}
}

func (m *MockHubMessage) SendToUser(userID uuid.UUID, msg websocket.Message) {
	m.messages[userID] = append(m.messages[userID], msg)
}

func (m *MockHubMessage) IsUserOnline(userID uuid.UUID) bool {
	return false
}

func (m *MockHubMessage) GetMessages(userID uuid.UUID) []websocket.Message {
	return m.messages[userID]
}

func setupMessageTestRouter(t *testing.T, db *sql.DB) (*gin.Engine, *MockHubMessage) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	messageRepo := repository.NewMessageRepository(db)
	chatRepo := repository.NewChatRepository(db)
	hub := NewMockHubMessage()

	messageService := service.NewMessageService(messageRepo, chatRepo, hub)
	messageHandler := handler.NewMessageHandler(messageService)

	authMiddleware := func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			return
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return
		}
		c.Set("userID", userID)
	}

	router.POST("/messages", authMiddleware, func(c *gin.Context) {
		messageHandler.SendMessage(c)
	})

	router.GET("/chats/:chat_id/messages", authMiddleware, func(c *gin.Context) {
		messageHandler.GetMessages(c)
	})

	router.PUT("/chats/:chat_id/mark-as-read", authMiddleware, func(c *gin.Context) {
		messageHandler.MarkAsRead(c)
	})

	router.PUT("/messages/:message_id", authMiddleware, func(c *gin.Context) {
		messageHandler.EditMessage(c)
	})

	router.DELETE("/messages/:message_id", authMiddleware, func(c *gin.Context) {
		messageHandler.DeleteMessage(c)
	})

	return router, hub
}

func createTestChat(t *testing.T, db *sql.DB, user1ID, user2ID uuid.UUID) uuid.UUID {
	chatID := uuid.New()
	_, err := db.Exec(
		"INSERT INTO chats (id, type) VALUES ($1, $2)",
		chatID, model.TypePrivate,
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3)",
		chatID, user1ID, user2ID,
	)
	require.NoError(t, err)
	return chatID
}

func TestSendMessageSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, hub := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	body := []byte(fmt.Sprintf(`{"chat_id":"%s","content":"Hello world"}`, chatID.String()))
	req := httptest.NewRequest("POST", "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response model.Message
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, chatID, response.ChatID)
	assert.Equal(t, user1ID, response.SenderID)
	assert.Equal(t, "Hello world", response.Content)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM messages WHERE chat_id = $1", chatID).Scan(&count)
	assert.Equal(t, 1, count, "Message should be in database")

	msgsSent := hub.GetMessages(user2ID)
	assert.Greater(t, len(msgsSent), 0, "Should notify other members")
}

func TestSendMessageInvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	userID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")

	body := []byte(`{invalid}`)
	req := httptest.NewRequest("POST", "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendMessageUnauthorized(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	body := []byte(fmt.Sprintf(`{"chat_id":"%s","content":"Hello"}`, chatID.String()))
	req := httptest.NewRequest("POST", "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSendMessageNotChatMember(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	user3ID := createTestUser(t, db, "user3", "user3@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	body := []byte(fmt.Sprintf(`{"chat_id":"%s","content":"Hello"}`, chatID.String()))
	req := httptest.NewRequest("POST", "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user3ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "доступ запрещен")
}

func TestSendMessageEmptyContent(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	body := []byte(fmt.Sprintf(`{"chat_id":"%s","content":""}`, chatID.String()))
	req := httptest.NewRequest("POST", "/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGetMessagesSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	msgID := uuid.New()
	_, err := db.Exec(
		"INSERT INTO messages (id, chat_id, sender_id, content) VALUES ($1, $2, $3, $4)",
		msgID, chatID, user1ID, "Test message",
	)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", fmt.Sprintf("/chats/%s/messages", chatID.String()), nil)
	req.Header.Set("X-User-ID", user2ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.Message
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 1, len(response))
	assert.Equal(t, "Test message", response[0].Content)
	assert.Equal(t, "user1", response[0].SenderName)

	var readAt sql.NullTime
	db.QueryRow("SELECT read_at FROM messages WHERE id = $1", msgID).Scan(&readAt)
	assert.True(t, readAt.Valid, "Message should be marked as read")
}

func TestGetMessagesMultiple(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	for i := 1; i <= 3; i++ {
		_, err := db.Exec(
			"INSERT INTO messages (id, chat_id, sender_id, content) VALUES ($1, $2, $3, $4)",
			uuid.New(), chatID, user1ID, fmt.Sprintf("Message %d", i),
		)
		require.NoError(t, err)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/chats/%s/messages", chatID.String()), nil)
	req.Header.Set("X-User-ID", user2ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.Message
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 3, len(response), "Should get all messages in order")
	assert.Equal(t, "Message 1", response[0].Content)
	assert.Equal(t, "Message 3", response[2].Content)
}

func TestGetMessagesInvalidChatID(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	userID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")

	req := httptest.NewRequest("GET", "/chats/invalid-uuid/messages", nil)
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMessagesNonexistentChat(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	userID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	nonexistentChatID := uuid.New()

	req := httptest.NewRequest("GET", fmt.Sprintf("/chats/%s/messages", nonexistentChatID.String()), nil)
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "не существует")
}

func TestGetMessagesEmpty(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	req := httptest.NewRequest("GET", fmt.Sprintf("/chats/%s/messages", chatID.String()), nil)
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.Message
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 0, len(response), "Should return empty messages")
}

func TestMarkAsReadSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, hub := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	msgID := uuid.New()
	_, err := db.Exec(
		"INSERT INTO messages (id, chat_id, sender_id, content, read_at) VALUES ($1, $2, $3, $4, NULL)",
		msgID, chatID, user1ID, "Test message",
	)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/chats/%s/mark-as-read", chatID.String()), nil)
	req.Header.Set("X-User-ID", user2ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "success", response["status"])

	var readAt sql.NullTime
	db.QueryRow("SELECT read_at FROM messages WHERE id = $1", msgID).Scan(&readAt)
	assert.True(t, readAt.Valid, "Message should be marked as read")

	msgsSent := hub.GetMessages(user1ID)
	assert.Greater(t, len(msgsSent), 0, "Should notify about messages read")
}

func TestMarkAsReadDoesNotMarkOwnMessages(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	msgID := uuid.New()
	_, err := db.Exec(
		"INSERT INTO messages (id, chat_id, sender_id, content, read_at) VALUES ($1, $2, $3, $4, NULL)",
		msgID, chatID, user1ID, "Test message",
	)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/chats/%s/mark-as-read", chatID.String()), nil)
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var readAt sql.NullTime
	db.QueryRow("SELECT read_at FROM messages WHERE id = $1", msgID).Scan(&readAt)
	assert.False(t, readAt.Valid, "Should not mark own messages as read")
}

func TestMarkAsReadInvalidChatID(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	userID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")

	req := httptest.NewRequest("PUT", "/chats/invalid-uuid/mark-as-read", nil)
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "неверный идентификатор")
}

func TestMarkAsReadUnauthorized(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/chats/%s/mark-as-read", chatID.String()), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "неавторизован")
}

func TestDeleteMessageSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, hub := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	msgID := uuid.New()
	_, err := db.Exec(
		"INSERT INTO messages (id, chat_id, sender_id, content) VALUES ($1, $2, $3, $4)",
		msgID, chatID, user1ID, "Message to delete",
	)
	require.NoError(t, err)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/messages/%s", msgID.String()), nil)
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "сообщение удалено", response["status"])

	var count int
	db.QueryRow("SELECT COUNT(*) FROM messages WHERE id = $1", msgID).Scan(&count)
	assert.Equal(t, 0, count, "Message should be deleted from DB")

	msgsSent := hub.GetMessages(user2ID)
	assert.Greater(t, len(msgsSent), 0, "Should notify other members about deletion")
}

func TestDeleteMessageForbidden(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "pass")
	chatID := createTestChat(t, db, user1ID, user2ID)

	msgID := uuid.New()
	_, err := db.Exec(
		"INSERT INTO messages (id, chat_id, sender_id, content) VALUES ($1, $2, $3, $4)",
		msgID, chatID, user1ID, "Message to delete",
	)
	require.NoError(t, err)

	// User 2 tries to delete User 1's message
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/messages/%s", msgID.String()), nil)
	req.Header.Set("X-User-ID", user2ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "не являетесь его автором")
}

func TestDeleteMessageNotFound(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupMessageTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "pass")
	nonexistentMsgID := uuid.New()

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/messages/%s", nonexistentMsgID.String()), nil)
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "сообщение не найдено")
}
