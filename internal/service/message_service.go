package service

import (
	"errors"
	"messenger/internal/model"
	"messenger/internal/service/websocket"
	"sync"

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

	// Сохранение и получение участников — параллельно
	var members []uuid.UUID
	var sendErr, membersErr error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		sendErr = s.repo.SendMessage(message)
	}()
	go func() {
		defer wg.Done()
		members, membersErr = s.chatRepo.GetChatMembers(message.ChatID)
	}()
	wg.Wait()

	if sendErr != nil {
		return sendErr
	}
	if membersErr != nil {
		return membersErr
	}

	s.broadcastToMembers(members, websocket.Message{
		Type:    "new_message",
		Content: message,
	})
	return nil
}

func (s *MessageService) GetMessagesByChatID(chatID, viewerID uuid.UUID) ([]model.Message, error) {
	// Проверка существования и загрузка сообщений — параллельно
	var msgs []model.Message
	var exists bool
	var msgsErr, existsErr error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		exists, existsErr = s.chatRepo.Exists(chatID)
	}()
	go func() {
		defer wg.Done()
		msgs, msgsErr = s.repo.GetMessagesByChatID(chatID)
	}()
	wg.Wait()

	if existsErr != nil {
		return nil, existsErr
	}
	if !exists {
		return nil, errors.New("чат не существует")
	}
	if msgsErr != nil {
		return nil, msgsErr
	}

	// Голоса — параллельно с ничем (уже есть msgs)
	if s.voteEnricher != nil && len(msgs) > 0 {
		votes, err := s.voteEnricher.GetVotesForMessagesFixed(msgs, viewerID)
		if err == nil {
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

	// Получаем участников и рассылаем WS асинхронно — не блокируем ответ
	go func() {
		members, err := s.chatRepo.GetChatMembers(msg.ChatID)
		if err != nil {
			return
		}
		s.broadcastToMembers(members, websocket.Message{
			Type:    "message_edited",
			Content: msg,
		})
	}()

	return msg, nil
}

func (s *MessageService) MarkChatAsRead(chatID, userID uuid.UUID) error {
	if err := s.repo.MarkAsRead(chatID, userID); err != nil {
		return err
	}

	// Рассылка — асинхронно, не блокируем ответ клиенту
	go func() {
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
	}()

	return nil
}

func (s *MessageService) DeleteMessage(messageID, senderID uuid.UUID) error {
	chatID, err := s.repo.DeleteMessage(messageID, senderID)
	if err != nil {
		return err
	}

	// Рассылка — асинхронно
	go func() {
		members, err := s.chatRepo.GetChatMembers(chatID)
		if err != nil {
			return
		}
		s.broadcastToMembers(members, websocket.Message{
			Type: "message_deleted",
			Content: map[string]interface{}{
				"message_id": messageID,
				"chat_id":    chatID,
			},
		})
	}()

	return nil
}

// broadcastToMembers — рассылает WS-сообщение всем участникам параллельно
func (s *MessageService) broadcastToMembers(members []uuid.UUID, msg websocket.Message) {
	var wg sync.WaitGroup
	for _, userID := range members {
		wg.Add(1)
		go func(uid uuid.UUID) {
			defer wg.Done()
			s.hub.SendToUser(uid, msg)
		}(userID)
	}
	wg.Wait()
}
