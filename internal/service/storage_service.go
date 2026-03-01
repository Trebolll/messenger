package service

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	Upload(ctx context.Context, objectName string, file io.Reader, size int64, contentType string) (string, error)
}

type StorageService struct {
	client     *minio.Client
	bucketName string
	endpoint   string
}

func NewStorageService(endpoint, accessKey, secretKey, bucket string) (*StorageService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil || !exists {
		return nil, fmt.Errorf("bucket %s не найден", bucket)
	}

	return &StorageService{
		client:     client,
		bucketName: bucket,
		endpoint:   endpoint,
	}, nil
}

func (s *StorageService) Upload(
	ctx context.Context,
	objectName string,
	file io.Reader,
	size int64,
	contentType string,
) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	url := "http://" + s.endpoint + "/" + s.bucketName + "/" + objectName

	return url, nil
}
