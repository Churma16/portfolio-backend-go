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
	var tagRequest CreateTagRequest
	if err := ctx.ShouldBind(&tagRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	// Prepare the parameters for creating a new tag
	createTagParams := db.CreateTagParams{
		Name:       tagRequest.Name,
		Slug:       slug.Make(tagRequest.Name),
		Color:      convertToNullString(tagRequest.Color),
		CategoryID: sql.NullInt64{Int64: tagRequest.CategoryId, Valid: true},
	}

	// Call the service to create a new tag
	createdTag, creationError := server.store.CreateTag(ctx, createTagParams)
	if creationError != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", creationError.Error()))
		return
	}

	// Prepare and send the response
	responseData := tagResponse(createdTag)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Tag berhasil dibuat")
	responsePayload := response.NewSingleDataResponse(responseMeta, responseData)
	ctx.JSON(http.StatusOK, responsePayload)
}

func (server *Server) showTags(ctx *gin.Context) {
	// Call the service to get all tags
	allTags, retrievalError := server.store.GetTags(ctx)
	if retrievalError != nil {
		ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Tags not found"))
		return
	}

	responseData := tagsResponses(allTags)
	responseMeta := response.NewMetaWithCount(http.StatusOK, "success", "Tags ditemukan", len(allTags))
	responsePayload := response.NewMultipleDataResponse(responseMeta, responseData)
	ctx.JSON(http.StatusOK, responsePayload)
}

func (server *Server) showTag(ctx *gin.Context) {
	// Parse the tag ID from the URL parameter
	tagIDParam := ctx.Param("id")
	tagID, parseError := strconv.ParseInt(tagIDParam, 10, 64)
	if parseError != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "invalid id"))
		return
	}

	tagDetails, retrievalError := server.store.GetTag(ctx, tagID)
	if retrievalError != nil {
		if retrievalError == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Category not found"))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", "Internal server error"))
		}
		return
	}
	responseData := tagResponse(tagDetails)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Tag ditemukan")
	responsePayload := response.NewSingleDataResponse(responseMeta, responseData)
	ctx.JSON(http.StatusOK, responsePayload)
}

func (server *Server) updateTag(ctx *gin.Context) {
	var req UpdateTagRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "invalid id"))
		return
	}

	existingTag, err := server.store.GetTag(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Tag not found"))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
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
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Tag not found"))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
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
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "invalid id"))
		return
	}

	tag, err := server.store.DeleteTag(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Tag not found"))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
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
