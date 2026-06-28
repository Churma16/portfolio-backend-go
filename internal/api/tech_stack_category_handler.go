package api

import (
	"encoding/json"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"
)

type CreateTechStackCategoryRequest struct {
	Name  string `form:"name" json:"name" binding:"required"`
	Slug  string `form:"slug" json:"slug"`
	Color string `form:"color" json:"color"`
}

type UpdateTechStackCategoryRequest struct {
	Name  string `form:"name" json:"name" binding:"required"`
	Slug  string `form:"slug" json:"slug"`
	Color string `form:"color" json:"color"`
}

type TechStackCategoryData struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (server *Server) createTechStackCategory(ctx *gin.Context) {
	var req CreateTechStackCategoryRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	createParams := db.CreateTechStackCategoryParams{
		Name:  req.Name,
		Slug:  slug.Make(req.Name),
		Color: convertToNullString(req.Color),
	}

	category, err := server.store.CreateTechStackCategory(ctx, createParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "tech_stack_categories:")
	responseData := techStackCategoryResponse(category)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Tech stack category berhasil dibuat")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, responseData))
}

func (server *Server) showTechStackCategories(ctx *gin.Context) {
	cacheKey := "tech_stack_categories:list:all"
	cacheValue, cacheErr := server.redisClient.Get(ctx, cacheKey).Result()
	if cacheErr == nil {
		var cachedCategories []TechStackCategoryData
		if err := json.Unmarshal([]byte(cacheValue), &cachedCategories); err == nil {
			responseMeta := response.NewMetaWithCount(http.StatusOK, "success", "Tech stack categories (Cached)", len(cachedCategories))
			ctx.JSON(http.StatusOK, response.NewMultipleDataResponse(responseMeta, cachedCategories))
			return
		}
	}

	categories, err := server.store.GetTechStackCategories(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	responseData := techStackCategoriesResponse(categories)
	cachedData, _ := json.Marshal(responseData)
	server.redisClient.Set(ctx, cacheKey, cachedData, 1*time.Hour)

	responseMeta := response.NewMetaWithCount(http.StatusOK, "success", "Get all tech stack categories", len(categories))
	ctx.JSON(http.StatusOK, response.NewMultipleDataResponse(responseMeta, responseData))
}

func (server *Server) showTechStackCategory(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid tech stack category ID"))
		return
	}

	category, err := server.store.GetTechStackCategory(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Tech stack category not found"))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		}
		return
	}

	responseData := techStackCategoryResponse(category)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Get tech stack category")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, responseData))
}

func (server *Server) updateTechStackCategory(ctx *gin.Context) {
	var req UpdateTechStackCategoryRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid tech stack category ID"))
		return
	}

	existingCategory, err := server.store.GetTechStackCategory(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Tech stack category not found"))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		}
		return
	}

	updateParams := db.UpdateTechStackCategoryParams{
		ID:    id,
		Name:  req.Name,
		Color: convertToNullString(req.Color),
	}

	if req.Name != existingCategory.Name {
		updateParams.Slug = slug.Make(req.Name)
	} else {
		updateParams.Slug = existingCategory.Slug
	}

	category, err := server.store.UpdateTechStackCategory(ctx, updateParams)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "tech_stack_categories:")
	util.DeleteCacheByPrefix(server.redisClient, "tech_stacks:")
	responseData := techStackCategoryResponse(category)
	meta := response.NewMeta(http.StatusOK, "success", "Tech stack category berhasil diperbarui")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, responseData))
}

func (server *Server) deleteTechStackCategory(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid tech stack category ID"))
		return
	}

	category, err := server.store.DeleteTechStackCategory(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "tech_stack_categories:")
	util.DeleteCacheByPrefix(server.redisClient, "tech_stacks:")
	responseData := techStackCategoryResponse(category)
	meta := response.NewMeta(http.StatusOK, "success", "Tech stack category berhasil dihapus")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, responseData))
}

func techStackCategoryResponse(category db.TechStackCategory) TechStackCategoryData {
	return TechStackCategoryData{
		ID:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		Color:     category.Color.String,
		CreatedAt: category.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		UpdatedAt: category.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
	}
}

func techStackCategoriesResponse(categories []db.TechStackCategory) []TechStackCategoryData {
	data := make([]TechStackCategoryData, len(categories))
	for i, category := range categories {
		data[i] = techStackCategoryResponse(category)
	}
	return data
}
