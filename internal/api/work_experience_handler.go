package api

import (
	"database/sql"
	"encoding/json"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sqlc-dev/pqtype"
)

type workExperienceRequest struct {
	Company      string      `json:"company" binding:"required"`
	Position     string      `json:"position" binding:"required"`
	Location     string      `json:"location"`
	StartDate    string      `json:"start_date"`
	EndDate      string      `json:"end_date"`
	IsCurrent    bool        `json:"is_current"`
	Description  string      `json:"description"`
	TechStackIDs interface{} `json:"tech_stack_ids"`
	TagIDs       interface{} `json:"tag_ids"`
	Achievements interface{} `json:"achievements"`
}

type workExperienceData struct {
	ID          int64           `json:"id"`
	Company     string          `json:"company"`
	Position    string          `json:"position"`
	Location    string          `json:"location"`
	StartDate   string          `json:"start_date"`
	EndDate     string          `json:"end_date"`
	IsCurrent   bool            `json:"is_current"`
	Description string          `json:"description"`
	Achievements json.RawMessage `json:"achievements"`
	ColumnOrder int32           `json:"column_order"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`

	TechStacks []TechStackData `json:"tech_stack,omitempty"`
	Tags       []TagData       `json:"tags,omitempty"`
}

func (server *Server) createWorkExperience(ctx *gin.Context) {
	var workExpRequest workExperienceRequest
	if err := ctx.ShouldBind(&workExpRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid request"))
		return
	}

	techStackIDs := util.ParseInterfaceToIntArray(workExpRequest.TechStackIDs)
	tagIDs := util.ParseInterfaceToIntArray(workExpRequest.TagIDs)

	var newWorkExperience db.WorkExperience
	transactionError := server.store.ExecTx(ctx, func(queries *db.Queries) error {
		var executionError error

		var achievementsRaw []byte
		if workExpRequest.Achievements != nil {
			achievementsRaw, _ = json.Marshal(workExpRequest.Achievements)
		}

		createWorkExpParam := db.CreateWorkExperienceParams{
			Company:      workExpRequest.Company,
			Position:     workExpRequest.Position,
			Location:     convertToNullString(workExpRequest.Location),
			StartDate:    convertToNullString(workExpRequest.StartDate),
			EndDate:      convertToNullString(workExpRequest.EndDate),
			IsCurrent:    convertToNullBool(workExpRequest.IsCurrent),
			Description:  convertToNullString(workExpRequest.Description),
			Achievements: pqtype.NullRawMessage{RawMessage: achievementsRaw, Valid: len(achievementsRaw) > 0},
		}

		newWorkExperience, executionError = queries.CreateWorkExperience(ctx, createWorkExpParam)
		if executionError != nil {
			return executionError
		}
		for _, techStackID := range techStackIDs {
			executionError = queries.AddTechStackToWorkExperience(ctx, db.AddTechStackToWorkExperienceParams{
				WorkExperienceID: newWorkExperience.ID,
				TechStackID:      techStackID,
			})
			if executionError != nil {
				return executionError
			}
		}

		for _, tagID := range tagIDs {
			executionError = queries.AddTagToWorkExperience(ctx, db.AddTagToWorkExperienceParams{
				WorkExperienceID: newWorkExperience.ID,
				TagID:            tagID,
			})
			if executionError != nil {
				return executionError
			}
		}
		return nil
	})

	if transactionError != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", transactionError.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "workExperiences:")
	responseData := workExperienceResponse(newWorkExperience)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Work Experience created successfully")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, responseData))
}

