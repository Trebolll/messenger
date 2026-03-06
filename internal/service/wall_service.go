package service

import (
	"messenger/internal/model"
	"messenger/internal/repository"

	"github.com/google/uuid"
)

type WallService struct {
	repo *repository.WallRepository
}

func NewWallService(repo *repository.WallRepository) *WallService {
	return &WallService{repo: repo}
}

func (s *WallService) InitWall(userID uuid.UUID) error {
	return s.repo.CreateWall(userID)
}

func (s *WallService) UpdateSettings(userID uuid.UUID, bio string) error {
	return s.repo.UpdateWallSettings(userID, bio)
}

func (s *WallService) CreatePost(post *model.WallPost) error {
	return s.repo.CreatePost(post)
}

func (s *WallService) GetWall(userID uuid.UUID, viewerID uuid.UUID) (*model.WallResponse, error) {
	wall, err := s.repo.GetWallByUserID(userID)
	if err != nil {
		s.repo.CreateWall(userID)
		wall = &model.Wall{UserID: userID}
	}

	posts, err := s.repo.GetPostsByUserID(userID, viewerID)
	if err != nil {
		return nil, err
	}

	media, err := s.repo.GetAllMediaByUserID(userID)
	if err != nil {
		media = []model.WallAttachment{}
	}

	return &model.WallResponse{
		Wall:  *wall,
		Posts: posts,
		Media: media,
	}, nil
}

func (s *WallService) GetPostChat(postID uuid.UUID) (uuid.UUID, error) {
	return s.repo.GetPostChat(postID)
}

func (s *WallService) ToggleLike(postID uuid.UUID, userID uuid.UUID) (bool, int, error) {
	return s.repo.ToggleLike(postID, userID)
}

func (s *WallService) GetPostOwner(postID uuid.UUID) (uuid.UUID, error) {
	return s.repo.GetPostOwner(postID)
}

func (s *WallService) DeletePost(postID uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
	return s.repo.DeletePost(postID, userID)
}

func (s *WallService) DeleteAttachment(attID uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteAttachment(attID, userID)
}

func (s *WallService) AddAttachment(att *model.WallAttachment) error {
	return s.repo.AddAttachment(att)
}
