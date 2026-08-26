package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go-portfolio-api/internal/dto"
	"go-portfolio-api/internal/mapper"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"

	"github.com/gin-gonic/gin"
)

func (server *Server) createCategory(ctx *gin.Context) {
	var req dto.CreateCategoryRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := server.categoryService.CreateCategory(ctx, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := mapper.MapCategoryToData(category)
	meta := response.NewMeta(http.StatusOK, "success", "Kategori berhasil dibuat")
	resp := response.NewSingleDataResponse(meta, data)
	util.DeleteCacheByPrefix(server.redisClient, "categories:")
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) showCategories(ctx *gin.Context) {
	cacheKey := "categories:list:all"
	if server.redisClient != nil {
		cacheValue, cacheErr := server.redisClient.Get(ctx, cacheKey).Result()
		if cacheErr == nil {
			var cachedCategories []dto.CategoryData
			if err := json.Unmarshal([]byte(cacheValue), &cachedCategories); err == nil {
				meta := response.NewMetaWithCount(http.StatusOK, "success", "Kategori ditemukan (Cached)", len(cachedCategories))
				resp := response.NewMultipleDataResponse(meta, cachedCategories)
				ctx.JSON(http.StatusOK, resp)
				return
			}
		}
	}

	categories, err := server.categoryService.GetCategories(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
		return
	}

	data := mapper.MapCategoriesToData(categories)
	if server.redisClient != nil {
		cachedData, _ := json.Marshal(data)
		server.redisClient.Set(ctx, cacheKey, cachedData, 1*time.Hour)
	}

	meta := response.NewMetaWithCount(http.StatusOK, "success", "Kategori ditemukan", len(categories))
	resp := response.NewMultipleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) showCategory(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// Define cache key
	cacheKey := "categories:single:" + idParam

	// Check if data exists in cache
	if server.redisClient != nil {
		cacheValue, cacheErr := server.redisClient.Get(ctx, cacheKey).Result()
		if cacheErr == nil {
			var cachedCategory dto.CategoryData
			if err := json.Unmarshal([]byte(cacheValue), &cachedCategory); err == nil {
				meta := response.NewMeta(http.StatusOK, "success", "Kategori ditemukan (Cached)")
				resp := response.NewSingleDataResponse(meta, cachedCategory)
				ctx.JSON(http.StatusOK, resp)
				return
			}
		}
	}

	category, err := server.categoryService.GetCategory(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	data := mapper.MapCategoryToData(category)
	if server.redisClient != nil {
		cachedData, _ := json.Marshal(data)
		server.redisClient.Set(ctx, cacheKey, cachedData, 1*time.Hour)
	}

	meta := response.NewMeta(http.StatusOK, "success", "Kategori ditemukan")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) updateCategory(ctx *gin.Context) {

	var req dto.UpdateCategoryRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	category, err := server.categoryService.UpdateCategory(ctx, id, req)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "categories:")
	data := mapper.MapCategoryToData(category)
	meta := response.NewMeta(http.StatusOK, "success", "Kategori berhasil diperbarui")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) deleteCategory(ctx *gin.Context) {
	idParam := ctx.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "invalid id"))
		return
	}

	category, err := server.categoryService.DeleteCategory(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Category not found"))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", "Internal server error"))
		}
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "categories:")
	data := mapper.MapCategoryToData(category)
	meta := response.NewMeta(http.StatusOK, "success", "Kategori berhasil dihapus")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}
