package main

import (
	"context"
	"go-portfolio-api/internal/storage"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load("app.env"); err != nil {
		log.Println("[WARN] Could not load app.env, reading system environment variables")
	}

	r2Config := storage.LoadR2ConfigFromEnv()
	if err := r2Config.Validate(); err != nil {
		log.Fatalf("[ERROR] %v", err)
	}

	storageDir := os.Getenv("STORAGE_BASE_DIR")
	if storageDir == "" {
		storageDir = "storage"
	}

	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		log.Fatalf("[ERROR] Source storage directory '%s' does not exist", storageDir)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 2. Initialize AWS / R2 Client
	s3Client, err := storage.NewS3Client(ctx, r2Config)
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize Cloudflare R2 client: %v", err)
	}

	log.Printf("[INFO] Starting asset migration from '%s/' to Cloudflare R2 bucket '%s'...", storageDir, r2Config.BucketName)

	uploadedCount := 0
	errorCount := 0

	err = filepath.Walk(storageDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		// Calculate relative object key (e.g. storage/projects/test.jpg -> projects/test.jpg)
		relPath, relErr := filepath.Rel(storageDir, path)
		if relErr != nil {
			log.Printf("[WARN] Failed to determine relative path for %s: %v", path, relErr)
			errorCount++
			return nil
		}

		// Normalize to forward slashes for S3/R2 keys
		objectKey := filepath.ToSlash(relPath)

		fileData, readErr := os.Open(path)
		if readErr != nil {
			log.Printf("[ERROR] Failed to open file %s: %v", path, readErr)
			errorCount++
			return nil
		}
		defer fileData.Close()

		// Detect Content-Type
		buffer := make([]byte, 512)
		n, _ := fileData.Read(buffer)
		contentType := http.DetectContentType(buffer[:n])

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
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
		}

		// Rewind file pointer
		if _, seekErr := fileData.Seek(0, 0); seekErr != nil {
			log.Printf("[ERROR] Failed to seek file %s: %v", path, seekErr)
			errorCount++
			return nil
		}

		log.Printf("[SYNCING] Uploading '%s' -> key: '%s' (Type: %s, Size: %d bytes)...", path, objectKey, contentType, info.Size())

		_, putErr := s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:       aws.String(r2Config.BucketName),
			Key:          aws.String(objectKey),
			Body:         fileData,
			ContentType:  aws.String(contentType),
			CacheControl: aws.String("public, max-age=31536000, immutable"),
		})

		if putErr != nil {
			log.Printf("[ERROR] Failed to upload '%s': %v", objectKey, putErr)
			errorCount++
		} else {
			log.Printf("[SUCCESS] Successfully uploaded '%s'", objectKey)
			uploadedCount++
		}

		return nil
	})

	if err != nil {
		log.Fatalf("[ERROR] Walk error: %v", err)
	}

	log.Println("\n================ SYNC SUMMARY ================")
	log.Printf("Total files successfully uploaded: %d", uploadedCount)
	log.Printf("Total errors encountered:          %d", errorCount)
	log.Println("==============================================")
}
