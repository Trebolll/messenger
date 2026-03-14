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

func TestWallService_InitWall(t *testing.T) {
	mockRepo := new(MockWallRepo)
	userID := uuid.New()

	mockRepo.On("CreateWall", userID).Return(nil)

	s := service.NewWallService(mockRepo)
	err := s.InitWall(userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWallService_UpdateSettings(t *testing.T) {
	mockRepo := new(MockWallRepo)
	userID := uuid.New()
	bio := "new bio"

	mockRepo.On("UpdateWallSettings", userID, bio).Return(nil)

	s := service.NewWallService(mockRepo)
	err := s.UpdateSettings(userID, bio)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWallService_CreatePost(t *testing.T) {
	mockRepo := new(MockWallRepo)
	post := &model.WallPost{ID: uuid.New(), Content: "test post"}

	mockRepo.On("CreatePost", post).Return(nil)

	s := service.NewWallService(mockRepo)
	err := s.CreatePost(post)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWallService_GetWall_WallNotFound(t *testing.T) {
	mockRepo := new(MockWallRepo)
	userID := uuid.New()
	viewerID := uuid.New()

	mockRepo.On("GetWallByUserID", userID).Return(nil, assert.AnError)
	mockRepo.On("CreateWall", userID).Return(nil)
	mockRepo.On("GetPostsByUserID", userID, viewerID).Return([]model.WallPost{}, nil)
	mockRepo.On("GetAllMediaByUserID", userID).Return([]model.WallAttachment{}, nil)

	s := service.NewWallService(mockRepo)
	res, err := s.GetWall(userID, viewerID)

	assert.NoError(t, err)
	assert.Equal(t, userID, res.Wall.UserID)
	mockRepo.AssertExpectations(t)
}

func TestWallService_GetWall_PostsError(t *testing.T) {
	mockRepo := new(MockWallRepo)
	userID := uuid.New()
	viewerID := uuid.New()

	mockRepo.On("GetWallByUserID", userID).Return(&model.Wall{UserID: userID}, nil)
	mockRepo.On("GetPostsByUserID", userID, viewerID).Return([]model.WallPost{}, assert.AnError)

	s := service.NewWallService(mockRepo)
	res, err := s.GetWall(userID, viewerID)

	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}

func TestWallService_GetWall_MediaError(t *testing.T) {
	mockRepo := new(MockWallRepo)
	userID := uuid.New()
	viewerID := uuid.New()

	mockRepo.On("GetWallByUserID", userID).Return(&model.Wall{UserID: userID}, nil)
	mockRepo.On("GetPostsByUserID", userID, viewerID).Return([]model.WallPost{}, nil)
	mockRepo.On("GetAllMediaByUserID", userID).Return([]model.WallAttachment{}, assert.AnError)

	s := service.NewWallService(mockRepo)
	res, err := s.GetWall(userID, viewerID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res.Media)
	mockRepo.AssertExpectations(t)
}

func TestWallService_GetPostChat(t *testing.T) {
	mockRepo := new(MockWallRepo)
	postID := uuid.New()
	chatID := uuid.New()

	mockRepo.On("GetPostChat", postID).Return(chatID, nil)

	s := service.NewWallService(mockRepo)
	res, err := s.GetPostChat(postID)

	assert.NoError(t, err)
	assert.Equal(t, chatID, res)
	mockRepo.AssertExpectations(t)
}

func TestWallService_GetTotalWallLikes(t *testing.T) {
	mockRepo := new(MockWallRepo)
	userID := uuid.New()

	mockRepo.On("GetTotalWallLikes", userID).Return(10, nil)

	s := service.NewWallService(mockRepo)
	res, err := s.GetTotalWallLikes(userID)

	assert.NoError(t, err)
	assert.Equal(t, 10, res)
	mockRepo.AssertExpectations(t)
}

func TestWallService_GetPostOwner(t *testing.T) {
	mockRepo := new(MockWallRepo)
	postID := uuid.New()
	ownerID := uuid.New()

	mockRepo.On("GetPostOwner", postID).Return(ownerID, nil)

	s := service.NewWallService(mockRepo)
	res, err := s.GetPostOwner(postID)

	assert.NoError(t, err)
	assert.Equal(t, ownerID, res)
	mockRepo.AssertExpectations(t)
}

func TestWallService_DeletePost(t *testing.T) {
	mockRepo := new(MockWallRepo)
	postID := uuid.New()
	userID := uuid.New()
	deletedID := uuid.New()

	mockRepo.On("DeletePost", postID, userID).Return(deletedID, nil)

	s := service.NewWallService(mockRepo)
	res, err := s.DeletePost(postID, userID)

	assert.NoError(t, err)
	assert.Equal(t, deletedID, res)
	mockRepo.AssertExpectations(t)
}

func TestWallService_DeleteAttachment(t *testing.T) {
	mockRepo := new(MockWallRepo)
	attID := uuid.New()
	userID := uuid.New()

	mockRepo.On("DeleteAttachment", attID, userID).Return(nil)

	s := service.NewWallService(mockRepo)
	err := s.DeleteAttachment(attID, userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWallService_AddAttachment(t *testing.T) {
	mockRepo := new(MockWallRepo)
	att := &model.WallAttachment{ID: uuid.New()}

	mockRepo.On("AddAttachment", att).Return(nil)

	s := service.NewWallService(mockRepo)
	err := s.AddAttachment(att)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWallService_GetGlobalMediaFeed(t *testing.T) {
	mockRepo := new(MockWallRepo)
	viewerID := uuid.New()
	posts := []model.WallPost{{ID: uuid.New()}}

	mockRepo.On("GetGlobalMediaFeed", viewerID).Return(posts, nil)

	s := service.NewWallService(mockRepo)
	res, err := s.GetGlobalMediaFeed(viewerID)

	assert.NoError(t, err)
	assert.Equal(t, posts, res)
	mockRepo.AssertExpectations(t)
}
