package api

import (
	"database/sql"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

type createProjectRequest struct {
	Title       string `form:"title" binding:"required"`
	Thumbnail   string `form:"thumbnail"`
	Content     string `form:"content"`
	DemoUrl     string `form:"demo_url"`
	RepoUrl     string `form:"repo_url"`
	CategoryID  int64  `form:"category_id" binding:"required"`
	PublishedAt string `form:"published_at"`

	TechStackIDs string `form:"tech_stack_ids"`
	TagIDs       string `form:"tag_ids"`
}

type projectData struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Thumbnail   string `json:"thumbnail"`
	Content     string `json:"content"`
	DemoUrl     string `json:"demo_url"`
	RepoUrl     string `json:"repo_url"`
	ColumnOrder int32  `json:"column_order"`
	CategoryID  int64  `json:"category_id"`
	PublishedAt string `json:"published_at"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`

	// Field Tambahan (Pointers & Omitted if Empty)
	// Kita pakai Pointer (*) supaya kalau tidak diminta, nilainya null (tidak muncul di JSON)
	Category   *db.Category    `json:"category,omitempty"`
	TechStacks []TechStackData `json:"tech_stacks,omitempty"`
	Tags       []TagData       `json:"tags,omitempty"`
}

func (server *Server) createProject(ctx *gin.Context) {
	var req createProjectRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//  Handle Upload Thumbnail
	var thumbnailURL string
	file, err := ctx.FormFile("thumbnail")
	folderName := "projects"
	if err == nil {
		url, errSave := util.SaveUploadedFile(ctx, file, folderName) // Reuse fungsi helper yg kita buat di profile
		if errSave != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload thumbnail"})
			return
		}
		thumbnailURL = url
	}

	techStackIDs, _ := util.ParseStringToIntArray(req.TechStackIDs)
	tagIDs, _ := util.ParseStringToIntArray(req.TagIDs)

	// 3. TRANSACTION BLOCK (ExecTx)
	// Kita pakai var 'result' untuk menampung data project yg berhasil dibuat
	var createdProject db.Project
	transactionError := server.store.ExecTx(ctx, func(queries *db.Queries) error {
		var executionError error

		// A. Insert Main Project
		projectParams := db.CreateProjectParams{
			Title:      req.Title,
			Slug:       slug.Make(req.Title),
			Content:    convertToNullString(req.Content),
			Thumbnail:  convertToNullString(thumbnailURL),
			RepoUrl:    convertToNullString(req.RepoUrl),
			DemoUrl:    convertToNullString(req.DemoUrl),
			CategoryID: sql.NullInt64{Int64: req.CategoryID, Valid: true},
		}

		createdProject, executionError = queries.CreateProject(ctx, projectParams)
		if executionError != nil {
			return executionError
		}

		// B. Loop & Insert Tech Stacks (Pivot)
		for _, techStackID := range techStackIDs {
			executionError = queries.AddTechStackToProject(ctx, db.AddTechStackToProjectParams{
				ProjectID:   createdProject.ID,
				TechStackID: techStackID, // Cast to int32 if SQLC generates int32
			})
			if executionError != nil {
				return executionError
			}
		}

		// C. Loop & Insert Tags (Pivot)
		for _, tagID := range tagIDs {
			executionError = queries.AddTagToProject(ctx, db.AddTagToProjectParams{
				ProjectID: createdProject.ID,
				TagID:     tagID, // Cast to int32
			})
			if executionError != nil {
				return executionError
			}
		}

		return nil // Commit Transaction
	})

	if transactionError != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": transactionError.Error()})
		return
	}

	data := projectResponse(createdProject)
	meta := response.NewMeta(http.StatusOK, "success", "Project created successfully")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, data))
}

func projectResponse(project db.Project) projectData {
	return projectData{
		ID:          project.ID,
		Title:       project.Title,
		Slug:        project.Slug,
		Thumbnail:   project.Thumbnail.String,
		Content:     project.Content.String,
		DemoUrl:     project.DemoUrl.String,
		RepoUrl:     project.RepoUrl.String,
		ColumnOrder: project.ColumnOrder,
		CategoryID:  project.CategoryID.Int64,
		PublishedAt: project.PublishedAt.Time.Format("2006-01-02 15:04:05"),
		CreatedAt:   project.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   project.UpdatedAt.Format("2006-01-02 15:04:05"),

		// Field Tambahan (Pointers & Omitted if Empty)
		// Kita pakai Pointer (*) supaya kalau tidak diminta, nilainya null (tidak muncul di JSON)
		Category:   nil,
		TechStacks: nil,
		Tags:       nil,
	}
}

func projectsResponse(projects []db.Project) []projectData {
	data := make([]projectData, len(projects))
	for i, project := range projects {
		data[i] = projectData{
			ID:          project.ID,
			Title:       project.Title,
			Slug:        project.Slug,
			Thumbnail:   project.Thumbnail.String,
			Content:     project.Content.String,
			DemoUrl:     project.DemoUrl.String,
			RepoUrl:     project.RepoUrl.String,
			ColumnOrder: project.ColumnOrder,
			CategoryID:  project.CategoryID.Int64,
			PublishedAt: project.PublishedAt.Time.Format("2006-01-02 15:04:05"),
			CreatedAt:   project.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   project.UpdatedAt.Format("2006-01-02 15:04:05"),

			Category:   nil,
			TechStacks: nil,
			Tags:       nil,
		}
	}
	return data
}
