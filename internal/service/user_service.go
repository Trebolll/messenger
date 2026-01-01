package service

import (
	"errors"
	"messenger/internal/model"
	"messenger/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(u *model.User) error {

	existingUser, _ := s.repo.GetByUsername(u.Username)
	if existingUser != nil {
		return errors.New("user with this username already exists")
	}

	existingEmail, _ := s.repo.GetByEmail(u.Email)
	if existingEmail != nil {
		return errors.New("user with this email already exists")
	}

	if len(u.Password) < 6 {
		return errors.New("password must contain at least 6 characters")
	}

	hashedPassword, err := hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword

	return s.repo.Create(u)
}

func hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