func (server *Server) showWorkExperiences(ctx *gin.Context) {
	cacheKey := "workExperiences:list:" + ctx.Request.URL.RequestURI()
	cacheValue, cacheErr := server.redisClient.Get(ctx, cacheKey).Result()
	if cacheErr == nil {
		// Cache hit: Data found in Redis
		var cachedWorkExperiences []workExperienceData
		unmarshalErr := json.Unmarshal([]byte(cacheValue), &cachedWorkExperiences)

		if unmarshalErr == nil {
			// Return cached data to the user
			meta := response.NewMetaWithCount(http.StatusOK, "success", "List Work Experiences retrieved (Cached)", len(cachedWorkExperiences))
			ctx.JSON(http.StatusOK, response.NewMultipleDataResponse(meta, cachedWorkExperiences))
			return
		}
	}

	queryParam := ctx.Query("with")

	workExperiences, retrievalError := server.store.GetWorkExperiences(ctx)
	if retrievalError != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", retrievalError.Error()))
		return
	}

	// Populate workExperienceIDs
	var workExperienceIDs []int32
	for _, workExperience := range workExperiences {
		workExperienceIDs = append(workExperienceIDs, int32(workExperience.ID))
	}

	techStackLookup := make(map[int64][]TechStackData)
	tagLookup := make(map[int64][]TagData)

	if strings.Contains(queryParam, "techStacks") && len(workExperienceIDs) > 0 {
		techStackResults, techStackError := server.store.GetTechStacksByWorkExperienceID(ctx, workExperienceIDs)
		if techStackError == nil {
			for _, techStackRow := range techStackResults {
				techStackData := TechStackData{
					ID:   techStackRow.ID,
					Name: techStackRow.Name,
					Slug: techStackRow.Slug,
					Icon: techStackRow.Icon.String,
				}
				techStackLookup[techStackRow.WorkExperienceID] = append(techStackLookup[techStackRow.WorkExperienceID], techStackData)
			}
		}
	}

	if strings.Contains(queryParam, "tags") && len(workExperienceIDs) > 0 {
		tagResults, tagError := server.store.GetTagsByWorkExperienceID(ctx, workExperienceIDs)
		if tagError == nil {
			for _, tagRow := range tagResults {
				tagData := TagData{
					ID:    tagRow.ID,
					Name:  tagRow.Name,
					Slug:  tagRow.Slug,
					Color: tagRow.Color.String,
				}
				tagLookup[tagRow.WorkExperienceID] = append(tagLookup[tagRow.WorkExperienceID], tagData)
			}
		}
	}

	var responseData []workExperienceData
	for _, workExperience := range workExperiences {
		workExpItem := workExperienceResponse(workExperience)
		if techStacks, exists := techStackLookup[workExperience.ID]; exists {
			workExpItem.TechStacks = techStacks
		}
		if tags, exists := tagLookup[workExperience.ID]; exists {
			workExpItem.Tags = tags
		}
		responseData = append(responseData, workExpItem)
	}

	cachedJsonData, _ := json.Marshal(responseData)
	server.redisClient.Set(ctx, cacheKey, cachedJsonData, 1*time.Hour)

	responseMeta := response.NewMetaWithCount(http.StatusOK, "success", "Work Experiences retrieved successfully", len(responseData))
	//data := workExperiencesResponse(workExperiences)
	ctx.JSON(http.StatusOK, response.NewMultipleDataResponse(responseMeta, responseData))
}

