package service

import (
	"context"
	"io"
	"messenger/internal/model"
	"messenger/internal/service"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAttachmentRepository struct {
	mock.Mock
}

func (m *MockAttachmentRepository) Create(a *model.Attachment) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockAttachmentRepository) GetByID(id uuid.UUID) (*model.Attachment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Attachment), args.Error(1)
}

func (m *MockAttachmentRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Upload(ctx context.Context, objectName string, file io.Reader, size int64, contentType string) (string, error) {
	args := m.Called(ctx, objectName, file, size, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) Download(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	args := m.Called(ctx, objectName)
	if args.Get(0) == nil {
		return nil, 0, "", args.Error(3)
	}
	return args.Get(0).(io.ReadCloser), int64(args.Int(1)), args.String(2), args.Error(3)
}

func (m *MockStorage) Delete(ctx context.Context, objectName string) error {
	args := m.Called(ctx, objectName)
	return args.Error(0)
}

func (m *MockStorage) GetURL(objectName string) string {
	args := m.Called(objectName)
	return args.String(0)
}

func TestAttachmentUpload_Success(t *testing.T) {
	mockRepo := new(MockAttachmentRepository)
	mockStorage := new(MockStorage)
	s := service.NewAttachmentService(mockRepo, mockStorage)

	senderID := uuid.New()
	chatID := uuid.New()
	file := strings.NewReader("test data")
	filename := "test.txt"
	size := int64(len("test data"))
	mimeType := "text/plain"
	url := "http://example.com/test.txt"

	mockStorage.On("Upload", mock.Anything, mock.Anything, file, size, mimeType).Return(url, nil)
	mockRepo.On("Create", mock.MatchedBy(func(a *model.Attachment) bool {
		return a.Url == url && a.Filename == filename && a.ChatID == chatID
	})).Return(nil)

	result, err := s.Upload(context.Background(), senderID, chatID, nil, file, filename, size, mimeType)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, url, result.Url)
	mockStorage.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAttachmentDownload_Success(t *testing.T) {
	mockRepo := new(MockAttachmentRepository)
	mockStorage := new(MockStorage)
	s := service.NewAttachmentService(mockRepo, mockStorage)

	attachmentID := uuid.New()
	url := "http://example.com/test.txt"
	filename := "test.txt"
	mimeType := "text/plain"
	size := int64(10)
	content := io.NopCloser(strings.NewReader("some data"))

	mockRepo.On("GetByID", attachmentID).Return(&model.Attachment{
		ID:       attachmentID,
		Url:      url,
		Filename: filename,
	}, nil)

	mockStorage.On("Download", mock.Anything, "test.txt").Return(content, 10, mimeType, nil)

	resContent, resSize, resMime, resFilename, err := s.Download(context.Background(), attachmentID)

	assert.NoError(t, err)
	assert.Equal(t, content, resContent)
	assert.Equal(t, size, resSize)
	assert.Equal(t, mimeType, resMime)
	assert.Equal(t, filename, resFilename)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestAttachmentDelete_Success(t *testing.T) {
	mockRepo := new(MockAttachmentRepository)
	mockStorage := new(MockStorage)
	s := service.NewAttachmentService(mockRepo, mockStorage)

	attachmentID := uuid.New()
	url := "http://example.com/test.txt"

	mockRepo.On("GetByID", attachmentID).Return(&model.Attachment{
		ID:  attachmentID,
		Url: url,
	}, nil)

	mockStorage.On("Delete", mock.Anything, "test.txt").Return(nil)
	mockRepo.On("Delete", attachmentID).Return(nil)

	err := s.Delete(context.Background(), attachmentID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}
