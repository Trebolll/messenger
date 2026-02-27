package handler

import (
	"bytes"
	"encoding/json"
	"messenger/internal/handler"
	"messenger/internal/model"
	"messenger/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsernameAndEmail(username string, email string) (*model.User, error) {
	args := m.Called(username, email)
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

func setupTestRouter(mockRepo *MockUserRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	userService := service.NewUserService(mockRepo)
	userHandler := handler.NewUserHandler(userService, nil, nil)

	router.POST("/register", userHandler.Register)
	router.POST("/login", userHandler.Login)
	router.GET("/search", userHandler.SearchUsers)

	return router
}

func TestRegisterSuccess(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com").Return(nil, nil)
	mockRepo.On("Create", mock.MatchedBy(func(u *model.User) bool {
		return u.Username == "testuser" && u.Email == "test@example.com"
	})).Return(nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"username":"testuser","email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "пользователь успешно создан", response["сообщение"])
	mockRepo.AssertExpectations(t)
}

func TestRegisterInvalidJSON(t *testing.T) {
	mockRepo := new(MockUserRepository)
	router := setupTestRouter(mockRepo)
	body := []byte(`{invalid json}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["ошибка"])
}

func TestRegisterInvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "invalid-email").Return(nil, nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"username":"testuser","email":"invalid-email","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "неверный формат электронной почты")
}

func TestRegisterUsernameTooShort(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByUsernameAndEmail", "ab", "test@example.com").Return(nil, nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"username":"ab","email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "от 3 до 50 символов")
}

func TestRegisterUsernameTooLong(t *testing.T) {
	mockRepo := new(MockUserRepository)
	longUsername := "a"
	for i := 0; i < 51; i++ {
		longUsername += "a"
	}
	mockRepo.On("GetByUsernameAndEmail", longUsername, "test@example.com").Return(nil, nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"username":"` + longUsername + `","email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "от 3 до 50 символов")
}

func TestRegisterPasswordTooShort(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com").Return(nil, nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"username":"testuser","email":"test@example.com","password":"pass"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "не менее 6 символов")
}

func TestRegisterDuplicateUsername(t *testing.T) {
	existingUser := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "existing@example.com",
	}
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com").Return(existingUser, nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"username":"testuser","email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "имен")
}

func TestRegisterDuplicateEmail(t *testing.T) {
	existingUser := &model.User{
		ID:       uuid.New(),
		Username: "existinguser",
		Email:    "test@example.com",
	}
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByUsernameAndEmail", "testuser", "test@example.com").Return(existingUser, nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"username":"testuser","email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["ошибка"], "адресом электронной почты")
}

func TestLoginSuccess(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 14)
	user := &model.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"email":"test@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
	assert.NotNil(t, response["user"])
	mockRepo.AssertExpectations(t)
}

func TestLoginInvalidJSON(t *testing.T) {
	mockRepo := new(MockUserRepository)
	router := setupTestRouter(mockRepo)
	body := []byte(`{invalid json}`)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["error"])
}

func TestLoginUserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"email":"nonexistent@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid credentials")
}

func TestLoginInvalidPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 14)
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}
	mockRepo := new(MockUserRepository)
	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

	router := setupTestRouter(mockRepo)
	body := []byte(`{"email":"test@example.com","password":"wrongpassword"}`)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid credentials")
	mockRepo.AssertExpectations(t)
}

func TestSearchUsersWithQParameter(t *testing.T) {
	users := []model.User{
		{ID: uuid.New(), Username: "testuser1", Email: "test1@example.com"},
		{ID: uuid.New(), Username: "testuser2", Email: "test2@example.com"},
	}
	mockRepo := new(MockUserRepository)
	mockRepo.On("SearchByUsername", "test").Return(users, nil)

	router := setupTestRouter(mockRepo)
	req := httptest.NewRequest("GET", "/search?q=test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 2)
	assert.Equal(t, "testuser1", response[0].Username)
	mockRepo.AssertExpectations(t)
}

func TestSearchUsersWithUsernameParameter(t *testing.T) {
	users := []model.User{
		{ID: uuid.New(), Username: "john", Email: "john@example.com"},
	}
	mockRepo := new(MockUserRepository)
	mockRepo.On("SearchByUsername", "john").Return(users, nil)

	router := setupTestRouter(mockRepo)
	req := httptest.NewRequest("GET", "/search?username=john", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 1)
	assert.Equal(t, "john", response[0].Username)
	mockRepo.AssertExpectations(t)
}

func TestSearchUsersMissingQuery(t *testing.T) {
	mockRepo := new(MockUserRepository)
	router := setupTestRouter(mockRepo)
	req := httptest.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "required")
}

func TestSearchUsersQueryTooShort(t *testing.T) {
	mockRepo := new(MockUserRepository)
	router := setupTestRouter(mockRepo)
	req := httptest.NewRequest("GET", "/search?q=ab", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "не менее 3 символов")
}

func TestSearchUsersEmptyResults(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("SearchByUsername", "nonexistent").Return([]model.User{}, nil)

	router := setupTestRouter(mockRepo)
	req := httptest.NewRequest("GET", "/search?q=nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response []model.User
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 0)
	mockRepo.AssertExpectations(t)
}

func TestSearchUsersServiceError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("SearchByUsername", "test").Return(nil, assert.AnError)

	router := setupTestRouter(mockRepo)
	req := httptest.NewRequest("GET", "/search?q=test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["error"])
	mockRepo.AssertExpectations(t)
}
