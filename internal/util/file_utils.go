package util

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SaveUploadedFile saves an uploaded file to the specified folder and returns the file path.
func SaveUploadedFile(context *gin.Context, uploadedFile *multipart.FileHeader, targetFolder string) (string, error) {
	println("\n========== FILE UPLOAD DEBUG ==========")
	println("TARGET FOLDER:", targetFolder)
	println("ORIGINAL FILENAME:", uploadedFile.Filename)
	println("FILE SIZE (bytes):", uploadedFile.Size)

	// Generate a unique filename using the current timestamp and the original file name
	uniqueFilename := fmt.Sprintf("%d_%s", time.Now().Unix(), uploadedFile.Filename)
	uniqueFilename = strings.ReplaceAll(uniqueFilename, " ", "_")
	println("UNIQUE FILENAME:", uniqueFilename)

	// Construct the full save path for the file
	savePath := filepath.Join("storage/"+targetFolder, uniqueFilename)
	println("SAVE PATH:", savePath)

	// Save the uploaded file to the specified path
	println("ATTEMPTING TO SAVE FILE...")
	if err := context.SaveUploadedFile(uploadedFile, savePath); err != nil {
		println("ERROR SAVING FILE:", err.Error())
		println("=======================================\n")
		return "", err
	}

	println("FILE SAVED SUCCESSFULLY!")

	// Return the relative file path for further use
	relativePath := fmt.Sprintf("%s/%s", targetFolder, uniqueFilename)
	println("RETURNING RELATIVE PATH:", relativePath)
	println("=======================================\n")
	return relativePath, nil
}

// DeleteFile removes a file from the filesystem.
func DeleteFile(filePath string) error {
	return os.Remove("./storage/" + filePath)
}
