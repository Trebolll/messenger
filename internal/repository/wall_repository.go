package repository

import (
	"database/sql"
	"fmt"
	"log"
	"messenger/internal/model"

	"github.com/google/uuid"
)

type WallRepository struct {
	db *sql.DB
}

func NewWallRepository(db *sql.DB) *WallRepository {
	return &WallRepository{db: db}
}

func (r *WallRepository) CreateWall(userID uuid.UUID) error {
	query := `INSERT INTO walls(user_id) VALUES($1) ON CONFLICT (user_id) DO NOTHING;`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *WallRepository) GetWallByUserID(userID uuid.UUID) (*model.Wall, error) {
	w := new(model.Wall)
	var bio, banner, avatar, status, location, profession sql.NullString
	query := `
		SELECT w.id, w.user_id, w.bio, w.banner_url, u.username, u.avatar_url, u.status,
		       GREATEST(0, u.rating) as rating, u.location, u.created_at, u.profession, u.birth_date,
		       COALESCE((
		           SELECT COUNT(*) FROM wall_post_likes wpl
		           JOIN wall_posts wp ON wpl.post_id = wp.id
		           WHERE wp.user_id = u.id
		       ), 0) as total_wall_likes
		FROM walls w
		JOIN users u ON w.user_id = u.id
		WHERE w.user_id = $1`
	err := r.db.QueryRow(query, userID).Scan(
		&w.ID, &w.UserID, &bio, &banner, &w.Username, &avatar, &status,
		&w.UserRating, &location, &w.UserCreatedAt, &profession, &w.UserBirthDate,
		&w.TotalWallLikes,
	)
	if err != nil {
		return nil, err
	}
	w.Bio = bio.String
	w.BannerUrl = banner.String
	w.AvatarUrl = avatar.String
	w.Status = status.String
	w.UserLocation = location.String
	w.UserProfession = profession.String
	return w, nil
}

func (r *WallRepository) UpdateWallSettings(userID uuid.UUID, bio string) error {
	query := `UPDATE walls SET bio = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.db.Exec(query, bio, userID)
	return err
}

func (r *WallRepository) CreatePost(post *model.WallPost) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Создаём пост
	err = tx.QueryRow(`
		WITH inserted_post AS (
			INSERT INTO wall_posts(user_id, content)
			VALUES ($1, $2)
			RETURNING id, user_id, content, created_at, updated_at
		)
		SELECT p.id, p.user_id, p.content, p.created_at, p.updated_at, u.username, COALESCE(u.avatar_url, '')
		FROM inserted_post p
		JOIN users u ON p.user_id = u.id`,
		post.UserID, post.Content,
	).Scan(
		&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.UpdatedAt,
		&post.AuthorName, &post.AuthorAvatar,
	)
	if err != nil {
		return err
	}

	// 2. Создаём публичный чат для поста
	var chatID uuid.UUID
	err = tx.QueryRow(
		`INSERT INTO chats(type, name, creator_id) VALUES('public', $1, $2) RETURNING id`,
		post.Content, post.UserID,
	).Scan(&chatID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO chat_members(chat_id, user_id) VALUES($1,$2)`, chatID, post.UserID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE wall_posts SET chat_id=$1 WHERE id=$2`, chatID, post.ID)
	if err != nil {
		return err
	}

	post.ChatID = &chatID
	return tx.Commit()
}

func (r *WallRepository) AddAttachment(att *model.WallAttachment) error {
	query := `
		INSERT INTO wall_attachments(post_id, url, filename, mime_type, size_bytes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.db.QueryRow(query, att.PostID, att.Url, att.Filename, att.MimeType, att.SizeBytes).Scan(&att.ID, &att.CreatedAt)
}

func (r *WallRepository) GetPostsByUserID(userID uuid.UUID, viewerID uuid.UUID) ([]model.WallPost, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.created_at, p.updated_at, u.username, COALESCE(u.avatar_url, ''),
		       COUNT(DISTINCT l.user_id) as likes_count,
		       COALESCE(BOOL_OR(l.user_id = $2), false) as is_liked,
		       p.chat_id,
		       (SELECT COUNT(*) FROM messages m WHERE m.chat_id = p.chat_id) as comments_count
		FROM wall_posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN wall_post_likes l ON l.post_id = p.id
		WHERE p.user_id = $1
		GROUP BY p.id, p.user_id, p.content, p.created_at, p.updated_at, p.chat_id, u.username, u.avatar_url
		ORDER BY p.created_at DESC`

	rows, err := r.db.Query(query, userID, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.WallPost = []model.WallPost{}
	for rows.Next() {
		var p model.WallPost
		var chatID sql.NullString
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Content, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorName, &p.AuthorAvatar, &p.LikesCount, &p.IsLiked, &chatID, &p.CommentsCount,
		); err != nil {
			return nil, err
		}
		if chatID.Valid && chatID.String != "" {
			uid, err := uuid.Parse(chatID.String)
			if err == nil {
				p.ChatID = &uid
			}
		}
		attachments, err := r.GetAttachmentsByPostID(p.ID)
		if err != nil {
			log.Printf("Error fetching attachments for post %s: %v", p.ID, err)
		} else {
			p.Attachments = attachments
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func (r *WallRepository) GetTotalWallLikes(userID uuid.UUID) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM wall_post_likes wpl
		JOIN wall_posts wp ON wpl.post_id = wp.id
		WHERE wp.user_id = $1`
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (r *WallRepository) ToggleLike(postID uuid.UUID, userID uuid.UUID) (liked bool, count int, err error) {
	// Проверяем есть ли лайк
	var exists bool
	err = r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM wall_post_likes WHERE post_id=$1 AND user_id=$2)`, postID, userID).Scan(&exists)
	if err != nil {
		return
	}
	if exists {
		_, err = r.db.Exec(`DELETE FROM wall_post_likes WHERE post_id=$1 AND user_id=$2`, postID, userID)
		liked = false
	} else {
		_, err = r.db.Exec(`INSERT INTO wall_post_likes(post_id, user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, postID, userID)
		liked = true
	}
	if err != nil {
		return
	}
	err = r.db.QueryRow(`SELECT COUNT(*) FROM wall_post_likes WHERE post_id=$1`, postID).Scan(&count)
	return
}

