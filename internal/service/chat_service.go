package service

import (
	"errors"
	"fmt"
	"strings"

	"messenger/internal/model"
	wsmodel "messenger/internal/service/websocket"

	"github.com/google/uuid"
)

type ChatRepositoryForService interface {
	CreatePrivateChat(initiatorID, targetUserID uuid.UUID) (*model.Chat, error)
	CreateGroupChat(name string, memberIDs []uuid.UUID) (*model.Chat, error)
	GetUserChats(userID uuid.UUID) ([]model.ChatListItem, error)
	UpdateGroupAvatarUrl(chatID uuid.UUID, url string) error
	GetChatByID(chatID uuid.UUID) (*model.Chat, error)
	UpdateGroupChat(chatID uuid.UUID, name string, avatarUrl string) error
	RemoveChatMember(chatID, userID uuid.UUID) error
	AddChatMember(chatID, userID uuid.UUID) error
	IsChatMember(chatID uuid.UUID, userID uuid.UUID) (bool, error)
	GetMembersInfo(chatID uuid.UUID) ([]model.ChatMemberInfo, error)
}

type UserRepositoryForService interface {
	GetById(id uuid.UUID) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
}

type HubForService interface {
	IsUserOnline(userID uuid.UUID) bool
	BroadcastToUsers(userIDs []uuid.UUID, message wsmodel.Message)
}

type ChatService struct {
	repo     ChatRepositoryForService
	userRepo UserRepositoryForService
	hub      HubForService
}

func NewChatService(repo ChatRepositoryForService, userRepo UserRepositoryForService, hub HubForService) *ChatService {
	return &ChatService{repo: repo, userRepo: userRepo, hub: hub}
}

func (s *ChatService) CreatePrivateChat(userId0 uuid.UUID, userId1 uuid.UUID) (*model.Chat, error) {
	for _, id := range []uuid.UUID{userId0, userId1} {
		user, err := s.userRepo.GetById(id)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, fmt.Errorf("пользователь с ID %s не найден", id)
		}
	}
	return s.repo.CreatePrivateChat(userId0, userId1)
}

func (s *ChatService) CreateGroupChatByUsernames(name string, usernames []string, creatorID uuid.UUID) (*model.Chat, error) {
	seenIDs := make(map[uuid.UUID]bool)
	seenIDs[creatorID] = true
	userIDs := []uuid.UUID{creatorID}
	resolvedNames := []string{}

	for _, username := range usernames {
		user, err := s.userRepo.GetByUsername(username)
		if err != nil || user == nil {
			return nil, fmt.Errorf("пользователь %s не найден", username)
		}
		if !seenIDs[user.ID] {
			seenIDs[user.ID] = true
			userIDs = append(userIDs, user.ID)
			resolvedNames = append(resolvedNames, user.Username)
		}
	}

	if name == "" {
		name = strings.Join(resolvedNames, ", ")
	}

	return s.repo.CreateGroupChat(name, userIDs)
}

func (s *ChatService) GetUserChats(userID uuid.UUID) ([]model.ChatListItem, error) {
	chats, err := s.repo.GetUserChats(userID)
	if err != nil {
		return nil, err
	}

	for i := range chats {
		if chats[i].InterlocutorID != nil {
			chats[i].IsOnline = s.hub.IsUserOnline(*chats[i].InterlocutorID)
		}
	}

	if len(chats) == 0 || chats == nil {
		return nil, errors.New("у пользователя пока нет чатов")
	}

	return chats, nil
}

func (s *ChatService) UpdateGroupAvatarUrl(chatID uuid.UUID, url string) error {
	return s.repo.UpdateGroupAvatarUrl(chatID, url)
}

// GetChatByID — возвращает чат с проверкой членства
func (s *ChatService) GetChatByID(chatID uuid.UUID, requestingUserID uuid.UUID) (*model.Chat, error) {
	isMember, err := s.repo.IsChatMember(chatID, requestingUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("access denied")
	}
	return s.repo.GetChatByID(chatID)
}

