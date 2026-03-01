package service

import (
	"errors"
	"messenger/internal/model"
	"messenger/internal/service"
	"messenger/internal/service/websocket"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) SendMessage(message *model.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetMessagesByChatID(chatID uuid.UUID) ([]model.Message, error) {
	args := m.Called(chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Message), args.Error(1)
}

func (m *MockMessageRepository) MarkAsRead(chatID, userID uuid.UUID) error {
	args := m.Called(chatID, userID)
	return args.Error(0)
}

func (m *MockMessageRepository) EditMessage(messageID, senderID uuid.UUID, content string) (*model.Message, error) {
	args := m.Called(messageID, senderID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Message), args.Error(1)
}

func (m *MockMessageRepository) DeleteMessage(messageID, senderID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(messageID, senderID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) IsChatMember(chatID, userID uuid.UUID) (bool, error) {
	args := m.Called(chatID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatRepository) GetChatMembers(chatID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockChatRepository) Exists(chatID uuid.UUID) (bool, error) {
	args := m.Called(chatID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatRepository) CreatePrivateChat(initiatorID, targetUserID uuid.UUID) (*model.Chat, error) {
	args := m.Called(initiatorID, targetUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepository) CreateGroupChat(name string, initiatorID uuid.UUID, memberIDs []uuid.UUID) (*model.Chat, error) {
	args := m.Called(name, initiatorID, memberIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepository) GetUserChats(userID uuid.UUID) ([]model.Chat, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Chat), args.Error(1)
}

type MockHub struct {
	mock.Mock
}

func (m *MockHub) SendToUser(userID uuid.UUID, msg websocket.Message) {
	m.Called(userID, msg)
}

func (m *MockHub) Run() {
	m.Called()
}

func TestSendMessage_Success(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	senderID := uuid.New()
	memberID1 := uuid.New()
	memberID2 := senderID

	message := &model.Message{
		ID:         uuid.New(),
		ChatID:     chatID,
		SenderID:   senderID,
		SenderName: "testuser",
		Content:    "Hello, world!",
		CreatedAt:  time.Now(),
	}

	mockChatRepo.On("IsChatMember", chatID, senderID).Return(true, nil)
	mockMessageRepo.On("SendMessage", message).Return(nil)
	mockChatRepo.On("GetChatMembers", chatID).Return([]uuid.UUID{memberID1, memberID2}, nil)
	mockHub.On("SendToUser", memberID1, mock.MatchedBy(func(msg websocket.Message) bool {
		return msg.Type == "new_message"
	})).Return()
	mockHub.On("SendToUser", memberID2, mock.MatchedBy(func(msg websocket.Message) bool {
		return msg.Type == "new_message"
	})).Return()

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.SendMessage(message)

	assert.NoError(t, err)
	mockChatRepo.AssertCalled(t, "IsChatMember", chatID, senderID)
	mockMessageRepo.AssertCalled(t, "SendMessage", message)
	mockChatRepo.AssertCalled(t, "GetChatMembers", chatID)
	mockHub.AssertCalled(t, "SendToUser", memberID1, mock.MatchedBy(func(msg websocket.Message) bool {
		return msg.Type == "new_message"
	}))
}

func TestSendMessage_NotChatMember(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	senderID := uuid.New()

	message := &model.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  "Hello, world!",
	}

	mockChatRepo.On("IsChatMember", chatID, senderID).Return(false, nil)

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.SendMessage(message)

	assert.Error(t, err)
	assert.Equal(t, "доступ запрещен: вы не являетесь участником этого чата", err.Error())
	mockChatRepo.AssertCalled(t, "IsChatMember", chatID, senderID)
	mockMessageRepo.AssertNotCalled(t, "SendMessage")
	mockChatRepo.AssertNotCalled(t, "GetChatMembers")
}

func TestSendMessage_IsChatMemberError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	senderID := uuid.New()

	message := &model.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  "Hello, world!",
	}

	mockChatRepo.On("IsChatMember", chatID, senderID).Return(false, errors.New("database error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.SendMessage(message)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockChatRepo.AssertCalled(t, "IsChatMember", chatID, senderID)
	mockMessageRepo.AssertNotCalled(t, "SendMessage")
}

func TestSendMessage_SendMessageError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	senderID := uuid.New()

	message := &model.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  "Hello, world!",
	}

	mockChatRepo.On("IsChatMember", chatID, senderID).Return(true, nil)
	mockMessageRepo.On("SendMessage", message).Return(errors.New("database error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.SendMessage(message)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockChatRepo.AssertCalled(t, "IsChatMember", chatID, senderID)
	mockMessageRepo.AssertCalled(t, "SendMessage", message)
	mockChatRepo.AssertNotCalled(t, "GetChatMembers")
}

func TestSendMessage_GetChatMembersError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	senderID := uuid.New()

	message := &model.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  "Hello, world!",
	}

	mockChatRepo.On("IsChatMember", chatID, senderID).Return(true, nil)
	mockMessageRepo.On("SendMessage", message).Return(nil)
	mockChatRepo.On("GetChatMembers", chatID).Return(nil, errors.New("database error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.SendMessage(message)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockChatRepo.AssertCalled(t, "IsChatMember", chatID, senderID)
	mockMessageRepo.AssertCalled(t, "SendMessage", message)
	mockChatRepo.AssertCalled(t, "GetChatMembers", chatID)
}

func TestGetMessagesByChatID_Success(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	messages := []model.Message{
		{
			ID:        uuid.New(),
			ChatID:    chatID,
			SenderID:  uuid.New(),
			Content:   "Message 1",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			ChatID:    chatID,
			SenderID:  uuid.New(),
			Content:   "Message 2",
			CreatedAt: time.Now(),
		},
	}

	mockChatRepo.On("Exists", chatID).Return(true, nil)
	mockMessageRepo.On("GetMessagesByChatID", chatID).Return(messages, nil)

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	result, err := messageService.GetMessagesByChatID(chatID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "Message 1", result[0].Content)
	assert.Equal(t, "Message 2", result[1].Content)
	mockChatRepo.AssertCalled(t, "Exists", chatID)
	mockMessageRepo.AssertCalled(t, "GetMessagesByChatID", chatID)
	mockChatRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestGetMessagesByChatID_ChatNotFound(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()

	mockChatRepo.On("Exists", chatID).Return(false, nil)

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	result, err := messageService.GetMessagesByChatID(chatID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "чат не существует", err.Error())
	mockChatRepo.AssertCalled(t, "Exists", chatID)
	mockMessageRepo.AssertNotCalled(t, "GetMessagesByChatID")
	mockChatRepo.AssertExpectations(t)
}

func TestGetMessagesByChatID_ExistsError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()

	mockChatRepo.On("Exists", chatID).Return(false, errors.New("database error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	result, err := messageService.GetMessagesByChatID(chatID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())
	mockChatRepo.AssertCalled(t, "Exists", chatID)
	mockMessageRepo.AssertNotCalled(t, "GetMessagesByChatID")
	mockChatRepo.AssertExpectations(t)
}

func TestGetMessagesByChatID_GetMessagesError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()

	mockChatRepo.On("Exists", chatID).Return(true, nil)
	mockMessageRepo.On("GetMessagesByChatID", chatID).Return(nil, errors.New("database error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	result, err := messageService.GetMessagesByChatID(chatID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())
	mockChatRepo.AssertCalled(t, "Exists", chatID)
	mockMessageRepo.AssertCalled(t, "GetMessagesByChatID", chatID)
	mockMessageRepo.AssertExpectations(t)
}

func TestGetMessagesByChatID_EmptyMessages(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()

	mockChatRepo.On("Exists", chatID).Return(true, nil)
	mockMessageRepo.On("GetMessagesByChatID", chatID).Return([]model.Message{}, nil)

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	result, err := messageService.GetMessagesByChatID(chatID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	mockChatRepo.AssertCalled(t, "Exists", chatID)
	mockMessageRepo.AssertCalled(t, "GetMessagesByChatID", chatID)
	mockChatRepo.AssertExpectations(t)
	mockMessageRepo.AssertExpectations(t)
}

func TestMarkChatAsRead_Success(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	userID := uuid.New()
	memberID1 := uuid.New()
	memberID2 := uuid.New()

	mockMessageRepo.On("MarkAsRead", chatID, userID).Return(nil)
	mockChatRepo.On("GetChatMembers", chatID).Return([]uuid.UUID{userID, memberID1, memberID2}, nil)
	mockHub.On("SendToUser", memberID1, mock.MatchedBy(func(msg websocket.Message) bool {
		return msg.Type == "messages_read"
	})).Return()
	mockHub.On("SendToUser", memberID2, mock.MatchedBy(func(msg websocket.Message) bool {
		return msg.Type == "messages_read"
	})).Return()

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.MarkChatAsRead(chatID, userID)

	assert.NoError(t, err)
	mockMessageRepo.AssertCalled(t, "MarkAsRead", chatID, userID)
	mockChatRepo.AssertCalled(t, "GetChatMembers", chatID)
	mockHub.AssertCalled(t, "SendToUser", memberID1, mock.MatchedBy(func(msg websocket.Message) bool {
		return msg.Type == "messages_read"
	}))
	mockHub.AssertCalled(t, "SendToUser", memberID2, mock.MatchedBy(func(msg websocket.Message) bool {
		return msg.Type == "messages_read"
	}))
	mockHub.AssertNotCalled(t, "SendToUser", userID, mock.Anything)
}

func TestMarkChatAsRead_MarkAsReadError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	userID := uuid.New()

	mockMessageRepo.On("MarkAsRead", chatID, userID).Return(errors.New("database error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.MarkChatAsRead(chatID, userID)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockMessageRepo.AssertCalled(t, "MarkAsRead", chatID, userID)
	mockChatRepo.AssertNotCalled(t, "GetChatMembers")
	mockHub.AssertNotCalled(t, "SendToUser", mock.Anything, mock.Anything)
}

func TestMarkChatAsRead_GetChatMembersError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	userID := uuid.New()

	mockMessageRepo.On("MarkAsRead", chatID, userID).Return(nil)
	mockChatRepo.On("GetChatMembers", chatID).Return(nil, errors.New("database error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.MarkChatAsRead(chatID, userID)

	assert.NoError(t, err)
	mockMessageRepo.AssertCalled(t, "MarkAsRead", chatID, userID)
	mockChatRepo.AssertCalled(t, "GetChatMembers", chatID)
	mockHub.AssertNotCalled(t, "SendToUser", mock.Anything, mock.Anything)
}

func TestMarkChatAsRead_SingleMember(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	userID := uuid.New()

	mockMessageRepo.On("MarkAsRead", chatID, userID).Return(nil)
	mockChatRepo.On("GetChatMembers", chatID).Return([]uuid.UUID{userID}, nil)

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.MarkChatAsRead(chatID, userID)

	assert.NoError(t, err)
	mockMessageRepo.AssertCalled(t, "MarkAsRead", chatID, userID)
	mockChatRepo.AssertCalled(t, "GetChatMembers", chatID)
	mockHub.AssertNotCalled(t, "SendToUser", mock.Anything, mock.Anything)
}

func TestMarkChatAsRead_NoOtherMembers(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	chatID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()

	mockMessageRepo.On("MarkAsRead", chatID, userID).Return(nil)
	mockChatRepo.On("GetChatMembers", chatID).Return([]uuid.UUID{userID, memberID}, nil)
	mockHub.On("SendToUser", memberID, mock.MatchedBy(func(msg websocket.Message) bool {
		return msg.Type == "messages_read"
	})).Return()

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.MarkChatAsRead(chatID, userID)

	assert.NoError(t, err)
	mockMessageRepo.AssertCalled(t, "MarkAsRead", chatID, userID)
	mockChatRepo.AssertCalled(t, "GetChatMembers", chatID)
	mockHub.AssertCalled(t, "SendToUser", memberID, mock.MatchedBy(func(msg websocket.Message) bool {
		return msg.Type == "messages_read"
	}))
	mockHub.AssertNotCalled(t, "SendToUser", userID, mock.Anything)
}

func TestDeleteMessage_Success(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	messageID := uuid.New()
	senderID := uuid.New()
	chatID := uuid.New()
	memberID1 := uuid.New()
	memberID2 := uuid.New()

	mockMessageRepo.On("DeleteMessage", messageID, senderID).Return(chatID, nil)
	mockChatRepo.On("GetChatMembers", chatID).Return([]uuid.UUID{memberID1, memberID2}, nil)
	mockHub.On("SendToUser", memberID1, mock.MatchedBy(func(msg websocket.Message) bool {
		content, ok := msg.Content.(map[string]interface{})
		return ok && msg.Type == "message_deleted" && content["message_id"] == messageID && content["chat_id"] == chatID
	})).Return()
	mockHub.On("SendToUser", memberID2, mock.MatchedBy(func(msg websocket.Message) bool {
		content, ok := msg.Content.(map[string]interface{})
		return ok && msg.Type == "message_deleted" && content["message_id"] == messageID && content["chat_id"] == chatID
	})).Return()

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.DeleteMessage(messageID, senderID)

	assert.NoError(t, err)
	mockMessageRepo.AssertCalled(t, "DeleteMessage", messageID, senderID)
	mockChatRepo.AssertCalled(t, "GetChatMembers", chatID)
	mockHub.AssertCalled(t, "SendToUser", memberID1, mock.Anything)
	mockHub.AssertCalled(t, "SendToUser", memberID2, mock.Anything)
}

func TestDeleteMessage_DeleteError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	messageID := uuid.New()
	senderID := uuid.New()

	mockMessageRepo.On("DeleteMessage", messageID, senderID).Return(uuid.Nil, errors.New("delete error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.DeleteMessage(messageID, senderID)

	assert.Error(t, err)
	assert.Equal(t, "delete error", err.Error())
	mockMessageRepo.AssertCalled(t, "DeleteMessage", messageID, senderID)
	mockChatRepo.AssertNotCalled(t, "GetChatMembers")
}

func TestDeleteMessage_GetChatMembersError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	messageID := uuid.New()
	senderID := uuid.New()
	chatID := uuid.New()

	mockMessageRepo.On("DeleteMessage", messageID, senderID).Return(chatID, nil)
	mockChatRepo.On("GetChatMembers", chatID).Return(nil, errors.New("db error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	err := messageService.DeleteMessage(messageID, senderID)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	mockMessageRepo.AssertCalled(t, "DeleteMessage", messageID, senderID)
	mockChatRepo.AssertCalled(t, "GetChatMembers", chatID)
	mockHub.AssertNotCalled(t, "SendToUser", mock.Anything, mock.Anything)
}

func TestEditMessage_Success(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	messageID := uuid.New()
	senderID := uuid.New()
	chatID := uuid.New()
	content := "Updated content"
	msg := &model.Message{ID: messageID, ChatID: chatID, SenderID: senderID, Content: content}

	mockMessageRepo.On("EditMessage", messageID, senderID, content).Return(msg, nil)
	mockChatRepo.On("GetChatMembers", chatID).Return([]uuid.UUID{senderID}, nil)
	mockHub.On("SendToUser", senderID, mock.MatchedBy(func(m websocket.Message) bool {
		return m.Type == "message_edited" && m.Content.(*model.Message).Content == content
	})).Return()

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	result, err := messageService.EditMessage(messageID, senderID, content)

	assert.NoError(t, err)
	assert.Equal(t, content, result.Content)
	mockMessageRepo.AssertExpectations(t)
	mockChatRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)
}

func TestEditMessage_EmptyContent(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	messageID := uuid.New()
	senderID := uuid.New()

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	_, err := messageService.EditMessage(messageID, senderID, "")

	assert.Error(t, err)
	assert.Equal(t, "содержимое сообщения не может быть пустым", err.Error())
}

func TestEditMessage_RepoError(t *testing.T) {
	mockMessageRepo := new(MockMessageRepository)
	mockChatRepo := new(MockChatRepository)
	mockHub := new(MockHub)

	messageID := uuid.New()
	senderID := uuid.New()
	content := "Updated content"

	mockMessageRepo.On("EditMessage", messageID, senderID, content).Return(nil, errors.New("db error"))

	messageService := service.NewMessageService(mockMessageRepo, mockChatRepo, mockHub)
	_, err := messageService.EditMessage(messageID, senderID, content)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}
