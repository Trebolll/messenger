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

// ChatMemberInfo — краткая информация об участнике для фронтенда
type ChatMemberInfo struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarUrl string    `json:"avatar_url"`
}

type ChatListItem struct {
	ID              uuid.UUID        `json:"id"`
	Type            TypeChat         `json:"type"`
	Name            string           `json:"name"` // Имя собеседника или группы
	LastMessage     string           `json:"last_message"`
	LastMessageTime time.Time        `json:"last_message_time"`
	IsOnline        bool             `json:"is_online"`
	UserStatus      string           `json:"user_status"`       // Текстовый статус собеседника
	InterlocutorID  *uuid.UUID       `json:"interlocutor_id"`   // ID собеседника для проверки онлайна
	AvatarUrl       string           `json:"avatar_url"`        // Аватар собеседника или группы
	IsGroup         bool             `json:"is_group"`          // Флаг группового чата
	Members         []ChatMemberInfo `json:"members,omitempty"` // Участники (только для групп)
}
