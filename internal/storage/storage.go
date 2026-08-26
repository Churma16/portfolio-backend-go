package storage

import (
	"context"
	"mime/multipart"
)

// StorageService defines the contract for file storage operations
type StorageService interface {
	// UploadFile uploads a file to the storage provider and returns the relative path / key
	UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, targetFolder string) (string, error)
	// DeleteFile removes a file by its relative path / key from the storage provider
	DeleteFile(ctx context.Context, objectKey string) error
	// GetFileURL returns the full public URL for accessing the object
	GetFileURL(objectKey string) string
}
