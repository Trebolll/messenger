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

func (r *ChatRepository) GetUserChats(userID uuid.UUID) ([]model.ChatListItem, error) {
	query := `
		SELECT 
			c.id, 
			c.type, 
			COALESCE(c.name, u.username) as name,
			COALESCE(m.content, '') as last_message,
			COALESCE(m.created_at, c.created_at) as last_message_time,
			u.id as interlocutor_id,
			COALESCE(u.status, '') as user_status,
			COALESCE(CASE WHEN c.type = 'group' THEN c.avatar_url ELSE u.avatar_url END, '') as avatar_url,
			c.creator_id
		FROM chats c
		JOIN chat_members cm ON c.id = cm.chat_id
		LEFT JOIN users u ON c.type = 'private' AND EXISTS (
			SELECT 1 FROM chat_members cm2 
			WHERE cm2.chat_id = c.id AND cm2.user_id != $1
		) AND u.id = (
			SELECT user_id FROM chat_members cm2 
			WHERE cm2.chat_id = c.id AND cm2.user_id != $1 
			LIMIT 1
		)
		LEFT JOIN LATERAL (
			SELECT content, created_at 
			FROM messages 
			WHERE chat_id = c.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) m ON true
		WHERE cm.user_id = $1
		ORDER BY last_message_time DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []model.ChatListItem
	for rows.Next() {
		var chat model.ChatListItem
		var userStatus sql.NullString
		var creatorID sql.NullString
		if err := rows.Scan(&chat.ID, &chat.Type, &chat.Name, &chat.LastMessage, &chat.LastMessageTime, &chat.InterlocutorID, &userStatus, &chat.AvatarUrl, &creatorID); err != nil {
			return nil, err
		}
		chat.UserStatus = userStatus.String
		if chat.Type == model.TypeGroup {
			chat.IsGroup = true
		}
		if creatorID.String != "" {
			uid, err := uuid.Parse(creatorID.String)
			if err == nil {
				chat.CreatorID = &uid
			}
		}
		chats = append(chats, chat)
	}

	// Для групп подгружаем участников
	for i := range chats {
		if chats[i].IsGroup {
			members, err := r.GetMembersInfo(chats[i].ID)
			if err == nil {
				chats[i].Members = members
			}
		}
	}

	return chats, nil
}

// GetMembersInfo возвращает краткую информацию об участниках чата
func (r *ChatRepository) GetMembersInfo(chatID uuid.UUID) ([]model.ChatMemberInfo, error) {
	query := `
		SELECT u.id, u.username, COALESCE(u.full_name, ''), COALESCE(u.avatar_url, '')
		FROM chat_members cm
		JOIN users u ON u.id = cm.user_id
		WHERE cm.chat_id = $1
		ORDER BY cm.joined_at ASC`

	rows, err := r.db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.ChatMemberInfo
	for rows.Next() {
		var m model.ChatMemberInfo
		if err := rows.Scan(&m.ID, &m.Username, &m.FullName, &m.AvatarUrl); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

// UpdateGroupAvatarUrl обновляет аватар группового чата
func (r *ChatRepository) UpdateGroupAvatarUrl(chatID uuid.UUID, url string) error {
	_, err := r.db.Exec(`UPDATE chats SET avatar_url = $1 WHERE id = $2`, url, chatID)
	return err
}

// GetChatByID возвращает чат по ID
func (r *ChatRepository) GetChatByID(chatID uuid.UUID) (*model.Chat, error) {
	var chat model.Chat
	var creatorID sql.NullString
	var avatarUrl sql.NullString
	query := `SELECT id, type, COALESCE(name,''), COALESCE(creator_id::text, ''), COALESCE(avatar_url, ''), created_at FROM chats WHERE id = $1`
	err := r.db.QueryRow(query, chatID).Scan(&chat.ID, &chat.Type, &chat.Name, &creatorID, &avatarUrl, &chat.CreatedAt)
	if err != nil {
		return nil, err
	}
	if creatorID.String != "" {
		uid, err := uuid.Parse(creatorID.String)
		if err == nil {
			chat.CreatorID = &uid
		}
	}
	chat.AvatarUrl = avatarUrl.String
	return &chat, nil
}

// UpdateGroupChat обновляет имя и аватар группы
func (r *ChatRepository) UpdateGroupChat(chatID uuid.UUID, name string, avatarUrl string) error {
	_, err := r.db.Exec(`UPDATE chats SET name = $1, avatar_url = NULLIF($2, '') WHERE id = $3 AND type = 'group'`, name, avatarUrl, chatID)
	return err
}

// RemoveChatMember удаляет участника из группы
func (r *ChatRepository) RemoveChatMember(chatID, userID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2`, chatID, userID)
	return err
}

// AddChatMember добавляет участника в группу
func (r *ChatRepository) AddChatMember(chatID, userID uuid.UUID) error {
	_, err := r.db.Exec(`INSERT INTO chat_members(chat_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, chatID, userID)
	return err
}

func (r *ChatRepository) CreatePrivateChat(userId0 uuid.UUID, userId1 uuid.UUID) (*model.Chat, error) {
	if userId0 == userId1 {
		return nil, errors.New("не могу создать чат с самим собой")
	}

	existingChat, err := r.ExistPrivateChatByUsers(userId0, userId1, model.TypePrivate)
	if err != nil {
		return nil, err
	}
	if existingChat != nil {
		return existingChat, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var chat model.Chat
	chat.Type = model.TypePrivate

	query := `insert into chats(type) values ($1) returning id, created_at`
	err = tx.QueryRow(query, chat.Type).Scan(&chat.ID, &chat.CreatedAt)
	if err != nil {
		return nil, err
	}

	memberQuery := `insert into chat_members(chat_id, user_id) values ($1, $2)`
	members := []uuid.UUID{userId0, userId1}
	for _, uID := range members {
		if _, err = tx.Exec(memberQuery, chat.ID, uID); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *ChatRepository) CreateGroupChat(name string, userIDs []uuid.UUID) (*model.Chat, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var chat model.Chat
	chat.Type = model.TypeGroup
	chat.Name = name

	// Первый в списке — создатель
	var creatorID *uuid.UUID
	if len(userIDs) > 0 {
		id := userIDs[0]
		creatorID = &id
	}

	query := `INSERT INTO chats(type, name, creator_id) VALUES ($1, $2, $3) RETURNING id, created_at`
	err = tx.QueryRow(query, chat.Type, chat.Name, creatorID).Scan(&chat.ID, &chat.CreatedAt)
	if err != nil {
		return nil, err
	}
	chat.CreatorID = creatorID

	memberQuery := `INSERT INTO chat_members(chat_id, user_id) VALUES ($1, $2)`
	for _, uID := range userIDs {
		if _, err = tx.Exec(memberQuery, chat.ID, uID); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *ChatRepository) ExistPrivateChatByUsers(userId0 uuid.UUID, userId1 uuid.UUID, chatType model.TypeChat) (*model.Chat, error) {
	query := `
	select c.id, c.type, c.created_at
    from chats c
    join chat_members cm1 ON c.id = cm1.chat_id
    join chat_members cm2 ON c.id = cm2.chat_id
    where c.type = $3 
    and cm1.user_id = $1 
    and cm2.user_id = $2`

	var chat model.Chat
	err := r.db.QueryRow(query, userId0, userId1, chatType).Scan(&chat.ID, &chat.Type, &chat.CreatedAt)
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
	err := r.db.QueryRow(query, chatID, userID).Scan(&exists)
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
