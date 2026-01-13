package api

import (
	db "go-portfolio-api/db/sqlc"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

type createCategoryRequest struct {
	Name  string `form:"name" binding:"required"`
	Slug  string `form:"slug"`
	Color string `form:"color"`
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

	response := categoryResponse(category)
	ctx.JSON(http.StatusOK, response)
}

func categoryResponse(category db.Category) gin.H {
	return gin.H{
		"id":         category.ID,
		"name":       category.Name,
		"slug":       category.Slug,
		"color":      category.Color.String,
		"created_at": category.CreatedAt,
		"updated_at": category.UpdatedAt,
	}
}
