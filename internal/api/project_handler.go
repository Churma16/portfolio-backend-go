package api

import (
	"database/sql"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

type createProjectRequest struct {
	Title       string `form:"title" binding:"required"`
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
	println("\n\n========== CREATE PROJECT DEBUG START ==========")
	println("STEP 1: Checking Content-Type")
	contentType := ctx.Request.Header.Get("Content-Type")
	println("  Content-Type:", contentType)

	//  Handle Upload Thumbnail
	var req createProjectRequest
	println("STEP 2: Attempting to ShouldBind()")
	if err := ctx.ShouldBind(&req); err != nil {
		println("  ERROR BINDING:", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		println("========== CREATE PROJECT DEBUG END (ERROR BINDING) ==========\n\n")
		return
	}
	println("  ✓ ShouldBind SUCCESS")
	println("  Parsed title:", req.Title)

	println("STEP 3: Attempting to get thumbnail file")
	var thumbnailURL string
	file, err := ctx.FormFile("thumbnail")
	folderName := "projects"
	if err == nil {
		println("  ✓ Thumbnail file found!")
		println("  Thumbnail filename:", file.Filename)
		println("  Thumbnail size:", file.Size)

		println("  Calling SaveUploadedFile for thumbnail...")
		url, errSave := util.SaveUploadedFile(ctx, file, folderName)
		if errSave != nil {
			println("  ERROR saving thumbnail:", errSave.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload thumbnail"})
			println("========== CREATE PROJECT DEBUG END (THUMBNAIL ERROR) ==========\n\n")
			return
		}
		thumbnailURL = url
		println("  ✓ Thumbnail saved with URL:", thumbnailURL)
	} else {
		println("  ✗ Thumbnail NOT found:", err.Error())
	}

	techStackIDs, _ := util.ParseStringToIntArray(req.TechStackIDs)
	tagIDs, _ := util.ParseStringToIntArray(req.TagIDs)

	// 3. TRANSACTION BLOCK (ExecTx)
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
		println("  ERROR in transaction:", transactionError.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": transactionError.Error()})
		println("========== CREATE PROJECT DEBUG END (TX ERROR) ==========\n\n")
		return
	}

	println("STEP 4: Project created successfully")
	println("  Project ID:", createdProject.ID)
	println("  Project thumbnail from DB:", createdProject.Thumbnail.String)

	data := projectResponse(createdProject)
	meta := response.NewMeta(http.StatusOK, "success", "Project created successfully")
	println("========== CREATE PROJECT DEBUG END (SUCCESS) ==========\n\n")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(meta, data))
}
func (server *Server) GetProjects(ctx *gin.Context) {
	// 1. Ambil Query Parameter "with"
	// Contoh: /projects?with=category,techStacks,tags
	withParam := ctx.Query("with")

func (server *Server) showProjects(context *gin.Context) {
	queryParam := context.Query("with")

	// 1. Retrieve Projects
	projects, retrievalError := server.store.GetProjects(context)
	if retrievalError != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": retrievalError.Error()})
		return
	}

	// 2. Collect IDs
	// Collect project IDs and category IDs for batch queries
	var projectIDList []int32
	var categoryIDList []int32

	for _, project := range projects {
		projectIDList = append(projectIDList, int32(project.ID))
		if project.CategoryID.Valid {
			categoryIDList = append(categoryIDList, int32(project.CategoryID.Int64))
		}
	}

	// 3. Prepare Maps for Temporary Storage
	// These maps allow quick data lookup (O(1)) without additional looping
	techStackLookup := make(map[int64][]TechStackData) // Key: ProjectID, Value: List of TechStacks
	tagLookup := make(map[int64][]TagData)             // Key: ProjectID, Value: List of Tags
	categoryLookup := make(map[int64]*db.Category)     // Key: CategoryID, Value: Category

	// --- Batch Query: Categories ---
	if strings.Contains(queryParam, "category") && len(categoryIDList) > 0 {
		categories, categoryError := server.store.GetCategoriesByIDs(context, categoryIDList)
		if categoryError == nil {
			for _, category := range categories {
				categoryCopy := category // Copy variable to ensure pointer safety
				categoryLookup[category.ID] = &categoryCopy
			}
		}
	}

	// --- Batch Query: Tech Stacks ---
	if strings.Contains(queryParam, "techStacks") && len(projectIDList) > 0 {
		techStackResults, techStackError := server.store.GetTechStacksByProjectID(context, projectIDList)
		if techStackError == nil {
			for _, techStackRow := range techStackResults {
				// Convert row to response struct
				techStackData := TechStackData{
					ID:   techStackRow.ID,
					Name: techStackRow.Name,
					Slug: techStackRow.Slug,
					Icon: techStackRow.Icon.String,
					// Add other fields as needed
				}
				item.TechStacks = tsData
			}
		}
				// Add to map based on Project ID
				techStackLookup[techStackRow.ProjectID] = append(techStackLookup[techStackRow.ProjectID], techStackData)
			}
		}
	}

	// --- Batch Query: Tags ---
	if strings.Contains(queryParam, "tags") && len(projectIDList) > 0 {
		tagResults, tagError := server.store.GetTagsByProjectID(context, projectIDList)
		if tagError == nil {
			for _, tagRow := range tagResults {
				tagData := TagData{
					ID:    tagRow.ID,
					Name:  tagRow.Name,
					Slug:  tagRow.Slug,
					Color: tagRow.Color.String,
					// Add other fields as needed
				}
				tagLookup[tagRow.ProjectID] = append(tagLookup[tagRow.ProjectID], tagData)
			}
		}
	}

	// 4. Assemble Final Response
	var projectResponseList []projectData

	for _, project := range projects {
		projectItem := projectResponse(project)

		// Attach Category (from map)
		if category, exists := categoryLookup[project.CategoryID.Int64]; exists {
			projectItem.Category = category
		}

		// Attach TechStacks (from map)
		if techStacks, exists := techStackLookup[project.ID]; exists {
			projectItem.TechStacks = techStacks
		}

		// Attach Tags (from map)
		if tags, exists := tagLookup[project.ID]; exists {
			projectItem.Tags = tags
		}

		projectResponseList = append(projectResponseList, projectItem)
	}

	// 5. Return JSON Response
	responseMetadata := response.NewMetaWithCount(http.StatusOK, "success", "List projects retrieved", len(projectResponseList))
	context.JSON(http.StatusOK, response.NewMultipleDataResponse(responseMetadata, projectResponseList))
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
