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

// S3Config contains configuration for the S3 storage
type S3Config struct {
	BucketName      string
	Region          string
	Timeout         time.Duration
	AccessKeyID     string
	SecretAccessKey string
}

// S3Storage implements storage operations using AWS S3
type S3Storage struct {
	client     *s3.Client
	uploader   *manager.Uploader
	bucketName string
	timeout    time.Duration
}

// FileInfo contains information about a stored file
type FileInfo struct {
	URL      string
	Key      string
	Size     int64
	Checksum string
}

// New creates a new S3Storage instance
func New(cfg S3Config) (*S3Storage, error) {
	// Set default timeout if not provided
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Create AWS credential provider with static credentials
	var opts []func(*config.LoadOptions) error

	// Set region
	opts = append(opts, config.WithRegion(cfg.Region))

	// Add static credentials if provided
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

	// Load AWS configuration with provided options
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Create S3 client
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
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Generate a unique file key with original extension
	ext := filepath.Ext(fileName)
	baseKey := uuid.New().String()
	key := fmt.Sprintf("documents/%s%s", baseKey, ext)

	// Upload file to S3
	result, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Return file information
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
