package api

import (
	"encoding/json"
	db "go-portfolio-api/db/sqlc"
	"net/http"

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

	//	Sukses
	ctx.JSON(http.StatusOK, profile)
}
