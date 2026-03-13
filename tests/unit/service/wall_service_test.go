package service

import (
	"messenger/internal/model"
	"messenger/internal/service"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWallRepo struct {
	mock.Mock
}

func (m *MockWallRepo) CreateWall(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockWallRepo) GetWallByUserID(userID uuid.UUID) (*model.Wall, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Wall), args.Error(1)
}

func (m *MockWallRepo) UpdateWallSettings(userID uuid.UUID, bio string) error {
	args := m.Called(userID, bio)
	return args.Error(0)
}

func (m *MockWallRepo) CreatePost(post *model.WallPost) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockWallRepo) GetPostsByUserID(userID uuid.UUID, viewerID uuid.UUID) ([]model.WallPost, error) {
	args := m.Called(userID, viewerID)
	return args.Get(0).([]model.WallPost), args.Error(1)
}

func (m *MockWallRepo) GetAllMediaByUserID(userID uuid.UUID) ([]model.WallAttachment, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.WallAttachment), args.Error(1)
}

func (m *MockWallRepo) GetPostChat(postID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(postID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockWallRepo) GetTotalWallLikes(userID uuid.UUID) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockWallRepo) ToggleLike(postID uuid.UUID, userID uuid.UUID) (bool, int, error) {
	args := m.Called(postID, userID)
	return args.Bool(0), args.Int(1), args.Error(2)
}

func (m *MockWallRepo) GetPostOwner(postID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(postID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockWallRepo) DeletePost(postID uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(postID, userID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockWallRepo) DeleteAttachment(attID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(attID, userID)
	return args.Error(0)
}

func (m *MockWallRepo) AddAttachment(att *model.WallAttachment) error {
	args := m.Called(att)
	return args.Error(0)
}

func (m *MockWallRepo) GetGlobalMediaFeed(viewerID uuid.UUID) ([]model.WallPost, error) {
	args := m.Called(viewerID)
	return args.Get(0).([]model.WallPost), args.Error(1)
}

func TestWallService_GetWall_Success(t *testing.T) {
	mockRepo := new(MockWallRepo)
	userID := uuid.New()
	viewerID := uuid.New()

	wall := &model.Wall{UserID: userID, Username: "testuser"}
	posts := []model.WallPost{{ID: uuid.New(), Content: "Post 1"}}
	media := []model.WallAttachment{{ID: uuid.New(), Url: "url1"}}

	mockRepo.On("GetWallByUserID", userID).Return(wall, nil)
	mockRepo.On("GetPostsByUserID", userID, viewerID).Return(posts, nil)
	mockRepo.On("GetAllMediaByUserID", userID).Return(media, nil)

	s := service.NewWallService(mockRepo)
	res, err := s.GetWall(userID, viewerID)

	assert.NoError(t, err)
	assert.Equal(t, "testuser", res.Wall.Username)
	assert.Len(t, res.Posts, 1)
	assert.Len(t, res.Media, 1)
	mockRepo.AssertExpectations(t)
}

func TestWallService_ToggleLike_Success(t *testing.T) {
	mockRepo := new(MockWallRepo)
	postID := uuid.New()
	userID := uuid.New()

	mockRepo.On("ToggleLike", postID, userID).Return(true, 5, nil)

	s := service.NewWallService(mockRepo)
	liked, count, err := s.ToggleLike(postID, userID)

	assert.NoError(t, err)
	assert.True(t, liked)
	assert.Equal(t, 5, count)
	mockRepo.AssertExpectations(t)
}
