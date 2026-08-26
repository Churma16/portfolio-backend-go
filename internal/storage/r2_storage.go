package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2StorageService implements StorageService using Cloudflare R2
type R2StorageService struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

// NewR2StorageService creates a new R2StorageService instance using R2Config
func NewR2StorageService(ctx context.Context, r2Config R2Config) (*R2StorageService, error) {
	s3Client, err := NewS3Client(ctx, r2Config)
	if err != nil {
		return nil, err
	}

	return NewR2StorageServiceWithClient(s3Client, r2Config.BucketName, r2Config.PublicURL), nil
}

// NewR2StorageServiceWithClient creates a new R2StorageService instance with an existing S3 client
func NewR2StorageServiceWithClient(client *s3.Client, bucketName, publicURL string) *R2StorageService {
	return &R2StorageService{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}
}

// UploadFile streams the uploaded file directly to Cloudflare R2
func (r *R2StorageService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, targetFolder string) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Detect content type from first 512 bytes
	headerBuffer := make([]byte, 512)
	bytesRead, _ := src.Read(headerBuffer)
	contentType := http.DetectContentType(headerBuffer[:bytesRead])

	// Explicit override for common extensions
	fileExtension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	switch fileExtension {
	case ".pdf":
		contentType = "application/pdf"
	case ".svg":
		contentType = "image/svg+xml"
	case ".webp":
		contentType = "image/webp"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	}

	// Reset stream to beginning
	if _, err := src.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to seek stream: %w", err)
	}

	uniqueFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename)
	uniqueFilename = strings.ReplaceAll(uniqueFilename, " ", "_")

	objectKey := fmt.Sprintf("%s/%s", strings.Trim(targetFolder, "/"), uniqueFilename)

	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(r.bucketName),
		Key:          aws.String(objectKey),
		Body:         src,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object to Cloudflare R2: %w", err)
	}

	return objectKey, nil
}

// DeleteFile deletes an object by its key from Cloudflare R2
func (r *R2StorageService) DeleteFile(ctx context.Context, objectKey string) error {
	if objectKey == "" {
		return nil
	}

	cleanedKey := strings.TrimPrefix(objectKey, "/")
	if strings.HasPrefix(cleanedKey, "http://") || strings.HasPrefix(cleanedKey, "https://") {
		if r.publicURL != "" && strings.HasPrefix(cleanedKey, r.publicURL) {
			cleanedKey = strings.TrimPrefix(cleanedKey, r.publicURL)
			cleanedKey = strings.TrimPrefix(cleanedKey, "/")
		}
	}

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(cleanedKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from Cloudflare R2: %w", err)
	}

	return nil
}

// GetFileURL returns the full public CDN URL for the object
func (r *R2StorageService) GetFileURL(objectKey string) string {
	if objectKey == "" {
		return ""
	}
	if strings.HasPrefix(objectKey, "http://") || strings.HasPrefix(objectKey, "https://") {
		return objectKey
	}
	base := strings.TrimRight(r.publicURL, "/")
	key := strings.TrimLeft(objectKey, "/")
	return fmt.Sprintf("%s/%s", base, key)
}
