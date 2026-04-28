package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"messenger/internal/handler"
	"messenger/internal/model"
	"messenger/internal/service"
	"messenger/internal/service/websocket"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) CreatePrivateChat(initiatorID, targetUserID uuid.UUID) (*model.Chat, error) {
	args := m.Called(initiatorID, targetUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepository) CreateGroupChat(name string, memberIDs []uuid.UUID) (*model.Chat, error) {
	args := m.Called(name, memberIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepository) GetUserChats(userID uuid.UUID) ([]model.ChatListItem, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ChatListItem), args.Error(1)
}

func (m *MockChatRepository) UpdateGroupAvatarUrl(chatID uuid.UUID, url string) error {
	args := m.Called(chatID, url)
	return args.Error(0)
}

func (m *MockChatRepository) GetChatByID(chatID uuid.UUID) (*model.Chat, error) {
	args := m.Called(chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepository) UpdateGroupChat(chatID uuid.UUID, name string, avatarUrl string) error {
	args := m.Called(chatID, name, avatarUrl)
	return args.Error(0)
}

func (m *MockChatRepository) RemoveChatMember(chatID, userID uuid.UUID) error {
	args := m.Called(chatID, userID)
	return args.Error(0)
}

func (m *MockChatRepository) AddChatMember(chatID, userID uuid.UUID) error {
	args := m.Called(chatID, userID)
	return args.Error(0)
}

func (m *MockChatRepository) IsChatMember(chatID uuid.UUID, userID uuid.UUID) (bool, error) {
	args := m.Called(chatID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatRepository) GetMembersInfo(chatID uuid.UUID) ([]model.ChatMemberInfo, error) {
	args := m.Called(chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ChatMemberInfo), args.Error(1)
}

type MockUserRepositoryForChat struct {
	mock.Mock
}

func (m *MockUserRepositoryForChat) GetById(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepositoryForChat) GetByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

type MockHubForChat struct {
	mock.Mock
}

func (m *MockHubForChat) IsUserOnline(userID uuid.UUID) bool {
	args := m.Called(userID)
	return args.Bool(0)
}

func (m *MockHubForChat) BroadcastToUsers(userIDs []uuid.UUID, message websocket.Message) {
	m.Called(userIDs, message)
}

type MockStorageForChat struct {
	mock.Mock
}

func (m *MockStorageForChat) Upload(ctx context.Context, objectName string, file io.Reader, size int64, contentType string) (string, error) {
	args := m.Called(ctx, objectName, file, size, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockStorageForChat) Download(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	args := m.Called(ctx, objectName)
	if args.Get(0) == nil {
		return nil, 0, "", args.Error(3)
	}
	return args.Get(0).(io.ReadCloser), int64(args.Int(1)), args.String(2), args.Error(3)
}

func (m *MockStorageForChat) Delete(ctx context.Context, objectName string) error {
	args := m.Called(ctx, objectName)
	return args.Error(0)
}

func (m *MockStorageForChat) GetURL(objectName string) string {
	args := m.Called(objectName)
	return args.String(0)
}

func setupChatTestRouter(mockRepo *MockChatRepository, mockUserRepo *MockUserRepositoryForChat, mockHub *MockHubForChat, mockStorage *MockStorageForChat) (*gin.Engine, *handler.ChatHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	chatService := service.NewChatService(mockRepo, mockUserRepo, mockHub)
	chatHandler := handler.NewChatHandler(chatService, mockStorage)

	return router, chatHandler
}

func TestCreatePrivateChat(t *testing.T) {
	currentUserID := uuid.New()
	targetUserID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		mockUserRepo := new(MockUserRepositoryForChat)
		mockHub := new(MockHubForChat)
		mockStorage := new(MockStorageForChat)
		_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

		mockUserRepo.On("GetById", currentUserID).Return(&model.User{ID: currentUserID}, nil).Once()
		mockUserRepo.On("GetById", targetUserID).Return(&model.User{ID: targetUserID}, nil).Once()
		mockRepo.On("CreatePrivateChat", currentUserID, targetUserID).Return(&model.Chat{ID: uuid.New(), Type: model.TypePrivate}, nil).Once()

		body := map[string]uuid.UUID{"user_id": targetUserID}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", currentUserID)

		chatHandler.CreatePrivateChat(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		mockUserRepo := new(MockUserRepositoryForChat)
		mockHub := new(MockHubForChat)
		mockStorage := new(MockStorageForChat)
		_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

		req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader([]byte(`invalid`)))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		chatHandler.CreatePrivateChat(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		mockUserRepo := new(MockUserRepositoryForChat)
		mockHub := new(MockHubForChat)
		mockStorage := new(MockStorageForChat)
		_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

		body := map[string]uuid.UUID{"user_id": targetUserID}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		chatHandler.CreatePrivateChat(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InitiatorNotFound", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		mockUserRepo := new(MockUserRepositoryForChat)
		mockHub := new(MockHubForChat)
		mockStorage := new(MockStorageForChat)
		_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

		mockUserRepo.On("GetById", currentUserID).Return(nil, errors.New("user not found")).Once()

		body := map[string]uuid.UUID{"user_id": targetUserID}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", currentUserID)

		chatHandler.CreatePrivateChat(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "user not found")
	})

	t.Run("TargetUserNotFound", func(t *testing.T) {
		mockRepo := new(MockChatRepository)
		mockUserRepo := new(MockUserRepositoryForChat)
		mockHub := new(MockHubForChat)
		mockStorage := new(MockStorageForChat)
		_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

		mockUserRepo.On("GetById", currentUserID).Return(&model.User{ID: currentUserID}, nil).Once()
		mockUserRepo.On("GetById", targetUserID).Return(nil, errors.New("target user not found")).Once()

		body := map[string]uuid.UUID{"user_id": targetUserID}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", currentUserID)

		chatHandler.CreatePrivateChat(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "target user not found")
	})
}

func TestCreateGroupChat(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)
	mockStorage := new(MockStorageForChat)

	_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

	currentUserID := uuid.New()
	targetUserID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.On("GetByUsername", "user1").Return(&model.User{ID: targetUserID, Username: "user1"}, nil).Once()
		mockRepo.On("CreateGroupChat", "Group Name", []uuid.UUID{currentUserID, targetUserID}).Return(&model.Chat{ID: uuid.New(), Name: "Group Name", Type: model.TypeGroup}, nil).Once()

		body := map[string]interface{}{
			"name":      "Group Name",
			"usernames": []string{"user1"},
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/group", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", currentUserID)

		chatHandler.CreateGroupChat(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockUserRepo.On("GetByUsername", "nonexistent").Return(nil, errors.New("not found")).Once()

		body := map[string]interface{}{
			"usernames": []string{"nonexistent"},
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/group", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", currentUserID)

		chatHandler.CreateGroupChat(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "пользователь nonexistent не найден")
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/chats/group", bytes.NewReader([]byte(`invalid`)))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		chatHandler.CreateGroupChat(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetUserChats(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)
	mockStorage := new(MockStorageForChat)

	_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

	currentUserID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		chats := []model.ChatListItem{
			{ID: uuid.New(), Name: "Chat 1"},
		}
		mockRepo.On("GetUserChats", currentUserID).Return(chats, nil).Once()

		req := httptest.NewRequest("GET", "/chats", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", currentUserID)

		chatHandler.GetUserChats(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []model.ChatListItem
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp, 1)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		chatHandler.GetUserChats(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("InvalidUserIDType", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "not-a-uuid")
		chatHandler.GetUserChats(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUpdateGroupAvatar(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)
	mockStorage := new(MockStorageForChat)

	_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

	chatID := uuid.New()
	currentUserID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="avatar"; filename="test.jpg"`)
		h.Set("Content-Type", "image/jpeg")
		part, _ := writer.CreatePart(h)
		part.Write([]byte("fake image content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/chats/"+chatID.String()+"/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything, "image/jpeg").Return("http://avatar.url", nil).Once()
		mockRepo.On("UpdateGroupAvatarUrl", chatID, "http://avatar.url").Return(nil).Once()

		chatHandler.UpdateGroupAvatar(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "http://avatar.url")
	})

	t.Run("InvalidChatID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "chat_id", Value: "invalid"}}
		chatHandler.UpdateGroupAvatar(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		chatHandler.UpdateGroupAvatar(c)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("NoFile", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/chats/"+chatID.String()+"/avatar", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.UpdateGroupAvatar(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "файл не найден")
	})

	t.Run("UploadError", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="avatar"; filename="test.jpg"`)
		h.Set("Content-Type", "image/jpeg")
		part, _ := writer.CreatePart(h)
		part.Write([]byte("fake image content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/chats/"+chatID.String()+"/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything, "image/jpeg").Return("", errors.New("upload failed")).Once()

		chatHandler.UpdateGroupAvatar(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "не удалось загрузить файл")
	})

	t.Run("ServiceError", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="avatar"; filename="test.jpg"`)
		h.Set("Content-Type", "image/jpeg")
		part, _ := writer.CreatePart(h)
		part.Write([]byte("fake image content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/chats/"+chatID.String()+"/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, mock.Anything, "image/jpeg").Return("http://avatar.url", nil).Once()
		mockRepo.On("UpdateGroupAvatarUrl", chatID, "http://avatar.url").Return(errors.New("db error")).Once()

		chatHandler.UpdateGroupAvatar(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})
}

func TestUpdateGroupInfo(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)
	mockStorage := new(MockStorageForChat)

	_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

	chatID := uuid.New()
	currentUserID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetChatByID", chatID).Return(&model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &currentUserID, Name: "Old Name"}, nil).Once()
		mockRepo.On("UpdateGroupChat", chatID, "New Name", "").Return(nil).Once()

		body := map[string]string{"name": "New Name"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/"+chatID.String(), bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.UpdateGroupInfo(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("InvalidChatID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "chat_id", Value: "invalid"}}
		c.Set("userID", currentUserID)
		chatHandler.UpdateGroupInfo(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/chats/"+chatID.String(), bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)
		chatHandler.UpdateGroupInfo(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		mockRepo.On("GetChatByID", chatID).Return(&model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &uuid.Nil}, nil).Once()

		body := map[string]string{"name": "New Name"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/"+chatID.String(), bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.UpdateGroupInfo(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAddChatMember(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)
	mockStorage := new(MockStorageForChat)

	_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

	chatID := uuid.New()
	currentUserID := uuid.New()
	targetUserID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetChatByID", chatID).Return(&model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &currentUserID}, nil).Once()
		mockUserRepo.On("GetByUsername", "newuser").Return(&model.User{ID: targetUserID, Username: "newuser"}, nil).Once()
		mockRepo.On("AddChatMember", chatID, targetUserID).Return(nil).Once()
		mockRepo.On("GetMembersInfo", chatID).Return([]model.ChatMemberInfo{{ID: currentUserID}, {ID: targetUserID}}, nil).Once()
		mockHub.On("BroadcastToUsers", mock.Anything, mock.Anything).Return().Once()

		body := map[string]string{"username": "newuser"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/"+chatID.String()+"/members", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.AddChatMember(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockHub.AssertExpectations(t)
	})

	t.Run("InvalidChatID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "chat_id", Value: "invalid"}}
		c.Set("userID", currentUserID)
		chatHandler.AddChatMember(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/chats/"+chatID.String()+"/members", bytes.NewReader([]byte("invalid")))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)
		chatHandler.AddChatMember(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		mockRepo.On("GetChatByID", chatID).Return(&model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &uuid.Nil}, nil).Once()

		body := map[string]string{"username": "newuser"}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/"+chatID.String()+"/members", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.AddChatMember(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRemoveChatMember(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)
	mockStorage := new(MockStorageForChat)

	_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

	chatID := uuid.New()
	currentUserID := uuid.New()
	targetUserID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("GetChatByID", chatID).Return(&model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &currentUserID}, nil).Once()
		mockRepo.On("GetMembersInfo", chatID).Return([]model.ChatMemberInfo{{ID: currentUserID}, {ID: targetUserID}}, nil).Once()
		mockRepo.On("RemoveChatMember", chatID, targetUserID).Return(nil).Once()
		mockHub.On("BroadcastToUsers", mock.Anything, mock.Anything).Return().Once()

		req := httptest.NewRequest("DELETE", "/chats/"+chatID.String()+"/members/"+targetUserID.String(), nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{
			{Key: "chat_id", Value: chatID.String()},
			{Key: "user_id", Value: targetUserID.String()},
		}
		c.Set("userID", currentUserID)

		chatHandler.RemoveChatMember(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
		mockHub.AssertExpectations(t)
	})

	t.Run("InvalidChatID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{
			{Key: "chat_id", Value: "invalid"},
			{Key: "user_id", Value: targetUserID.String()},
		}
		c.Set("userID", currentUserID)
		chatHandler.RemoveChatMember(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{
			{Key: "chat_id", Value: chatID.String()},
			{Key: "user_id", Value: "invalid"},
		}
		c.Set("userID", currentUserID)
		chatHandler.RemoveChatMember(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		mockRepo.On("GetChatByID", chatID).Return(&model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &uuid.Nil}, nil).Once()

		req := httptest.NewRequest("DELETE", "/chats/"+chatID.String()+"/members/"+targetUserID.String(), nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{
			{Key: "chat_id", Value: chatID.String()},
			{Key: "user_id", Value: targetUserID.String()},
		}
		c.Set("userID", currentUserID)

		chatHandler.RemoveChatMember(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestGetGroupMembers_Error(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)
	mockStorage := new(MockStorageForChat)

	_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

	chatID := uuid.New()
	currentUserID := uuid.New()

	t.Run("InvalidChatID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "chat_id", Value: "invalid"}}
		c.Set("userID", currentUserID)
		chatHandler.GetGroupMembers(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ServiceError", func(t *testing.T) {
		mockRepo.On("IsChatMember", chatID, currentUserID).Return(true, nil).Once()
		mockRepo.On("GetMembersInfo", chatID).Return(nil, errors.New("db error")).Once()

		req := httptest.NewRequest("GET", "/chats/"+chatID.String()+"/members", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.GetGroupMembers(c)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestGetGroupMembers(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)
	mockStorage := new(MockStorageForChat)

	_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

	chatID := uuid.New()
	currentUserID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		mockRepo.On("IsChatMember", chatID, currentUserID).Return(true, nil).Once()
		members := []model.ChatMemberInfo{{ID: currentUserID, Username: "user1"}}
		mockRepo.On("GetMembersInfo", chatID).Return(members, nil).Once()
		mockHub.On("IsUserOnline", currentUserID).Return(true).Once()

		req := httptest.NewRequest("GET", "/chats/"+chatID.String()+"/members", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.GetGroupMembers(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp []model.ChatMemberInfo
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp, 1)
		assert.True(t, resp[0].IsOnline)
	})

	t.Run("Forbidden", func(t *testing.T) {
		mockRepo.On("IsChatMember", chatID, currentUserID).Return(false, nil).Once()

		req := httptest.NewRequest("GET", "/chats/"+chatID.String()+"/members", nil)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.GetGroupMembers(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestChatHandler_AdditionalScenarios(t *testing.T) {
	mockRepo := new(MockChatRepository)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)
	mockStorage := new(MockStorageForChat)

	_, chatHandler := setupChatTestRouter(mockRepo, mockUserRepo, mockHub, mockStorage)

	currentUserID := uuid.New()

	t.Run("CreatePrivateChat_ServiceError", func(t *testing.T) {
		targetUserID := uuid.New()
		mockUserRepo.On("GetById", currentUserID).Return(&model.User{ID: currentUserID}, nil).Once()
		mockUserRepo.On("GetById", targetUserID).Return(&model.User{ID: targetUserID}, nil).Once()
		mockRepo.On("CreatePrivateChat", currentUserID, targetUserID).Return(nil, errors.New("db error")).Once()

		body := map[string]uuid.UUID{"user_id": targetUserID}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest("POST", "/chats/private", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", currentUserID)

		chatHandler.CreatePrivateChat(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("UpdateGroupAvatar_FileTooLarge", func(t *testing.T) {
		chatID := uuid.New()
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "test.jpg")
		// 6MB > 5MB limit
		part.Write(make([]byte, 6<<20))
		writer.Close()

		req := httptest.NewRequest("POST", "/chats/"+chatID.String()+"/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.UpdateGroupAvatar(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "файл слишком большой")
	})

	t.Run("UpdateGroupAvatar_InvalidMimeType", func(t *testing.T) {
		chatID := uuid.New()
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "test.txt")
		part.Write([]byte("not an image"))
		writer.Close()

		// Change Content-Type manually because CreateFormFile sets it based on extension sometimes
		// but gin.FormFile uses the header.
		req := httptest.NewRequest("POST", "/chats/"+chatID.String()+"/avatar", body)
		// We need to simulate the multi-part header with a wrong content type
		// But let's just use a .txt extension and see what happens.
		// Actually c.FormFile(name) returns *multipart.FileHeader which has Header map.

		// To truly test MIME type check in chat_handler.go:
		// mimeType := fileHeader.Header.Get("Content-Type")
		// We need a more manual way to build the request.

		boundary := "boundary"
		body2 := &bytes.Buffer{}
		fmt.Fprintf(body2, "--%s\r\n", boundary)
		fmt.Fprintf(body2, "Content-Disposition: form-data; name=\"avatar\"; filename=\"test.txt\"\r\n")
		fmt.Fprintf(body2, "Content-Type: text/plain\r\n\r\n")
		body2.WriteString("not an image")
		fmt.Fprintf(body2, "\r\n--%s--\r\n", boundary)

		req = httptest.NewRequest("POST", "/chats/"+chatID.String()+"/avatar", body2)
		req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = []gin.Param{{Key: "chat_id", Value: chatID.String()}}
		c.Set("userID", currentUserID)

		chatHandler.UpdateGroupAvatar(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "разрешены только изображения")
	})
}
