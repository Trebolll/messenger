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

type MockWallManager struct {
	mock.Mock
}

func (m *MockWallManager) InitWall(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetByUsernameAndEmail(username, email, phone string) (*model.User, error) {
	args := m.Called(username, email, phone)
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

func (m *MockUserRepository) GetByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmailOrPhone(login string) (*model.User, error) {
	args := m.Called(login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	args := m.Called(userID, hashedPassword)
	return args.Error(0)
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

func (m *MockUserRepository) GetByPhone(phone string) (*model.User, error) {
	args := m.Called(phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) CreateByPhone(phone, username string) (*model.User, error) {
	args := m.Called(phone, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func TestCreateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com", "").Return(nil, nil)
	mockRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "testuser" && u.Email == "test@example.com"
	})).Return(nil)
	mockWall.On("InitWall", mock.Anything).Return(nil)

	userService := service.NewUserService(mockRepo, mockWall)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com", "")
	mockRepo.AssertCalled(t, "Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "testuser" && u.Email == "test@example.com"
	}))
	mockRepo.AssertExpectations(t)
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)

	userService := service.NewUserService(mockRepo, mockWall)
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
	mockWall := new(MockWallManager)

	userService := service.NewUserService(mockRepo, mockWall)
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
	mockWall := new(MockWallManager)

	userService := service.NewUserService(mockRepo, mockWall)
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
	mockWall := new(MockWallManager)
	existingUser := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "existing@example.com",
	}
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com", "").Return(existingUser, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "имя пользователя уже занято", err.Error())
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com", "")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	existingUser := &model.User{
		ID:       uuid.New(),
		Username: "otheruser",
		Email:    "test@example.com",
	}
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com", "").Return(existingUser, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "пользователь с таким адресом электронной почты уже существует", err.Error())
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com", "")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_PasswordTooShort(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com", "").Return(nil, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "short",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "пароль должен содержать не менее 6 символов", err.Error())
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com", "")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateUser_DatabaseError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com", "").Return(nil, nil)
	mockRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "testuser"
	})).Return(errors.New("database connection error"))

	userService := service.NewUserService(mockRepo, mockWall)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	err := userService.CreateUser(user)

	assert.Error(t, err)
	assert.Equal(t, "database connection error", err.Error())
	mockRepo.AssertCalled(t, "GetByUsernameAndEmail", "testuser", "test@example.com", "")
	mockRepo.AssertCalled(t, "Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "testuser"
	}))
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	hashedPassword := hashPassword("password123")
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
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
	mockWall := new(MockWallManager)
	mockRepo.On("GetByEmail", "notfound@example.com").Return(nil, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.LoginUser("notfound@example.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "неверные учетные данные электронной почты или пароль", err.Error())
	mockRepo.AssertCalled(t, "GetByEmail", "notfound@example.com")
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	mockRepo.On("GetByEmail", "test@example.com").Return(nil, errors.New("database connection error"))

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.LoginUser("test@example.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database connection error", err.Error())
	mockRepo.AssertCalled(t, "GetByEmail", "test@example.com")
	mockRepo.AssertExpectations(t)
}

func TestLoginUser_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	hashedPassword := hashPassword("correctpassword")
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.LoginUser("test@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "неверные учетные данные электронной почты или пароль", err.Error())
	mockRepo.AssertCalled(t, "GetByEmail", "test@example.com")
	mockRepo.AssertExpectations(t)
}

func TestSearchUsers_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
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

	userService := service.NewUserService(mockRepo, mockWall)
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
	mockWall := new(MockWallManager)
	mockRepo.On("SearchByUsername", "nonexistent").Return([]model.User{}, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.SearchUsers("nonexistent")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	mockRepo.AssertCalled(t, "SearchByUsername", "nonexistent")
	mockRepo.AssertExpectations(t)
}

func TestSearchUsers_QueryTooShort(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.SearchUsers("ab")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "поисковый запрос должен содержать не менее 3 символов", err.Error())
	mockRepo.AssertNotCalled(t, "SearchByUsername")
	mockRepo.AssertExpectations(t)
}

