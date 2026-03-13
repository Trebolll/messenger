package service

import (
	"database/sql"
	"messenger/internal/model"
	"messenger/internal/service"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFeedRepo struct {
	mock.Mock
}

func (m *MockFeedRepo) TrackEventBatch(events []*model.FeedEvent) error {
	args := m.Called(events)
	return args.Error(0)
}

func (m *MockFeedRepo) UpsertScore(postID, userID uuid.UUID, eventType string, watchSec int) error {
	args := m.Called(postID, userID, eventType, watchSec)
	return args.Error(0)
}

func (m *MockFeedRepo) UpdatePreferences(userID uuid.UUID, mimeType string) error {
	args := m.Called(userID, mimeType)
	return args.Error(0)
}

func (m *MockFeedRepo) GetPersonalFeedAndGlobal(userID uuid.UUID, limit, offset int, wi, wv, wt float64) ([]model.WallPost, []model.WallPost, error) {
	args := m.Called(userID, limit, offset, wi, wv, wt)
	return args.Get(0).([]model.WallPost), args.Get(1).([]model.WallPost), args.Error(2)
}

func (m *MockFeedRepo) GetPreferences(userID uuid.UUID) (float64, float64, float64, error) {
	args := m.Called(userID)
	return args.Get(0).(float64), args.Get(1).(float64), args.Get(2).(float64), args.Error(3)
}

func (m *MockFeedRepo) DB() *sql.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*sql.DB)
}

type MockFeedWallRepo struct {
	mock.Mock
}

func TestFeedService_GetFeed_Success(t *testing.T) {
	mockRepo := new(MockFeedRepo)
	userID := uuid.New()
	posts := []model.WallPost{{ID: uuid.New(), Content: "Personal post"}}

	// Setup mock to return default prefs
	mockRepo.On("GetPreferences", userID).Return(0.33, 0.33, 0.34, sql.ErrNoRows)
	mockRepo.On("GetPersonalFeedAndGlobal", userID, 30, 0, 0.33, 0.33, 0.34).Return(posts, []model.WallPost{}, nil)

	s := service.NewFeedService(mockRepo, nil)
	// Stop flushLoop to avoid background calls that might interfere with assertions
	s.Stop()

	res, err := s.GetFeed(userID, 30, 0)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "Personal post", res[0].Content)
	mockRepo.AssertExpectations(t)
}

func TestFeedService_GetFeed_ColdStart(t *testing.T) {
	mockRepo := new(MockFeedRepo)
	userID := uuid.New()
	personal := []model.WallPost{{ID: uuid.New(), Content: "Personal post"}}
	global := []model.WallPost{
		{ID: personal[0].ID, Content: "Personal post"}, // Duplicate
		{ID: uuid.New(), Content: "Global post"},
	}

	mockRepo.On("GetPreferences", userID).Return(0.33, 0.33, 0.34, sql.ErrNoRows)
	mockRepo.On("GetPersonalFeedAndGlobal", userID, 10, 0, 0.33, 0.33, 0.34).Return(personal, global, nil)

	s := service.NewFeedService(mockRepo, nil)
	s.Stop()

	res, err := s.GetFeed(userID, 10, 0)

	assert.NoError(t, err)
	// Should contain personal post + 1 global post (skipping duplicate)
	assert.Len(t, res, 2)
	mockRepo.AssertExpectations(t)
}
