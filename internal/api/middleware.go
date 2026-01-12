package api

import (
	"go-portfolio-api/internal/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func authMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. Ambil Header Authorization
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// 2. Cek Format "Bearer <token>"
		fields := strings.Fields(authHeader)
		if len(fields) < 2 || strings.ToLower(fields[0]) != "bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		// 3. Verifikasi Token
		accessToken := fields[1]
		userID, err := util.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 4. Simpan User ID ke Context (Biar bisa dipakai di handler selanjutnya)
		ctx.Set("user_id", userID)
		ctx.Next()
	}
}
