package util

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SaveUploadedFile saves an uploaded file to the specified folder and returns the file path.
func SaveUploadedFile(context *gin.Context, uploadedFile *multipart.FileHeader, targetFolder string) (string, error) {
	// Generate a unique filename using the current timestamp and the original file name
	uniqueFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), uploadedFile.Filename)
	uniqueFilename = strings.ReplaceAll(uniqueFilename, " ", "_")

	// Construct the full save path for the file
	savePath := filepath.Join("storage/"+targetFolder, uniqueFilename)

	// Save the uploaded file to the specified path
	if err := context.SaveUploadedFile(uploadedFile, savePath); err != nil {
		return "", err
	}

	// Return the relative file path for further use
	return fmt.Sprintf("/%s/%s", targetFolder, uniqueFilename), nil
}
