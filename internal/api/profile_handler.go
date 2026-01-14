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
	println("\n\n========== CREATE PROFILE DEBUG START ==========")
	println("STEP 1: Checking Content-Type")
	contentType := ctx.Request.Header.Get("Content-Type")
	println("  Content-Type:", contentType)

	// DEBUG: Lihat ALL form fields yang diterima sebelum binding
	println("STEP 1.5: Inspecting RAW request body")
	err := ctx.Request.ParseMultipartForm(32 << 20) // 32MB
	if err == nil && ctx.Request.MultipartForm != nil {
		println("  Form Fields received:")
		for key, values := range ctx.Request.Form {
			println("    -", key, "=", values[0])
		}
		println("  File Fields received:")
		for key, fileHeaders := range ctx.Request.MultipartForm.File {
			println("    -", key, "| Count:", len(fileHeaders))
			for i, fh := range fileHeaders {
				println("      [", i, "]", fh.Filename, "|", fh.Size, "bytes")
			}
		}
	} else if err != nil {
		println("  WARNING: ParseMultipartForm error:", err.Error())
	}

	// 1. Bind data - akan auto-detect berdasarkan Content-Type
	var req createProfileRequest

	println("STEP 2: Attempting to ShouldBind()")
	if err := ctx.ShouldBind(&req); err != nil {
		println("  ERROR BINDING:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"debug": gin.H{
				"content_type": contentType,
				"message":      "Pastikan Content-Type sesuai (multipart/form-data untuk file upload, atau application/json untuk JSON)",
			},
		})
		println("========== CREATE PROFILE DEBUG END (ERROR BINDING) ==========\n\n")
		return
	}
	println("  ✓ ShouldBind SUCCESS")
	println("  Parsed name:", req.Name)
	println("  Parsed headline:", req.Headline)

	userID := ctx.MustGet("user_id").(int64)
	println("STEP 3: Got user_id:", userID)

	// 2. Handle Upload AVATAR (SETELAH Bind)
	println("STEP 4: Attempting to get avatar file")
	var avatarUrl string
	fileAvatar, errAvatar := ctx.FormFile("avatar")
	if errAvatar == nil { // File ada
		println("  ✓ Avatar file found!")
		println("  Avatar filename:", fileAvatar.Filename)
		println("  Avatar size:", fileAvatar.Size)

		println("  Calling SaveUploadedFile for avatar...")
		url, errSave := util.SaveUploadedFile(ctx, fileAvatar, "avatar")
		if errSave != nil {
			println("  ERROR saving avatar:", errSave.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload avatar: " + errSave.Error()})
			println("========== CREATE PROFILE DEBUG END (AVATAR ERROR) ==========\n\n")
			return
		}
		avatarUrl = url
		println("  ✓ Avatar saved with URL:", avatarUrl)
	} else {
		println("  ✗ Avatar NOT found (error or missing):", errAvatar.Error())
	}

	// 3. Handle Upload CV (SETELAH Bind)
	println("STEP 5: Attempting to get cv_files file")
	var cvUrl string
	fileCV, errCV := ctx.FormFile("cv_files")
	if errCV == nil { // File ada
		println("  ✓ CV file found!")
		println("  CV filename:", fileCV.Filename)
		println("  CV size:", fileCV.Size)

		println("  Calling SaveUploadedFile for cv...")
		url, errSave := util.SaveUploadedFile(ctx, fileCV, "cv_files")
		if errSave != nil {
			println("  ERROR saving CV:", errSave.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload CV: " + errSave.Error()})
			println("========== CREATE PROFILE DEBUG END (CV ERROR) ==========\n\n")
			return
		}
		cvUrl = url
		println("  ✓ CV saved with URL:", cvUrl)
	} else {
		println("  ✗ CV NOT found (error or missing):", errCV.Error())
	}

	// 4. Handle Socials (String JSON -> RawMessage)
	println("STEP 6: Processing socials")
	var socialsRaw json.RawMessage
	if req.Socials != "" {
		socialsRaw = json.RawMessage(req.Socials)
		println("  ✓ Socials found:", string(socialsRaw))
	} else {
		println("  ✗ Socials empty")
	}

	// 5. Masukkan ke DB
	println("STEP 7: Preparing database insert params")
	println("  avatarUrl to insert:", avatarUrl)
	println("  cvUrl to insert:", cvUrl)

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
		Socials:        pqtype.NullRawMessage{RawMessage: socialsRaw, Valid: len(socialsRaw) > 0},
	}

	println("STEP 8: Calling CreateProfile database")
	profile, err := server.store.CreateProfile(ctx, arg)
	if err != nil {
		println("  ERROR creating profile:", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		println("========== CREATE PROFILE DEBUG END (DB ERROR) ==========\n\n")
		return
	}

	println("  ✓ Profile created successfully with ID:", profile.ID)
	println("  Profile avatar from DB:", profile.Avatar.String)
	println("  Profile cv_files from DB:", profile.CvFiles.String)

	//	Sukses
	rsp := newProfileResponse(profile)
	println("STEP 9: Sending response")
	println("========== CREATE PROFILE DEBUG END (SUCCESS) ==========\n\n")
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
	// 1. Bind data form - auto-detect Content-Type
	var req createProfileRequest
	if err := ctx.ShouldBind(&req); err != nil {
		// Debug: Log content-type yang diterima
		contentType := ctx.Request.Header.Get("Content-Type")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"debug": gin.H{
				"content_type": contentType,
				"message":      "Pastikan Content-Type sesuai (multipart/form-data untuk file upload)",
			},
		})
		return
	}

	userID := ctx.MustGet("user_id").(int64)

	// 2. AMBIL DATA LAMA (PENTING!)
	oldProfile, err := server.store.GetProfileByUserId(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Profile belum dibuat, silahkan Create dulu"})
		return
	}

	// 3. Handle Upload AVATAR (SETELAH Bind)
	finalAvatar := oldProfile.Avatar.String // Default: Pakai yang lama
	fileAvatar, errAvatar := ctx.FormFile("avatar")
	if errAvatar == nil {
		// User upload file baru -> Ganti!
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
