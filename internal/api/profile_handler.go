package api

import (
	"encoding/json"
	"fmt"
	db "go-portfolio-api/db/sqlc"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sqlc-dev/pqtype"
)

// Helper internal buat simpan file
func saveUploadedFile(ctx *gin.Context, fileHeader *multipart.FileHeader, folderName string) (string, error) {
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	savePath := filepath.Join("storage/"+folderName, filename)

	if err := ctx.SaveUploadedFile(fileHeader, savePath); err != nil {
		return "", err
	}
	return fmt.Sprintf("/%s/%s", folderName, filename), nil
}

// Struct request pakai tag 'form' bukan 'json'
type createProfileRequest struct {
	Name           string `form:"name" binding:"required"`
	Headline       string `form:"headline"`
	Role           string `form:"role"`
	BioShort       string `form:"bio_short"`
	BioLong        string `form:"bio_long"`
	Location       string `form:"location"`
	IsHireable     bool   `form:"is_hireable"`
	HeroImageCodes string `form:"hero_image_codes"`

	// Socials kita terima sebagai STRING JSON, nanti kita convert manual
	Socials string `form:"socials"`
}

type profileResponse struct {
	ID             int64           `json:"id"`
	UserID         int64           `json:"user_id"`
	Name           string          `json:"name"`
	Headline       string          `json:"headline"`
	Role           string          `json:"role"`
	BioShort       string          `json:"bio_short"`
	BioLong        string          `json:"bio_long"`
	Location       string          `json:"location"`
	IsHireable     bool            `json:"is_hireable"`
	Avatar         string          `json:"avatar"`
	CvFiles        string          `json:"cv_files"`
	HeroImageCodes string          `json:"hero_image_codes"`
	Socials        json.RawMessage `json:"socials"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

func (server *Server) createProfile(ctx *gin.Context) {
	var req createProfileRequest

	// 1. Baca data Text (Form)
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.MustGet("user_id").(int64)

	// 2. Handle Upload AVATAR (Langsung disini)
	var avatarUrl string
	fileAvatar, err := ctx.FormFile("avatar") // Ambil file dari key 'avatar'
	folderNameAvatar := "avatar"
	if err == nil { // Kalau user upload file
		url, errSave := saveUploadedFile(ctx, fileAvatar, folderNameAvatar)
		if errSave != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload avatar"})
			return
		}
		avatarUrl = url
	}

	// 3. Handle Upload CV (Langsung disini)
	var cvUrl string
	fileCV, err := ctx.FormFile("cv_files") // Ambil file dari key 'cv_files'
	folderNameCv := "cv_files"

	if err == nil {
		url, errSave := saveUploadedFile(ctx, fileCV, folderNameCv)
		if errSave != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload CV"})
			return
		}
		cvUrl = url
	}

	// 4. Handle Socials (String JSON -> RawMessage)
	// Kita ubah string "{"github": "..."}" menjadi RawMessage
	var socialsRaw json.RawMessage
	if req.Socials != "" {
		socialsRaw = json.RawMessage(req.Socials)
	}

	// 5. Masukkan ke DB
	arg := db.CreateProfileParams{
		UserID:         userID,
		Name:           req.Name,
		Headline:       convertToNullString(req.Headline),
		Role:           convertToNullString(req.Role),
		BioShort:       convertToNullString(req.BioShort),
		BioLong:        convertToNullString(req.BioLong),
		Location:       convertToNullString(req.Location),
		IsHireable:     convertToNullBool(req.IsHireable),
		Avatar:         convertToNullString(avatarUrl), // Pakai URL hasil upload tadi
		CvFiles:        convertToNullString(cvUrl),     // Pakai URL hasil upload tadi
		HeroImageCodes: convertToNullString(req.HeroImageCodes),
		Socials:        pqtype.NullRawMessage{RawMessage: socialsRaw, Valid: len(socialsRaw) > 0},
	}

	profile, err := server.store.CreateProfile(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rsp := profileResponse{
		ID:             profile.ID,
		UserID:         profile.UserID,
		Name:           profile.Name,
		Headline:       profile.Headline.String, // Ambil String-nya saja
		Role:           profile.Role.String,
		BioShort:       profile.BioShort.String,
		BioLong:        profile.BioLong.String,
		Location:       profile.Location.String,
		IsHireable:     profile.IsHireable.Bool,
		Avatar:         profile.Avatar.String,
		CvFiles:        profile.CvFiles.String,
		HeroImageCodes: profile.HeroImageCodes.String,
		Socials:        profile.Socials.RawMessage,
		CreatedAt:      profile.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      profile.UpdatedAt.Format(time.RFC3339),
	}
	//	Sukses
	ctx.JSON(http.StatusOK, rsp)
}
