package api

import (
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Server melayani request HTTP dan koneksi DB
type Server struct {
	store           *db.Store // Ini struct hasil generate SQLC
	categoryService service.CategoryService
	router          *gin.Engine
	redisClient     *redis.Client
	tokenKey        string // <--- TAMBAHKAN INI (Untuk menyimpan TOKEN_SYMMETRIC_KEY)
}

// NewServer membuat instance server baru
func NewServer(store *db.Store, redisClient *redis.Client, tokenKey string) *Server {
	server := &Server{
		store:           store,
		redisClient:     redisClient,
		tokenKey:        tokenKey,                          // Simpan ke struct
		categoryService: service.NewCategoryService(store), // Initialize categoryService
	}
	router := gin.Default()
	// ADD CORS CONFIGURATION
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://churma.codes"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(config))

	//rute UTama
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API is running"})
	})

	// SERVE STATIC FILES (Storage folder for uploaded images)
	router.Static("/storage", "./storage")

	// RUTE DENGAN AUTHENTIKASI
	authRoutes := router.Group("/api").Use(server.authMiddleware())
	apiRoutes := router.Group("/api")

	// RUTE PUBLIC (Siapapun boleh akses)
	apiRoutes.POST("/users", server.createUser)
	apiRoutes.POST("/login", server.loginUser)
	apiRoutes.GET("/health", server.health)

	// RUTE PRIVATE (Harus bawa Token)

	// Rute Profile
	authRoutes.POST("/profiles", server.createProfile)
	apiRoutes.GET("/profiles/:user_id", server.getProfile)
	apiRoutes.GET("/profiles", server.getProfiles)
	authRoutes.PUT("/profiles", server.updateProfile)
	authRoutes.POST("/upload", server.uploadFile)
	authRoutes.DELETE("/profiles", server.deleteProfile)

	// Rute Category
	authRoutes.POST("/categories", server.createCategory)
	apiRoutes.GET("/categories", server.showCategories)
	apiRoutes.GET("/categories/:id", server.showCategory)
	authRoutes.PUT("/categories/:id", server.updateCategory)
	authRoutes.DELETE("/categories/:id", server.deleteCategory)

	// Rute Tag
	authRoutes.POST("/tags", server.createTag)
	apiRoutes.GET("/tags", server.showTags)
	apiRoutes.GET("/tags/:id", server.showTag)
	authRoutes.PUT("/tags/:id", server.updateTag)
	authRoutes.DELETE("/tags/:id", server.deleteTag)

	// Rute Tech Stack
	authRoutes.POST("/tech-stacks", server.createTechStack)
	apiRoutes.GET("/tech-stacks", server.showTechStacks)
	apiRoutes.GET("/tech-stacks/:id", server.showTechStack)
	authRoutes.PUT("/tech-stacks/:id", server.updateTechStack)
	authRoutes.DELETE("/tech-stacks/:id", server.deleteTechStack)

	// Rute Project
	authRoutes.POST("/projects", server.createProject)
	apiRoutes.GET("/projects", server.showProjects)
	apiRoutes.GET("/projects/:id", server.showProject)
	authRoutes.PUT("/projects/:id", server.updateProject)
	authRoutes.DELETE("/projects/:id", server.deleteProject)

	// Rute Work Experience
	authRoutes.POST("/work-experiences", server.createWorkExperience)
	apiRoutes.GET("/work-experiences", server.showWorkExperiences)
	apiRoutes.GET("/work-experiences/:id", server.showWorkExperience)
	authRoutes.PUT("/work-experiences/:id", server.updateWorkExperience)
	authRoutes.DELETE("/work-experiences/:id", server.deleteWorkExperience)

	// Rute Message
	apiRoutes.POST("/messages", server.rateLimiterMiddleware("3-H"), server.createMessage)
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
