package repository

import (
	"database/sql"
	"errors"
	"messenger/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *model.User) error {
	query := `INSERT INTO users(username, email, password) VALUES($1,$2,$3) RETURNING id;`
	return r.db.QueryRow(query, u.Username, u.Email, u.Password).Scan(&u.ID)
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	u := new(model.User)
	var phone, fullName, location, status sql.NullString
	var birthDate sql.NullTime
	err := r.db.QueryRow(
		"SELECT id, username, email, password, phone, full_name, birth_date, location, status, COALESCE(avatar_url,'') FROM users WHERE email = $1", email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &phone, &fullName, &birthDate, &location, &status, &u.AvatarUrl)
	if err != nil {
		return nil, err
	}
	u.Phone = phone.String
	u.FullName = fullName.String
	u.Location = location.String
	u.Status = status.String
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	return u, nil
}

func (r *UserRepository) GetById(id uuid.UUID) (*model.User, error) {
	u := new(model.User)
	var phone, fullName, location, status sql.NullString
	var birthDate sql.NullTime
	err := r.db.QueryRow(
		"SELECT id, username, email, phone, full_name, birth_date, location, status, COALESCE(avatar_url,'') FROM users WHERE id = $1", id,
	).Scan(&u.ID, &u.Username, &u.Email, &phone, &fullName, &birthDate, &location, &status, &u.AvatarUrl)
	if err != nil {
		return nil, err
	}
	u.Phone = phone.String
	u.FullName = fullName.String
	u.Location = location.String
	u.Status = status.String
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	return u, nil
}

func (r *UserRepository) GetByUsernameAndEmail(username string, email string) (*model.User, error) {
	var u model.User
	query := `SELECT id, username, email FROM users WHERE username = $1 OR email = $2 LIMIT 1`
	err := r.db.QueryRow(query, username, email).Scan(&u.ID, &u.Username, &u.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	u := new(model.User)
	var phone, fullName, location, status sql.NullString
	var birthDate sql.NullTime
	query := `SELECT id, username, email, phone, full_name, birth_date, location, status, COALESCE(avatar_url,'') FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&u.ID, &u.Username, &u.Email, &phone, &fullName, &birthDate, &location, &status, &u.AvatarUrl)
	if err != nil {
		return nil, err
	}
	u.Phone = phone.String
	u.FullName = fullName.String
	u.Location = location.String
	u.Status = status.String
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	return u, nil
}

func (r *UserRepository) SearchByUsername(username string) ([]model.User, error) {
	query := `SELECT id, username, email FROM users WHERE username ILIKE $1 LIMIT 10`
	rows, err := r.db.Query(query, "%"+username+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) VerifyPassword(email, password string) (*model.User, error) {
	user, err := r.GetByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("пользователь не найден")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("неверный пароль")
	}

	user.Password = ""
	return user, nil
}

func (r *UserRepository) UpdateProfile(u *model.User) error {
	query := `UPDATE users SET phone = $1, full_name = $2, birth_date = $3, location = $4, status = $5, username = $6 , updated_at = NOW() WHERE id = $7`
	_, err := r.db.Exec(query, u.Phone, u.FullName, u.BirthDate, u.Location, u.Status, u.Username, u.ID)
	return err
}

func (r *UserRepository) UpdateStatus(id uuid.UUID, status string) (*model.User, error) {
	query := `UPDATE users SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return nil, err
	}

	return r.GetById(id)
}

func (r *UserRepository) UpdateAvatarUrl(userID uuid.UUID, url string) error {
	_, err := r.db.Exec(`UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2`, url, userID)
	return err
}
