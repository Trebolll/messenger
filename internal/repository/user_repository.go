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
	err := r.db.QueryRow("SELECT id, username, email, password FROM users WHERE email = $1", email).Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetById(id uuid.UUID) (*model.User, error) {
	u := new(model.User)
	err := r.db.QueryRow("SELECT id, username, email FROM users WHERE id = $1", id).Scan(&u.ID, &u.Username, &u.Email)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var u model.User
	query := `SELECT id, username, email FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&u.ID, &u.Username, &u.Email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) VerifyPassword(email, password string) (*model.User, error) {
	user, err := r.GetByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("пользователь не найден")
	}

	// Сравниваем хеши
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("неверный пароль")
	}

	// Очищаем пароль из объекта
	user.Password = ""
	return user, nil
}
