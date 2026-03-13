package service

import (
	"errors"
	"fmt"
	"messenger/internal/model"
	_ "messenger/internal/repository"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetByUsernameAndEmail(username, email, phone string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetById(id uuid.UUID) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	Create(u *model.User) error
	SearchByUsername(username string) ([]model.User, error)
	UpdateProfile(u *model.User) error
	UpdateStatus(id uuid.UUID, status string) (*model.User, error)
	UpdateAvatarUrl(userID uuid.UUID, url string) error
	GetByEmailOrPhone(login string) (*model.User, error)
	UpdatePassword(userID uuid.UUID, hashedPassword string) error
}

type WallManager interface {
	InitWall(userID uuid.UUID) error
}

type UserService struct {
	repo UserRepository
	wall WallManager
}

func NewUserService(repo UserRepository, wall WallManager) *UserService {
	return &UserService{repo: repo, wall: wall}
}

func (s *UserService) CreateUser(u *model.User) error {
	if u.Email != "" && !isValidEmail(u.Email) {
		return errors.New("неверный формат электронной почты, формат должен быть в виде example@example.com")
	}

	if len(u.Username) < 3 || len(u.Username) > 50 {
		return errors.New("имя пользователя должно содержать от 3 до 50 символов")
	}

	existingUser, _ := s.repo.GetByUsernameAndEmail(u.Username, u.Email, u.Phone)
	if existingUser != nil {
		if existingUser.Username == u.Username {
			return errors.New("имя пользователя уже занято")
		}
		if u.Email != "" && existingUser.Email == u.Email {
			return errors.New("пользователь с таким адресом электронной почты уже существует")
		}
		if u.Phone != "" && existingUser.Phone == u.Phone {
			return errors.New("пользователь с таким номером телефона уже существует")
		}
		return errors.New("пользователь с такими данными уже существует")
	}

	if len(u.Password) < 6 && u.Password != "" {
		return errors.New("пароль должен содержать не менее 6 символов")
	}

	hashedPassword, err := hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword

	if err := s.repo.Create(u); err != nil {
		return err
	}

	// Инициализируем стену для нового пользователя
	return s.wall.InitWall(u.ID)
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

func (s *UserService) UpdateProfile(userID interface{}, phone *string, fullName *string, username *string, birthDate *string, location *string, profession *string, status *string) (*model.User, error) {
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
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("user not found")
	}

	// Обновляем только переданные поля (nil = поле не передано = не трогаем)
	if phone != nil {
		existing.Phone = *phone
	}
	if fullName != nil {
		existing.FullName = *fullName
	}
	if username != nil && len(*username) >= 3 {
		existing.Username = *username
	}
	// location, profession, status: nil = не трогаем, "" = очищаем, "value" = устанавливаем
	if location != nil {
		existing.Location = *location
	}
	if profession != nil {
		existing.Profession = *profession
	}
	if status != nil {
		existing.Status = *status
	}
	if birthDate != nil {
		if *birthDate == "" {
			// Пустая строка = очищаем дату рождения
			existing.BirthDate = nil
		} else {
			parsedDate, err := time.Parse("2006-01-02", *birthDate)
			if err != nil {
				return nil, errors.New("invalid birth date format, expected YYYY-MM-DD")
			}
			existing.BirthDate = &parsedDate
		}
	}

	fmt.Printf("[UpdateProfile] BEFORE SAVE: location=%q profession=%q birthDate=%v\n", existing.Location, existing.Profession, existing.BirthDate)
	if err := s.repo.UpdateProfile(existing); err != nil {
		fmt.Printf("[UpdateProfile] DB ERROR: %v\n", err)
		return nil, err
	}
	fmt.Printf("[UpdateProfile] AFTER SAVE: location=%q profession=%q birthDate=%v\n", existing.Location, existing.Profession, existing.BirthDate)

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

func (s *UserService) UpdateAvatarUrl(userID uuid.UUID, url string) error {
	return s.repo.UpdateAvatarUrl(userID, url)
}

func (s *UserService) GetUserByID(id uuid.UUID) (*model.User, error) {
	return s.repo.GetById(id)
}

// GetOrCreateByPhone — ищет юзера по телефону, если нет — создаёт нового.
// username нужен только при создании; если пустой — используем номер как временное имя.
func (s *UserService) GetOrCreateByPhone(phone, username string) (*model.User, error) {
	// Интерфейс UserRepository нужно расширить — используем type assertion
	type phoneRepo interface {
		GetByPhone(phone string) (*model.User, error)
		CreateByPhone(phone, username string) (*model.User, error)
	}
	pr, ok := s.repo.(phoneRepo)
	if !ok {
		return nil, errors.New("репозиторий не поддерживает авторизацию по телефону")
	}

	user, err := pr.GetByPhone(phone)
	if err != nil {
		return nil, err
	}

	// Юзер уже есть — логин
	if user != nil {
		user.Password = ""
		return user, nil
	}

	// Новый юзер — регистрация
	if username == "" {
		// Временное имя из номера телефона пока пользователь не заполнит профиль
		username = "user_" + phone[len(phone)-4:]
	}
	if len(username) < 3 || len(username) > 50 {
		return nil, errors.New("имя пользователя должно содержать от 3 до 50 символов")
	}

	// Проверяем уникальность username
	existing, _ := s.repo.GetByUsernameAndEmail(username, "", "")
	if existing != nil {
		username = username + "_" + phone[len(phone)-4:]
	}

	user, err = pr.CreateByPhone(phone, username)
	if err != nil {
		return nil, err
	}

	// Инициализируем стену
	_ = s.wall.InitWall(user.ID)
	return user, nil
}

// GetOrCreateByEmail — ищет юзера по email, если нет — создаёт без пароля.
func (s *UserService) GetOrCreateByEmail(email, username string) (*model.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err == nil && user != nil {
		user.Password = ""
		return user, nil
	}

	// Новый пользователь
	if username == "" {
		username = strings.Split(email, "@")[0]
	}
	if len(username) < 3 {
		username = username + "_user"
	}

	// Проверяем уникальность username
	if existing, _ := s.repo.GetByUsernameAndEmail(username, "", ""); existing != nil {
		username = username + "_" + email[:3]
	}

	u := &model.User{Email: email, Username: username}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	_ = s.wall.InitWall(u.ID)
	return u, nil
}

// ExistsByLogin — проверяет существует ли юзер по email или телефону
func (s *UserService) ExistsByLogin(dest, method string) (bool, error) {
	var user *model.User
	var err error
	if method == "phone" {
		type phoneRepo interface {
			GetByPhone(string) (*model.User, error)
		}
		if pr, ok := s.repo.(phoneRepo); ok {
			user, err = pr.GetByPhone(dest)
		}
	} else {
		user, err = s.repo.GetByEmail(dest)
	}
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

// LoginByPhone — вход по номеру телефона и паролю
func (s *UserService) LoginByPhone(phone, password string) (*model.User, error) {
	type phoneRepo interface {
		GetByPhone(string) (*model.User, error)
	}
	pr, ok := s.repo.(phoneRepo)
	if !ok {
		return nil, errors.New("не поддерживается")
	}
	user, err := pr.GetByPhone(phone)
	if err != nil || user == nil {
		return nil, errors.New("неверные данные")
	}
	if !checkPasswordHash(password, user.Password) {
		return nil, errors.New("неверные данные")
	}
	user.Password = ""
	return user, nil
}

// LoginByUsername — вход по username и паролю
func (s *UserService) LoginByUsername(username, password string) (*model.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil || user == nil {
		return nil, errors.New("неверные данные")
	}
	if !checkPasswordHash(password, user.Password) {
		return nil, errors.New("неверные данные")
	}
	user.Password = ""
	return user, nil
}

// SetBirthDate — устанавливает дату рождения после регистрации
func (s *UserService) SetBirthDate(userID uuid.UUID, dateStr string) error {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errors.New("неверный формат даты, ожидается YYYY-MM-DD")
	}
	existing, err := s.repo.GetById(userID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("пользователь не найден")
	}
	existing.BirthDate = &t
	return s.repo.UpdateProfile(existing)
}

// UpdatePasswordByLogin — находит пользователя по email или phone и меняет пароль
func (s *UserService) UpdatePasswordByLogin(login, plainPassword string) error {
	user, err := s.repo.GetByEmailOrPhone(login)
	if err != nil || user == nil {
		return errors.New("пользователь не найден")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(user.ID, string(hash))
}
