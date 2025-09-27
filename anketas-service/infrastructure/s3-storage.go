package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	// "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	// "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
}

func NewS3Storage() (*S3Storage, error) {
	// Загружаем конфигурацию из .aws/credentials и .aws/config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ru-central1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: "https://storage.yandexcloud.net",
				}, nil
			},
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить конфигурацию AWS: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		client:     client,
		bucketName: "u2-bucket-new",
	}, nil
}

// Генерирует presigned URL для загрузки фотографии
func (s *S3Storage) GenerateUploadURL(ctx context.Context, userID string) (string, error) {
	photoID := uuid.New().String()
	key := fmt.Sprintf("photos/%s/%s.jpg", userID, photoID)

	// Создаем presigned URL для PUT запроса
	presignClient := s3.NewPresignClient(s.client)
	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(15 * time.Minute) // URL действителен 15 минут
	})

	if err != nil {
		return "", fmt.Errorf("не удалось создать presigned URL: %v", err)
	}

	log.Printf("Создан presigned URL для загрузки: %s", key)
	return request.URL, nil
}

// Получает URL для просмотра фотографии
func (s *S3Storage) GetPhotoURL(ctx context.Context, key string) (string, error) {
	// Создаем presigned URL для GET запроса
	presignClient := s3.NewPresignClient(s.client)
	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(24 * time.Hour) // URL действителен 24 часа
	})

	if err != nil {
		return "", fmt.Errorf("не удалось создать URL для просмотра: %v", err)
	}

	return request.URL, nil
}


// Удаляет фотографию из S3
func (s *S3Storage) DeletePhoto(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("не удалось удалить фотографию: %v", err)
	}

	log.Printf("Фотография удалена: %s", key)
	return nil
}

// Проверяет существование bucket
func (s *S3Storage) CheckBucketExists(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})

	if err != nil {
		return fmt.Errorf("bucket не найден или недоступен: %v", err)
	}

	log.Printf("Bucket %s доступен", s.bucketName)
	return nil
}
