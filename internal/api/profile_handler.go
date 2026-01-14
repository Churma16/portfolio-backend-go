package api

import (
	"encoding/json"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sqlc-dev/pqtype"
)

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
		url, errSave := util.SaveUploadedFile(ctx, fileAvatar, folderNameAvatar)
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
		url, errSave := util.SaveUploadedFile(ctx, fileCV, folderNameCv)
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

	//	Sukses
	rsp := newProfileResponse(profile)
	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) getProfile(ctx *gin.Context) {
	// 1. Ambil User ID.
	// Kita bisa ambil dari Token (kalau rute Private /me)
	// ATAU ambil dari URL parameter (kalau rute Public /:user_id)

	// Skenario: PUBLIC ACCESS (via URL param id)
	// Contoh: GET /profile/1
	var req struct {
		ID int64 `uri:"user_id" binding:"required,min=1"`
	}

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Panggil Database
	profile, err := server.store.GetProfileByUserId(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Profile tidak ditemukan"})
		return
	}

	// 3. Mapping ke JSON & Return
	ctx.JSON(http.StatusOK, newProfileResponse(profile))
}

func (server *Server) updateProfile(ctx *gin.Context) {
	var req createProfileRequest // Kita pakai struct yang sama dengan Create (Reuse)

	// 1. Bind Data Form
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.MustGet("user_id").(int64)

	// 2. AMBIL DATA LAMA (PENTING!)
	// Kita butuh ini untuk tahu "Avatar Lama" kalau user nggak upload baru
	oldProfile, err := server.store.GetProfileByUserId(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Profile belum dibuat, silahkan Create dulu"})
		return
	}

	// 3. LOGIKA AVATAR (Ganti atau Tetap?)
	finalAvatar := oldProfile.Avatar.String // Default: Pakai yang lama
	fileAvatar, err := ctx.FormFile("avatar")
	folderNameAvatar := "avatar"
	if err == nil {
		// User upload file baru -> Ganti!
		url, errSave := util.SaveUploadedFile(ctx, fileAvatar, folderNameAvatar)
		if errSave != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload avatar baru"})
			return
		}
		finalAvatar = url
	}

	// 4. LOGIKA CV (Ganti atau Tetap?)
	finalCV := oldProfile.CvFiles.String // Default: Pakai yang lama
	fileCV, err := ctx.FormFile("cv_files")
	folderNameCv := "cv_files"
	if err == nil {
		url, errSave := util.SaveUploadedFile(ctx, fileCV, folderNameCv)
		if errSave != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload CV baru"})
			return
		}
		finalCV = url
	}

	// 5. Handle Socials (Sama kayak Create)
	var socialsRaw json.RawMessage
	if req.Socials != "" {
		socialsRaw = json.RawMessage(req.Socials)
	} else {
		// Kalau user kirim string kosong, kita pakai socials yang lama
		socialsRaw = oldProfile.Socials.RawMessage
	}

	// 6. Eksekusi Update ke DB
	arg := db.UpdateProfileParams{
		UserID:         userID,
		Name:           req.Name,
		Headline:       convertToNullString(req.Headline),
		Role:           convertToNullString(req.Role),
		BioShort:       convertToNullString(req.BioShort),
		BioLong:        convertToNullString(req.BioLong),
		Location:       convertToNullString(req.Location),
		IsHireable:     convertToNullBool(req.IsHireable),
		Avatar:         convertToNullString(finalAvatar), // <-- Pakai variabel final
		CvFiles:        convertToNullString(finalCV),     // <-- Pakai variabel final
		HeroImageCodes: convertToNullString(req.HeroImageCodes),
		Socials:        pqtype.NullRawMessage{RawMessage: socialsRaw, Valid: len(socialsRaw) > 0},
	}

	updatedProfile, err := server.store.UpdateProfile(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rsp := newProfileResponse(updatedProfile)
	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) deleteProfile(ctx *gin.Context) {
	// 1. Ambil ID User dari Token (Wajib Login!)
	userID := ctx.MustGet("user_id").(int64)

	// 2. Eksekusi Hapus
	_, err := server.store.DeleteProfile(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Response Sukses
	ctx.JSON(http.StatusOK, gin.H{"message": "Profile berhasil dihapus"})
}

func newProfileResponse(profile db.Profile) profileResponse {
	return profileResponse{
		ID:             profile.ID,
		UserID:         profile.UserID,
		Name:           profile.Name,
		Headline:       profile.Headline.String,
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
}