func (server *Server) showWorkExperience(ctx *gin.Context) {

	cacheKey := "workExperiences:single:" + ctx.Request.URL.RequestURI()
	cacheValue, cacheErr := server.redisClient.Get(ctx, cacheKey).Result()
	if cacheErr == nil {
		// Cache hit: Data found in Redis
		var cachedWorkExperience workExperienceData
		unmarshalErr := json.Unmarshal([]byte(cacheValue), &cachedWorkExperience)
		if unmarshalErr == nil {
			// Return cached data to the user
			meta := response.NewMeta(http.StatusOK, "success", "Work experience retrieved successfully (Cached)")
			ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, cachedWorkExperience))
			return
		}
	}
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid ID"))
		return
	}
	queryParam := ctx.Query("with")

	// Retrieve work experience
	workExperience, err := server.store.GetWorkExperience(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Work experience not found"))
		} else {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		}
		return
	}
	techStackLookup := make(map[int64][]TechStackData)
	tagLookup := make(map[int64][]TagData)

	if strings.Contains(queryParam, "techStacks") {
		techStackResults, techStackError := server.store.GetTechStacksByWorkExperienceID(ctx, []int32{int32(workExperience.ID)})
		if techStackError == nil {
			for _, techStackRow := range techStackResults {
				techStackData := TechStackData{
					ID:   techStackRow.ID,
					Name: techStackRow.Name,
					Slug: techStackRow.Slug,
					Icon: techStackRow.Icon.String,
				}
				techStackLookup[techStackRow.WorkExperienceID] = append(techStackLookup[techStackRow.WorkExperienceID], techStackData)
			}
		}
	}

	if strings.Contains(queryParam, "tags") {
		tagResults, tagError := server.store.GetTagsByWorkExperienceID(ctx, []int32{int32(workExperience.ID)})
		if tagError == nil {
			for _, tagRow := range tagResults {
				tagData := TagData{
					ID:    tagRow.ID,
					Name:  tagRow.Name,
					Slug:  tagRow.Slug,
					Color: tagRow.Color.String,
				}
				tagLookup[tagRow.WorkExperienceID] = append(tagLookup[tagRow.WorkExperienceID], tagData)
			}
		}
	}
	workExperienceItem := workExperienceResponse(workExperience)

	// Attach TechStacks
	if techStacks, exists := techStackLookup[workExperience.ID]; exists {
		workExperienceItem.TechStacks = techStacks
	}
	// Attach Tags
	if tags, exists := tagLookup[workExperience.ID]; exists {
		workExperienceItem.Tags = tags
	}

	cachedJsonData, _ := json.Marshal(workExperienceItem)
	server.redisClient.Set(ctx, cacheKey, cachedJsonData, 1*time.Hour)

	responseMeta := response.NewMeta(http.StatusOK, "success", "Work experience retrieved successfully")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, workExperienceItem))
}

func (server *Server) updateWorkExperience(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid ID"))
		return
	}

	var workExpRequest workExperienceRequest
	if err := ctx.ShouldBind(&workExpRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid request"))
		return
	}

	var updatedWorkExp db.WorkExperience
	txErr := server.store.ExecTx(ctx, func(queries *db.Queries) error {
		StartDate := util.FormatDate(convertToNullString(workExpRequest.StartDate), "2006-01-02")
		EndDate := util.FormatDate(convertToNullString(workExpRequest.EndDate), "2006-01-02")

		var achievementsRaw []byte
		if workExpRequest.Achievements != nil {
			achievementsRaw, _ = json.Marshal(workExpRequest.Achievements)
		}

		updateParams := db.UpdateWorkExperienceParams{
			ID:           id,
			Company:      workExpRequest.Company,
			Position:     workExpRequest.Position,
			Location:     convertToNullString(workExpRequest.Location),
			StartDate:    convertToNullString(StartDate),
			EndDate:      convertToNullString(EndDate),
			IsCurrent:    convertToNullBool(workExpRequest.IsCurrent),
			Description:  convertToNullString(workExpRequest.Description),
			Achievements: pqtype.NullRawMessage{RawMessage: achievementsRaw, Valid: len(achievementsRaw) > 0},
		}

		updatedWorkExp, err = queries.UpdateWorkExperience(ctx, updateParams)
		if err != nil {
			return err
		}

		// Update TechStacks and Tags (if provided)
		techStackIDs := util.ParseInterfaceToIntArray(workExpRequest.TechStackIDs)
		tagIDs := util.ParseInterfaceToIntArray(workExpRequest.TagIDs)

		if err := queries.DeleteWorkExperienceTechStacks(ctx, id); err != nil {
			return err
		}
		if err := queries.DeleteWorkExperienceTags(ctx, id); err != nil {
			return err
		}

		for _, techStackID := range techStackIDs {
			if err := queries.AddTechStackToWorkExperience(ctx, db.AddTechStackToWorkExperienceParams{
				WorkExperienceID: id,
				TechStackID:      techStackID,
			}); err != nil {
				return err
			}
		}

		for _, tagID := range tagIDs {
			if err := queries.AddTagToWorkExperience(ctx, db.AddTagToWorkExperienceParams{
				WorkExperienceID: id,
				TagID:            tagID,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	if txErr != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", txErr.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "workExperiences:")

	meta := response.NewMeta(http.StatusOK, "success", "Work experience updated successfully")
	data := workExperienceResponse(updatedWorkExp)
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, data))
}

func (server *Server) deleteWorkExperience(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", "Invalid ID"))
		return
	}

	deletedWorkExp, err := server.store.DeleteWorkExperience(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "workExperiences:")
	meta := response.NewMeta(http.StatusOK, "success", "Work Experience deleted successfully")
	data := workExperienceResponse(deletedWorkExp)
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, data))
}

