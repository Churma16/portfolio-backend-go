package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
)

// Struct untuk validasi input JSON dari Postman/Frontend
type createUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Struct agar response JSON user tidak menampilkan Password (Bahaya!)
type userResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"` // String biar aman formatnya
}

// Struct untuk validasi input JSON login
type loginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Struct untuk response login
type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest

	// 1. Validasi Input
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	// 2. Hash Password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", "failed to hash password"))
		return
	}

	// 3. Simpan ke Database (Panggil SQLC)
	arg := db.CreateUserParams{
		Email:    req.Email,
		Password: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		// TODO: Nanti kita handle error "Email Already Exists" disini
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	// 4. Kirim Response (Tanpa Password)
	rsp := userResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest

	// 1. Validasi Input JSON
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	// 2. Cari User di Database berdasarkan Email
	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(http.StatusUnauthorized, "error", "User tidak ditemukan / Salah Password"))
		return
	}

	// 3. Cek Password (Bandingkan input vs Hash di DB)
	err = util.CheckPassword(req.Password, user.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(http.StatusUnauthorized, "error", "Password Salah"))
		return
	}

	// 4. Bikin Token (Berlaku 24 Jam)
	accessToken, err := util.CreateToken(user.ID, 24*time.Hour, server.tokenKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", "Gagal membuat token"))
		return
	}

	// 5. Kirim Response Token + Data User
	rsp := loginUserResponse{
		AccessToken: accessToken,
		User: userResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	}

	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) changePassword(ctx *gin.Context) {

	// Get userID dari context (di-set oleh middleware)
	userID := ctx.MustGet("user_id").(int64)

	var req changePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	// 1. Ambil data user dari DB
	user, err := server.store.GetUserByID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", "User tidak ditemukan"))
		return
	}

	// 2. Cek Old Password
	err = util.CheckPassword(req.OldPassword, user.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, response.ErrorResponse(http.StatusUnauthorized, "error", "Old Password Salah"))
		return
	}

	// 3. Hash New Password
	hashedPassword, err := util.HashPassword(req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", "Gagal hash password baru"))
		return
	}

	arg := db.UpdateUserPasswordParams{
		ID:       userID,
		Password: hashedPassword,
	}

	// 4. Update Password di DB
	_, err = server.store.UpdateUserPassword(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", "Gagal memperbarui password"))
		return
	}

	meta := response.NewMeta(http.StatusOK, "success", "Password berhasil diperbarui")
	ctx.JSON(http.StatusOK, meta)

}
