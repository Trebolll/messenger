package model

import (
	"time"

	"github.com/google/uuid"
)

type MessageVote struct {
	MessageID uuid.UUID `json:"message_id"`
	VoterID   uuid.UUID `json:"voter_id"`
	Vote      int       `json:"vote"` // 1 = лайк, -1 = дизлайк
	CreatedAt time.Time `json:"created_at"`
}

// RatingVoteResult — результат операции голосования
type RatingVoteResult struct {
	MessageID    uuid.UUID
	ChatID       uuid.UUID
	SenderID     uuid.UUID
	Likes        int
	Dislikes     int
	MyVote       int
	SenderRating int
}

// Ранги по порогам публичного рейтинга (max(0, rating))
type RatingRank struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Min   int    `json:"min"`
}

var RatingRanks = []RatingRank{
	{Name: "Легенда", Color: "#f59e0b", Min: 1000},
	{Name: "Авторитет", Color: "#8b5cf6", Min: 500},
	{Name: "Активный", Color: "#22c55e", Min: 200},
	{Name: "Участник", Color: "#3b82f6", Min: 50},
	{Name: "Новичок", Color: "#9ca3af", Min: 0},
}

func GetRank(rating int) RatingRank {
	public := rating
	if public < 0 {
		public = 0
	}
	for _, r := range RatingRanks {
		if public >= r.Min {
			return r
		}
	}
	return RatingRanks[len(RatingRanks)-1]
}
