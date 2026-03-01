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
