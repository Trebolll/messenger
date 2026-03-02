package repository

import (
	"database/sql"
	"errors"
	"log"
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
			RETURNING id, chat_id, sender_id, content, created_at, read_at, edited_at, likes, dislikes
		)
		SELECT m.id, m.chat_id, m.sender_id, u.username, COALESCE(u.avatar_url, ''), u.rating, m.content, m.created_at, m.read_at, m.edited_at, m.likes, m.dislikes
		FROM inserted_msg m
		JOIN users u ON m.sender_id = u.id`

	err := r.db.QueryRow(query, message.ChatID,
		message.SenderID,
		message.Content).Scan(
		&message.ID, &message.ChatID, &message.SenderID,
		&message.SenderName, &message.SenderAvatarURL, &message.SenderRating,
		&message.Content, &message.CreatedAt, &message.ReadAt, &message.EditedAt,
		&message.Likes, &message.Dislikes,
	)
	if err != nil {
		log.Printf("SQL Error in SendMessage: %v\nQuery: %s\nArgs: %v, %v, %v", err, query, message.ChatID, message.SenderID, message.Content)
	}
	return err
}

func (r *MessageRepository) GetMessagesByChatID(chatID uuid.UUID) ([]model.Message, error) {
	query := `
		SELECT m.id, m.chat_id, m.sender_id, u.username, COALESCE(u.avatar_url, ''), u.rating, m.content, m.created_at, m.read_at, m.edited_at, m.likes, m.dislikes
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.chat_id = $1
		ORDER BY m.created_at ASC`

	rows, err := r.db.Query(query, chatID)
	if err != nil {
		log.Printf("SQL Error: %v\nQuery: %s\nArgs: %v", err, query, chatID)
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		if err := rows.Scan(
			&m.ID, &m.ChatID, &m.SenderID,
			&m.SenderName, &m.SenderAvatarURL, &m.SenderRating,
			&m.Content, &m.CreatedAt, &m.ReadAt, &m.EditedAt,
			&m.Likes, &m.Dislikes,
		); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessageRepository) EditMessage(messageID, senderID uuid.UUID, content string) (*model.Message, error) {
	query := `
		WITH updated AS (
			UPDATE messages SET content=$1, edited_at=CURRENT_TIMESTAMP
			WHERE id=$2 AND sender_id=$3
			RETURNING id, chat_id, sender_id, content, created_at, read_at, edited_at, likes, dislikes
		)
		SELECT u.id, u.chat_id, u.sender_id, usr.username, COALESCE(usr.avatar_url, ''), usr.rating, u.content, u.created_at, u.read_at, u.edited_at, u.likes, u.dislikes
		FROM updated u
		JOIN users usr ON u.sender_id = usr.id`

	var m model.Message
	err := r.db.QueryRow(query, content, messageID, senderID).Scan(
		&m.ID, &m.ChatID, &m.SenderID,
		&m.SenderName, &m.SenderAvatarURL, &m.SenderRating,
		&m.Content, &m.CreatedAt, &m.ReadAt, &m.EditedAt,
		&m.Likes, &m.Dislikes,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("сообщение не найдено или нет прав на редактирование")
	}
	log.Printf("SQL in EditMessage: %v\nQuery: %s\nArgs: %v, %v, %v", err, query, content, messageID, senderID)
	return &m, err
}

func (r *MessageRepository) MarkAsRead(chatID, userID uuid.UUID) error {
	_, err := r.db.Exec(`
		UPDATE messages SET read_at=CURRENT_TIMESTAMP
		WHERE chat_id=$1 AND sender_id!=$2 AND read_at IS NULL`, chatID, userID)
	return err
}

func (r *MessageRepository) DeleteMessage(messageID, senderID uuid.UUID) (uuid.UUID, error) {
	var chatID uuid.UUID
	err := r.db.QueryRow(`DELETE FROM messages WHERE id=$1 AND sender_id=$2 RETURNING chat_id`, messageID, senderID).Scan(&chatID)
	if err == sql.ErrNoRows {
		return uuid.Nil, errors.New("сообщение не найдено или вы не являетесь его автором")
	}
	return chatID, err
}
