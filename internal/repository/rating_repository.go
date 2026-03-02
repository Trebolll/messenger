package repository

import (
	"database/sql"
	"messenger/internal/model"
	"strings"

	"github.com/google/uuid"
)

type RatingRepository struct {
	db *sql.DB
}

func NewRatingRepository(db *sql.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) Vote(messageID, voterID uuid.UUID, vote int) (*model.RatingVoteResult, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var prevVote int
	err = tx.QueryRow(`SELECT vote FROM message_votes WHERE message_id=$1 AND voter_id=$2`, messageID, voterID).Scan(&prevVote)
	noVote := err == sql.ErrNoRows

	var senderID, chatID uuid.UUID
	if err2 := tx.QueryRow(`SELECT sender_id, chat_id FROM messages WHERE id=$1`, messageID).Scan(&senderID, &chatID); err2 != nil {
		return nil, err2
	}

	if senderID == voterID {
		return nil, sql.ErrNoRows
	}

	var ratingDelta int
	newVote := vote

	if noVote {
		_, err = tx.Exec(`INSERT INTO message_votes(message_id, voter_id, vote) VALUES($1,$2,$3)`, messageID, voterID, vote)
		ratingDelta = vote
	} else if prevVote == vote {
		_, err = tx.Exec(`DELETE FROM message_votes WHERE message_id=$1 AND voter_id=$2`, messageID, voterID)
		ratingDelta = -vote
		newVote = 0
	} else {
		_, err = tx.Exec(`UPDATE message_votes SET vote=$1 WHERE message_id=$2 AND voter_id=$3`, vote, messageID, voterID)
		ratingDelta = vote - prevVote
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`UPDATE messages SET likes=(SELECT COUNT(*) FROM message_votes WHERE message_id=$1 AND vote=1), dislikes=(SELECT COUNT(*) FROM message_votes WHERE message_id=$1 AND vote=-1) WHERE id=$1`, messageID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`UPDATE users SET rating = rating + $1 WHERE id=$2`, ratingDelta, senderID)
	if err != nil {
		return nil, err
	}

	result := &model.RatingVoteResult{MessageID: messageID, ChatID: chatID, SenderID: senderID, MyVote: newVote}
	_ = tx.QueryRow(`SELECT likes, dislikes FROM messages WHERE id=$1`, messageID).Scan(&result.Likes, &result.Dislikes)
	_ = tx.QueryRow(`SELECT rating FROM users WHERE id=$1`, senderID).Scan(&result.SenderRating)

	return result, tx.Commit()
}

func (r *RatingRepository) GetUserRating(userID uuid.UUID) (int, error) {
	var rating int
	err := r.db.QueryRow(`SELECT rating FROM users WHERE id=$1`, userID).Scan(&rating)
	return rating, err
}

func (r *RatingRepository) GetVotesForMessagesFixed(msgs []model.Message, voterID uuid.UUID) (map[uuid.UUID]int, error) {
	result := make(map[uuid.UUID]int)
	if len(msgs) == 0 {
		return result, nil
	}
	ids := make([]string, len(msgs))
	for i, m := range msgs {
		ids[i] = m.ID.String()
	}
	rows, err := r.db.Query(`SELECT message_id, vote FROM message_votes WHERE voter_id=$1 AND message_id = ANY($2::uuid[])`, voterID, "{"+strings.Join(ids, ",")+"}")
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var msgID uuid.UUID
		var v int
		if err := rows.Scan(&msgID, &v); err == nil {
			result[msgID] = v
		}
	}
	return result, nil
}
