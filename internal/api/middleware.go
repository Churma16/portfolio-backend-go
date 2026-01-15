package api

import (
	"go-portfolio-api/internal/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
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

func (server *Server) rateLimiterMiddleware(rateFormat string) gin.HandlerFunc {
	// 1. Tentukan Rate (Batasannya)
	rate, err := limiter.NewRateFromFormatted(rateFormat)
	if err != nil {
		// Kalau format salah, panic aja biar ketahuan pas dev
		panic(err)
	}

	// 2. Setup Store (Simpan hitungan di Redis)
	// Kita pakai server.redisClient yang sudah ada
	store, err := sredis.NewStoreWithOptions(server.redisClient, limiter.StoreOptions{
		Prefix:   "limiter_contact:", // Prefix key di Redis
		MaxRetry: 3,
	})
	if err != nil {
		panic(err)
	}

	// 3. Buat Instance Limiter
	middleware := mgin.NewMiddleware(limiter.New(store, rate), mgin.WithKeyGetter(func(c *gin.Context) string {
		// Kita batasi berdasarkan IP Address user
		return c.ClientIP()
	}))

	// 4. Return Handler
	return func(ctx *gin.Context) {
		middleware(ctx) // Jalankan limiter

		// Middleware ulule biasanya otomatis return 429 kalau limit habis.
		// Tapi kalau mau custom response JSON, kita bisa cek statusnya disini (opsional).
		// Secara default dia sudah cukup pintar.
	}
}
