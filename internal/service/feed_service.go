package service

import (
	"messenger/internal/model"
	"messenger/internal/repository"

	"github.com/google/uuid"
)

type FeedService struct {
	repo     *repository.FeedRepository
	wallRepo *repository.WallRepository // для fallback на глобальную ленту
}

func NewFeedService(repo *repository.FeedRepository, wallRepo *repository.WallRepository) *FeedService {
	return &FeedService{repo: repo, wallRepo: wallRepo}
}

// TrackEvent записывает событие и сразу обновляет скор + предпочтения.
// Вызывается асинхронно из хендлера — не блокирует ответ клиенту.
func (s *FeedService) TrackEvent(userID uuid.UUID, req *model.TrackEventRequest, mimeType string) {
	e := &model.FeedEvent{
		UserID:       userID,
		PostID:       req.PostID,
		EventType:    req.EventType,
		WatchSeconds: req.WatchSeconds,
	}
	_ = s.repo.TrackEvent(e)
	_ = s.repo.UpsertScore(req.PostID, userID, req.EventType, req.WatchSeconds)

	// Обновляем предпочтения только для активных событий (не skip)
	if req.EventType != "skip" && mimeType != "" {
		_ = s.repo.UpdatePreferences(userID, mimeType)
	}
}

// GetFeed возвращает персональную ленту. Если постов меньше limit/2 —
// добивает глобальной лентой (cold start / мало контента).
func (s *FeedService) GetFeed(userID uuid.UUID, limit, offset int) ([]model.WallPost, error) {
	if limit <= 0 {
		limit = 30
	}

	posts, err := s.repo.GetPersonalFeed(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Cold start: если персональных постов мало — добавляем глобальные
	if len(posts) < limit/2 && offset == 0 {
		global, err := s.wallRepo.GetGlobalMediaFeed(userID)
		if err == nil {
			// Дедупликация по ID
			seen := make(map[uuid.UUID]bool, len(posts))
			for _, p := range posts {
				seen[p.ID] = true
			}
			for _, p := range global {
				if !seen[p.ID] {
					posts = append(posts, p)
					if len(posts) >= limit {
						break
					}
				}
			}
		}
	}

	return posts, nil
}
