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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockHub struct {
	onlineUsers map[uuid.UUID]bool
}

func NewMockHub() *MockHub {
	return &MockHub{onlineUsers: make(map[uuid.UUID]bool)}
}

func (m *MockHub) IsUserOnline(userID uuid.UUID) bool {
	return m.onlineUsers[userID]
}

func (m *MockHub) SetUserOnline(userID uuid.UUID, online bool) {
	m.onlineUsers[userID] = online
}

func setupChatTestRouter(t *testing.T, db *sql.DB) (*gin.Engine, *MockHub) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	chatRepo := repository.NewChatRepository(db)
	userRepo := repository.NewUserRepository(db)
	hub := NewMockHub()

	chatService := service.NewChatService(chatRepo, userRepo, hub)
	chatHandler := handler.NewChatHandler(chatService, nil)

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

	router.POST("/chats/private", authMiddleware, func(c *gin.Context) {
		chatHandler.CreatePrivateChat(c)
	})

	router.POST("/chats/group", authMiddleware, func(c *gin.Context) {
		chatHandler.CreateGroupChat(c)
	})

	router.GET("/chats", authMiddleware, func(c *gin.Context) {
		chatHandler.GetUserChats(c)
	})

	router.DELETE("/chats/:chat_id/members/:user_id", authMiddleware, func(c *gin.Context) {
		chatHandler.RemoveChatMember(c)
	})

	router.GET("/chats/:chat_id/members", authMiddleware, func(c *gin.Context) {
		chatHandler.GetGroupMembers(c)
	})

	router.POST("/chats/:chat_id/members", authMiddleware, func(c *gin.Context) {
		chatHandler.AddChatMember(c)
	})

	router.PUT("/chats/:chat_id", authMiddleware, func(c *gin.Context) {
		chatHandler.UpdateGroupInfo(c)
	})

	return router, hub
}

func createTestUser(t *testing.T, db *sql.DB, username, email, password string) uuid.UUID {
	id := uuid.New()
	_, err := db.Exec(
		"INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)",
		id, username, email, password,
	)
	require.NoError(t, err)
	return id
}

func TestCreatePrivateChatSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")

	body := []byte(fmt.Sprintf(`{"user_id":"%s"}`, user2ID.String()))
	req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response model.Chat
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, model.TypePrivate, response.Type)
	assert.NotEqual(t, uuid.Nil, response.ID)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM chat_members WHERE chat_id = $1", response.ID).Scan(&count)
	assert.Equal(t, 2, count, "Private chat should have 2 members")
}

func TestCreatePrivateChatWithSameUser(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	userID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")

	body := []byte(fmt.Sprintf(`{"user_id":"%s"}`, userID.String()))
	req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "самим собой")
}

func TestCreatePrivateChatUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	nonexistentID := uuid.New()

	body := []byte(fmt.Sprintf(`{"user_id":"%s"}`, nonexistentID.String()))
	req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "не найден")
}