func workExperienceResponse(workExperience db.WorkExperience) workExperienceData {
	return workExperienceData{
		ID:           workExperience.ID,
		Company:      workExperience.Company,
		Position:     workExperience.Position,
		Location:     workExperience.Location.String,
		StartDate:    util.FormatDate(workExperience.StartDate, "Jan 2006"),
		EndDate:      util.FormatDate(workExperience.EndDate, "Jan 2006"),
		IsCurrent:    workExperience.IsCurrent.Bool,
		Description:  workExperience.Description.String,
		Achievements: workExperience.Achievements.RawMessage,
		ColumnOrder:  workExperience.ColumnOrder,
		CreatedAt:    workExperience.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    workExperience.UpdatedAt.Format("2006-01-02 15:04:05"),

		TechStacks: nil,
		Tags:       nil,
	}
}

func workExperiencesResponse(workExperiences []db.WorkExperience) []workExperienceData {
	data := make([]workExperienceData, len(workExperiences))
	for i, workExperience := range workExperiences {
		data[i] = workExperienceResponse(workExperience)
	}
	return data
}

type reorderWorkExperienceRequest struct {
	Direction string `json:"direction" binding:"required,oneof=up down"`
}

func (server *Server) reorderWorkExperiences(ctx *gin.Context) {
	idParam := ctx.Param("id")
	workExpID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid work experience ID"})
		return
	}

	var req reorderWorkExperienceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentWorkExp, err := server.store.GetWorkExperience(ctx, workExpID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Work experience not found"})
		return
	}

	currentOrder := currentWorkExp.ColumnOrder

	var targetOrder int32
	var adjacentWorkExp db.WorkExperience

	if req.Direction == "up" {
		targetOrder = currentOrder - 1
		if targetOrder < 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot move work experience further up"})
			return
		}
		adjacentWorkExp, err = server.store.GetWorkExperienceByColumnOrder(ctx, targetOrder)
	} else { // down
		targetOrder = currentOrder + 1
		adjacentWorkExp, err = server.store.GetWorkExperienceByColumnOrder(ctx, targetOrder)
	}

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot move work experience further in that direction"})
		return
	}

	errTx := server.store.ExecTx(ctx, func(q *db.Queries) error {
		swapParams1 := db.UpdateWorkExperienceColumnOrderParams{
			ID:          currentWorkExp.ID,
			ColumnOrder: targetOrder,
		}

		swapParams2 := db.UpdateWorkExperienceColumnOrderParams{
			ID:          adjacentWorkExp.ID,
			ColumnOrder: currentOrder,
		}

		if _, err := q.UpdateWorkExperienceColumnOrder(ctx, swapParams1); err != nil {
			return err
		}
		if _, err := q.UpdateWorkExperienceColumnOrder(ctx, swapParams2); err != nil {
			return err
		}

		return nil
	})

	if errTx != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errTx.Error()})
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "workExperiences:")
	meta := response.NewMeta(http.StatusOK, "success", "Work experiences reordered successfully")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, gin.H{"message": "Work experiences swapped"}))
}
