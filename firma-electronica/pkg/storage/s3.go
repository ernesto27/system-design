package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type FileInfo struct {
	URL      string
	Key      string
	Size     int64
	Checksum string
}

// Storage defines the interface for storage operations
type Storage interface {
	UploadFile(ctx context.Context, reader io.Reader, fileName string, contentType string) (*FileInfo, error)
}

type Config struct {
	Timeout time.Duration
}

type Service struct {
	config   Config
	provider Storage
}

func New(config Config, provider Storage) *Service {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Service{
		config:   config,
		provider: provider,
	}
}

// UploadFile uploads a file using the configured storage provider
func (s *Service) UploadFile(ctx context.Context, reader io.Reader, fileName string, contentType string) (*FileInfo, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.config.Timeout)
		defer cancel()
	}

	return s.provider.UploadFile(ctx, reader, fileName, contentType)
}

type S3Config struct {
	BucketName      string
	Region          string
	Timeout         time.Duration
	AccessKeyID     string
	SecretAccessKey string
}

// S3Storage implements Storage interface using AWS S3
type S3Storage struct {
	client     *s3.Client
	uploader   *manager.Uploader
	bucketName string
	timeout    time.Duration
}

// NewS3Provider creates a new S3Storage instance
func NewS3Provider(cfg S3Config) (*S3Storage, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	var opts []func(*config.LoadOptions) error

	opts = append(opts, config.WithRegion(cfg.Region))

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     cfg.AccessKeyID,
					SecretAccessKey: cfg.SecretAccessKey,
				}, nil
			}),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)
	uploader := manager.NewUploader(client)

	return &S3Storage{
		client:     client,
		uploader:   uploader,
		bucketName: cfg.BucketName,
		timeout:    cfg.Timeout,
	}, nil
}

// UploadFile uploads a file to S3 and returns information about the uploaded file
func (s *S3Storage) UploadFile(ctx context.Context, reader io.Reader, fileName string, contentType string) (*FileInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ext := filepath.Ext(fileName)
	baseKey := uuid.New().String()
	key := fmt.Sprintf("documents/%s%s", baseKey, ext)

	result, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &FileInfo{
		URL:  result.Location,
		Key:  key,
		Size: 0, // Size is not returned by the uploader, would need to track separately
	}, nil
}

// IsS3URL checks if a URL is an S3 URL
func IsS3URL(url string) bool {
	return strings.Contains(url, "amazonaws.com") || strings.HasPrefix(url, "s3://")
}
