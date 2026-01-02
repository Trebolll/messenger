package repository

import (
	"database/sql"
	"messenger/internal/model"

	"github.com/google/uuid"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) SendMessage(message *model.Message) error {
	query := `insert into messages(chat_id, sender_id, content) values ($1, $2, $3) returning id, created_at`
	return r.db.QueryRow(query, message.ChatID, message.SenderID, message.Content).Scan(&message.ID, &message.CreatedAt)
}

func (r *MessageRepository) GetMessagesByChatID(chatID uuid.UUID) ([]model.Message, error) {
	query := `select id, chat_id, sender_id, content, created_at from messages where chat_id = $1 order by created_at asc`
	rows, err := r.db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}
