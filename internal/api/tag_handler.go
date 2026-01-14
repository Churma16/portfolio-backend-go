package api

import (
	"database/sql"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

type CreateTagRequest struct {
	Name       string `form:"name" binding:"required"`
	Slug       string `form:"slug"`
	Color      string `form:"color"`
	CategoryId int64  `form:"category_id" binding:"required"`
}
type UpdateTagRequest struct {
	Name       string `form:"name" binding:"required"`
	Slug       string `form:"slug"`
	Color      string `form:"color"`
	CategoryId int64  `form:"category_id" binding:"required"`
}

type TagData struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Color      string `json:"color"`
	CategoryId int64  `json:"category_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func (server *Server) createTag(ctx *gin.Context) {
	// Bind and validate the request body
	var req CreateTagRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prepare the parameters for creating a new tag
	arguments := db.CreateTagParams{
		Name:       req.Name,
		Slug:       slug.Make(req.Name),
		Color:      convertToNullString(req.Color),
		CategoryID: sql.NullInt64{Int64: req.CategoryId, Valid: true},
	}

	// Call the service to create a new tag
	tag, err := server.store.CreateTag(ctx, arguments)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Prepare and send the response
	data := tagResponse(tag)
	meta := response.NewMeta(http.StatusOK, "success", "Tag berhasil dibuat")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) showTags(ctx *gin.Context) {
	// Call the service to get all tags
	tags, err := server.store.GetTags(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Tags not found"})
		return
	}

	data := tagsResponses(tags)
	meta := response.NewMetaWithCount(http.StatusOK, "success", "Tags ditemukan", len(tags))
	resp := response.NewMultipleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) showTag(ctx *gin.Context) {
	// Parse the tag ID from the URL parameter
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	tags, err := server.store.GetTag(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}
	data := tagResponse(tags)
	meta := response.NewMeta(http.StatusOK, "success", "Tag ditemukan")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) updateTag(ctx *gin.Context) {
	var req UpdateTagRequest

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

	existingTag, err := server.store.GetTag(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	arguments := db.UpdateTagParams{
		ID:         id,
		Name:       req.Name,
		Color:      convertToNullString(req.Color),
		CategoryID: sql.NullInt64{Int64: req.CategoryId, Valid: true},
	}

	if req.Name != existingTag.Name {
		arguments.Name = req.Name
		arguments.Slug = slug.Make(req.Name)
	} else {
		arguments.Name = existingTag.Name
		arguments.Slug = existingTag.Slug
	}

	tag, err := server.store.UpdateTag(ctx, arguments)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	data := tagResponse(tag)
	meta := response.NewMeta(http.StatusOK, "success", "Tag berhasil diperbarui")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func (server *Server) deleteTag(ctx *gin.Context) {
	idParam := ctx.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	tag, err := server.store.DeleteTag(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	meta := response.NewMeta(http.StatusOK, "success", "Tag berhasil dihapus")
	data := tagResponse(tag)
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func tagResponse(tag db.Tag) TagData {
	return TagData{
		ID:         tag.ID,
		Name:       tag.Name,
		Slug:       tag.Slug,
		Color:      tag.Color.String,
		CategoryId: tag.CategoryID.Int64,
		CreatedAt:  tag.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  tag.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func tagsResponses(tags []db.Tag) []TagData {
	data := make([]TagData, len(tags))
	for i, tag := range tags {
		data[i] = TagData{
			ID:         tag.ID,
			Name:       tag.Name,
			Slug:       tag.Slug,
			Color:      tag.Color.String,
			CategoryId: tag.CategoryID.Int64,
			CreatedAt:  tag.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  tag.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return data
}
