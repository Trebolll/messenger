package service

import (
	"errors"
	"messenger/internal/model"
	"messenger/internal/repository"
	"messenger/internal/service/websocket"

	"github.com/google/uuid"
)

type MessageService struct {
	repo     *repository.MessageRepository
	chatRepo *repository.ChatRepository
	hub      *websocket.Hub
}

func NewMessageService(repo *repository.MessageRepository, chatRepo *repository.ChatRepository, hub *websocket.Hub) *MessageService {
	return &MessageService{
		repo:     repo,
		chatRepo: chatRepo,
		hub:      hub,
	}
}

func (s *MessageService) SendMessage(message *model.Message) error {

	isMember, err := s.chatRepo.IsChatMember(message.ChatID, message.SenderID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("доступ запрещен: вы не являетесь участником этого чата")
	}

	err = s.repo.SendMessage(message)
	if err != nil {
		return err
	}

	// Уведомляем участников чата
	members, err := s.chatRepo.GetChatMembers(message.ChatID)
	if err != nil {
		return err
	}
	for _, userID := range members {
		s.hub.SendToUser(userID, websocket.Message{
			Type:    "new_message",
			Content: message,
		})
	}
	return nil
}

func (s *MessageService) GetMessagesByChatID(chatID uuid.UUID) ([]model.Message, error) {
	exists, err := s.chatRepo.Exists(chatID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("чат не существует")
	}

	return s.repo.GetMessagesByChatID(chatID)
}
