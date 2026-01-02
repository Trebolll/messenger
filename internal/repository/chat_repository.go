package repository

import (
	"database/sql"
	"errors"
	"messenger/internal/model"

	"github.com/google/uuid"
)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) GetUserChats(userID uuid.UUID) ([]model.Chat, error) {

	query := `select c.id, c.type, c.created_at
	from chats c
	left join chat_members cm on c.id = cm.chat_id
	where cm.user_id = $1
	order by c.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []model.Chat
	for rows.Next() {
		var chat model.Chat
		if err := rows.Scan(&chat.ID, &chat.Type, &chat.CreatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	return chats, nil

}

func (r *ChatRepository) CreatePrivateChat(
	userId0 uuid.UUID,
	userId1 uuid.UUID) (*model.Chat, error) {
	if userId0 == userId1 {
		return nil, errors.New("не могу создать чат с самим собой")
	}

	// 1. Проверку существующего чата можно оставить вне транзакции
	existingChat, err := r.ExistPrivateChatByUsers(userId0, userId1, model.TypePrivate)
	if err != nil {
		return nil, err
	}
	if existingChat != nil {
		return existingChat, nil
	}

	// 2. Начало транзакции
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	var chat model.Chat
	chat.Type = model.TypePrivate

	// 3. создаем чат
	query := `insert into chats(type) values ($1) returning id, created_at`
	err = tx.QueryRow(query, chat.Type).Scan(&chat.ID, &chat.CreatedAt)
	if err != nil {
		return nil, err
	}

	// 4. вставка участников в чат
	memberQuery := `insert into chat_members(chat_id, user_id) values ($1, $2)`
	members := []uuid.UUID{userId0, userId1}

	for _, uID := range members {
		if _, err = tx.Exec(memberQuery, chat.ID, uID); err != nil {
			return nil, err
		}
	}

	// 5. Фиксация изменений
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *ChatRepository) ExistPrivateChatByUsers(
	userId0 uuid.UUID,
	userId1 uuid.UUID,
	chatType model.TypeChat) (*model.Chat, error) {
	query := `
	select c.id, c.type, c.created_at
    from chats c
    join chat_members cm1 ON c.id = cm1.chat_id
    join chat_members cm2 ON c.id = cm2.chat_id
    where c.type = $3 
    and cm1.user_id = $1 
    and cm2.user_id = $2`

	var chat model.Chat

	err := r.db.QueryRow(query, userId0, userId1, chatType).Scan(
		&chat.ID,
		&chat.Type,
		&chat.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &chat, err

}

func (r *ChatRepository) IsChatMember(chatID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `select exists(select 1 from chat_members where chat_id=$1 AND user_id=$2)`
	row := r.db.QueryRow(query, chatID, userID)
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *ChatRepository) Exists(chatID uuid.UUID) (bool, error) {
	var exists bool
	query := `select exists(select 1 from chats where id = $1)`
	err := r.db.QueryRow(query, chatID).Scan(&exists)
	return exists, err
}

func (r *ChatRepository) GetChatMembers(chatID uuid.UUID) ([]uuid.UUID, error) {
	query := `select user_id from chat_members where chat_id = $1`
	rows, err := r.db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}
