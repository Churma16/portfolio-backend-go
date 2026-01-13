package api

import (
	"database/sql"
	db "go-portfolio-api/db/sqlc"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

type createCategoryRequest struct {
	Name  string `form:"name" binding:"required"`
	Slug  string `form:"slug"`
	Color string `form:"color"`
}
type Meta struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
type CategoryData struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type SingleCategoryResponse struct {
	Meta Meta         `json:"meta"`
	Data CategoryData `json:"data"`
}

type MultipleCategoriesResponse struct {
	Meta Meta           `json:"meta"`
	Data []CategoryData `json:"data"`
}

func (server *Server) createCategory(ctx *gin.Context) {
	// Bind JSON request ke struct
	var req createCategoryRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Panggil method CreateCategory di db/sqlc
	arg := db.CreateCategoryParams{
		Name:  req.Name,
		Slug:  slug.Make(req.Name),
		Color: convertToNullString(req.Color),
	}

	// Execute query
	category, err := server.store.CreateCategory(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := categoryResponse(category, "Kategori berhasil dibuat")
	ctx.JSON(http.StatusOK, response)
}

func (server *Server) showCategories(ctx *gin.Context) {

	categories, err := server.store.GetCategories(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
		return
	}

	ctx.JSON(http.StatusOK, categoriesResponse(categories, "Kategori ditemukan"))
}

func (server *Server) showCategory(ctx *gin.Context) {
	idParam := ctx.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	categories, err := server.store.GetCategory(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, categoryResponse(categories, "Kategori ditemukan"))
}

func (server *Server) deleteCategory(ctx *gin.Context) {
	idParam := ctx.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	category, err := server.store.DeleteCategory(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, categoryResponse(category, "Kategori berhasil dihapus"))

}

func mapCategoryToData(category db.Category) CategoryData {
	return CategoryData{
		ID:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		Color:     category.Color.String,
		CreatedAt: category.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: category.UpdatedAt.Format("2006-01-02 15:04:05")}
}

func categoryResponse(category db.Category, message string) gin.H {
	return gin.H{
		"meta": gin.H{
			"code":    200,
			"status":  "success",
			"message": message,
		},
		"data": mapCategoryToData(category),
	}
}

func categoriesResponse(categories []db.Category, message string) gin.H {
	data := make([]CategoryData, len(categories))
	for i, category := range categories {
		data[i] = mapCategoryToData(category)
	}

	return gin.H{
		"meta": gin.H{
			"code":    200,
			"status":  "success",
			"message": message,
			"count":   len(categories),
		},
		"data": data,
	}
}
