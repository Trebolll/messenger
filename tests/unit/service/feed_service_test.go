package service

import (
	"database/sql"
	"messenger/internal/model"
	"messenger/internal/service"
	"testing"
	"time"

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

func (m *MockFeedWallRepo) GetWall(userID uuid.UUID, limit, offset int) ([]model.WallPost, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]model.WallPost), args.Error(1)
}

func TestFeedService_GetCachedPrefs_CacheHit(t *testing.T) {
	mockRepo := new(MockFeedRepo)
	userID := uuid.New()
	s := service.NewFeedService(mockRepo, nil)
	defer s.Stop()

	// First call - cache miss, should call repo
	mockRepo.On("GetPreferences", userID).Return(0.1, 0.2, 0.7, nil).Once()
	wi, wv, wt := s.GetCachedPrefs(userID)
	assert.Equal(t, 0.1, wi)
	assert.Equal(t, 0.2, wv)
	assert.Equal(t, 0.7, wt)

	// Second call - cache hit, should NOT call repo again
	wi, wv, wt = s.GetCachedPrefs(userID)
	assert.Equal(t, 0.1, wi)
	assert.Equal(t, 0.2, wv)
	assert.Equal(t, 0.7, wt)

	mockRepo.AssertExpectations(t)
}

func TestFeedService_TrackEvent_And_Flush(t *testing.T) {
	mockRepo := new(MockFeedRepo)
	userID := uuid.New()
	postID := uuid.New()
	s := service.NewFeedService(mockRepo, nil)

	req := &model.TrackEventRequest{
		PostID:       postID,
		EventType:    "click",
		WatchSeconds: 10,
	}

	// We expect these to be called when flushed
	mockRepo.On("TrackEventBatch", mock.MatchedBy(func(events []*model.FeedEvent) bool {
		return len(events) == 1 && events[0].PostID == postID
	})).Return(nil).Once()
	mockRepo.On("UpsertScore", postID, userID, "click", 10).Return(nil).Once()
	mockRepo.On("UpdatePreferences", userID, "video/mp4").Return(nil).Once()

	s.TrackEvent(userID, req, "video/mp4")

	// Trigger flush by stopping
	s.Stop()

	// Give it some time for the goroutine to finish writeBatch
	time.Sleep(200 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

func TestFeedService_TrackEvent_SkipUpdatePrefs(t *testing.T) {
	mockRepo := new(MockFeedRepo)
	userID := uuid.New()
	postID := uuid.New()
	s := service.NewFeedService(mockRepo, nil)

	req := &model.TrackEventRequest{
		PostID:       postID,
		EventType:    "skip",
		WatchSeconds: 0,
	}

	mockRepo.On("TrackEventBatch", mock.Anything).Return(nil).Once()
	mockRepo.On("UpsertScore", postID, userID, "skip", 0).Return(nil).Once()

	s.TrackEvent(userID, req, "video/mp4")
	s.Stop()
	time.Sleep(200 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

func TestFeedService_GetFeed_RepoError(t *testing.T) {
	mockRepo := new(MockFeedRepo)
	userID := uuid.New()

	mockRepo.On("GetPreferences", userID).Return(0.33, 0.33, 0.34, nil)
	mockRepo.On("GetPersonalFeedAndGlobal", userID, 30, 0, 0.33, 0.33, 0.34).Return([]model.WallPost{}, []model.WallPost{}, assert.AnError)

	s := service.NewFeedService(mockRepo, nil)
	defer s.Stop()

	res, err := s.GetFeed(userID, 30, 0)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestFeedService_GetFeed_Success(t *testing.T) {
	mockRepo := new(MockFeedRepo)
	userID := uuid.New()
	posts := []model.WallPost{{ID: uuid.New(), Content: "Personal post"}}

	// Setup mock to return default prefs
	mockRepo.On("GetPreferences", userID).Return(0.33, 0.33, 0.34, sql.ErrNoRows)
	mockRepo.On("GetPersonalFeedAndGlobal", userID, 30, 0, 0.33, 0.33, 0.34).Return(posts, []model.WallPost{}, nil)

	s := service.NewFeedService(mockRepo, nil)
	defer s.Stop()

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
	defer s.Stop()

	res, err := s.GetFeed(userID, 10, 0)

	assert.NoError(t, err)
	// Should contain personal post + 1 global post (skipping duplicate)
	assert.Len(t, res, 2)
	mockRepo.AssertExpectations(t)
}
