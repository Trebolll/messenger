package service

import (
	"database/sql"
	"errors"
	"messenger/internal/model"
	"messenger/internal/service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockChatRepositoryForChat struct {
	mock.Mock
}

func (m *MockChatRepositoryForChat) CreatePrivateChat(initiatorID, targetUserID uuid.UUID) (*model.Chat, error) {
	args := m.Called(initiatorID, targetUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepositoryForChat) CreateGroupChat(name string, memberIDs []uuid.UUID) (*model.Chat, error) {
	args := m.Called(name, memberIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepositoryForChat) GetUserChats(userID uuid.UUID) ([]model.ChatListItem, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ChatListItem), args.Error(1)
}

func (m *MockChatRepositoryForChat) IsChatMember(chatID, userID uuid.UUID) (bool, error) {
	args := m.Called(chatID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChatRepositoryForChat) UpdateGroupAvatarUrl(chatID uuid.UUID, url string) error {
	args := m.Called(chatID, url)
	return args.Error(0)
}

func (m *MockChatRepositoryForChat) GetChatByID(chatID uuid.UUID) (*model.Chat, error) {
	args := m.Called(chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Chat), args.Error(1)
}

func (m *MockChatRepositoryForChat) UpdateGroupChat(chatID uuid.UUID, name string, avatarUrl string) error {
	args := m.Called(chatID, name, avatarUrl)
	return args.Error(0)
}

func (m *MockChatRepositoryForChat) RemoveChatMember(chatID, userID uuid.UUID) error {
	args := m.Called(chatID, userID)
	return args.Error(0)
}

func (m *MockChatRepositoryForChat) AddChatMember(chatID, userID uuid.UUID) error {
	args := m.Called(chatID, userID)
	return args.Error(0)
}

func (m *MockChatRepositoryForChat) GetMembersInfo(chatID uuid.UUID) ([]model.ChatMemberInfo, error) {
	args := m.Called(chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ChatMemberInfo), args.Error(1)
}

type MockUserRepositoryForChat struct {
	mock.Mock
}

func (m *MockUserRepositoryForChat) GetByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepositoryForChat) GetByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepositoryForChat) Create(u *model.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockUserRepositoryForChat) GetById(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepositoryForChat) SearchByUsername(username string) ([]model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

type MockHubForChat struct {
	mock.Mock
}

func (m *MockHubForChat) SendToUser(userID uuid.UUID, msg interface{}) {
	m.Called(userID, msg)
}

func (m *MockHubForChat) Run() {
	m.Called()
}

func (m *MockHubForChat) IsUserOnline(userID uuid.UUID) bool {
	args := m.Called(userID)
	return args.Bool(0)
}

func TestCreatePrivateChat_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	user0ID := uuid.New()
	user1ID := uuid.New()
	chatID := uuid.New()

	user0 := &model.User{
		ID:       user0ID,
		Username: "user0",
		Email:    "user0@example.com",
	}
	user1 := &model.User{
		ID:       user1ID,
		Username: "user1",
		Email:    "user1@example.com",
	}
	chat := &model.Chat{
		ID:        chatID,
		Type:      model.TypePrivate,
		CreatedAt: time.Now(),
	}

	mockUserRepo.On("GetById", user0ID).Return(user0, nil)
	mockUserRepo.On("GetById", user1ID).Return(user1, nil)
	mockChatRepo.On("CreatePrivateChat", user0ID, user1ID).Return(chat, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreatePrivateChat(user0ID, user1ID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, chatID, result.ID)
	assert.Equal(t, model.TypePrivate, result.Type)
	mockUserRepo.AssertCalled(t, "GetById", user0ID)
	mockUserRepo.AssertCalled(t, "GetById", user1ID)
	mockChatRepo.AssertCalled(t, "CreatePrivateChat", user0ID, user1ID)
	mockUserRepo.AssertExpectations(t)
	mockChatRepo.AssertExpectations(t)
}

func TestCreatePrivateChat_User0NotFound(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	user0ID := uuid.New()
	user1ID := uuid.New()

	mockUserRepo.On("GetById", user0ID).Return(nil, sql.ErrNoRows)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreatePrivateChat(user0ID, user1ID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "пользователь с ID")
	assert.Contains(t, err.Error(), "не найден")
	mockUserRepo.AssertCalled(t, "GetById", user0ID)
	mockUserRepo.AssertNotCalled(t, "GetById", user1ID)
	mockChatRepo.AssertNotCalled(t, "CreatePrivateChat")
}

func TestCreatePrivateChat_User1NotFound(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	user0ID := uuid.New()
	user1ID := uuid.New()

	user0 := &model.User{
		ID:       user0ID,
		Username: "user0",
		Email:    "user0@example.com",
	}

	mockUserRepo.On("GetById", user0ID).Return(user0, nil)
	mockUserRepo.On("GetById", user1ID).Return(nil, sql.ErrNoRows)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreatePrivateChat(user0ID, user1ID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "пользователь с ID")
	assert.Contains(t, err.Error(), "не найден")
	mockUserRepo.AssertCalled(t, "GetById", user0ID)
	mockUserRepo.AssertCalled(t, "GetById", user1ID)
	mockChatRepo.AssertNotCalled(t, "CreatePrivateChat")
}

func TestCreatePrivateChat_GetByIdError(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	user0ID := uuid.New()
	user1ID := uuid.New()

	mockUserRepo.On("GetById", user0ID).Return(nil, errors.New("database error"))

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreatePrivateChat(user0ID, user1ID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())
	mockUserRepo.AssertCalled(t, "GetById", user0ID)
	mockUserRepo.AssertNotCalled(t, "GetById", user1ID)
	mockChatRepo.AssertNotCalled(t, "CreatePrivateChat")
}

func TestCreatePrivateChat_CreateChatError(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	user0ID := uuid.New()
	user1ID := uuid.New()

	user0 := &model.User{
		ID:       user0ID,
		Username: "user0",
		Email:    "user0@example.com",
	}
	user1 := &model.User{
		ID:       user1ID,
		Username: "user1",
		Email:    "user1@example.com",
	}

	mockUserRepo.On("GetById", user0ID).Return(user0, nil)
	mockUserRepo.On("GetById", user1ID).Return(user1, nil)
	mockChatRepo.On("CreatePrivateChat", user0ID, user1ID).Return(nil, errors.New("database error"))

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreatePrivateChat(user0ID, user1ID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())
	mockUserRepo.AssertCalled(t, "GetById", user0ID)
	mockUserRepo.AssertCalled(t, "GetById", user1ID)
	mockChatRepo.AssertCalled(t, "CreatePrivateChat", user0ID, user1ID)
	mockChatRepo.AssertExpectations(t)
}

func TestCreateGroupChatByUsernames_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	creatorID := uuid.New()
	user1ID := uuid.New()
	user2ID := uuid.New()
	chatID := uuid.New()

	user1 := &model.User{
		ID:       user1ID,
		Username: "user1",
		Email:    "user1@example.com",
	}
	user2 := &model.User{
		ID:       user2ID,
		Username: "user2",
		Email:    "user2@example.com",
	}
	chat := &model.Chat{
		ID:        chatID,
		Type:      model.TypeGroup,
		Name:      "Test Group",
		CreatedAt: time.Now(),
	}

	mockUserRepo.On("GetByUsername", "user1").Return(user1, nil)
	mockUserRepo.On("GetByUsername", "user2").Return(user2, nil)
	mockChatRepo.On("CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 3 && ids[0] == creatorID
	})).Return(chat, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreateGroupChatByUsernames("Test Group", []string{"user1", "user2"}, creatorID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, chatID, result.ID)
	assert.Equal(t, "Test Group", result.Name)
	mockUserRepo.AssertCalled(t, "GetByUsername", "user1")
	mockUserRepo.AssertCalled(t, "GetByUsername", "user2")
	mockChatRepo.AssertCalled(t, "CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 3 && ids[0] == creatorID
	}))
	mockChatRepo.AssertExpectations(t)
}

