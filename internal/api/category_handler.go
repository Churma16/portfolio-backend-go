package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"go-portfolio-api/internal/dto"
	"go-portfolio-api/internal/mapper"
	"go-portfolio-api/internal/response"

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
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) showCategories(ctx *gin.Context) {
	categories, err := server.categoryService.GetCategories(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Kategori tidak ditemukan"})
		return
	}

	data := mapper.MapCategoriesToData(categories)
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

	data := mapper.MapCategoryToData(category)
	meta := response.NewMeta(http.StatusOK, "success", "Kategori berhasil diperbarui")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) deleteCategory(ctx *gin.Context) {
	idParam := ctx.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	category, err := server.categoryService.DeleteCategory(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	data := mapper.MapCategoryToData(category)
	meta := response.NewMeta(http.StatusOK, "success", "Kategori berhasil dihapus")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}
