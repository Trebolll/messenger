package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"messenger/internal/handler"
	"messenger/internal/model"
	"messenger/internal/repository"
	"messenger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStorage struct{}

func (m *mockStorage) Upload(ctx context.Context, objectName string, file io.Reader, size int64, contentType string) (string, error) {
	return fmt.Sprintf("http://mock-storage/%s", objectName), nil
}

func (m *mockStorage) Download(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	return io.NopCloser(bytes.NewReader([]byte("mock content"))), 12, "application/octet-stream", nil
}

func (m *mockStorage) Delete(ctx context.Context, objectName string) error {
	return nil
}

func (m *mockStorage) GetURL(objectName string) string {
	return fmt.Sprintf("http://mock-storage/%s", objectName)
}

func TestUploadAttachmentSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	userRepo := repository.NewUserRepository(db)
	attachRepo := repository.NewAttachmentRepository(db)

	mockStorage := &mockStorage{}
	attachService := service.NewAttachmentService(attachRepo, mockStorage)

	user1 := &model.User{ID: uuid.New(), Username: "user1", Email: "user1@test.com", Password: "password"}
	err := userRepo.Create(user1)
	require.NoError(t, err)

	// Create chat manually in DB for simplicity in test
	chatID := uuid.New()
	_, err = db.Exec("INSERT INTO chats (id, type, creator_id) VALUES ($1, $2, $3)", chatID, model.TypePrivate, user1.ID)
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2)", chatID, user1.ID)
	require.NoError(t, err)

	chat := &model.Chat{ID: chatID, Type: model.TypePrivate, CreatorID: &user1.ID}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock middleware to set user in context
	router.Use(func(c *gin.Context) {
		c.Set("userID", user1.ID)
		c.Next()
	})

	attachHandler := handler.NewAttachmentHandler(attachService)
	router.POST("/api/chats/:chat_id/attachments", attachHandler.Upload)

	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", "test.png"))
	h.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(h)
	require.NoError(t, err)
	_, err = part.Write([]byte("fake image content"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/chats/%s/attachments", chat.ID), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.Attachment
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test.png", response.Filename)
	assert.Equal(t, chat.ID, response.ChatID)
	assert.Equal(t, user1.ID, response.SenderID)
	assert.Contains(t, response.Url, ".png")
}

func TestUploadAttachmentInvalidChat(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	userRepo := repository.NewUserRepository(db)
	attachRepo := repository.NewAttachmentRepository(db)

	attachService := service.NewAttachmentService(attachRepo, &mockStorage{})

	user1 := &model.User{ID: uuid.New(), Username: "user1", Password: "password"}
	userRepo.Create(user1)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", user1.ID)
		c.Next()
	})

	attachHandler := handler.NewAttachmentHandler(attachService)
	router.POST("/api/chats/:chat_id/attachments", attachHandler.Upload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.CreateFormFile("file", "test.png")
	writer.Close()

	req, _ := http.NewRequest("POST", "/api/chats/not-a-uuid/attachments", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDownloadAttachmentSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	userRepo := repository.NewUserRepository(db)
	attachRepo := repository.NewAttachmentRepository(db)
	mockStorage := &mockStorage{}
	attachService := service.NewAttachmentService(attachRepo, mockStorage)

	user1 := &model.User{ID: uuid.New(), Username: "user1", Email: "user1@test.com", Password: "password"}
	err := userRepo.Create(user1)
	require.NoError(t, err)

	chatID := uuid.New()
	_, err = db.Exec("INSERT INTO chats (id, type, creator_id) VALUES ($1, $2, $3)", chatID, model.TypePrivate, user1.ID)
	require.NoError(t, err)

	attachment := &model.Attachment{
		ChatID:    chatID,
		SenderID:  user1.ID,
		Url:       "http://mock-storage/test-file.png",
		Filename:  "test-file.png",
		MimeType:  "image/png",
		SizeBytes: 12,
	}
	err = attachRepo.Create(attachment)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	attachHandler := handler.NewAttachmentHandler(attachService)
	router.GET("/api/attachments/:id", attachHandler.Download)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/attachments/%s", attachment.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "attachment; filename=test-file.png", w.Header().Get("Content-Disposition"))
	assert.Equal(t, "mock content", w.Body.String())
}

func TestDeleteAttachmentSuccess(t *testing.T) {
	db := setupTestDB(t)
	createTestTables(t, db)
	defer cleanupTestTables(t, db)

	userRepo := repository.NewUserRepository(db)
	attachRepo := repository.NewAttachmentRepository(db)
	mockStorage := &mockStorage{}
	attachService := service.NewAttachmentService(attachRepo, mockStorage)

	user1 := &model.User{ID: uuid.New(), Username: "user1", Email: "user1@test.com", Password: "password"}
	err := userRepo.Create(user1)
	require.NoError(t, err)

	chatID := uuid.New()
	_, err = db.Exec("INSERT INTO chats (id, type, creator_id) VALUES ($1, $2, $3)", chatID, model.TypePrivate, user1.ID)
	require.NoError(t, err)

	attachment := &model.Attachment{
		ChatID:    chatID,
		SenderID:  user1.ID,
		Url:       "http://mock-storage/test-file.png",
		Filename:  "test-file.png",
		MimeType:  "image/png",
		SizeBytes: 12,
	}
	err = attachRepo.Create(attachment)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	attachHandler := handler.NewAttachmentHandler(attachService)
	router.DELETE("/api/attachments/:id", attachHandler.Delete)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/attachments/%s", attachment.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's deleted from DB
	_, err = attachRepo.GetByID(attachment.ID)
	assert.Error(t, err)
}