func TestCreateGroupChatByUsernames_UserNotFound(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	creatorID := uuid.New()

	mockUserRepo.On("GetByUsername", "user1").Return(nil, errors.New("user not found"))

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreateGroupChatByUsernames("Test Group", []string{"user1", "user2"}, creatorID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "пользователь user1 не найден")
	mockUserRepo.AssertCalled(t, "GetByUsername", "user1")
	mockUserRepo.AssertNotCalled(t, "GetByUsername", "user2")
	mockChatRepo.AssertNotCalled(t, "CreateGroupChat")
}

func TestCreateGroupChatByUsernames_DuplicateUsernames(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	creatorID := uuid.New()
	user1ID := uuid.New()
	chatID := uuid.New()

	user1 := &model.User{
		ID:       user1ID,
		Username: "user1",
		Email:    "user1@example.com",
	}
	chat := &model.Chat{
		ID:        chatID,
		Type:      model.TypeGroup,
		Name:      "Test Group",
		CreatedAt: time.Now(),
	}

	mockUserRepo.On("GetByUsername", "user1").Return(user1, nil)
	mockChatRepo.On("CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2 && ids[0] == creatorID && ids[1] == user1ID
	})).Return(chat, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreateGroupChatByUsernames("Test Group", []string{"user1", "user1"}, creatorID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockUserRepo.AssertCalled(t, "GetByUsername", "user1")
	mockChatRepo.AssertCalled(t, "CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2
	}))
}

