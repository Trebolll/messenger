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
	_ "time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var testDB *sql.DB

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestDB(t *testing.T) *sql.DB {
	// Теперь мы используем БД из контейнера, поднятого в TestMain
	return containerDB
}

func createTestTables(t *testing.T, db *sql.DB) {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`DROP TABLE IF EXISTS messages CASCADE`,
		`DROP TABLE IF EXISTS chat_members CASCADE`,
		`DROP TABLE IF EXISTS chats CASCADE`,
		`DROP TABLE IF EXISTS users CASCADE`,
		`CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password TEXT NOT NULL,
			phone VARCHAR(20),
			full_name VARCHAR(255),
			birth_date DATE,
			location VARCHAR(255),
			status VARCHAR(255),
			avatar_url TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users (username)`,
		`CREATE TABLE chats (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			type VARCHAR(10) NOT NULL CHECK (type IN ('private', 'group')),
			name VARCHAR(100),
			creator_id UUID REFERENCES users(id),
			avatar_url TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE chat_members (
			chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (chat_id, user_id)
		)`,
		`CREATE TABLE messages (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
			sender_id UUID REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			read_at TIMESTAMP,
			edited_at TIMESTAMP
		)`,
		`CREATE TABLE attachments (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
			sender_id UUID REFERENCES users(id) ON DELETE CASCADE,
			message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
			url TEXT NOT NULL,
			filename TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			size_bytes BIGINT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_chat_members_user_id ON chat_members (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_chat_id_created_at ON messages (chat_id, created_at ASC)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		require.NoError(t, err, "Failed to create test tables: %s", query)
	}
}

func cleanupTestTables(t *testing.T, db *sql.DB) {
	queries := []string{
		`DROP TABLE IF EXISTS attachments CASCADE`,
		`DROP TABLE IF EXISTS messages CASCADE`,
		`DROP TABLE IF EXISTS chat_members CASCADE`,
		`DROP TABLE IF EXISTS chats CASCADE`,
		`DROP TABLE IF EXISTS users CASCADE`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		require.NoError(t, err, "Failed to cleanup test tables")
	}
}

func setupTestRouter(t *testing.T, db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService, nil, nil)

	router.POST("/register", userHandler.Register)
	router.POST("/login", userHandler.Login)
	router.GET("/search", userHandler.SearchUsers)

	return router
}

func TestRegisterSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router := setupTestRouter(t, db)

	body := []byte(`{"username":"testuser","email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "пользователь успешно создан", response["сообщение"])

	var user model.User
	err = db.QueryRow("SELECT id, username, email FROM users WHERE username = $1", "testuser").
		Scan(&user.ID, &user.Username, &user.Email)
	require.NoError(t, err, "User should be created in database")
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestRegisterInvalidEmail(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router := setupTestRouter(t, db)

	body := []byte(`{"username":"testuser","email":"invalid-email","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "неверный формат электронной почты")

	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	assert.Equal(t, 0, count, "No user should be created with invalid email")
}

func TestRegisterUsernameTooShort(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router := setupTestRouter(t, db)

	body := []byte(`{"username":"ab","email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "от 3 до 50 символов")
}

func TestRegisterUsernameTooLong(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router := setupTestRouter(t, db)

	longUsername := "a"
	for i := 0; i < 51; i++ {
		longUsername += "a"
	}

	body := []byte(fmt.Sprintf(`{"username":"%s","email":"test@example.com","password":"password123"}`, longUsername))
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "от 3 до 50 символов")
}

func TestRegisterPasswordTooShort(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router := setupTestRouter(t, db)

	body := []byte(`{"username":"testuser","email":"test@example.com","password":"pass"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "не менее 6 символов")
}

func TestRegisterDuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	hashedPassword := "$2a$14$testhashedpassword"
	id := uuid.New()
	_, err := db.Exec(
		"INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)",
		id, "testuser", "existing@example.com", hashedPassword,
	)
	require.NoError(t, err)

	router := setupTestRouter(t, db)

	body := []byte(`{"username":"testuser","email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "имен")

	var count int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", "testuser").Scan(&count)
	assert.Equal(t, 1, count, "Should have exactly one user with this username")
}

func TestRegisterDuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	hashedPassword := "$2a$14$testhashedpassword"
	id := uuid.New()
	_, err := db.Exec(
		"INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)",
		id, "existinguser", "test@example.com", hashedPassword,
	)
	require.NoError(t, err)

	router := setupTestRouter(t, db)

	body := []byte(`{"username":"testuser","email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "адресом электронной почты")

	var count int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "test@example.com").Scan(&count)
	assert.Equal(t, 1, count, "Should have exactly one user with this email")
}

func TestLoginSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), 14)
	require.NoError(t, err)
	userID := uuid.New()
	_, err = db.Exec(
		"INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)",
		userID, "testuser", "test@example.com", string(hashedPassword),
	)
	require.NoError(t, err)

	router := setupTestRouter(t, db)

	body := []byte(`{"email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response["token"], "Token should be returned")
	assert.NotNil(t, response["user"], "User should be returned")

	user := response["user"].(map[string]interface{})
	assert.Equal(t, "testuser", user["username"])
	assert.Equal(t, "test@example.com", user["email"])
	assert.Empty(t, user["password"], "Password should be cleared from response")
}

func TestLoginUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router := setupTestRouter(t, db)

	body := []byte(`{"email":"nonexistent@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid credentials")
}

func TestLoginInvalidPassword(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), 14)
	require.NoError(t, err)
	userID := uuid.New()
	_, err = db.Exec(
		"INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)",
		userID, "testuser", "test@example.com", string(hashedPassword),
	)
	require.NoError(t, err)

	router := setupTestRouter(t, db)

	body := []byte(`{"email":"test@example.com","password":"wrongpassword"}`)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid credentials")
}

func TestSearchUsersWithQParameter(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	hashedPassword := "$2a$14$testhashedpassword"
	for i := 1; i <= 3; i++ {
		id := uuid.New()
		_, err := db.Exec(
			"INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)",
			id, fmt.Sprintf("testuser%d", i), fmt.Sprintf("test%d@example.com", i), hashedPassword,
		)
		require.NoError(t, err)
	}

	router := setupTestRouter(t, db)

	req := httptest.NewRequest("GET", "/search?q=test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 3, len(response), "Should find all users with 'test' in username")

	for _, user := range response {
		assert.Contains(t, user.Username, "test")
		assert.Empty(t, user.Password, "Password should not be returned in search results")
	}
}

func TestSearchUsersWithUsernameParameter(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	hashedPassword := "$2a$14$testhashedpassword"
	id := uuid.New()
	_, err := db.Exec(
		"INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)",
		id, "john", "john@example.com", hashedPassword,
	)
	require.NoError(t, err)

	router := setupTestRouter(t, db)

	req := httptest.NewRequest("GET", "/search?username=john", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.User
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 1, len(response))
	assert.Equal(t, "john", response[0].Username)
}

func TestSearchUsersMissingQuery(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router := setupTestRouter(t, db)

	req := httptest.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "required")
}

func TestSearchUsersQueryTooShort(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router := setupTestRouter(t, db)

	req := httptest.NewRequest("GET", "/search?q=ab", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "не менее 3 символов")
}

func TestSearchUsersEmptyResults(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	router := setupTestRouter(t, db)

	req := httptest.NewRequest("GET", "/search?q=nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if response == nil {
		response = []map[string]interface{}{}
	}
	assert.Equal(t, 0, len(response), "Should return empty results for nonexistent user")
}

func TestSearchUsersPartialMatch(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	hashedPassword := "$2a$14$testhashedpassword"
	usernames := []string{"alice", "bob", "alison"}
	for _, username := range usernames {
		id := uuid.New()
		_, err := db.Exec(
			"INSERT INTO users (id, username, email, password) VALUES ($1, $2, $3, $4)",
			id, username, username+"@example.com", hashedPassword,
		)
		require.NoError(t, err)
	}

	router := setupTestRouter(t, db)

	req := httptest.NewRequest("GET", "/search?q=ali", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 2, len(response), "Should find alice and alison")

	foundUsernames := make(map[string]bool)
	for _, user := range response {
		foundUsernames[user.Username] = true
	}
	assert.True(t, foundUsernames["alice"])
	assert.True(t, foundUsernames["alison"])
	assert.False(t, foundUsernames["bob"])
}
