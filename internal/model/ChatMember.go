package model

import "github.com/google/uuid"

type ChatMember struct {
	ChatID uuid.UUID
	UserID uuid.UUID
}
