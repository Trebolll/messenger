package model

import (
	"time"

	"github.com/google/uuid"
)

type Wall struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	Bio            string     `json:"bio"`
	BannerUrl      string     `json:"banner_url"`
	Username       string     `json:"username,omitempty"`
	AvatarUrl      string     `json:"avatar_url,omitempty"`
	Status         string     `json:"status,omitempty"`
	UserRating     int        `json:"user_rating,omitempty"`
	UserLocation   string     `json:"user_location,omitempty"`
	UserProfession string     `json:"user_profession,omitempty"`
	UserBirthDate  *time.Time `json:"user_birth_date,omitempty"`
	UserCreatedAt  time.Time  `json:"user_created_at,omitempty"`
}

type WallResponse struct {
	Wall  Wall       `json:"wall"`
	Posts []WallPost `json:"posts"`
}

type WallPost struct {
	ID           uuid.UUID        `json:"id"`
	UserID       uuid.UUID        `json:"user_id"`
	Content      string           `json:"content"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    *time.Time       `json:"updated_at,omitempty"`
	AuthorName   string           `json:"author_name"`
	AuthorAvatar string           `json:"author_avatar"`
	Attachments  []WallAttachment `json:"attachments"`
}

type WallAttachment struct {
	ID        uuid.UUID `json:"id"`
	PostID    uuid.UUID `json:"post_id"`
	Url       string    `json:"url"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}