func (r *WallRepository) GetPostChat(postID uuid.UUID) (uuid.UUID, error) {
	var chatIDStr sql.NullString
	err := r.db.QueryRow(`SELECT chat_id FROM wall_posts WHERE id=$1`, postID).Scan(&chatIDStr)
	if err != nil {
		return uuid.Nil, err
	}
	if !chatIDStr.Valid || chatIDStr.String == "" {
		return uuid.Nil, fmt.Errorf("чат не найден для поста %s", postID)
	}
	return uuid.Parse(chatIDStr.String)
}

func (r *WallRepository) GetAttachmentsByPostID(postID uuid.UUID) ([]model.WallAttachment, error) {
	query := `
		SELECT id, post_id, url, filename, mime_type, size_bytes, created_at
		FROM wall_attachments
		WHERE post_id = $1`

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []model.WallAttachment
	for rows.Next() {
		var a model.WallAttachment
		if err := rows.Scan(
			&a.ID, &a.PostID, &a.Url, &a.Filename, &a.MimeType, &a.SizeBytes, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	return attachments, nil
}

func (r *WallRepository) GetPostOwnerByChatID(chatID uuid.UUID) (uuid.UUID, error) {
	var ownerID uuid.UUID
	err := r.db.QueryRow(`SELECT user_id FROM wall_posts WHERE chat_id = $1`, chatID).Scan(&ownerID)
	return ownerID, err
}

func (r *WallRepository) GetPostOwner(postID uuid.UUID) (uuid.UUID, error) {
	var ownerID uuid.UUID
	err := r.db.QueryRow(`SELECT user_id FROM wall_posts WHERE id = $1`, postID).Scan(&ownerID)
	return ownerID, err
}

func (r *WallRepository) DeletePost(postID uuid.UUID, userID uuid.UUID) (uuid.UUID, error) {
	var ownerID uuid.UUID
	err := r.db.QueryRow(`SELECT user_id FROM wall_posts WHERE id = $1`, postID).Scan(&ownerID)
	if err != nil {
		return uuid.Nil, err
	}

	query := `DELETE FROM wall_posts WHERE id = $1 AND user_id = $2`
	res, err := r.db.Exec(query, postID, userID)
	if err != nil {
		return uuid.Nil, err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return uuid.Nil, fmt.Errorf("пост не найден или нет прав")
	}
	return ownerID, nil
}

func (r *WallRepository) DeleteAttachment(attID uuid.UUID, userID uuid.UUID) error {
	// Удаляем только если вложение принадлежит посту этого пользователя
	query := `
		DELETE FROM wall_attachments wa
		USING wall_posts wp
		WHERE wa.post_id = wp.id AND wa.id = $1 AND wp.user_id = $2`
	res, err := r.db.Exec(query, attID, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("вложение не найдено или нет прав")
	}
	return nil
}

func (r *WallRepository) GetAllMediaByUserID(userID uuid.UUID) ([]model.WallAttachment, error) {
	query := `
		SELECT wa.id, wa.post_id, wa.url, wa.filename, wa.mime_type, wa.size_bytes, wa.created_at
		FROM wall_attachments wa
		JOIN wall_posts wp ON wa.post_id = wp.id
		WHERE wp.user_id = $1 
		  AND (wa.mime_type LIKE 'image/%' OR wa.mime_type LIKE 'video/%')
		  AND (wp.content IS NULL OR TRIM(wp.content) = '')
		ORDER BY wa.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []model.WallAttachment = []model.WallAttachment{}
	for rows.Next() {
		var a model.WallAttachment
		if err := rows.Scan(
			&a.ID, &a.PostID, &a.Url, &a.Filename, &a.MimeType, &a.SizeBytes, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		media = append(media, a)
	}
	return media, nil
}

func (r *WallRepository) GetPostByID(postID uuid.UUID) (*model.WallPost, error) {
	var p model.WallPost
	var chatID sql.NullString
	err := r.db.QueryRow(`
		SELECT p.id, p.user_id, p.content, p.created_at, p.updated_at, u.username, COALESCE(u.avatar_url, ''),
		       COUNT(DISTINCT l.user_id) as likes_count,
		       false as is_liked,
		       p.chat_id,
		       (SELECT COUNT(*) FROM messages m WHERE m.chat_id = p.chat_id) as comments_count
		FROM wall_posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN wall_post_likes l ON l.post_id = p.id
		WHERE p.id = $1
		GROUP BY p.id, p.user_id, p.content, p.created_at, p.updated_at, p.chat_id, u.username, u.avatar_url`,
		postID,
	).Scan(
		&p.ID, &p.UserID, &p.Content, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorName, &p.AuthorAvatar, &p.LikesCount, &p.IsLiked, &chatID, &p.CommentsCount,
	)
	if err != nil {
		return nil, err
	}
	if chatID.Valid && chatID.String != "" {
		uid, err := uuid.Parse(chatID.String)
		if err == nil {
			p.ChatID = &uid
		}
	}
	attachments, err := r.GetAttachmentsByPostID(p.ID)
	if err == nil {
		p.Attachments = attachments
	}
	return &p, nil
}

func (r *WallRepository) GetGlobalMediaFeed(viewerID uuid.UUID) ([]model.WallPost, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.created_at, p.updated_at, u.username, COALESCE(u.avatar_url, ''),
		       COUNT(DISTINCT l.user_id) as likes_count,
		       COALESCE(BOOL_OR(l.user_id = $1), false) as is_liked,
		       p.chat_id,
		       (SELECT COUNT(*) FROM messages m WHERE m.chat_id = p.chat_id) as comments_count
		FROM wall_posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN wall_post_likes l ON l.post_id = p.id
		WHERE EXISTS (
			SELECT 1 FROM wall_attachments wa 
			WHERE wa.post_id = p.id 
			AND (wa.mime_type LIKE 'image/%' OR wa.mime_type LIKE 'video/%')
		)
		GROUP BY p.id, p.user_id, p.content, p.created_at, p.updated_at, p.chat_id, u.username, u.avatar_url
		ORDER BY p.created_at DESC
		LIMIT 50`

	rows, err := r.db.Query(query, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.WallPost = []model.WallPost{}
	for rows.Next() {
		var p model.WallPost
		var chatID sql.NullString
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Content, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorName, &p.AuthorAvatar, &p.LikesCount, &p.IsLiked, &chatID, &p.CommentsCount,
		); err != nil {
			return nil, err
		}
		if chatID.Valid && chatID.String != "" {
			uid, err := uuid.Parse(chatID.String)
			if err == nil {
				p.ChatID = &uid
			}
		}
		attachments, err := r.GetAttachmentsByPostID(p.ID)
		if err != nil {
			log.Printf("Error fetching attachments for post %s: %v", p.ID, err)
		} else {
			p.Attachments = attachments
		}
		posts = append(posts, p)
	}
	return posts, nil
}
