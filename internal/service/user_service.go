package service

import (
	"errors"
	"messenger/internal/model"
	_ "messenger/internal/repository"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetByUsernameAndEmail(username string, email string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetById(id uuid.UUID) (*model.User, error)
	Create(u *model.User) error
	SearchByUsername(username string) ([]model.User, error)
	UpdateProfile(u *model.User) error
	UpdateStatus(id uuid.UUID, status string) (*model.User, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(u *model.User) error {

	if !isValidEmail(u.Email) {
		return errors.New("неверный формат электронной почты, формат должен быть в виде example@example.com")
	}

	if len(u.Username) < 3 || len(u.Username) > 50 {
		return errors.New("имя пользователя должно содержать от 3 до 50 символов")
	}

	existingUser, _ := s.repo.GetByUsernameAndEmail(u.Username, u.Email)
	if existingUser != nil {
		return errors.New("пользователь с таким именем или таким адресом электронной почты пользователя уже существует")
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

func (s *UserService) SearchUsers(username string) ([]model.User, error) {
	if len(username) < 3 {
		return nil, errors.New("поисковый запрос должен содержать не менее 3 символов")
	}
	return s.repo.SearchByUsername(username)
}

func hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
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

func (s *UserService) UpdateProfile(userID interface{}, phone *string, fullName *string, username *string, birthDate *string, location *string, status *string) (*model.User, error) {
	var id uuid.UUID
	switch v := userID.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, errors.New("invalid user id")
		}
		id = parsed
	case uuid.UUID:
		id = v
	default:
		return nil, errors.New("invalid user id type")
	}

	// Получаем текущие данные пользователя
	existing, err := s.repo.GetById(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Обновляем только переданные поля
	if phone != nil {
		existing.Phone = *phone
	}
	if fullName != nil {
		existing.FullName = *fullName
	}
	if username != nil && len(*username) >= 3 {
		existing.Username = *username
	}
	if location != nil {
		existing.Location = *location
	}
	if status != nil {
		existing.Status = *status
	}
	if birthDate != nil && *birthDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *birthDate)
		if err != nil {
			return nil, errors.New("invalid birth date format, expected YYYY-MM-DD")
		}
		existing.BirthDate = &parsedDate
	}

	if err := s.repo.UpdateProfile(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *UserService) UpdateStatus(userID interface{}, status string) (*model.User, error) {
	var id uuid.UUID
	switch v := userID.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, errors.New("invalid user id")
		}
		id = parsed
	case uuid.UUID:
		id = v
	default:
		return nil, errors.New("invalid user id type")
	}

	return s.repo.UpdateStatus(id, status)
}
