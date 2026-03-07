package model

import (
	"time"

	"github.com/google/uuid"
)

// FeedEvent — одно событие взаимодействия пользователя с постом.
// Собирается на каждый просмотр, лайк, комментарий, досмотр видео.
type FeedEvent struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	PostID    uuid.UUID `json:"post_id"`
	EventType string    `json:"event_type"` // view | like | comment | video_complete | skip
	// Сколько секунд пользователь смотрел пост (для видео/фото)
	WatchSeconds int       `json:"watch_seconds,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// FeedScore — агрегированный скор поста для конкретного пользователя.
// Пересчитывается при каждом событии и используется при построении ленты.
type FeedScore struct {
	PostID        uuid.UUID `json:"post_id"`
	UserID        uuid.UUID `json:"user_id"`
	Score         float64   `json:"score"`
	ViewCount     int       `json:"view_count"`
	LikeCount     int       `json:"like_count"`
	CommentCount  int       `json:"comment_count"`
	TotalWatchSec int       `json:"total_watch_sec"`
	LastSeenAt    time.Time `json:"last_seen_at"`
}

// FeedRequest — параметры запроса ленты от клиента.
type FeedRequest struct {
	Limit  int `form:"limit"`
	Offset int `form:"offset"`
}

// TrackEventRequest — тело запроса на запись события.
type TrackEventRequest struct {
	PostID       uuid.UUID `json:"post_id"       binding:"required"`
	EventType    string    `json:"event_type"    binding:"required"`
	WatchSeconds int       `json:"watch_seconds"`
}
