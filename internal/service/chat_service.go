package service

import (
	"database/sql"
	"errors"
	"fmt"

	"messenger/internal/model"
	"messenger/internal/repository"

	"github.com/google/uuid"
)

type ChatService struct {
	repo     *repository.ChatRepository
	userRepo *repository.UserRepository
}

func NewChatService(repo *repository.ChatRepository, userRepo *repository.UserRepository) *ChatService {
	return &ChatService{repo: repo, userRepo: userRepo}
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

func (s *ChatService) GetUserChats(userID uuid.UUID) ([]model.Chat, error) {
	chats, err := s.repo.GetUserChats(userID)
	if err != nil {
		return nil, err
	}

	if len(chats) == 0 || chats == nil {
		return nil, errors.New("у пользователя пока нет чатов")
	}

	return chats, nil
}