func TestCreateGroupChatByUsernames_CreatorInUsernames(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	creatorID := uuid.New()
	user1ID := uuid.New()
	chatID := uuid.New()

	creator := &model.User{
		ID:       creatorID,
		Username: "creator",
		Email:    "creator@example.com",
	}
	user1 := &model.User{
		ID:       user1ID,
		Username: "user1",
		Email:    "user1@example.com",
	}
	chat := &model.Chat{
		ID:        chatID,
		Type:      model.TypeGroup,
		Name:      "Test Group",
		CreatedAt: time.Now(),
	}

	mockUserRepo.On("GetByUsername", "creator").Return(creator, nil)
	mockUserRepo.On("GetByUsername", "user1").Return(user1, nil)
	mockChatRepo.On("CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2 && ids[0] == creatorID && ids[1] == user1ID
	})).Return(chat, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreateGroupChatByUsernames("Test Group", []string{"creator", "user1"}, creatorID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockUserRepo.AssertCalled(t, "GetByUsername", "creator")
	mockUserRepo.AssertCalled(t, "GetByUsername", "user1")
	mockChatRepo.AssertCalled(t, "CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2
	}))
}

func TestCreateGroupChatByUsernames_EmptyUsernames(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	creatorID := uuid.New()
	chatID := uuid.New()

	chat := &model.Chat{
		ID:        chatID,
		Type:      model.TypeGroup,
		Name:      "Test Group",
		CreatedAt: time.Now(),
	}

	mockChatRepo.On("CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 1 && ids[0] == creatorID
	})).Return(chat, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreateGroupChatByUsernames("Test Group", []string{}, creatorID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockUserRepo.AssertNotCalled(t, "GetByUsername")
	mockChatRepo.AssertCalled(t, "CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 1 && ids[0] == creatorID
	}))
}

func TestCreateGroupChatByUsernames_CreateChatError(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	creatorID := uuid.New()
	user1ID := uuid.New()

	user1 := &model.User{
		ID:       user1ID,
		Username: "user1",
		Email:    "user1@example.com",
	}

	mockUserRepo.On("GetByUsername", "user1").Return(user1, nil)
	mockChatRepo.On("CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2 && ids[0] == creatorID && ids[1] == user1ID
	})).Return(nil, errors.New("database error"))

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.CreateGroupChatByUsernames("Test Group", []string{"user1"}, creatorID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())
	mockUserRepo.AssertCalled(t, "GetByUsername", "user1")
	mockChatRepo.AssertCalled(t, "CreateGroupChat", "Test Group", mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2
	}))
	mockChatRepo.AssertExpectations(t)
}

