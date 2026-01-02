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
	query := `
		WITH inserted_msg AS (
			INSERT INTO messages(chat_id, sender_id, content) 
			VALUES ($1, $2, $3) 
			RETURNING id, chat_id, sender_id, content, created_at
		)
		SELECT m.id, m.chat_id, m.sender_id, u.username, m.content, m.created_at
		FROM inserted_msg m
		JOIN users u ON m.sender_id = u.id`

	return r.db.QueryRow(query, message.ChatID, message.SenderID, message.Content).Scan(
		&message.ID,
		&message.ChatID,
		&message.SenderID,
		&message.SenderName,
		&message.Content,
		&message.CreatedAt,
	)
}

func (r *MessageRepository) GetMessagesByChatID(chatID uuid.UUID) ([]model.Message, error) {
	query := `
		SELECT m.id, m.chat_id, m.sender_id, u.username, m.content, m.created_at 
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.chat_id = $1 
		ORDER BY m.created_at ASC`
	rows, err := r.db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.SenderName, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessageRepository) MarkAsRead(chatID, userID uuid.UUID) error {
	// Помечаем прочитанными все сообщения в чате, где отправитель НЕ текущий пользователь
	query := `
        UPDATE messages 
        SET read_at = CURRENT_TIMESTAMP 
        WHERE chat_id = $1 AND sender_id != $2 AND read_at IS NULL`

	_, err := r.db.Exec(query, chatID, userID)
	return err
}
