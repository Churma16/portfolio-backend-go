package storage

import (
	"context"
	"log"
	"os"
	"strings"
)

// InitStorageService initializes the storage driver based on STORAGE_DRIVER env (r2 or local)
func InitStorageService(ctx context.Context) StorageService {
	driver := strings.ToLower(os.Getenv("STORAGE_DRIVER"))

	if driver == "r2" {
		r2Config := LoadR2ConfigFromEnv()
		r2Service, err := NewR2StorageService(ctx, r2Config)
		if err != nil {
			log.Printf("[WARN] Failed to initialize Cloudflare R2 storage: %v. Falling back to local storage.", err)
		} else {
			log.Printf("[INFO] Initialized Cloudflare R2 storage provider (Bucket: %s)", r2Config.BucketName)
			return r2Service
		}
	}

	baseDir := os.Getenv("STORAGE_BASE_DIR")
	if baseDir == "" {
		baseDir = "storage"
	}
	publicURL := os.Getenv("STORAGE_PUBLIC_URL")
	if publicURL == "" {
		publicURL = "/storage"
	}

	log.Printf("[INFO] Initialized Local storage provider (BaseDir: %s)", baseDir)
	return NewLocalStorageService(baseDir, publicURL)
}
