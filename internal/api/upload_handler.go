package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// uploadResponse untuk balikan JSON setelah upload sukses
type uploadResponse struct {
	FileUrl string `json:"file_url"`
}

func (server *Server) uploadFile(ctx *gin.Context) {
	// 1. Ambil file dari form-data dengan key "file"
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File tidak ditemukan, gunakan key 'file'"})
		return
	}

	// 2. Validasi Ekstensi (Security dasar)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, // Untuk Avatar
		".pdf": true, // Untuk CV
	}

	if !validExtensions[ext] {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung (hanya jpg, png, pdf)"})
		return
	}

	// 3. Simpan File via StorageService
	objectKey, err := server.storage.UploadFile(ctx.Request.Context(), file, "uploads")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file: " + err.Error()})
		return
	}

	// 4. Balikkan URL Path / Key
	fileUrl := fmt.Sprintf("/%s", strings.TrimPrefix(objectKey, "/"))

	ctx.JSON(http.StatusOK, uploadResponse{
		FileUrl: fileUrl,
	})
}
