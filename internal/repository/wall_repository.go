package repository

import (
	"database/sql"
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
	var bio, banner sql.NullString
	query := `SELECT id, user_id, bio, banner_url FROM walls WHERE user_id = $1`
	err := r.db.QueryRow(query, userID).Scan(&w.ID, &w.UserID, &bio, &banner)
	if err != nil {
		return nil, err
	}
	w.Bio = bio.String
	w.BannerUrl = banner.String
	return w, nil
}

func (r *WallRepository) UpdateWallSettings(userID uuid.UUID, bio string) error {
	query := `UPDATE walls SET bio = $1, updated_at = NOW() WHERE user_id = $2`
	_, err := r.db.Exec(query, bio, userID)
	return err
}

func (r *WallRepository) CreatePost(post *model.WallPost) error {
	query := `
		WITH inserted_post AS (
			INSERT INTO wall_posts(user_id, content)
			VALUES ($1, $2)
			RETURNING id, user_id, content, created_at, updated_at
		)
		SELECT p.id, p.user_id, p.content, p.created_at, p.updated_at, u.username, COALESCE(u.avatar_url, '')
		FROM inserted_post p
		JOIN users u ON p.user_id = u.id`

	return r.db.QueryRow(query, post.UserID, post.Content).Scan(
		&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.UpdatedAt,
		&post.AuthorName, &post.AuthorAvatar,
	)
}

func (r *WallRepository) AddAttachment(att *model.WallAttachment) error {
	query := `
		INSERT INTO wall_attachments(post_id, url, filename, mime_type, size_bytes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.db.QueryRow(query, att.PostID, att.Url, att.Filename, att.MimeType, att.SizeBytes).Scan(&att.ID, &att.CreatedAt)
}

func (r *WallRepository) GetPostsByUserID(userID uuid.UUID) ([]model.WallPost, error) {
	query := `
		SELECT p.id, p.user_id, p.content, p.created_at, p.updated_at, u.username, COALESCE(u.avatar_url, '')
		FROM wall_posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.WallPost
	for rows.Next() {
		var p model.WallPost
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Content, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorName, &p.AuthorAvatar,
		); err != nil {
			return nil, err
		}

		// Fetch attachments for each post
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
