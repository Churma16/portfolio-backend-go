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

type CreateTechStackRequest struct {
	Name                string `form:"name" json:"name" binding:"required"`
	Slug                string `form:"slug" json:"slug"`
	Icon                string `form:"icon" json:"icon"`
	TechStackCategoryID int64  `form:"tech_stack_category_id" json:"tech_stack_category_id"`
}

type updateTechStackRequest struct {
	Name                string `form:"name" json:"name" binding:"required"`
	Slug                string `form:"slug" json:"slug"`
	Icon                string `form:"icon" json:"icon"`
	TechStackCategoryID int64  `form:"tech_stack_category_id" json:"tech_stack_category_id"`
}

type TechStackData struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	Slug                string `json:"slug"`
	Icon                string `json:"icon"`
	ColumnOrder         int32  `json:"column_order"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	TechStackCategoryID *int64 `json:"tech_stack_category_id"`
}

func (server *Server) createTechStack(ctx *gin.Context) {
	// Bind and validate the request body
	var techStackRequest CreateTechStackRequest
	if err := ctx.ShouldBind(&techStackRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	// Prepare the parameters for creating a new tech stack
	createTechStackParams := db.CreateTechStackParams{
		Name:                techStackRequest.Name,
		Slug:                slug.Make(techStackRequest.Name),
		Icon:                convertToNullString(techStackRequest.Icon),
		TechStackCategoryID: convertToNullInt64(techStackRequest.TechStackCategoryID),
	}

	// Call the service to create a new tech stack
	createdTechStack, creationError := server.store.CreateTechStack(ctx, createTechStackParams)
	if creationError != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", creationError.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "tech_stacks:")
	responseData := techStackResponse(createdTechStack)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Create tech stack")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, responseData))
}

func (server *Server) showTechStacks(ctx *gin.Context) {
	cacheKey := "tech_stacks:list:all"
	cacheValue, cacheErr := server.redisClient.Get(ctx, cacheKey).Result()
	if cacheErr == nil {
		var cachedTechStacks []TechStackData
		if err := json.Unmarshal([]byte(cacheValue), &cachedTechStacks); err == nil {
			responseMeta := response.NewMetaWithCount(http.StatusOK,
				"success",
				"Get all tech stacks (Cached)",
				len(cachedTechStacks))
			ctx.JSON(http.StatusOK, response.NewMultipleDataResponse(responseMeta, cachedTechStacks))
			return
		}
	}

	allTechStacks, retrievalError := server.store.GetTechStacks(ctx)
	if retrievalError != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", retrievalError.Error()))
		return
	}

	responseData := TechStacksResponse(allTechStacks)
	cachedData, _ := json.Marshal(responseData)
	server.redisClient.Set(ctx, cacheKey, cachedData, 1*time.Hour)

	responseMeta := response.NewMetaWithCount(http.StatusOK, "success", "Get all tech stacks", len(allTechStacks))
	ctx.JSON(http.StatusOK, response.NewMultipleDataResponse(responseMeta, responseData))
}

func (server *Server) showTechStack(ctx *gin.Context) {
	techStackIDParam := ctx.Param("id")
	techStackID, parseError := strconv.ParseInt(techStackIDParam, 10, 64)
	if parseError != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid tech stack ID"))
		return
	}

	techStackDetails, retrievalError := server.store.GetTechStack(ctx, techStackID)
	if retrievalError != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", retrievalError.Error()))
		return
	}

	responseData := techStackResponse(techStackDetails)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Get tech stack")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, responseData))

}

func (server *Server) updateTechStack(ctx *gin.Context) {

	var req updateTechStackRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid tech stack ID"))
		return
	}

	existingTechStack, err := server.store.GetTechStack(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Tech Stack not found"))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		}
		return
	}

	arguments := db.UpdateTechStackParams{
		ID:                  id,
		Name:                req.Name,
		Icon:                convertToNullString(req.Icon),
		TechStackCategoryID: convertToNullInt64(req.TechStackCategoryID),
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
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "tech_stacks:")
	data := techStackResponse(techStack)
	meta := response.NewMeta(http.StatusOK, "success", "Tech stack berhasil diperbarui")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)

}

func (server *Server) deleteTechStack(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid tech stack ID"))
		return
	}

	techStack, err := server.store.DeleteTechStack(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "tech_stacks:")
	data := techStackResponse(techStack)
	meta := response.NewMeta(http.StatusOK, "success", "Tech stack berhasil dihapus")
	resp := response.NewSingleDataResponse(meta, data)
	ctx.JSON(http.StatusOK, resp)
}

type reorderTechStackRequest struct {
	Direction string `json:"direction" binding:"required,oneof=up down"`
}

func (server *Server) reorderTechStack(ctx *gin.Context) {
	// 1. Extract Tech Stack ID from URI and Direction from Body
	techStackIDStr := ctx.Param("id")
	techStackID, err := strconv.ParseInt(techStackIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid tech stack ID"))
		return
	}

	var req reorderTechStackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	// 2. Retrieve Current Tech Stack
	currentTechStack, err := server.store.GetTechStack(ctx, techStackID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Tech stack not found"))
		return
	}

	currentOrder := currentTechStack.ColumnOrder

	// 3. Determine Target Order & Fetch Adjacent Tech Stack
	var targetOrder int32
	var adjacentTechStack db.TechStack

	if req.Direction == "up" {
		targetOrder = currentOrder - 1
		if targetOrder < 1 {
			ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Cannot move tech stack further up"))
			return
		}
		adjacentTechStack, err = server.store.GetTechStackByColumnOrder(ctx, targetOrder)
	} else { // down
		targetOrder = currentOrder + 1
		adjacentTechStack, err = server.store.GetTechStackByColumnOrder(ctx, targetOrder)
	}

	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Cannot move tech stack further in that direction"))
		return
	}

	// 4. TRANSACTION: Swap Column Orders
	errTx := server.store.ExecTx(ctx, func(q *db.Queries) error {
		// Swap: current -> target, adjacent -> current
		swapParams1 := db.UpdateTechStackColumnOrderParams{
			ID:          currentTechStack.ID,
			ColumnOrder: targetOrder, // NEW ORDER
		}

		swapParams2 := db.UpdateTechStackColumnOrderParams{
			ID:          adjacentTechStack.ID,
			ColumnOrder: currentOrder, // SWAP BACK
		}

		if _, err := q.UpdateTechStackColumnOrder(ctx, swapParams1); err != nil {
			return err
		}
		if _, err := q.UpdateTechStackColumnOrder(ctx, swapParams2); err != nil {
			return err
		}

		return nil
	})

	if errTx != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", errTx.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "tech_stacks:")
	meta := response.NewMeta(http.StatusOK, "success", "Tech stacks reordered successfully")
	resp := response.NewSingleDataResponse(meta, gin.H{"message": "Tech stacks swapped"})
	ctx.JSON(http.StatusOK, resp)
}

func techStackResponse(techStack db.TechStack) TechStackData {
	var catID *int64
	if techStack.TechStackCategoryID.Valid {
		v := techStack.TechStackCategoryID.Int64
		catID = &v
	}

	return TechStackData{
		ID:                  techStack.ID,
		Name:                techStack.Name,
		Slug:                techStack.Slug,
		Icon:                techStack.Icon.String,
		ColumnOrder:         techStack.ColumnOrder,
		CreatedAt:           techStack.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		UpdatedAt:           techStack.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
		TechStackCategoryID: catID,
	}
}

func TechStacksResponse(techStacks []db.TechStack) []TechStackData {
	data := make([]TechStackData, len(techStacks))
	for i, techStack := range techStacks {
		var catID *int64
		if techStack.TechStackCategoryID.Valid {
			v := techStack.TechStackCategoryID.Int64
			catID = &v
		}

		data[i] = TechStackData{
			ID:                  techStack.ID,
			Name:                techStack.Name,
			Slug:                techStack.Slug,
			Icon:                techStack.Icon.String,
			ColumnOrder:         techStack.ColumnOrder,
			CreatedAt:           techStack.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			UpdatedAt:           techStack.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
			TechStackCategoryID: catID,
		}
	}
	return data
}
