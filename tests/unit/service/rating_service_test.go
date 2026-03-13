package service

import (
	"messenger/internal/model"
	"messenger/internal/service"
	"messenger/internal/service/websocket"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRatingRepo struct {
	mock.Mock
}

func (m *MockRatingRepo) Vote(messageID, voterID uuid.UUID, vote int) (*model.RatingVoteResult, error) {
	args := m.Called(messageID, voterID, vote)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RatingVoteResult), args.Error(1)
}

func (m *MockRatingRepo) GetUserRating(userID uuid.UUID) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockRatingRepo) GetVotesForMessagesFixed(msgs []model.Message, voterID uuid.UUID) (map[uuid.UUID]int, error) {
	args := m.Called(msgs, voterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]int), args.Error(1)
}

type MockChatRepo struct {
	mock.Mock
}

func (m *MockChatRepo) GetChatMembers(chatID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

type MockRatingHub struct {
	mock.Mock
}

func (m *MockRatingHub) SendToUser(userID uuid.UUID, msg websocket.Message) {
	m.Called(userID, msg)
}

func TestRatingService_Vote_Success(t *testing.T) {
	mockRepo := new(MockRatingRepo)
	mockChatRepo := new(MockChatRepo)
	mockHub := new(MockRatingHub)

	msgID := uuid.New()
	voterID := uuid.New()
	senderID := uuid.New()
	chatID := uuid.New()

	result := &model.RatingVoteResult{
		MessageID:    msgID,
		ChatID:       chatID,
		SenderID:     senderID,
		Likes:        1,
		Dislikes:     0,
		MyVote:       1,
		SenderRating: 10,
	}

	mockRepo.On("Vote", msgID, voterID, 1).Return(result, nil)
	mockChatRepo.On("GetChatMembers", chatID).Return([]uuid.UUID{voterID, senderID}, nil)
	mockHub.On("SendToUser", mock.Anything, mock.Anything).Return()

	s := service.NewRatingService(mockRepo, mockChatRepo, mockHub)

	err := s.Vote(msgID, voterID, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockChatRepo.AssertExpectations(t)
	mockHub.AssertExpectations(t)
}

func TestRatingService_Vote_InvalidVoteValue(t *testing.T) {
	s := service.NewRatingService(nil, nil, nil)
	err := s.Vote(uuid.New(), uuid.New(), 2)
	assert.Error(t, err)
	assert.Equal(t, "голос должен быть 1 или -1", err.Error())
}

func TestRatingService_GetUserRating_Success(t *testing.T) {
	mockRepo := new(MockRatingRepo)
	userID := uuid.New()
	mockRepo.On("GetUserRating", userID).Return(15, nil)

	s := service.NewRatingService(mockRepo, nil, nil)
	rating, err := s.GetUserRating(userID)

	assert.NoError(t, err)
	assert.Equal(t, 15, rating)
}

func TestRatingService_GetUserRating_NegativeAsZero(t *testing.T) {
	mockRepo := new(MockRatingRepo)
	userID := uuid.New()
	mockRepo.On("GetUserRating", userID).Return(-5, nil)

	s := service.NewRatingService(mockRepo, nil, nil)
	rating, err := s.GetUserRating(userID)

	assert.NoError(t, err)
	assert.Equal(t, 0, rating)
}
