package model

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID        uuid.UUID `json:"id"`
	Type      TypeChat  `json:"type"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatListItem struct {
	ID              uuid.UUID  `json:"id"`
	Type            TypeChat   `json:"type"`
	Name            string     `json:"name"` // Имя собеседника или группы
	LastMessage     string     `json:"last_message"`
	LastMessageTime time.Time  `json:"last_message_time"`
	IsOnline        bool       `json:"is_online"`
	InterlocutorID  *uuid.UUID `json:"-"` // ID собеседника для проверки онлайна
}