// UpdateGroupChat — изменить имя/аватар группы (только создатель)
func (s *ChatService) UpdateGroupChat(chatID uuid.UUID, requestingUserID uuid.UUID, name string, avatarUrl string) (*model.Chat, error) {
	chat, err := s.repo.GetChatByID(chatID)
	if err != nil || chat == nil {
		return nil, fmt.Errorf("chat not found")
	}
	if chat.Type != model.TypeGroup {
		return nil, fmt.Errorf("not a group chat")
	}
	if chat.CreatorID == nil || *chat.CreatorID != requestingUserID {
		return nil, fmt.Errorf("only the creator can edit this group")
	}
	if name == "" {
		name = chat.Name
	}
	if avatarUrl == "" {
		avatarUrl = chat.AvatarUrl
	}
	if err := s.repo.UpdateGroupChat(chatID, name, avatarUrl); err != nil {
		return nil, err
	}
	chat.Name = name
	chat.AvatarUrl = avatarUrl
	return chat, nil
}

// RemoveChatMember — удалить участника (только создатель)
func (s *ChatService) RemoveChatMember(chatID uuid.UUID, requestingUserID uuid.UUID, targetUserID uuid.UUID) error {
	chat, err := s.repo.GetChatByID(chatID)
	if err != nil || chat == nil {
		return fmt.Errorf("chat not found")
	}
	if chat.Type != model.TypeGroup {
		return fmt.Errorf("not a group chat")
	}
	if chat.CreatorID == nil || *chat.CreatorID != requestingUserID {
		return fmt.Errorf("only the creator can remove members")
	}
	if targetUserID == requestingUserID {
		return fmt.Errorf("creator cannot remove themselves")
	}

	// Получаем список участников ДО удаления (чтобы уведомить их)
	membersBefore, _ := s.repo.GetMembersInfo(chatID)

	if err := s.repo.RemoveChatMember(chatID, targetUserID); err != nil {
		return err
	}

	// Рассылаем WS-событие всем оставшимся участникам (и удалённому)
	allUserIDs := make([]uuid.UUID, 0, len(membersBefore))
	for _, m := range membersBefore {
		allUserIDs = append(allUserIDs, m.ID)
	}
	s.hub.BroadcastToUsers(allUserIDs, wsmodel.Message{
		Type: "member_removed",
		Content: map[string]interface{}{
			"chat_id": chatID,
			"user_id": targetUserID,
		},
	})
	return nil
}

// AddChatMember — добавить участника по username (только создатель)
func (s *ChatService) AddChatMember(chatID uuid.UUID, requestingUserID uuid.UUID, username string) error {
	chat, err := s.repo.GetChatByID(chatID)
	if err != nil || chat == nil {
		return fmt.Errorf("chat not found")
	}
	if chat.Type != model.TypeGroup {
		return fmt.Errorf("not a group chat")
	}
	if chat.CreatorID == nil || *chat.CreatorID != requestingUserID {
		return fmt.Errorf("only the creator can add members")
	}
	user, err := s.userRepo.GetByUsername(username)
	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}

	if err := s.repo.AddChatMember(chatID, user.ID); err != nil {
		return err
	}

	// Получаем актуальный список участников ПОСЛЕ добавления
	membersAfter, _ := s.repo.GetMembersInfo(chatID)

	// Рассылаем WS-событие всем участникам (включая нового)
	allUserIDs := make([]uuid.UUID, 0, len(membersAfter))
	for _, m := range membersAfter {
		allUserIDs = append(allUserIDs, m.ID)
	}
	s.hub.BroadcastToUsers(allUserIDs, wsmodel.Message{
		Type: "member_added",
		Content: map[string]interface{}{
			"chat_id": chatID,
			"user": map[string]interface{}{
				"id":         user.ID,
				"username":   user.Username,
				"full_name":  user.FullName,
				"avatar_url": user.AvatarUrl,
			},
		},
	})
	return nil
}

// GetGroupMembers — получить список участников (только для членов чата)
func (s *ChatService) GetGroupMembers(chatID uuid.UUID, requestingUserID uuid.UUID) ([]model.ChatMemberInfo, error) {
	isMember, err := s.repo.IsChatMember(chatID, requestingUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, fmt.Errorf("access denied")
	}
	members, err := s.repo.GetMembersInfo(chatID)
	if err != nil {
		return nil, err
	}
	// Проставляем онлайн-статус через hub
	for i := range members {
		members[i].IsOnline = s.hub.IsUserOnline(members[i].ID)
	}
	return members, nil
}
