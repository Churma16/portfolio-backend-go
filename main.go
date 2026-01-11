package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// GANTI PASSWORD 'secret' DENGAN PASSWORD ANDA JUGA DISINI
const dbSource = "postgresql://postgres:root@localhost:5432/go_portfolio_db?sslmode=disable"

func main() {
	conn, err := sql.Open("postgres", dbSource)
	if err != nil {
		log.Fatal("Gagal connect DB:", err)
	}

	if err = conn.Ping(); err != nil {
		log.Fatal("DB tidak merespon:", err)
	}

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "Sukses! Database Terkoneksi via Make!"})
	})

	r.Run(":8080")
}
