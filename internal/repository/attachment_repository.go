package repository

import (
	"database/sql"
	"messenger/internal/model"
)

type AttachmentRepository struct {
	db *sql.DB
}

func NewAttachmentRepository(db *sql.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(a *model.Attachment) error {
	query := `INSERT INTO attachments (chat_id, sender_id, message_id, url, filename, mime_type, size_bytes)
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id,created_at`

	err := r.db.QueryRow(
		query,
		a.ChatID,
		a.SenderID,
		a.MessageID,
		a.Url,
		a.Filename,
		a.MimeType,
		a.SizeBytes,
	).Scan(&a.ID, &a.CreatedAt)

	if err != nil {
		return err
	}
	return nil
}
