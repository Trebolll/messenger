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
	Download(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error)
	Delete(ctx context.Context, objectName string) error
	GetURL(objectName string) string
}

type StorageService struct {
	client         *minio.Client
	bucketName     string
	endpoint       string
	publicEndpoint string
	useSSL         bool
}

// NewStorageService принимает внутренний endpoint для подключения к MinIO
// и publicEndpoint для генерации публичных URL (например, ngrok или внешний домен).
// Если publicEndpoint пустой — используется endpoint.
func NewStorageService(endpoint, accessKey, secretKey, bucket, publicEndpoint string) (*StorageService, error) {
	useSSL := false // Можно сделать настраиваемым, если нужно
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки бакета: %w", err)
	}
	if !exists {
		err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("не удалось создать бакет %s: %w", bucket, err)
		}
		policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Action":["s3:GetObject"],"Effect":"Allow","Principal":{"AWS":["*"]},"Resource":["arn:aws:s3:::%s/*"],"Sid":""}]}`, bucket)
		_ = client.SetBucketPolicy(context.Background(), bucket, policy)
	}

	pub := publicEndpoint
	if pub == "" {
		pub = endpoint
	}

	return &StorageService{
		client:         client,
		bucketName:     bucket,
		endpoint:       endpoint,
		publicEndpoint: pub,
		useSSL:         useSSL,
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

	return s.GetURL(objectName), nil
}

func (s *StorageService) Download(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, "", err
	}

	info, err := object.Stat()
	if err != nil {
		return nil, 0, "", err
	}

	return object, info.Size, info.ContentType, nil
}

func (s *StorageService) Delete(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
}

func (s *StorageService) GetURL(objectName string) string {
	schema := "http://"
	if s.useSSL || (s.publicEndpoint != "localhost" && s.publicEndpoint != "127.0.0.1") {
		schema = "https://"
	}

	return schema + s.publicEndpoint + "/" + s.bucketName + "/" + objectName
}
