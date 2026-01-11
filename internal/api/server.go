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

	// --> NANTI KITA DAFTARKAN RUTE DISINI <--
	router.POST("/users", server.createUser) // Endpoint Register

	server.router = router
	return server
}

// Start menjalankan server pada port tertentu
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
