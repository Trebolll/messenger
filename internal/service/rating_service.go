package service

import (
	"errors"
	"messenger/internal/model"
	"messenger/internal/service/websocket"

	"github.com/google/uuid"
)

type RatingRepositoryIface interface {
	Vote(messageID, voterID uuid.UUID, vote int) (*model.RatingVoteResult, error)
	GetUserRating(userID uuid.UUID) (int, error)
	GetVotesForMessagesFixed(msgs []model.Message, voterID uuid.UUID) (map[uuid.UUID]int, error)
}

type RatingChatRepo interface {
	GetChatMembers(chatID uuid.UUID) ([]uuid.UUID, error)
}

type RatingHub interface {
	SendToUser(userID uuid.UUID, msg websocket.Message)
}

type RatingService struct {
	repo     RatingRepositoryIface
	chatRepo RatingChatRepo
	hub      RatingHub
}

func NewRatingService(repo RatingRepositoryIface, chatRepo RatingChatRepo, hub RatingHub) *RatingService {
	return &RatingService{repo: repo, chatRepo: chatRepo, hub: hub}
}

func (s *RatingService) Vote(messageID, voterID uuid.UUID, vote int) error {
	if vote != 1 && vote != -1 {
		return errors.New("голос должен быть 1 или -1")
	}

	result, err := s.repo.Vote(messageID, voterID, vote)
	if err != nil {
		return errors.New("нельзя голосовать за своё сообщение")
	}

	publicRating := result.SenderRating
	if publicRating < 0 {
		publicRating = 0
	}

	members, err := s.chatRepo.GetChatMembers(result.ChatID)
	if err != nil {
		return err
	}

	for _, userID := range members {
		myVote := result.MyVote
		if userID != voterID {
			myVote = 0
		}
		s.hub.SendToUser(userID, websocket.Message{
			Type: "vote_updated",
			Content: map[string]interface{}{
				"message_id":    result.MessageID,
				"chat_id":       result.ChatID,
				"sender_id":     result.SenderID,
				"voter_id":      voterID,
				"likes":         result.Likes,
				"dislikes":      result.Dislikes,
				"my_vote":       myVote,
				"sender_rating": publicRating,
				"just_voted":    vote,
			},
		})
	}

	return nil
}

func (s *RatingService) GetUserRating(userID uuid.UUID) (int, error) {
	r, err := s.repo.GetUserRating(userID)
	if err != nil {
		return 0, err
	}
	if r < 0 {
		return 0, nil
	}
	return r, nil
}
