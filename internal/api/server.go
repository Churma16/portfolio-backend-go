package api

import (
	db "go-portfolio-api/db/sqlc"

	"github.com/gin-gonic/gin"
)

// Server melayani request HTTP dan koneksi DB
type Server struct {
	store  *db.Queries // Ini struct hasil generate SQLC
	router *gin.Engine
}

// NewServer membuat instance server baru
func NewServer(store *db.Queries) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// RUTE PUBLIC (Siapapun boleh akses)
	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	// RUTE PRIVATE (Harus bawa Token)
	authRoutes := router.Group("/").Use(authMiddleware())

	// Rute Profile
	authRoutes.POST("/profiles", server.createProfile)
	router.GET("/profiles/:user_id", server.getProfile)
	authRoutes.PUT("/profiles", server.updateProfile)
	authRoutes.POST("/upload", server.uploadFile)
	authRoutes.DELETE("/profiles", server.deleteProfile)

	// Rute Category
	router.GET("/categories", server.showCategories)
	router.GET("/categories/:id", server.showCategory)
	authRoutes.POST("/categories", server.createCategory)
	authRoutes.PUT("/categories/:id", server.updateCategory)
	authRoutes.DELETE("/categories/:id", server.deleteCategory)

	// Contoh: Rute Cek "Siapa Saya?" (Hanya bisa diakses kalau login)
	authRoutes.GET("/users/me", func(ctx *gin.Context) {
		userID, _ := ctx.Get("user_id")
		ctx.JSON(200, gin.H{
			"message":      "Kamu berhasil masuk area rahasia!",
			"your_user_id": userID,
		})
	})

	server.router = router
	return server
}

// Start menjalankan server pada port tertentu
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
