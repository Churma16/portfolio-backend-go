package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalStorageService implements StorageService for local disk storage
type LocalStorageService struct {
	baseDir   string
	publicURL string
}

// NewLocalStorageService creates a new instance of LocalStorageService
func NewLocalStorageService(baseDir string, publicURL string) *LocalStorageService {
	if baseDir == "" {
		baseDir = "storage"
	}
	if publicURL == "" {
		publicURL = "/storage"
	}
	return &LocalStorageService{
		baseDir:   baseDir,
		publicURL: publicURL,
	}
}

// UploadFile saves a multipart file to local disk and returns the relative path
func (l *LocalStorageService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, targetFolder string) (string, error) {
	targetDir := filepath.Join(l.baseDir, targetFolder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target directory: %w", err)
	}

	uniqueFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename)
	uniqueFilename = strings.ReplaceAll(uniqueFilename, " ", "_")

	savePath := filepath.Join(targetDir, uniqueFilename)

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file contents: %w", err)
	}

	relativePath := fmt.Sprintf("%s/%s", strings.Trim(targetFolder, "/"), uniqueFilename)
	return relativePath, nil
}

// DeleteFile removes a file from local disk
func (l *LocalStorageService) DeleteFile(ctx context.Context, objectKey string) error {
	if objectKey == "" {
		return nil
	}

	cleanedKey := filepath.Clean(objectKey)
	if strings.HasPrefix(cleanedKey, "..") {
		return fmt.Errorf("invalid file path: %s", objectKey)
	}

	fullPath := filepath.Join(l.baseDir, cleanedKey)
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete local file: %w", err)
	}
	return nil
}

// GetFileURL returns the full URL to the file
func (l *LocalStorageService) GetFileURL(objectKey string) string {
	if objectKey == "" {
		return ""
	}
	if strings.HasPrefix(objectKey, "http://") || strings.HasPrefix(objectKey, "https://") {
		return objectKey
	}
	base := strings.TrimRight(l.publicURL, "/")
	key := strings.TrimLeft(objectKey, "/")
	return fmt.Sprintf("%s/%s", base, key)
}
