package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

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

	// 3. Generate Nama File Unik (Timestamp + Nama Asli)
	// Contoh: 173546789_avatar.jpg
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)

	// Bersihkan nama file dari spasi aneh (opsional tapi bagus)
	filename = strings.ReplaceAll(filename, " ", "_")

	// 4. Tentukan lokasi simpan (Folder 'uploads' di root project)
	savePath := filepath.Join("uploads", filename)

	// 5. Simpan File ke Disk
	if err := ctx.SaveUploadedFile(file, savePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file ke server"})
		return
	}

	// 6. Balikkan URL Path supaya bisa disimpan Frontend
	// Format URL: http://localhost:8080/uploads/namafile.jpg
	fileUrl := fmt.Sprintf("/uploads/%s", filename)

	ctx.JSON(http.StatusOK, uploadResponse{
		FileUrl: fileUrl,
	})
}
