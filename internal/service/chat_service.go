package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"messenger/internal/model"
	_ "messenger/internal/service/websocket"

	"github.com/google/uuid"
)

type ChatRepositoryForService interface {
	CreatePrivateChat(initiatorID, targetUserID uuid.UUID) (*model.Chat, error)
	CreateGroupChat(name string, memberIDs []uuid.UUID) (*model.Chat, error)
	GetUserChats(userID uuid.UUID) ([]model.ChatListItem, error)
	UpdateGroupAvatarUrl(chatID uuid.UUID, url string) error
}

type UserRepositoryForService interface {
	GetById(id uuid.UUID) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
}

type HubForService interface {
	IsUserOnline(userID uuid.UUID) bool
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
		_, err := s.userRepo.GetById(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("пользователь с ID %s не найден", id)
			}
			return nil, err
		}
	}
	return s.repo.CreatePrivateChat(userId0, userId1)
}

func (s *ChatService) CreateGroupChatByUsernames(name string, usernames []string, creatorID uuid.UUID) (*model.Chat, error) {
	// Используем map для хранения уникальных ID пользователей
	seenIDs := make(map[uuid.UUID]bool)
	seenIDs[creatorID] = true

	userIDs := []uuid.UUID{creatorID}
	resolvedNames := []string{}

	for _, username := range usernames {
		user, err := s.userRepo.GetByUsername(username)
		if err != nil {
			return nil, fmt.Errorf("пользователь %s не найден", username)
		}

		// Добавляем только если этого пользователя еще нет в списке (включая создателя)
		if !seenIDs[user.ID] {
			seenIDs[user.ID] = true
			userIDs = append(userIDs, user.ID)
			resolvedNames = append(resolvedNames, user.Username)
		}
	}

	// Если имя не задано — используем никнеймы участников через запятую
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
