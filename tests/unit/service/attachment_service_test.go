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

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Upload(ctx context.Context, objectName string, file io.Reader, size int64, contentType string) (string, error) {
	args := m.Called(ctx, objectName, file, size, contentType)
	return args.String(0), args.Error(1)
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

	result, err := s.Upload(context.Background(), senderID, chatID, file, filename, size, mimeType)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, url, result.Url)
	mockStorage.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