func TestSearchUsers_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	mockRepo.On("SearchByUsername", "test").Return(nil, errors.New("database connection error"))

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.SearchUsers("test")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database connection error", err.Error())
	mockRepo.AssertCalled(t, "SearchByUsername", "test")
	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	existingUser := &model.User{ID: userID, Username: "olduser"}
	newName := "New Name"
	newUsername := "newuser"

	mockRepo.On("GetById", userID).Return(existingUser, nil)
	mockRepo.On("UpdateProfile", mock.Anything).Return(nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateProfile(userID, nil, &newName, &newUsername, nil, nil, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, newName, result.FullName)
	assert.Equal(t, newUsername, result.Username)
	mockRepo.AssertExpectations(t)
}

func TestUpdateStatus_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	status := "online"
	user := &model.User{ID: userID, Status: status}

	mockRepo.On("UpdateStatus", userID, status).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateStatus(userID, status)

	assert.NoError(t, err)
	assert.Equal(t, status, result.Status)
	mockRepo.AssertExpectations(t)
}

func TestUpdateAvatarUrl_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	url := "http://example.com/avatar.png"

	mockRepo.On("UpdateAvatarUrl", userID, url).Return(nil)

	userService := service.NewUserService(mockRepo, mockWall)
	err := userService.UpdateAvatarUrl(userID, url)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetUserByID_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	user := &model.User{ID: userID, Username: "testuser"}

	mockRepo.On("GetById", userID).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.GetUserByID(userID)

	assert.NoError(t, err)
	assert.Equal(t, user, result)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_UserIdAsString(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	existingUser := &model.User{ID: userID, Username: "olduser"}
	newName := "New Name"

	mockRepo.On("GetById", userID).Return(existingUser, nil)
	mockRepo.On("UpdateProfile", mock.Anything).Return(nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateProfile(userID.String(), nil, &newName, nil, nil, nil, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, newName, result.FullName)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_InvalidUserIdString(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateProfile("invalid-uuid", nil, nil, nil, nil, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid user id", err.Error())
}

func TestUpdateProfile_InvalidUserIdType(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateProfile(123, nil, nil, nil, nil, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid user id type", err.Error())
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	mockRepo.On("GetById", userID).Return(nil, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateProfile(userID, nil, nil, nil, nil, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())
}

func TestUpdateProfile_BirthDate(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	existingUser := &model.User{ID: userID}
	birthDate := "1990-01-01"

	mockRepo.On("GetById", userID).Return(existingUser, nil)
	mockRepo.On("UpdateProfile", mock.Anything).Return(nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateProfile(userID, nil, nil, nil, &birthDate, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result.BirthDate)
	assert.Equal(t, "1990-01-01", result.BirthDate.Format("2006-01-02"))
	mockRepo.AssertExpectations(t)
}

func TestUpdateProfile_BirthDateInvalid(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	existingUser := &model.User{ID: userID}
	birthDate := "01-01-1990"

	mockRepo.On("GetById", userID).Return(existingUser, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateProfile(userID, nil, nil, nil, &birthDate, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid birth date format, expected YYYY-MM-DD", err.Error())
}

func TestUpdateProfile_BirthDateEmpty(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	existingUser := &model.User{ID: userID}
	birthDate := ""

	mockRepo.On("GetById", userID).Return(existingUser, nil)
	mockRepo.On("UpdateProfile", mock.Anything).Return(nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateProfile(userID, nil, nil, nil, &birthDate, nil, nil, nil)

	assert.NoError(t, err)
	assert.Nil(t, result.BirthDate)
	mockRepo.AssertExpectations(t)
}

func TestUpdateStatus_UserIdAsString(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	status := "away"
	user := &model.User{ID: userID, Status: status}

	mockRepo.On("UpdateStatus", userID, status).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateStatus(userID.String(), status)

	assert.NoError(t, err)
	assert.Equal(t, status, result.Status)
	mockRepo.AssertExpectations(t)
}

func TestUpdateStatus_InvalidUserIdType(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.UpdateStatus(123, "online")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid user id type", err.Error())
}

func TestGetOrCreateByPhone_Success_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	phone := "+1234567890"
	user := &model.User{ID: uuid.New(), Phone: phone, Password: "hashedpassword"}

	mockRepo.On("GetByPhone", phone).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.GetOrCreateByPhone(phone, "")

	assert.NoError(t, err)
	assert.Equal(t, phone, result.Phone)
	assert.Empty(t, result.Password)
	mockRepo.AssertExpectations(t)
}

func TestGetOrCreateByPhone_Success_Create(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	phone := "+1234567890"
	username := "testuser"
	user := &model.User{ID: uuid.New(), Phone: phone, Username: username}

	mockRepo.On("GetByPhone", phone).Return(nil, nil)
	mockRepo.On("GetByUsernameAndEmail", username, "", "").Return(nil, nil)
	mockRepo.On("CreateByPhone", phone, username).Return(user, nil)
	mockWall.On("InitWall", user.ID).Return(nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.GetOrCreateByPhone(phone, username)

	assert.NoError(t, err)
	assert.Equal(t, username, result.Username)
	mockRepo.AssertExpectations(t)
	mockWall.AssertExpectations(t)
}

func TestGetOrCreateByEmail_Success_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	email := "test@example.com"
	user := &model.User{ID: uuid.New(), Email: email, Password: "hashedpassword"}

	mockRepo.On("GetByEmail", email).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.GetOrCreateByEmail(email, "")

	assert.NoError(t, err)
	assert.Equal(t, email, result.Email)
	assert.Empty(t, result.Password)
}

func TestExistsByLogin_Phone(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	phone := "+1234567890"
	user := &model.User{ID: uuid.New(), Phone: phone}

	mockRepo.On("GetByPhone", phone).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	exists, err := userService.ExistsByLogin(phone, "phone")

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestLoginByPhone_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	phone := "+1234567890"
	password := "password123"
	hashedPassword := hashPassword(password)
	user := &model.User{ID: uuid.New(), Phone: phone, Password: hashedPassword}

	mockRepo.On("GetByPhone", phone).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.LoginByPhone(phone, password)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Password)
}

func TestLoginByUsername_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	username := "testuser"
	password := "password123"
	hashedPassword := hashPassword(password)
	user := &model.User{ID: uuid.New(), Username: username, Password: hashedPassword}

	mockRepo.On("GetByUsername", username).Return(user, nil)

	userService := service.NewUserService(mockRepo, mockWall)
	result, err := userService.LoginByUsername(username, password)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Password)
}

func TestSetBirthDate_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	userID := uuid.New()
	user := &model.User{ID: userID}
	dateStr := "1990-01-01"

	mockRepo.On("GetById", userID).Return(user, nil)
	mockRepo.On("UpdateProfile", mock.MatchedBy(func(u *model.User) bool {
		return u.BirthDate != nil && u.BirthDate.Format("2006-01-02") == dateStr
	})).Return(nil)

	userService := service.NewUserService(mockRepo, mockWall)
	err := userService.SetBirthDate(userID, dateStr)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePasswordByLogin_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockWall := new(MockWallManager)
	login := "test@example.com"
	user := &model.User{ID: uuid.New(), Email: login}

	mockRepo.On("GetByEmailOrPhone", login).Return(user, nil)
	mockRepo.On("UpdatePassword", user.ID, mock.Anything).Return(nil)

	userService := service.NewUserService(mockRepo, mockWall)
	err := userService.UpdatePasswordByLogin(login, "newpassword")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
