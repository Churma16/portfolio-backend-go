package main

import (
	"database/sql"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/api"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Load file .env
	err := godotenv.Load("app.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbDriver := os.Getenv("DB_DRIVER")
	dbSource := os.Getenv("DB_SOURCE")
	serverAddress := os.Getenv("SERVER_ADDRESS")

	// 2. Konek Database
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	// 3. Init Store & Server
	//store := db.New(conn)
	store := db.NewStore(conn)
	server := api.NewServer(store)

	// 4. Jalankan Server
	log.Println("Server running on", serverAddress)
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
