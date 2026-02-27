package service

import (
	"errors"
	"messenger/internal/model"
	"messenger/internal/service"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

var _ service.UserRepository = (*MockUserRepository)(nil)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByUsernameAndEmail(username string, email string) (*model.User, error) {
	args := m.Called(username, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(u *model.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockUserRepository) SearchByUsername(username string) ([]model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) GetById(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateProfile(u *model.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateStatus(id uuid.UUID, status string) (*model.User, error) {
	args := m.Called(id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateAvatarUrl(userID uuid.UUID, url string) error {
	args := m.Called(userID, url)
	return args.Error(0)
}

func TestCreateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com").Return(nil, nil)
	mockRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "testuser" && u.Email == "test@example.com"
	})).Return(nil)

	userService := service.NewUserService(mockRepo)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com")
	mockRepo.AssertCalled(t, "Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "testuser" && u.Email == "test@example.com"
	}))
	mockRepo.AssertExpectations(t)
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)

	userService := service.NewUserService(mockRepo)
	user := &model.User{
		Username: "testuser",
		Email:    "invalidemail",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "неверный формат электронной почты, формат должен быть в виде example@example.com", err.Error())
	mockRepo.AssertNotCalled(t, "GetByEmail")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_UsernameTooShort(t *testing.T) {
	mockRepo := new(MockUserRepository)

	userService := service.NewUserService(mockRepo)
	user := &model.User{
		Username: "ab",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "имя пользователя должно содержать от 3 до 50 символов", err.Error())
	mockRepo.AssertNotCalled(t, "GetByEmail")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_UsernameTooLong(t *testing.T) {
	mockRepo := new(MockUserRepository)

	userService := service.NewUserService(mockRepo)
	user := &model.User{
		Username: "a" + string(make([]byte, 50)) + "a",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "имя пользователя должно содержать от 3 до 50 символов", err.Error())
	mockRepo.AssertNotCalled(t, "GetByEmail")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_UsernameAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	existingUser := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "existing@example.com",
	}
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com").Return(existingUser, nil)

	userService := service.NewUserService(mockRepo)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "пользователь с таким именем или таким адресом электронной почты пользователя уже существует", err.Error())
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	existingUser := &model.User{
		ID:       uuid.New(),
		Username: "otheruser",
		Email:    "test@example.com",
	}
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com").Return(existingUser, nil)

	userService := service.NewUserService(mockRepo)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "пользователь с таким именем или таким адресом электронной почты пользователя уже существует", err.Error())
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_PasswordTooShort(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com").Return(nil, nil)

	userService := service.NewUserService(mockRepo)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "short",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "пароль должен содержать не менее 6 символов", err.Error())
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_DatabaseError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com").Return(nil, nil)
	mockRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "testuser"
	})).Return(errors.New("database connection error"))

	userService := service.NewUserService(mockRepo)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "database connection error", err.Error())
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com")
	mockRepo.AssertCalled(t, "Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "testuser"
	}))
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	hashedPassword := hashPassword("password123")
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

	userService := service.NewUserService(mockRepo)
	result, err := userService.LoginUser("test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "testuser", result.Username)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Empty(t, result.Password)
	mockRepo.AssertCalled(t, "GetByEmail", "test@example.com")
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByEmail", "notfound@example.com").Return(nil, nil)

	userService := service.NewUserService(mockRepo)
	result, err := userService.LoginUser("notfound@example.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "неверные учетные данные электронной почты или пароль", err.Error())
	mockRepo.AssertCalled(t, "GetByEmail", "notfound@example.com")
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByEmail", "test@example.com").Return(nil, errors.New("database connection error"))

	userService := service.NewUserService(mockRepo)
	result, err := userService.LoginUser("test@example.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database connection error", err.Error())
	mockRepo.AssertCalled(t, "GetByEmail", "test@example.com")
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	hashedPassword := hashPassword("correctpassword")
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

	userService := service.NewUserService(mockRepo)
	result, err := userService.LoginUser("test@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "неверные учетные данные электронной почты или пароль", err.Error())
	mockRepo.AssertCalled(t, "GetByEmail", "test@example.com")
	mockRepo.AssertExpectations(t)
}

func TestSearchUsers_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	users := []model.User{
		{
			ID:       uuid.New(),
			Username: "testuser1",
			Email:    "test1@example.com",
		},
		{
			ID:       uuid.New(),
			Username: "testuser2",
			Email:    "test2@example.com",
		},
	}
	mockRepo.On("SearchByUsername", "test").Return(users, nil)

	userService := service.NewUserService(mockRepo)
	result, err := userService.SearchUsers("test")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "testuser1", result[0].Username)
	assert.Equal(t, "testuser2", result[1].Username)
	mockRepo.AssertCalled(t, "SearchByUsername", "test")
	mockRepo.AssertExpectations(t)
}

func TestSearchUsers_EmptyResult(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("SearchByUsername", "nonexistent").Return([]model.User{}, nil)

	userService := service.NewUserService(mockRepo)
	result, err := userService.SearchUsers("nonexistent")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	mockRepo.AssertCalled(t, "SearchByUsername", "nonexistent")
	mockRepo.AssertExpectations(t)
}

func TestSearchUsers_QueryTooShort(t *testing.T) {
	mockRepo := new(MockUserRepository)

	userService := service.NewUserService(mockRepo)
	result, err := userService.SearchUsers("ab")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "поисковый запрос должен содержать не менее 3 символов", err.Error())
	mockRepo.AssertNotCalled(t, "SearchByUsername")
	mockRepo.AssertExpectations(t)
}

func TestSearchUsers_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("SearchByUsername", "test").Return(nil, errors.New("database connection error"))

	userService := service.NewUserService(mockRepo)
	result, err := userService.SearchUsers("test")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database connection error", err.Error())
	mockRepo.AssertCalled(t, "SearchByUsername", "test")
	mockRepo.AssertExpectations(t)
}
