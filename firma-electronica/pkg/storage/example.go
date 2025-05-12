package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Example demonstrates how to use the storage service with S3
func Example() {
	// Configure the S3 storage provider
	s3Config := S3Config{
		BucketName:      "your-bucket-name",
		Region:          "us-east-1",
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Timeout:         time.Second * 30,
	}

	// Create a new S3 provider
	s3Provider, err := NewS3Provider(s3Config)
	if err != nil {
		fmt.Printf("Failed to create S3 provider: %v\n", err)
		return
	}

	// Create a storage service with the S3 provider
	storageService := New(Config{
		Timeout: time.Second * 30,
	}, s3Provider)

	// Create a sample file to upload
	fileContent := strings.NewReader("This is a test file content")
	fileName := "test-document.txt"
	contentType := "text/plain"

	// Upload the file
	ctx := context.Background()
	fileInfo, err := storageService.UploadFile(ctx, fileContent, fileName, contentType)
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		return
	}

	fmt.Printf("File uploaded successfully:\n")
	fmt.Printf("  URL: %s\n", fileInfo.URL)
	fmt.Printf("  Key: %s\n", fileInfo.Key)
}

// ExampleUploadDocument demonstrates how to upload a document file
func ExampleUploadDocument(filePath string) (*FileInfo, error) {
	// Configure the S3 storage provider
	s3Config := S3Config{
		BucketName:      "your-document-bucket",
		Region:          "us-east-1",
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Timeout:         time.Second * 30,
	}

	// Create a new S3 provider
	s3Provider, err := NewS3Provider(s3Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 provider: %w", err)
	}

	// Create a storage service with the S3 provider
	storageService := New(Config{
		Timeout: time.Second * 30,
	}, s3Provider)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Determine content type based on file extension
	contentType := "application/octet-stream"
	if strings.HasSuffix(filePath, ".pdf") {
		contentType = "application/pdf"
	} else if strings.HasSuffix(filePath, ".doc") || strings.HasSuffix(filePath, ".docx") {
		contentType = "application/msword"
	}

	// Upload the file
	ctx := context.Background()
	return storageService.UploadFile(ctx, file, filePath, contentType)
}

// ExampleLocalStorage shows a basic implementation of a local storage provider
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates a new local storage provider
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &LocalStorage{
		baseDir: baseDir,
	}, nil
}

// UploadFile implements the Storage interface for local file storage
func (l *LocalStorage) UploadFile(ctx context.Context, reader io.Reader, fileName string, contentType string) (*FileInfo, error) {
	// Generate a unique filename with original extension
	ext := filepath.Ext(fileName)
	baseKey := uuid.New().String()
	key := fmt.Sprintf("%s%s", baseKey, ext)
	filePath := filepath.Join(l.baseDir, key)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy the content to the file
	size, err := io.Copy(file, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Return file information
	return &FileInfo{
		URL:  fmt.Sprintf("file://%s", filePath),
		Key:  key,
		Size: size,
	}, nil
}

// ExampleWithLocalStorage demonstrates how to use a local storage provider
func ExampleWithLocalStorage() {
	// Create a local storage provider
	localStorage, err := NewLocalStorage("/tmp/firma-electronica")
	if err != nil {
		fmt.Printf("Failed to create local storage: %v\n", err)
		return
	}

	// Create a storage service with the local provider
	storageService := New(Config{
		Timeout: time.Second * 10,
	}, localStorage)

	// Create a sample file to upload
	fileContent := strings.NewReader("This is a test file content")
	fileName := "test-document.txt"
	contentType := "text/plain"

	// Upload the file
	ctx := context.Background()
	fileInfo, err := storageService.UploadFile(ctx, fileContent, fileName, contentType)
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		return
	}

	fmt.Printf("File uploaded successfully to local storage:\n")
	fmt.Printf("  URL: %s\n", fileInfo.URL)
	fmt.Printf("  Key: %s\n", fileInfo.Key)
	fmt.Printf("  Size: %d bytes\n", fileInfo.Size)
}
