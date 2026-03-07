package service

import (
	"log"
	"messenger/internal/model"
	"messenger/internal/repository"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ── Кеш preferences ──────────────────────────────────────────────────────────

type cachedPrefs struct {
	wi, wv, wt float64
	expiresAt  time.Time
}

// ── Сервис ───────────────────────────────────────────────────────────────────

type FeedService struct {
	repo     *repository.FeedRepository
	wallRepo *repository.WallRepository

	// Кеш предпочтений — TTL 5 минут, не ходим в БД на каждый GET /feed
	prefCache sync.Map // uuid.UUID → cachedPrefs

	// Буфер событий — флашим пачками каждые 2с или при 100 событиях
	eventBuf chan *eventJob
	stopCh   chan struct{}
}

type eventJob struct {
	event    *model.FeedEvent
	postID   uuid.UUID
	userID   uuid.UUID
	evtType  string
	watchSec int
	mimeType string
}

func NewFeedService(repo *repository.FeedRepository, wallRepo *repository.WallRepository) *FeedService {
	s := &FeedService{
		repo:     repo,
		wallRepo: wallRepo,
		eventBuf: make(chan *eventJob, 1000), // буфер на 1000 событий
		stopCh:   make(chan struct{}),
	}
	go s.flushLoop()
	return s
}

// Stop — корректная остановка (вызвать при shutdown)
func (s *FeedService) Stop() {
	close(s.stopCh)
}

// ── Батчинг событий ───────────────────────────────────────────────────────────

func (s *FeedService) flushLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	batch := make([]*eventJob, 0, 100)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		// Копируем и сбрасываем
		toFlush := make([]*eventJob, len(batch))
		copy(toFlush, batch)
		batch = batch[:0]

		// Флашим в горутине чтобы не блокировать тикер
		go s.writeBatch(toFlush)
	}

	for {
		select {
		case job := <-s.eventBuf:
			batch = append(batch, job)
			if len(batch) >= 100 {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-s.stopCh:
			// Финальный flush при остановке
			for len(s.eventBuf) > 0 {
				batch = append(batch, <-s.eventBuf)
			}
			flush()
			return
		}
	}
}

func (s *FeedService) writeBatch(jobs []*eventJob) {
	events := make([]*model.FeedEvent, len(jobs))
	for i, j := range jobs {
		events[i] = j.event
	}

	if err := s.repo.TrackEventBatch(events); err != nil {
		log.Printf("FeedService: ошибка записи батча событий: %v", err)
	}

	// UpsertScore и UpdatePreferences — по каждому событию отдельно
	// (они обновляют агрегаты, батчить сложнее)
	for _, j := range jobs {
		if err := s.repo.UpsertScore(j.postID, j.userID, j.evtType, j.watchSec); err != nil {
			log.Printf("FeedService: UpsertScore error: %v", err)
		}
		if j.evtType != "skip" && j.mimeType != "" {
			if err := s.repo.UpdatePreferences(j.userID, j.mimeType); err != nil {
				log.Printf("FeedService: UpdatePreferences error: %v", err)
			}
			// Инвалидируем кеш после обновления предпочтений
			s.prefCache.Delete(j.userID)
		}
	}
}

// ── Публичные методы ──────────────────────────────────────────────────────────

// TrackEvent — кладёт событие в буфер и возвращается мгновенно.
func (s *FeedService) TrackEvent(userID uuid.UUID, req *model.TrackEventRequest, mimeType string) {
	job := &eventJob{
		event: &model.FeedEvent{
			UserID:       userID,
			PostID:       req.PostID,
			EventType:    req.EventType,
			WatchSeconds: req.WatchSeconds,
		},
		postID:   req.PostID,
		userID:   userID,
		evtType:  req.EventType,
		watchSec: req.WatchSeconds,
		mimeType: mimeType,
	}

	select {
	case s.eventBuf <- job:
		// ok
	default:
		// Буфер переполнен — пишем напрямую чтобы не терять события
		log.Printf("FeedService: eventBuf full, writing directly")
		go s.writeBatch([]*eventJob{job})
	}
}

// GetFeed — персональная лента с кешем preferences + параллельным cold start.
func (s *FeedService) GetFeed(userID uuid.UUID, limit, offset int) ([]model.WallPost, error) {
	if limit <= 0 {
		limit = 30
	}

	// Берём веса из кеша (без запроса в БД если кеш тёплый)
	wi, wv, wt := s.GetCachedPrefs(userID)

	// Запускаем персональную и глобальную ленты параллельно —
	// глобальная нужна только для cold start (offset=0), горутина дешёвая
	personal, global, err := s.repo.GetPersonalFeedAndGlobal(userID, limit, offset, wi, wv, wt)
	if err != nil {
		return nil, err
	}

	// Cold start: добиваем глобальной лентой (она уже готова, без доп. запроса)
	if len(personal) < limit/2 && offset == 0 && len(global) > 0 {
		seen := make(map[uuid.UUID]bool, len(personal))
		for _, p := range personal {
			seen[p.ID] = true
		}
		for _, p := range global {
			if !seen[p.ID] {
				personal = append(personal, p)
				if len(personal) >= limit {
					break
				}
			}
		}
	}

	return personal, nil
}

// GetCachedPrefs — возвращает веса из кеша (используется репозиторием напрямую).
func (s *FeedService) GetCachedPrefs(userID uuid.UUID) (wi, wv, wt float64) {
	wi, wv, wt = 0.33, 0.33, 0.34
	if v, ok := s.prefCache.Load(userID); ok {
		p := v.(cachedPrefs)
		if time.Now().Before(p.expiresAt) {
			return p.wi, p.wv, p.wt
		}
	}
	// Промах — грузим из БД
	if err := s.repo.DB().QueryRow(
		`SELECT weight_image, weight_video, weight_text FROM feed_preferences WHERE user_id = $1`,
		userID,
	).Scan(&wi, &wv, &wt); err == nil {
		s.prefCache.Store(userID, cachedPrefs{wi, wv, wt, time.Now().Add(5 * time.Minute)})
	}
	return
}
