package api

import (
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

// createUser is a handler function for creating a new user.
// It validates the input, hashes the password, saves the user to the database,
// and returns a response without exposing the password.
//
// @param ctx *gin.Context - The Gin context, which contains the HTTP request and response.
//
// The function performs the following steps:
// 1. Validates the input JSON payload.
// 2. Hashes the user's password.
// 3. Saves the user to the database using SQLC.
// 4. Returns a JSON response with the user's ID, email, and creation timestamp.
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest

	// 1. Validasi Input
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Hash Password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
