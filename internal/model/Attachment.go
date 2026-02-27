package model

import (
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID        uuid.UUID  `json:"id"`
	ChatID    uuid.UUID  `json:"chat_id"`
	SenderID  uuid.UUID  `json:"sender_id"`
	MessageID *uuid.UUID `json:"message_id"`
	Url       string     `json:"url"`
	Filename  string     `json:"filename"`
	MimeType  string     `json:"mime_type"`
	SizeBytes int64      `json:"size_bytes"`
	CreatedAt time.Time  `json:"created_at"`
}
