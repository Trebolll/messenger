package repository

import (
	"database/sql"
	"fmt"
	"messenger/internal/model"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type FeedRepository struct {
	db *sql.DB
}

func NewFeedRepository(db *sql.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

// ── Запись события ────────────────────────────────────────────────────────────

func (r *FeedRepository) TrackEvent(e *model.FeedEvent) error {
	_, err := r.db.Exec(`
		INSERT INTO feed_events (user_id, post_id, event_type, watch_seconds)
		VALUES ($1, $2, $3, $4)`,
		e.UserID, e.PostID, e.EventType, e.WatchSeconds,
	)
	return err
}

// ── Обновление скора ──────────────────────────────────────────────────────────

func (r *FeedRepository) UpsertScore(postID, userID uuid.UUID, eventType string, watchSec int) error {
	weights := map[string]float64{
		"like": 10, "comment": 8, "video_complete": 5, "view": 1, "skip": -2,
	}
	delta := weights[eventType]
	likeDelta, commentDelta, viewDelta := 0, 0, 0
	switch eventType {
	case "like":
		likeDelta = 1
	case "comment":
		commentDelta = 1
	case "view", "video_complete":
		viewDelta = 1
	}
	_, err := r.db.Exec(`
		INSERT INTO feed_scores (post_id, user_id, score, view_count, like_count, comment_count, total_watch_sec, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (post_id, user_id) DO UPDATE SET
			score           = feed_scores.score + EXCLUDED.score,
			view_count      = feed_scores.view_count + $4,
			like_count      = feed_scores.like_count + $5,
			comment_count   = feed_scores.comment_count + $6,
			total_watch_sec = feed_scores.total_watch_sec + $7,
			last_seen_at    = NOW()`,
		postID, userID, delta, viewDelta, likeDelta, commentDelta, watchSec,
	)
	return err
}

// ── Обновление предпочтений ───────────────────────────────────────────────────

func (r *FeedRepository) UpdatePreferences(userID uuid.UUID, mimeType string) error {
	var col string
	switch {
	case strings.HasPrefix(mimeType, "image"):
		col = "weight_image"
	case strings.HasPrefix(mimeType, "video"):
		col = "weight_video"
	default:
		col = "weight_text"
	}
	_, err := r.db.Exec(`
		INSERT INTO feed_preferences (user_id, weight_image, weight_video, weight_text)
		VALUES ($1, 0.33, 0.33, 0.34)
		ON CONFLICT (user_id) DO UPDATE SET
			weight_image = CASE WHEN $2 = 'weight_image' THEN LEAST(feed_preferences.weight_image + 0.05, 1.0)
			                    ELSE GREATEST(feed_preferences.weight_image - 0.01, 0.0) END,
			weight_video = CASE WHEN $2 = 'weight_video' THEN LEAST(feed_preferences.weight_video + 0.05, 1.0)
			                    ELSE GREATEST(feed_preferences.weight_video - 0.01, 0.0) END,
			weight_text  = CASE WHEN $2 = 'weight_text'  THEN LEAST(feed_preferences.weight_text  + 0.05, 1.0)
			                    ELSE GREATEST(feed_preferences.weight_text  - 0.01, 0.0) END,
			updated_at   = NOW()`,
		userID, col,
	)
	return err
}

// ── Персональная лента ────────────────────────────────────────────────────────
//
// Оптимизации:
//   1. Preferences читаем отдельным точечным SELECT до основного запроса
//   2. Все агрегаты через JOIN — ноль коррелированных подзапросов
//   3. first_att через DISTINCT ON — один проход по индексу
//   4. Вложения грузятся одним батч-запросом после основного (нет N+1)

func (r *FeedRepository) GetPersonalFeed(userID uuid.UUID, limit, offset int) ([]model.WallPost, error) {
	return r.GetPersonalFeedWithPrefs(userID, limit, offset, 0.33, 0.33, 0.34)
}

// GetPersonalFeedWithPrefs — версия с уже известными весами (из кеша сервиса).
// Позволяет сервису передать preferences без лишнего SELECT.
func (r *FeedRepository) GetPersonalFeedWithPrefs(userID uuid.UUID, limit, offset int, wi, wv, wt float64) ([]model.WallPost, error) {
	if limit <= 0 {
		limit = 30
	}

	// 2. Основной запрос — только JOIN, никаких коррелированных подзапросов
	const query = `
		WITH
		likes_agg AS (
			SELECT post_id,
			       COUNT(*)            AS likes_count,
			       BOOL_OR(user_id = $1) AS is_liked
			FROM wall_post_likes
			GROUP BY post_id
		),
		comments_agg AS (
			SELECT wp.id AS post_id, COUNT(m.id) AS comments_count
			FROM wall_posts wp
			LEFT JOIN messages m ON m.chat_id = wp.chat_id
			GROUP BY wp.id
		),
		first_att AS (
			SELECT DISTINCT ON (post_id)
			       post_id, mime_type,
			       (mime_type LIKE 'image/%' OR mime_type LIKE 'video/%') AS is_media
			FROM wall_attachments
			ORDER BY post_id, created_at
		),
		user_score AS (
			SELECT post_id, view_count, last_seen_at
			FROM feed_scores
			WHERE user_id = $1
		)
		SELECT
			p.id, p.user_id, p.content, p.created_at, p.updated_at, p.chat_id,
			u.username,
			COALESCE(u.avatar_url, '')          AS author_avatar,
			COALESCE(la.likes_count, 0)         AS likes_count,
			COALESCE(la.is_liked, false)        AS is_liked,
			COALESCE(ca.comments_count, 0)      AS comments_count,
			(
				(COALESCE(la.likes_count, 0) * 10.0 + COALESCE(ca.comments_count, 0) * 8.0)
				* CASE
					WHEN fa.mime_type LIKE 'image%' THEN $2 + 0.67
					WHEN fa.mime_type LIKE 'video%' THEN $3 + 0.67
					ELSE                                 $4 + 0.66
				  END
				* EXP(-EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 604800.0)
				* CASE WHEN COALESCE(us.view_count, 0) > 0 THEN 0.3 ELSE 1.0 END
			) AS final_score
		FROM wall_posts p
		JOIN users u ON u.id = p.user_id
		JOIN first_att fa ON fa.post_id = p.id AND fa.is_media = true
		LEFT JOIN likes_agg la ON la.post_id = p.id
		LEFT JOIN comments_agg ca ON ca.post_id = p.id
		LEFT JOIN user_score us ON us.post_id = p.id
		WHERE COALESCE(us.last_seen_at, '1970-01-01') < NOW() - INTERVAL '10 minutes'
		ORDER BY final_score DESC
		LIMIT $5 OFFSET $6`

	rows, err := r.db.Query(query, userID, wi, wv, wt, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.WallPost
	var postIDs []uuid.UUID

	for rows.Next() {
		var p model.WallPost
		var chatID sql.NullString
		var score float64
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Content, &p.CreatedAt, &p.UpdatedAt,
			&chatID, &p.AuthorName, &p.AuthorAvatar,
			&p.LikesCount, &p.IsLiked, &p.CommentsCount, &score,
		); err != nil {
			return nil, err
		}
		if chatID.Valid && chatID.String != "" {
			if uid, err := uuid.Parse(chatID.String); err == nil {
				p.ChatID = &uid
			}
		}
		posts = append(posts, p)
		postIDs = append(postIDs, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return []model.WallPost{}, nil
	}

	// 3. Вложения — батч-запрос, запускаем как только есть postIDs
	attMap, err := r.getAttachmentsBatch(postIDs)
	if err == nil {
		for i := range posts {
			if atts := attMap[posts[i].ID]; atts != nil {
				posts[i].Attachments = atts
			} else {
				posts[i].Attachments = []model.WallAttachment{}
			}
		}
	}

	return posts, nil
}

// GetPersonalFeedAndGlobal — запускает персональную и глобальную ленты параллельно.
// Используется в сервисе для cold start без двух последовательных запросов.
func (r *FeedRepository) GetPersonalFeedAndGlobal(
	userID uuid.UUID, limit, offset int, wi, wv, wt float64,
) (personal []model.WallPost, global []model.WallPost, err error) {
	var wg sync.WaitGroup
	var personalErr, globalErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		personal, personalErr = r.GetPersonalFeedWithPrefs(userID, limit, offset, wi, wv, wt)
	}()

	go func() {
		defer wg.Done()
		// Глобальная лента — только если offset=0 (нужна для cold start)
		if offset == 0 {
			global, globalErr = r.getGlobalFeedLite(userID, limit)
		}
	}()

	wg.Wait()

	if personalErr != nil {
		return nil, nil, personalErr
	}
	_ = globalErr // глобальная лента — fallback, ошибка не критична
	return personal, global, nil
}

// getGlobalFeedLite — облегчённая глобальная лента без тяжёлого скоринга.
// Используется только как fallback при cold start.
func (r *FeedRepository) getGlobalFeedLite(userID uuid.UUID, limit int) ([]model.WallPost, error) {
	const query = `
		SELECT p.id, p.user_id, p.content, p.created_at, p.updated_at, p.chat_id,
		       u.username, COALESCE(u.avatar_url, ''),
		       COUNT(DISTINCT l.user_id) AS likes_count,
		       COALESCE(BOOL_OR(l.user_id = $1), false) AS is_liked,
		       (SELECT COUNT(*) FROM messages m WHERE m.chat_id = p.chat_id) AS comments_count
		FROM wall_posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN wall_post_likes l ON l.post_id = p.id
		WHERE EXISTS (
			SELECT 1 FROM wall_attachments wa
			WHERE wa.post_id = p.id
			AND (wa.mime_type LIKE 'image/%' OR wa.mime_type LIKE 'video/%')
		)
		GROUP BY p.id, p.user_id, p.content, p.created_at, p.updated_at, p.chat_id, u.username, u.avatar_url
		ORDER BY p.created_at DESC
		LIMIT $2`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.WallPost
	var postIDs []uuid.UUID
	for rows.Next() {
		var p model.WallPost
		var chatID sql.NullString
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Content, &p.CreatedAt, &p.UpdatedAt,
			&chatID, &p.AuthorName, &p.AuthorAvatar,
			&p.LikesCount, &p.IsLiked, &p.CommentsCount,
		); err != nil {
			return nil, err
		}
		if chatID.Valid && chatID.String != "" {
			if uid, err := uuid.Parse(chatID.String); err == nil {
				p.ChatID = &uid
			}
		}
		posts = append(posts, p)
		postIDs = append(postIDs, p.ID)
	}
	if len(posts) == 0 {
		return []model.WallPost{}, nil
	}
	attMap, _ := r.getAttachmentsBatch(postIDs)
	for i := range posts {
		if atts := attMap[posts[i].ID]; atts != nil {
			posts[i].Attachments = atts
		} else {
			posts[i].Attachments = []model.WallAttachment{}
		}
	}
	return posts, nil
}

// getAttachmentsBatch — один SELECT за все вложения через ANY($1::uuid[])
func (r *FeedRepository) getAttachmentsBatch(postIDs []uuid.UUID) (map[uuid.UUID][]model.WallAttachment, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}

	// Собираем {uuid1,uuid2,...} для передачи как uuid[]
	sb := strings.Builder{}
	sb.WriteByte('{')
	for i, id := range postIDs {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(id.String())
	}
	sb.WriteByte('}')

	rows, err := r.db.Query(`
		SELECT id, post_id, url, filename, mime_type, size_bytes, created_at
		FROM wall_attachments
		WHERE post_id = ANY($1::uuid[])
		ORDER BY post_id, created_at`,
		sb.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]model.WallAttachment, len(postIDs))
	for rows.Next() {
		var a model.WallAttachment
		if err := rows.Scan(&a.ID, &a.PostID, &a.Url, &a.Filename, &a.MimeType, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		result[a.PostID] = append(result[a.PostID], a)
	}
	return result, rows.Err()
}

// DB возвращает *sql.DB для использования в сервисе (кеш preferences)
func (r *FeedRepository) DB() *sql.DB {
	return r.db
}

// TrackEventBatch — вставляет пачку событий одним запросом.
func (r *FeedRepository) TrackEventBatch(events []*model.FeedEvent) error {
	if len(events) == 0 {
		return nil
	}

	args := make([]interface{}, 0, len(events)*4)
	var sb strings.Builder
	sb.WriteString("INSERT INTO feed_events (user_id, post_id, event_type, watch_seconds) VALUES ")

	for i, e := range events {
		if i > 0 {
			sb.WriteByte(',')
		}
		base := i * 4
		fmt.Fprintf(&sb, "($%d,$%d,$%d,$%d)", base+1, base+2, base+3, base+4)
		args = append(args, e.UserID, e.PostID, e.EventType, e.WatchSeconds)
	}

	_, err := r.db.Exec(sb.String(), args...)
	return err
}
