package service

import (
	"errors"
	"messenger/internal/model"
	"messenger/internal/service/websocket"

	"github.com/google/uuid"
)

type MessageRepository interface {
	SendMessage(message *model.Message) error
	GetMessagesByChatID(chatID uuid.UUID) ([]model.Message, error)
	MarkAsRead(chatID, userID uuid.UUID) error
	EditMessage(messageID, senderID uuid.UUID, content string) (*model.Message, error)
	DeleteMessage(messageID, senderID uuid.UUID) (uuid.UUID, error)
}

type ChatRepository interface {
	IsChatMember(chatID, userID uuid.UUID) (bool, error)
	GetChatMembers(chatID uuid.UUID) ([]uuid.UUID, error)
	Exists(chatID uuid.UUID) (bool, error)
}

type VoteEnricher interface {
	GetVotesForMessagesFixed(msgs []model.Message, voterID uuid.UUID) (map[uuid.UUID]int, error)
}

type Hub interface {
	SendToUser(userID uuid.UUID, msg websocket.Message)
}

type MessageService struct {
	repo         MessageRepository
	chatRepo     ChatRepository
	hub          Hub
	voteEnricher VoteEnricher
}

func NewMessageService(repo MessageRepository, chatRepo ChatRepository, hub Hub) *MessageService {
	return &MessageService{repo: repo, chatRepo: chatRepo, hub: hub}
}

func (s *MessageService) SetVoteEnricher(ve VoteEnricher) {
	s.voteEnricher = ve
}

func (s *MessageService) SendMessage(message *model.Message) error {
	isMember, err := s.chatRepo.IsChatMember(message.ChatID, message.SenderID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("доступ запрещен: вы не являетесь участником этого чата")
	}

	if err = s.repo.SendMessage(message); err != nil {
		return err
	}

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

func (s *MessageService) GetMessagesByChatID(chatID, viewerID uuid.UUID) ([]model.Message, error) {
	exists, err := s.chatRepo.Exists(chatID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("чат не существует")
	}
	msgs, err := s.repo.GetMessagesByChatID(chatID)
	if err != nil {
		return nil, err
	}
	// Проставляем my_vote для текущего пользователя
	if s.voteEnricher != nil && len(msgs) > 0 {
		votes, err2 := s.voteEnricher.GetVotesForMessagesFixed(msgs, viewerID)
		if err2 == nil {
			for i := range msgs {
				if v, ok := votes[msgs[i].ID]; ok {
					msgs[i].MyVote = v
				}
			}
		}
	}
	return msgs, nil
}

func (s *MessageService) EditMessage(messageID, senderID uuid.UUID, content string) (*model.Message, error) {
	if content == "" {
		return nil, errors.New("содержимое сообщения не может быть пустым")
	}

	msg, err := s.repo.EditMessage(messageID, senderID, content)
	if err != nil {
		return nil, err
	}

	// Уведомляем всех участников чата об изменении через WebSocket
	members, err := s.chatRepo.GetChatMembers(msg.ChatID)
	if err != nil {
		return nil, err
	}
	for _, userID := range members {
		s.hub.SendToUser(userID, websocket.Message{
			Type:    "message_edited",
			Content: msg,
		})
	}
	return msg, nil
}

func (s *MessageService) MarkChatAsRead(chatID, userID uuid.UUID) error {
	if err := s.repo.MarkAsRead(chatID, userID); err != nil {
		return err
	}

	members, _ := s.chatRepo.GetChatMembers(chatID)
	for _, memberID := range members {
		if memberID != userID {
			s.hub.SendToUser(memberID, websocket.Message{
				Type: "messages_read",
				Content: map[string]interface{}{
					"chat_id":   chatID,
					"reader_id": userID,
				},
			})
		}
	}
	return nil
}

func (s *MessageService) DeleteMessage(messageID, senderID uuid.UUID) error {
	chatID, err := s.repo.DeleteMessage(messageID, senderID)
	if err != nil {
		return err
	}

	// Уведомляем всех участников чата об удалении через WebSocket
	members, err := s.chatRepo.GetChatMembers(chatID)
	if err != nil {
		return err
	}

	for _, userID := range members {
		s.hub.SendToUser(userID, websocket.Message{
			Type: "message_deleted",
			Content: map[string]interface{}{
				"message_id": messageID,
				"chat_id":    chatID,
			},
		})
	}

	return nil
}
