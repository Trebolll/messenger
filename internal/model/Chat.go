package model

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID        uuid.UUID
	Type      TypeChat
	Name      string
	CreatedAt time.Time
}
