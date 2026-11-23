package api

import (
	"net/http"
	"time"

	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"

	"github.com/gin-gonic/gin"
)

type FlushCacheRequest struct {
	Prefix string `json:"prefix"`
}

func (server *Server) flushCache(ctx *gin.Context) {
	var requestPayload FlushCacheRequest
	_ = ctx.ShouldBindJSON(&requestPayload)

	if requestPayload.Prefix == "" || requestPayload.Prefix == "all" {
		targetPrefixes := []string{
			"categories:",
			"site_profile",
			"projects:",
			"tags:",
			"tech_stacks:",
			"tech_stack_categories:",
			"workExperiences:",
		}
		for _, targetPrefix := range targetPrefixes {
			util.DeleteCacheByPrefix(server.redisClient, targetPrefix)
		}
	} else {
		util.DeleteCacheByPrefix(server.redisClient, requestPayload.Prefix)
	}

	responseMeta := response.NewMeta(http.StatusOK, "success", "Go cache flushed successfully")
	responseData := gin.H{
		"backend":    "go",
		"flushed_at": time.Now().Format(time.RFC3339),
	}
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, responseData))
}
