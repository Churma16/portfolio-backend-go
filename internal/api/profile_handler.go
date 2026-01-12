package api

import (
	"encoding/json"
	db "go-portfolio-api/db/sqlc"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sqlc-dev/pqtype"
)

type createProfileRequest struct {
	Name           string          `json:"name" binding:"required"`
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
}

// Struct agar output JSON bersih (tanpa String/Valid)
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

	// Validasi Input JSON
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ambil user_id dari Token (yang ditaruh Middleware)
	// Kita pakai MustGet karena yakin middleware sudah jalan
	userID := ctx.MustGet("user_id").(int64)

	// Simpan ke Database
	arg := db.CreateProfileParams{
		UserID:         userID,
		Name:           req.Name,
		Headline:       convertToNullString(req.Headline),
		Role:           convertToNullString(req.Role),
		BioShort:       convertToNullString(req.BioShort),
		BioLong:        convertToNullString(req.BioLong),
		Location:       convertToNullString(req.Location),
		IsHireable:     convertToNullBool(req.IsHireable),
		Avatar:         convertToNullString(req.Avatar),
		CvFiles:        convertToNullString(req.CvFiles),
		HeroImageCodes: convertToNullString(req.HeroImageCodes),
		Socials:        pqtype.NullRawMessage{RawMessage: req.Socials, Valid: len(req.Socials) > 0},
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
