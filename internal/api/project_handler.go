package api

import (
	"database/sql"
	"encoding/json"
	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/dto"
	"go-portfolio-api/internal/response"
	"go-portfolio-api/internal/util"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

type createProjectRequest struct {
	Title       string `json:"title" form:"title" binding:"required"`
	Content     string `json:"content" form:"content"`
	DemoUrl     string `json:"demo_url" form:"demo_url"`
	RepoUrl     string `json:"repo_url" form:"repo_url"`
	CategoryID  int64  `json:"category_id" form:"category_id" binding:"required"`
	PublishedAt string `json:"published_at" form:"published_at"`

	TechStackIDs interface{} `json:"tech_stack_ids" form:"tech_stack_ids"` // Can be array or string
	TagIDs       interface{} `json:"tag_ids" form:"tag_ids"`               // Can be array or string
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
	Category   *dto.CategoryData `json:"category,omitempty"`
	TechStacks []TechStackData   `json:"tech_stack,omitempty"`
	Tags       []TagData         `json:"tags,omitempty"`
}

func (server *Server) createProject(ctx *gin.Context) {
	// Handle Upload Thumbnail
	var projectRequest createProjectRequest
	if err := ctx.ShouldBind(&projectRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var thumbnailFileURL string

	uploadedThumbnailFile, fileError := ctx.FormFile("thumbnail")
	projectFolderName := "projects"
	if fileError == nil {

		savedThumbnailURL, saveError := util.SaveUploadedFile(ctx, uploadedThumbnailFile, projectFolderName)
		if saveError != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload thumbnail"})
			return
		}
		thumbnailFileURL = savedThumbnailURL
	} else {
	}

	techStackIDs := util.ParseInterfaceToIntArray(projectRequest.TechStackIDs)
	tagIDs := util.ParseInterfaceToIntArray(projectRequest.TagIDs)

	println(techStackIDs, tagIDs)
	// 3. TRANSACTION BLOCK (ExecTx)
	var newProject db.Project
	transactionError := server.store.ExecTx(ctx, func(queries *db.Queries) error {
		var executionError error

		// A. Insert Main Project
		createProjectParams := db.CreateProjectParams{
			Title:     projectRequest.Title,
			Slug:      slug.Make(projectRequest.Title),
			Content:   convertToNullString(projectRequest.Content),
			Thumbnail: convertToNullString(thumbnailFileURL),
			RepoUrl:   convertToNullString(projectRequest.RepoUrl),
			DemoUrl:   convertToNullString(projectRequest.DemoUrl), CategoryID: sql.NullInt64{Int64: projectRequest.CategoryID, Valid: true},
		}

		newProject, executionError = queries.CreateProject(ctx, createProjectParams)
		if executionError != nil {
			return executionError
		}

		// B. Loop & Insert Tech Stacks (Pivot)
		for _, techStackID := range techStackIDs {
			executionError = queries.AddTechStackToProject(ctx, db.AddTechStackToProjectParams{
				ProjectID:   newProject.ID,
				TechStackID: techStackID,
			})
			if executionError != nil {
				return executionError
			}
		}

		// C. Loop & Insert Tags (Pivot)
		for _, tagID := range tagIDs {
			executionError = queries.AddTagToProject(ctx, db.AddTagToProjectParams{
				ProjectID: newProject.ID,
				TagID:     tagID,
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
	println("  Project ID:", newProject.ID)
	println("  Project thumbnail from DB:", newProject.Thumbnail.String)

	responseData := projectResponse(newProject)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Project created successfully")
	println("========== CREATE PROJECT DEBUG END (SUCCESS) ==========\n\n")

	util.DeleteCacheByPrefix(server.redisClient, "projects:")

	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, responseData))
}

func (server *Server) showProjects(context *gin.Context) {
	// Improved variable readability for Redis caching logic
	cacheKey := "projects:list:" + context.Request.URL.RequestURI()
	cacheValue, cacheErr := server.redisClient.Get(context, cacheKey).Result()
	if cacheErr == nil {
		// Cache hit: Data found in Redis
		var cachedProjects []projectData
		unmarshalErr := json.Unmarshal([]byte(cacheValue), &cachedProjects)

		if unmarshalErr == nil {
			// Return cached data to the user
			meta := response.NewMetaWithCount(http.StatusOK, "success", "List projects retrieved (Cached)", len(cachedProjects))
			context.JSON(http.StatusOK, response.NewMultipleDataResponse(meta, cachedProjects))
			return
		}
	}

	queryParam := context.Query("with")

	// 1. Retrieve Projects
	projects, retrievalError := server.store.GetProjects(context)
	if retrievalError != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": retrievalError.Error()})
		return
	}

	// 2. Collect IDs
	// Collect project IDs and category IDs for batch queries
	projectIDs := []int32{}
	categoryIDs := []int32{}

	for _, project := range projects {
		projectIDs = append(projectIDs, int32(project.ID))
		if project.CategoryID.Valid {
			categoryIDs = append(categoryIDs, int32(project.CategoryID.Int64))
		}
	}

	// 3. Prepare Maps for Temporary Storage
	// These maps allow quick data lookup (O(1)) without additional looping
	techStackLookup := make(map[int64][]TechStackData)  // Key: ProjectID, Value: List of TechStacks
	tagLookup := make(map[int64][]TagData)              // Key: ProjectID, Value: List of Tags
	categoryLookup := make(map[int64]*dto.CategoryData) // Key: CategoryID, Value: Category

	// --- Batch Query: Categories ---
	if strings.Contains(queryParam, "category") && len(categoryIDs) > 0 {
		categories, categoryError := server.store.GetCategoriesByIDs(context, categoryIDs)
		if categoryError == nil {
			for _, category := range categories {
				// Map DB Model -> DTO
				categoryLookup[category.ID] = &dto.CategoryData{
					ID:   category.ID,
					Name: category.Name,
					Slug: category.Slug,
					// THIS FIXES THE ISSUE: Extract the String value
					Color:     category.Color.String,
					CreatedAt: category.CreatedAt.Format("2006-01-02 15:04:05"),
					UpdatedAt: category.UpdatedAt.Format("2006-01-02 15:04:05"),
				}
			}
		}
	}

	// --- Batch Query: Tech Stacks ---
	if strings.Contains(queryParam, "techStacks") && len(projectIDs) > 0 {
		techStackResults, techStackError := server.store.GetTechStacksByProjectID(context, projectIDs)
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
				// Add to map based on Project ID
				techStackLookup[techStackRow.ProjectID] = append(techStackLookup[techStackRow.ProjectID], techStackData)
			}
		}
	}

	// --- Batch Query: Tags ---
	if strings.Contains(queryParam, "tags") && len(projectIDs) > 0 {
		tagResults, tagError := server.store.GetTagsByProjectID(context, projectIDs)
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

	// Store As Cached Data in Redis (with 1 hour expiration)
	cachedJsonData, _ := json.Marshal(projectResponseList)
	server.redisClient.Set(context, cacheKey, cachedJsonData, 1*time.Hour)

	// 5. Return JSON Response
	responseMetadata := response.NewMetaWithCount(http.StatusOK, "success", "List projects retrieved", len(projectResponseList))
	context.JSON(http.StatusOK, response.NewMultipleDataResponse(responseMetadata, projectResponseList))
}

func (server *Server) showProject(context *gin.Context) {
	// Improved Redis caching logic for showProject
	cacheKey := "projects:single:" + context.Request.URL.RequestURI()
	cacheValue, cacheErr := server.redisClient.Get(context, cacheKey).Result()
	if cacheErr == nil {
		// Cache hit: Data found in Redis
		var cachedProject projectData
		unmarshalErr := json.Unmarshal([]byte(cacheValue), &cachedProject)
		if unmarshalErr == nil {
			// Return cached data to the user
			meta := response.NewMeta(http.StatusOK, "success", "Project retrieved (Cached)")
			context.JSON(http.StatusOK, response.NewSingleDataResponse(meta, cachedProject))
			return
		}
	}
	idParam := context.Param("id")
	projectID, parseError := strconv.ParseInt(idParam, 10, 64)
	if parseError != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	queryParam := context.Query("with")

	// 1. Retrieve Project by ID
	project, retrievalError := server.store.GetProject(context, projectID)
	if retrievalError != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 2. Prepare Maps for Related Data
	techStackLookup := make(map[int64][]TechStackData)
	tagLookup := make(map[int64][]TagData)
	var category *dto.CategoryData // Update type to DTO

	// --- Query: Category ---
	if strings.Contains(queryParam, "category") && project.CategoryID.Valid {
		categoryResult, categoryError := server.store.GetCategory(context, project.CategoryID.Int64)
		if categoryError == nil {
			// Map DB Model -> DTO
			category = &dto.CategoryData{
				ID:   categoryResult.ID,
				Name: categoryResult.Name,
				Slug: categoryResult.Slug,
				// THIS FIXES THE ISSUE: Extract the String value
				Color:     categoryResult.Color.String,
				CreatedAt: categoryResult.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: categoryResult.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
		}
	}

	// --- Query: Tech Stacks ---
	if strings.Contains(queryParam, "techStacks") {
		techStackResults, techStackError := server.store.GetTechStacksByProjectID(context, []int32{int32(project.ID)})
		if techStackError == nil {
			for _, techStackRow := range techStackResults {
				techStackData := TechStackData{
					ID:   techStackRow.ID,
					Name: techStackRow.Name,
					Slug: techStackRow.Slug,
					Icon: techStackRow.Icon.String,
				}
				techStackLookup[techStackRow.ProjectID] = append(techStackLookup[techStackRow.ProjectID], techStackData)
			}
		}
	}

	// --- Query: Tags ---
	if strings.Contains(queryParam, "tags") {
		tagResults, tagError := server.store.GetTagsByProjectID(context, []int32{int32(project.ID)})
		if tagError == nil {
			for _, tagRow := range tagResults {
				tagData := TagData{
					ID:    tagRow.ID,
					Name:  tagRow.Name,
					Slug:  tagRow.Slug,
					Color: tagRow.Color.String,
				}
				tagLookup[tagRow.ProjectID] = append(tagLookup[tagRow.ProjectID], tagData)
			}
		}
	}

	// 3. Assemble Final Response
	projectItem := projectResponse(project)

	// Attach Category
	if category != nil {
		projectItem.Category = category
	}

	// Attach TechStacks
	if techStacks, exists := techStackLookup[project.ID]; exists {
		projectItem.TechStacks = techStacks
	}

	// Attach Tags
	if tags, exists := tagLookup[project.ID]; exists {
		projectItem.Tags = tags
	}

	//cachedJsonData, _ := json.Marshal(projectItem)
	//server.redisClient.Set(context, cacheKey, cachedJsonData, 1*time.Hour)
	// 4. Return JSON Response
	responseMetadata := response.NewMeta(http.StatusOK, "success", "Project retrieved successfully")
	context.JSON(http.StatusOK, response.NewSingleDataResponse(responseMetadata, projectItem))
}

func (server *Server) updateProject(ctx *gin.Context) {
	// 1. Extract Project ID from URL (/projects/:id)
	var projectID struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := ctx.ShouldBindUri(&projectID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Parse Form Data (Reuse struct createProjectRequest)
	var req createProjectRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Retrieve Existing Project Data (Needed for old thumbnail if no new upload)
	existingProject, err := server.store.GetProject(ctx, projectID.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// 4. Handle Thumbnail (Use new upload or fallback to old thumbnail)
	thumbnailURL := existingProject.Thumbnail.String
	uploadedFile, err := ctx.FormFile("thumbnail")
	if err == nil {
		// New thumbnail uploaded
		const folderName = "projects"
		// Delete the old thumbnail if it exists
		if existingProject.Thumbnail.Valid {
			if err := util.DeleteFile(existingProject.Thumbnail.String); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old thumbnail"})
				return
			}
		}
		url, saveErr := util.SaveUploadedFile(ctx, uploadedFile, folderName)
		if saveErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload thumbnail"})
			return
		}
		thumbnailURL = url
	}

	// 5. Convert Array IDs from Interface to Integers
	techStackIDs := util.ParseInterfaceToIntArray(req.TechStackIDs)
	tagIDs := util.ParseInterfaceToIntArray(req.TagIDs)

	// 6. TRANSACTION BLOCK (Wipe & Replace Strategy)
	var updatedProject db.Project

	errTx := server.store.ExecTx(ctx, func(q *db.Queries) error {
		var err error

		// A. Update Main Project Data
		updateArgs := db.UpdateProjectParams{
			ID:         projectID.ID,
			Title:      req.Title,
			Slug:       slug.Make(req.Title),
			Content:    convertToNullString(req.Content),
			Thumbnail:  convertToNullString(thumbnailURL),
			RepoUrl:    convertToNullString(req.RepoUrl),
			DemoUrl:    convertToNullString(req.DemoUrl),
			CategoryID: sql.NullInt64{Int64: req.CategoryID, Valid: true},
		}

		updatedProject, err = q.UpdateProject(ctx, updateArgs)
		if err != nil {
			return err
		}

		// B. WIPE: Remove All Old Relationships
		if err = q.DeleteProjectTechStacks(ctx, projectID.ID); err != nil {
			return err
		}
		if err = q.DeleteProjectTags(ctx, projectID.ID); err != nil {
			return err
		}

		// C. REPLACE: Insert New Relationships (Similar to Create)
		for _, techStackID := range techStackIDs {
			if err = q.AddTechStackToProject(ctx, db.AddTechStackToProjectParams{
				ProjectID:   projectID.ID,
				TechStackID: techStackID,
			}); err != nil {
				return err
			}
		}

		for _, tagID := range tagIDs {
			if err = q.AddTagToProject(ctx, db.AddTagToProjectParams{
				ProjectID: projectID.ID,
				TagID:     tagID,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	if errTx != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": errTx.Error()})
		return
	}

	util.DeleteCacheByPrefix(server.redisClient, "projects:")
	// 7. Return Response
	responseData := projectResponse(updatedProject)
	responseMeta := response.NewMeta(http.StatusOK, "success", "Project updated successfully")
	ctx.JSON(http.StatusOK, response.NewSingleDataResponse(responseMeta, responseData))
}

func (server *Server) deleteProject(ctx *gin.Context) {
	// 1. Ambil ID dari URL
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, response.ErrorResponse(http.StatusBadRequest, "error", err.Error()))
		return
	}

	// 2. Cek Apakah Project Ada? (Optional, tapi bagus buat UX biar bisa return 404)
	_, err := server.store.GetProject(ctx, uri.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, response.ErrorResponse(http.StatusNotFound, "error", "Project not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	// 3. Eksekusi Hapus
	// Berkat ON DELETE CASCADE, Tech Stack & Tags ikut terhapus otomatis.
	project, err := server.store.DeleteProject(ctx, uri.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	if project.Thumbnail.Valid {
		if err := util.DeleteFile(project.Thumbnail.String); err != nil {
			ctx.JSON(http.StatusInternalServerError, response.ErrorResponse(http.StatusInternalServerError, "error", "Failed to delete project thumbnail"))
			return
		}
	}

	// 4. Return Success
	util.DeleteCacheByPrefix(server.redisClient, "projects:")
	meta := response.NewMeta(http.StatusOK, "success", "Project deleted successfully")
	data := projectResponse(project)
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
