package service

import (
	"errors"
	"messenger/internal/model"
	"messenger/internal/repository"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(u *model.User) error {

	if !isValidEmail(u.Email) {
		return errors.New("неверный формат электронной почты, формат должен быть в виде example@example.com")
	}

	if len(u.Username) < 3 || len(u.Username) > 50 {
		return errors.New("имя пользователя должно содержать от 3 до 50 символов")
	}

	existingUser, _ := s.repo.GetByUsername(u.Username)
	if existingUser != nil {
		return errors.New("пользователь с таким именем пользователя уже существует")
	}

	existingEmail, _ := s.repo.GetByEmail(u.Email)
	if existingEmail != nil {
		return errors.New("пользователь с таким адресом электронной почты уже существует")
	}

	if len(u.Password) < 6 {
		return errors.New("пароль должен содержать не менее 6 символов")
	}

	hashedPassword, err := hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword

	return s.repo.Create(u)
}

func (s *UserService) LoginUser(email, password string) (*model.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("неверные учетные данные электронной почты или пароль")
	}

	if !checkPasswordHash(password, user.Password) {
		return nil, errors.New("неверные учетные данные электронной почты или пароль")
	}

	user.Password = ""
	return user, nil
}

func hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
