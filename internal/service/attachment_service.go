package service

import (
	"context"
	"io"
	"messenger/internal/model"
	"path/filepath"

	"github.com/google/uuid"
)

type AttachmentRepository interface {
	Create(a *model.Attachment) error
	GetByID(id uuid.UUID) (*model.Attachment, error)
	Delete(id uuid.UUID) error
}
type AttachmentService struct {
	attachmentRepo AttachmentRepository
	storageService Storage
}

func NewAttachmentService(repo AttachmentRepository, storage Storage) *AttachmentService {
	return &AttachmentService{
		attachmentRepo: repo,
		storageService: storage,
	}
}

func (s *AttachmentService) Upload(
	ctx context.Context,
	senderID uuid.UUID,
	chatID uuid.UUID,
	messageID *uuid.UUID,
	file io.Reader,
	filename string,
	size int64,
	mimeType string,
) (*model.Attachment, error) {
	ext := filepath.Ext(filename)
	uniqName := uuid.New().String() + ext

	url, err := s.storageService.Upload(ctx, uniqName, file, size, mimeType)
	if err != nil {
		return nil, err
	}

	attachment := &model.Attachment{
		ChatID:    chatID,
		SenderID:  senderID,
		MessageID: messageID,
		Url:       url,
		Filename:  filename,
		MimeType:  mimeType,
		SizeBytes: size,
	}

	if err := s.attachmentRepo.Create(attachment); err != nil {
		return nil, err
	}

	return attachment, nil

}

func (s *AttachmentService) Download(ctx context.Context, attachmentID uuid.UUID) (io.ReadCloser, int64, string, string, error) {
	attachment, err := s.attachmentRepo.GetByID(attachmentID)
	if err != nil {
		return nil, 0, "", "", err
	}

	// Извлекаем имя объекта из URL
	objectName := filepath.Base(attachment.Url)

	content, size, contentType, err := s.storageService.Download(ctx, objectName)
	if err != nil {
		return nil, 0, "", "", err
	}

	return content, size, contentType, attachment.Filename, nil
}

func (s *AttachmentService) Delete(ctx context.Context, attachmentID uuid.UUID) error {
	attachment, err := s.attachmentRepo.GetByID(attachmentID)
	if err != nil {
		return err
	}

	objectName := filepath.Base(attachment.Url)
	if err := s.storageService.Delete(ctx, objectName); err != nil {
		return err
	}

	return s.attachmentRepo.Delete(attachmentID)
}