func TestCreatePrivateChatDuplicate(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")

	bodyStr := fmt.Sprintf(`{"user_id":"%s"}`, user2ID.String())
	req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader([]byte(bodyStr)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("First request failed with status %d: %s", w.Code, w.Body.String())
	}

	var firstChat model.Chat
	json.Unmarshal(w.Body.Bytes(), &firstChat)

	req = httptest.NewRequest("POST", "/chats/private", bytes.NewReader([]byte(bodyStr)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", user1ID.String())
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Second request failed with status %d: %s", w.Code, w.Body.String())
	}

	var secondChat model.Chat
	json.Unmarshal(w.Body.Bytes(), &secondChat)

	assert.Equal(t, firstChat.ID, secondChat.ID, "Should return existing private chat")
}

func TestCreatePrivateChatInvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	userID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePrivateChatUnauthorized(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")

	body := []byte(fmt.Sprintf(`{"user_id":"%s"}`, user2ID.String()))
	req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateGroupChatSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	creatorID := createTestUser(t, db, "creator", "creator@example.com", "hashed_pass")
	createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")

	body := []byte(`{"name":"Test Group","usernames":["user1","user2"]}`)
	req := httptest.NewRequest("POST", "/chats/group", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", creatorID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response model.Chat
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, model.TypeGroup, response.Type)
	assert.Equal(t, "Test Group", response.Name)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM chat_members WHERE chat_id = $1", response.ID).Scan(&count)
	assert.Equal(t, 3, count, "Group chat should have creator + 2 members")
}

func TestCreateGroupChatWithDuplicateUsernames(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	creatorID := createTestUser(t, db, "creator", "creator@example.com", "hashed_pass")
	createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")

	body := []byte(`{"name":"Test Group","usernames":["user1","user1"]}`)
	req := httptest.NewRequest("POST", "/chats/group", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", creatorID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response model.Chat
	json.Unmarshal(w.Body.Bytes(), &response)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM chat_members WHERE chat_id = $1", response.ID).Scan(&count)
	assert.Equal(t, 2, count, "Should add user1 only once + creator")
}

func TestCreateGroupChatUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	creatorID := createTestUser(t, db, "creator", "creator@example.com", "hashed_pass")

	body := []byte(`{"name":"Test Group","usernames":["nonexistent"]}`)
	req := httptest.NewRequest("POST", "/chats/group", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", creatorID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "не найден")
}

func TestCreateGroupChatInvalidJSON(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	creatorID := createTestUser(t, db, "creator", "creator@example.com", "hashed_pass")

	body := []byte(`{invalid}`)
	req := httptest.NewRequest("POST", "/chats/group", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", creatorID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserChatsSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, hub := setupChatTestRouter(t, db)

	user1ID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")

	hub.SetUserOnline(user2ID, true)

	createPrivateChatInDB(t, db, user1ID, user2ID)

	req := httptest.NewRequest("GET", "/chats", nil)
	req.Header.Set("X-User-ID", user1ID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.ChatListItem
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Greater(t, len(response), 0, "User should have at least one chat")
	assert.Equal(t, "user2", response[0].Name)
	assert.True(t, response[0].IsOnline)
}

func TestGetUserChatsMultiple(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	userID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")
	user2ID := createTestUser(t, db, "user2", "user2@example.com", "hashed_pass")
	user3ID := createTestUser(t, db, "user3", "user3@example.com", "hashed_pass")

	createPrivateChatInDB(t, db, userID, user2ID)
	createPrivateChatInDB(t, db, userID, user3ID)

	chatID := uuid.New()
	_, err := db.Exec(
		"INSERT INTO chats (id, type, name) VALUES ($1, $2, $3)",
		chatID, model.TypeGroup, "Group Chat",
	)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3), ($1, $4)",
		chatID, userID, user2ID, user3ID,
	)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/chats", nil)
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.ChatListItem
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 3, len(response), "User should have 3 chats")
}

func TestGetUserChatsEmpty(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	userID := createTestUser(t, db, "user1", "user1@example.com", "hashed_pass")

	req := httptest.NewRequest("GET", "/chats", nil)
	req.Header.Set("X-User-ID", userID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "null", string(w.Body.Bytes()), "Should return empty result when user has no chats")
}

func TestGetUserChatsUnauthorized(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	req := httptest.NewRequest("GET", "/chats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func createPrivateChatInDB(t *testing.T, db *sql.DB, user1ID, user2ID uuid.UUID) {
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
}

func TestUpdateGroupInfoSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	creatorID := createTestUser(t, db, "creator", "creator@example.com", "pass")
	chatID := uuid.New()
	_, err := db.Exec("INSERT INTO chats (id, type, name, creator_id) VALUES ($1, $2, $3, $4)",
		chatID, model.TypeGroup, "Old Name", creatorID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)", chatID, creatorID)
	require.NoError(t, err)

	body := []byte(`{"name":"New Group Name"}`)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/chats/%s", chatID.String()), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", creatorID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response model.Chat
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "New Group Name", response.Name)

	var name string
	db.QueryRow("SELECT name FROM chats WHERE id = $1", chatID).Scan(&name)
	assert.Equal(t, "New Group Name", name)
}

func TestAddChatMemberSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	creatorID := createTestUser(t, db, "creator", "creator@example.com", "pass")
	createTestUser(t, db, "user1", "user1@example.com", "pass")
	chatID := uuid.New()
	_, err := db.Exec("INSERT INTO chats (id, type, name, creator_id) VALUES ($1, $2, $3, $4)",
		chatID, model.TypeGroup, "Group", creatorID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)", chatID, creatorID)
	require.NoError(t, err)

	body := []byte(`{"username":"user1"}`)
	req := httptest.NewRequest("POST", fmt.Sprintf("/chats/%s/members", chatID.String()), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", creatorID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM chat_members WHERE chat_id = $1", chatID).Scan(&count)
	assert.Equal(t, 2, count)
}

func TestRemoveChatMemberSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	creatorID := createTestUser(t, db, "creator", "creator@example.com", "pass")
	user1ID := createTestUser(t, db, "user1", "user1@example.com", "pass")
	chatID := uuid.New()
	_, err := db.Exec("INSERT INTO chats (id, type, name, creator_id) VALUES ($1, $2, $3, $4)",
		chatID, model.TypeGroup, "Group", creatorID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3)",
		chatID, creatorID, user1ID)
	require.NoError(t, err)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/chats/%s/members/%s", chatID.String(), user1ID.String()), nil)
	req.Header.Set("X-User-ID", creatorID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int
	db.QueryRow("SELECT COUNT(*) FROM chat_members WHERE chat_id = $1 AND user_id = $2", chatID, user1ID).Scan(&count)
	assert.Equal(t, 0, count)
}

func TestGetGroupMembersSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router, _ := setupChatTestRouter(t, db)

	creatorID := createTestUser(t, db, "creator", "creator@example.com", "pass")
	user1ID := createTestUser(t, db, "user1", "user1@example.com", "pass")
	chatID := uuid.New()
	_, err := db.Exec("INSERT INTO chats (id, type, name, creator_id) VALUES ($1, $2, $3, $4)",
		chatID, model.TypeGroup, "Group", creatorID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3)",
		chatID, creatorID, user1ID)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", fmt.Sprintf("/chats/%s/members", chatID.String()), nil)
	req.Header.Set("X-User-ID", creatorID.String())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.ChatMemberInfo
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 2)
}
