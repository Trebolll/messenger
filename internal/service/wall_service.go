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

func (s *WallService) GetWall(userID uuid.UUID) (*model.WallResponse, error) {
	wall, err := s.repo.GetWallByUserID(userID)
	if err != nil {
		// Если по какой-то причине стены нет, создаем её
		s.repo.CreateWall(userID)
		wall = &model.Wall{UserID: userID}
	}

	posts, err := s.repo.GetPostsByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &model.WallResponse{
		Wall:  *wall,
		Posts: posts,
	}, nil
}

func (s *WallService) AddAttachment(att *model.WallAttachment) error {
	return s.repo.AddAttachment(att)
}
