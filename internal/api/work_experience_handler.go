package api

import (
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type workExperienceRequest struct {
	Company     string `form:"company" binding:"required"`
	Position    string `form:"position" binding:"required"`
	Location    string `form:"location"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	isCurrent   bool   `form:"is_current"`
	Description string `form:"description"`

	TechStackIDs string `form:"tech_stack_ids"`
	TagIDs       string `form:"tag_ids"`
}

type workExperienceData struct {
	Company     string `json:"company"`
	Position    string `json:"position"`
	Location    string `json:"location"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	IsCurrent   bool   `json:"is_current"`
	Description string `json:"description"`

	TechStacks []TechStackData `json:"tech_stacks,omitempty"`
	Tags       []TagData       `json:"tags,omitempty"`
}

func (server *Server) createWorkExperience(ctx *gin.Context) {
	var workExpRequest workExperienceRequest
	if err := ctx.ShouldBind(&workExpRequest); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	techStackIDs, _ := util.ParseStringToIntArray(workExpRequest.TechStackIDs)
	tagIDs, _ := util.ParseStringToIntArray(workExpRequest.TagIDs)

	var newWorkExperience db.WorkExperience
	transactionError := server.store.ExecTx(ctx, func(queries *db.Queries) error {
		var executionError error

		createWorkExpParam := db.CreateWorkExperienceParams{
			Company:     workExpRequest.Company,
			Position:    workExpRequest.Position,
			Location:    convertToNullString(workExpRequest.Location),
			StartDate:   convertToNullString(workExpRequest.StartDate),
			EndDate:     convertToNullString(workExpRequest.EndDate),
			IsCurrent:   convertToNullBool(workExpRequest.isCurrent),
			Description: convertToNullString(workExpRequest.Description),
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": transactionError.Error()})
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "workExperiences:")
	responseData := workExperienceResponse(newWorkExperience)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Work Experience created successfully")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, responseData))
}

func workExperienceResponse(workExperience db.WorkExperience) workExperienceData {
	return workExperienceData{
		Company:     workExperience.Company,
		Position:    workExperience.Position,
		Location:    workExperience.Location.String,
		StartDate:   workExperience.StartDate.String,
		EndDate:     workExperience.EndDate.String,
		IsCurrent:   workExperience.IsCurrent.Bool,
		Description: workExperience.Description.String,

		TechStacks: nil,
		Tags:       nil,
	}
}
