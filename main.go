package main

import (
	"database/sql"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/api"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Load file .env

	err := godotenv.Load("app.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Sesuaikan alamat Redis kamu
		Password: "",               // Kosongkan jika tidak ada password
		DB:       0,                // Default DB
	})
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
	server := api.NewServer(store, redisClient)

	// 4. Jalankan Server
	log.Println("Server running on", serverAddress)
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
