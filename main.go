package main

import (
	"context"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/api"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Load file .env
	err := godotenv.Load("app.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	tokenKey := os.Getenv("TOKEN_SYMMETRIC_KEY")
	if len(tokenKey) != 32 {
		log.Fatal("TOKEN_SYMMETRIC_KEY must be 32 characters")
	}

	redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient := redis.NewClient(
		&redis.Options{
			Addr:     redisAddr,
			Password: "",
			DB:       0,
		},
	)
	dbSource := os.Getenv("DB_SOURCE")
	serverAddress := os.Getenv("SERVER_ADDRESS")

	// 2. Konek Database
	conn, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	defer conn.Close()

	// 3. Init Store & Server
	//store := db.New(conn)
	store := db.NewStore(conn)
	server := api.NewServer(store, redisClient, tokenKey)

	// 4. Jalankan Server
	log.Println("Server running on", serverAddress)
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
