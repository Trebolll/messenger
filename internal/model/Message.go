package model

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID              uuid.UUID  `json:"id"`
	ChatID          uuid.UUID  `json:"chat_id"`
	SenderID        uuid.UUID  `json:"sender_id"`
	SenderName      string     `json:"sender_name"`
	SenderAvatarURL string     `json:"sender_avatar_url"`
	SenderRating    int        `json:"sender_rating"`
	Content         string     `json:"content"`
	CreatedAt       time.Time  `json:"created_at"`
	ReadAt          *time.Time `json:"read_at"`
	EditedAt        *time.Time `json:"edited_at"`
	Likes           int        `json:"likes"`
	Dislikes        int        `json:"dislikes"`
	MyVote          int        `json:"my_vote"` // 0 = нет, 1 = лайк, -1 = дизлайк
}
