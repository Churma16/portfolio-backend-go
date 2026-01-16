package api

import (
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Status                string  `json:"status"`
	RedisAlive            bool    `json:"redis_alive"`
	ServerTime            float64 `json:"server_time"`
	RedisProcessingTimeMs float64 `json:"redis_processing_time"`
}

// Health checks the health status of the server and Redis connection
func (server *Server) health(ctx *gin.Context) {
	// 1. Start stopwatch
	startTime := time.Now()

	// 2. Check Redis connection (ping)
	redisAlive := false
	err := server.redisClient.Ping(ctx).Err()
	if err == nil {
		redisAlive = true
	}

	// 3. Stop stopwatch
	endTime := time.Now()

	// 4. Calculate duration in milliseconds
	processingTime := endTime.Sub(startTime).Seconds() * 1000

	// 5. Get current server time (Unix timestamp as float)
	serverTime := time.Now().Unix()

	// 6. Prepare response
	healthData := healthResponse{
		Status:                "ok",
		RedisAlive:            redisAlive,
		ServerTime:            float64(serverTime),
		RedisProcessingTimeMs: roundFloat(processingTime, 2),
	}

	//meta := response.NewMeta(http.StatusOK, "success", "Server is healthy")
	ctx.JSON(http.StatusOK, healthData)
}

// roundFloat rounds a float to a specified number of decimal places
func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