func TestGetUserChats_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	userID := uuid.New()
	interlocutorID1 := uuid.New()
	interlocutorID2 := uuid.New()

	chats := []model.ChatListItem{
		{
			ID:              uuid.New(),
			Type:            model.TypePrivate,
			Name:            "user1",
			LastMessage:     "Hello",
			LastMessageTime: time.Now(),
			InterlocutorID:  &interlocutorID1,
			IsOnline:        false,
		},
		{
			ID:              uuid.New(),
			Type:            model.TypePrivate,
			Name:            "user2",
			LastMessage:     "Hi",
			LastMessageTime: time.Now(),
			InterlocutorID:  &interlocutorID2,
			IsOnline:        false,
		},
	}

	mockChatRepo.On("GetUserChats", userID).Return(chats, nil)
	mockHub.On("IsUserOnline", interlocutorID1).Return(true)
	mockHub.On("IsUserOnline", interlocutorID2).Return(false)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.GetUserChats(userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.True(t, result[0].IsOnline)
	assert.False(t, result[1].IsOnline)
	mockChatRepo.AssertCalled(t, "GetUserChats", userID)
	mockHub.AssertCalled(t, "IsUserOnline", interlocutorID1)
	mockHub.AssertCalled(t, "IsUserOnline", interlocutorID2)
}

func TestGetUserChats_GetUserChatsError(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	userID := uuid.New()

	mockChatRepo.On("GetUserChats", userID).Return(nil, errors.New("database error"))

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.GetUserChats(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())
	mockChatRepo.AssertCalled(t, "GetUserChats", userID)
	mockHub.AssertNotCalled(t, "IsUserOnline")
}

