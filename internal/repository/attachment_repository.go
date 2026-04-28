package repository

import (
	"database/sql"
	"messenger/internal/model"

	"github.com/google/uuid"
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

func (r *AttachmentRepository) GetByID(id uuid.UUID) (*model.Attachment, error) {
	query := `SELECT id, chat_id, sender_id, message_id, url, filename, mime_type, size_bytes, created_at FROM attachments WHERE id = $1`
	a := &model.Attachment{}
	err := r.db.QueryRow(query, id).Scan(
		&a.ID, &a.ChatID, &a.SenderID, &a.MessageID, &a.Url, &a.Filename, &a.MimeType, &a.SizeBytes, &a.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AttachmentRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM attachments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
