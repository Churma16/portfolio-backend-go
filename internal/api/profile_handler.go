package api

import (
	"encoding/json"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
	// 1. Bind data - using FormMultipart for multipart/form-data
	var req createProfileRequest

	if err := ctx.ShouldBindWith(&req, binding.FormMultipart); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID := ctx.MustGet("user_id").(int64)

	// 2. Handle Upload AVATAR (SETELAH Bind)
	var avatarUrl string
	fileAvatar, errAvatar := ctx.FormFile("avatar")
	if errAvatar == nil { // File ada
		url, errSave := util.SaveUploadedFile(ctx, fileAvatar, "avatar")
		if errSave != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload avatar: " + errSave.Error()})
			return
		}
		avatarUrl = url
	} else {
	}

	// 3. Handle Upload CV (SETELAH Bind)
	var cvUrl string
	fileCV, errCV := ctx.FormFile("cv_files")
	if errCV == nil { // File ada
		url, errSave := util.SaveUploadedFile(ctx, fileCV, "cv_files")
		if errSave != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload CV: " + errSave.Error()})
			return
		}
		cvUrl = url
	} else {
	}

	// 4. Handle Socials (String JSON -> RawMessage)
	var socialsRaw json.RawMessage
	if req.Socials != "" {
		socialsRaw = json.RawMessage(req.Socials)
	} else {
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
		Avatar:         convertToNullString(avatarUrl),
		CvFiles:        convertToNullString(cvUrl),
		HeroImageCodes: convertToNullString(req.HeroImageCodes),
		Socials:        socialsRaw,
	}

	profile, err := server.store.CreateProfile(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//	Sukses
	server.redisClient.Del(ctx, "site_profile")
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

func (server *Server) getProfiles(ctx *gin.Context) {
	// 1. Ambil User ID.
	// Kita bisa ambil dari Token (kalau rute Private /me)
	// ATAU ambil dari URL parameter (kalau rute Public /:user_id)

	// Skenario: PUBLIC ACCESS (via URL param id)
	// Contoh: GET /profile/1

	// 2. Panggil Database
	profile, err := server.store.GetFirstProfile(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Profile tidak ditemukan"})
		return
	}

	// 3. Mapping ke JSON & Return
	meta := response.NewMeta(http.StatusOK, "success", "Profile berhasil diambil")
	data := newProfileResponse(profile)
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, data))
}

func (server *Server) updateProfile(ctx *gin.Context) {

	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid tech stack ID"))
		return
	}

	// 1. Bind data form - using FormMultipart for multipart/form-data
	var req createProfileRequest
	if err := ctx.ShouldBindWith(&req, binding.FormMultipart); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID := ctx.MustGet("user_id").(int64)

	// 2. AMBIL DATA LAMA (PENTING!)
	oldProfile, err := server.store.GetProfileById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Profile belum dibuat, silahkan Create dulu"})
		return
	}

	// 3. Handle Upload AVATAR (SETELAH Bind)
	finalAvatar := oldProfile.Avatar.String // Default: Pakai yang lama
	fileAvatar, errAvatar := ctx.FormFile("avatar")
	if errAvatar == nil {
		// User upload file baru -> Delete old file first
		if oldProfile.Avatar.Valid {
			if err := util.DeleteFile(oldProfile.Avatar.String); err != nil {
				ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
				return
			}
		}
		url, errSave := util.SaveUploadedFile(ctx, fileAvatar, "avatar")
		if errSave != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload avatar baru: " + errSave.Error()})
			return
		}
		finalAvatar = url
	}

	// 4. Handle Upload CV (SETELAH Bind)
	finalCV := oldProfile.CvFiles.String // Default: Pakai yang lama
	fileCV, errCV := ctx.FormFile("cv_files")
	if errCV == nil {
		// User upload file baru -> Delete old file first
		if oldProfile.CvFiles.Valid {
			if err := util.DeleteFile(oldProfile.CvFiles.String); err != nil {
				ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
				return
			}
		}
		url, errSave := util.SaveUploadedFile(ctx, fileCV, "cv_files")
		if errSave != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload CV baru: " + errSave.Error()})
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
		socialsRaw = oldProfile.Socials
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
		Socials:        socialsRaw,
	}

	updatedProfile, err := server.store.UpdateProfile(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	server.redisClient.Del(ctx, "site_profile")
	rsp := newProfileResponse(updatedProfile)
	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) deleteProfile(ctx *gin.Context) {
	// 1. Ambil ID User dari Token (Wajib Login!)
	userID := ctx.MustGet("user_id").(int64)

	// 2. Eksekusi Hapus
	_, err := server.store.DeleteProfile(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	// 3. Response Sukses
	server.redisClient.Del(ctx, "site_profile")
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
		Socials:        profile.Socials,
		CreatedAt:      profile.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      profile.UpdatedAt.Time.Format(time.RFC3339),
	}
}