func TestGetUserChats_NoChats(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	userID := uuid.New()

	mockChatRepo.On("GetUserChats", userID).Return([]model.ChatListItem{}, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.GetUserChats(userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "у пользователя пока нет чатов", err.Error())
	mockChatRepo.AssertCalled(t, "GetUserChats", userID)
	mockHub.AssertNotCalled(t, "IsUserOnline")
}

func TestGetUserChats_ChatWithoutInterlocutor(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	userID := uuid.New()

	chats := []model.ChatListItem{
		{
			ID:              uuid.New(),
			Type:            model.TypeGroup,
			Name:            "Group Chat",
			LastMessage:     "Group message",
			LastMessageTime: time.Now(),
			InterlocutorID:  nil,
			IsOnline:        false,
		},
	}

	mockChatRepo.On("GetUserChats", userID).Return(chats, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.GetUserChats(userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "Group Chat", result[0].Name)
	mockChatRepo.AssertCalled(t, "GetUserChats", userID)
	mockHub.AssertNotCalled(t, "IsUserOnline")
}

func TestGetUserChats_MixedChats(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	userID := uuid.New()
	interlocutorID := uuid.New()

	chats := []model.ChatListItem{
		{
			ID:              uuid.New(),
			Type:            model.TypePrivate,
			Name:            "user1",
			LastMessage:     "Hello",
			LastMessageTime: time.Now(),
			InterlocutorID:  &interlocutorID,
			IsOnline:        false,
		},
		{
			ID:              uuid.New(),
			Type:            model.TypeGroup,
			Name:            "Group Chat",
			LastMessage:     "Group message",
			LastMessageTime: time.Now(),
			InterlocutorID:  nil,
			IsOnline:        false,
		},
	}

	mockChatRepo.On("GetUserChats", userID).Return(chats, nil)
	mockHub.On("IsUserOnline", interlocutorID).Return(true)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.GetUserChats(userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.True(t, result[0].IsOnline)
	assert.False(t, result[1].IsOnline)
	mockChatRepo.AssertCalled(t, "GetUserChats", userID)
	mockHub.AssertCalled(t, "IsUserOnline", interlocutorID)
}

func TestGetUserChats_SingleChat(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	userID := uuid.New()
	interlocutorID := uuid.New()

	chats := []model.ChatListItem{
		{
			ID:              uuid.New(),
			Type:            model.TypePrivate,
			Name:            "user1",
			LastMessage:     "Last message",
			LastMessageTime: time.Now(),
			InterlocutorID:  &interlocutorID,
			IsOnline:        false,
		},
	}

	mockChatRepo.On("GetUserChats", userID).Return(chats, nil)
	mockHub.On("IsUserOnline", interlocutorID).Return(true)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.GetUserChats(userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "user1", result[0].Name)
	assert.True(t, result[0].IsOnline)
	mockChatRepo.AssertCalled(t, "GetUserChats", userID)
	mockHub.AssertCalled(t, "IsUserOnline", interlocutorID)
}

func TestUpdateGroupAvatarUrl_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	chatID := uuid.New()
	url := "http://example.com/avatar.png"

	mockChatRepo.On("UpdateGroupAvatarUrl", chatID, url).Return(nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	err := chatService.UpdateGroupAvatarUrl(chatID, url)

	assert.NoError(t, err)
	mockChatRepo.AssertExpectations(t)
}

func TestGetChatByID_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	chatID := uuid.New()
	userID := uuid.New()
	chat := &model.Chat{ID: chatID, Type: model.TypeGroup}

	mockChatRepo.On("IsChatMember", chatID, userID).Return(true, nil)
	mockChatRepo.On("GetChatByID", chatID).Return(chat, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.GetChatByID(chatID, userID)

	assert.NoError(t, err)
	assert.Equal(t, chat, result)
	mockChatRepo.AssertExpectations(t)
}

func TestGetChatByID_AccessDenied(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	chatID := uuid.New()
	userID := uuid.New()

	mockChatRepo.On("IsChatMember", chatID, userID).Return(false, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.GetChatByID(chatID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "access denied", err.Error())
}

func TestUpdateGroupChat_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	chatID := uuid.New()
	creatorID := uuid.New()
	chat := &model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &creatorID, Name: "Old Name"}

	mockChatRepo.On("GetChatByID", chatID).Return(chat, nil)
	mockChatRepo.On("UpdateGroupChat", chatID, "New Name", "").Return(nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.UpdateGroupChat(chatID, creatorID, "New Name", "")

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	mockChatRepo.AssertExpectations(t)
}

func TestUpdateGroupChat_NotCreator(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	chatID := uuid.New()
	creatorID := uuid.New()
	otherUserID := uuid.New()
	chat := &model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &creatorID}

	mockChatRepo.On("GetChatByID", chatID).Return(chat, nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	_, err := chatService.UpdateGroupChat(chatID, otherUserID, "New Name", "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only the creator")
}

func TestRemoveChatMember_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	chatID := uuid.New()
	creatorID := uuid.New()
	targetID := uuid.New()
	chat := &model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &creatorID}

	mockChatRepo.On("GetChatByID", chatID).Return(chat, nil)
	mockChatRepo.On("RemoveChatMember", chatID, targetID).Return(nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	err := chatService.RemoveChatMember(chatID, creatorID, targetID)

	assert.NoError(t, err)
	mockChatRepo.AssertExpectations(t)
}

func TestAddChatMember_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	chatID := uuid.New()
	creatorID := uuid.New()
	targetUser := &model.User{ID: uuid.New(), Username: "newuser"}
	chat := &model.Chat{ID: chatID, Type: model.TypeGroup, CreatorID: &creatorID}

	mockChatRepo.On("GetChatByID", chatID).Return(chat, nil)
	mockUserRepo.On("GetByUsername", "newuser").Return(targetUser, nil)
	mockChatRepo.On("AddChatMember", chatID, targetUser.ID).Return(nil)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	err := chatService.AddChatMember(chatID, creatorID, "newuser")

	assert.NoError(t, err)
	mockChatRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGetGroupMembers_Success(t *testing.T) {
	mockChatRepo := new(MockChatRepositoryForChat)
	mockUserRepo := new(MockUserRepositoryForChat)
	mockHub := new(MockHubForChat)

	chatID := uuid.New()
	userID := uuid.New()
	memberID := uuid.New()
	members := []model.ChatMemberInfo{
		{ID: memberID, Username: "member1"},
	}

	mockChatRepo.On("IsChatMember", chatID, userID).Return(true, nil)
	mockChatRepo.On("GetMembersInfo", chatID).Return(members, nil)
	mockHub.On("IsUserOnline", memberID).Return(true)

	chatService := service.NewChatService(mockChatRepo, mockUserRepo, mockHub)
	result, err := chatService.GetGroupMembers(chatID, userID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.True(t, result[0].IsOnline)
	mockChatRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)
}
