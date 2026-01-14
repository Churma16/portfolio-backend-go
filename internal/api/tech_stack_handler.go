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

type CreateTechStackRequest struct {
	Name string `form:"name" binding:"required"`
	Slug string `form:"slug"`
	Icon string `form:"icon"`
}

type updateTechStackRequest struct {
	Name string `form:"name" binding:"required"`
	Slug string `form:"slug"`
	Icon string `form:"icon"`
}

type TechStackData struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Icon        string `json:"icon"`
	ColumnOrder int32  `json:"column_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (server *Server) createTechStack(ctx *gin.Context) {
	// Bind and validate the request body
	var req CreateTechStackRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prepare the parameters for creating a new tech stack
	arguments := db.CreateTechStackParams{
		Name: req.Name,
		Slug: slug.Make(req.Name),
		Icon: convertToNullString(req.Icon),
	}

	// Call the service to create a new tech stack
	techStack, err := server.store.CreateTechStack(ctx, arguments)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := techStackResponse(techStack)
	meta := response.NewMeta(http.StatusOK, "success", "Create tech stack")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, data))
}

func (server *Server) showTechStacks(ctx *gin.Context) {
	techStacks, err := server.store.GetTechStacks(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := TechStacksResponse(techStacks)
	meta := response.NewMetaWithCount(http.StatusOK, "success", "Get all tech stacks", len(techStacks))
	ctx.JSON(http.StatusOK, response.NewMultipleDataResponse(meta, data))
}

func (server *Server) showTechStack(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tech stack ID"})
		return
	}

	techStacks, err := server.store.GetTechStack(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := techStackResponse(techStacks)
	meta := response.NewMeta(http.StatusOK, "success", "Get tech stack")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, data))

}

func (server *Server) updateTechStack(ctx *gin.Context) {
	var req updateTechStackRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tech stack ID"})
		return
	}

	existingTechStack, err := server.store.GetTechStack(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Tech Stack not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	arguments := db.UpdateTechStackParams{
		ID:   id,
		Name: req.Name,
		Icon: convertToNullString(req.Icon),
	}

	if req.Name != existingTechStack.Name {
		arguments.Name = req.Name
		arguments.Slug = slug.Make(req.Name)
	} else {
		arguments.Name = existingTechStack.Name
		arguments.Slug = existingTechStack.Slug
	}

	techStack, err := server.store.UpdateTechStack(ctx, arguments)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := techStackResponse(techStack)
	meta := response.NewMeta(http.StatusOK, "success", "Tech stack berhasil diperbarui")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)

}

func (server *Server) deleteTechStack(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tech stack ID"})
		return
	}

	techStack, err := server.store.DeleteTechStack(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := techStackResponse(techStack)
	meta := response.NewMeta(http.StatusOK, "success", "Tech stack berhasil dihapus")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

func techStackResponse(techStack db.TechStack) TechStackData {
	return TechStackData{
		ID:          techStack.ID,
		Name:        techStack.Name,
		Slug:        techStack.Slug,
		Icon:        techStack.Icon.String,
		ColumnOrder: techStack.ColumnOrder,
		CreatedAt:   techStack.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   techStack.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func TechStacksResponse(techStacks []db.TechStack) []TechStackData {
	data := make([]TechStackData, len(techStacks))
	for i, techStack := range techStacks {
		data[i] = TechStackData{
			ID:          techStack.ID,
			Name:        techStack.Name,
			Slug:        techStack.Slug,
			Icon:        techStack.Icon.String,
			ColumnOrder: techStack.ColumnOrder,
			CreatedAt:   techStack.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   techStack.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return data
}
