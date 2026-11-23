package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "go-portfolio-api/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestProjectResponse_IsFeaturedAndEmptyTags(t *testing.T) {
	now := time.Now()
	testProject := db.Project{
		ID:          10,
		Title:       "Test Portfolio Project",
		Slug:        "test-portfolio-project",
		Thumbnail:   pgtype.Text{String: "thumbnail.png", Valid: true},
		Content:     pgtype.Text{String: "Project description", Valid: true},
		DemoUrl:     pgtype.Text{String: "https://demo.example.com", Valid: true},
		RepoUrl:     pgtype.Text{String: "https://github.com/example/repo", Valid: true},
		IsFeatured:  pgtype.Bool{Bool: true, Valid: true},
		ColumnOrder: 5,
		CategoryID:  pgtype.Int8{Int64: 2, Valid: true},
		PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}

	response := projectResponse(testProject)

	// 1. Assert IsFeatured is restored
	if !response.IsFeatured {
		t.Errorf("expected IsFeatured to be true, got %v", response.IsFeatured)
	}

	// 2. Assert Tags is initialized as an empty slice (not nil)
	if response.Tags == nil {
		t.Errorf("expected Tags to be non-nil slice, got nil")
	}
	if len(response.Tags) != 0 {
		t.Errorf("expected empty Tags slice, got length %d", len(response.Tags))
	}

	// 3. Assert JSON serialization outputs "is_featured":true and "tags":[]
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal projectResponse: %v", err)
	}

	jsonString := string(jsonData)
	if !strings.Contains(jsonString, `"is_featured":true`) {
		t.Errorf("expected JSON to contain '\"is_featured\":true', got: %s", jsonString)
	}
	if !strings.Contains(jsonString, `"tags":[]`) {
		t.Errorf("expected JSON to contain '\"tags\":[]', got: %s", jsonString)
	}

	// Test unfeatured project
	testProject.IsFeatured = pgtype.Bool{Bool: false, Valid: true}
	unfeaturedResponse := projectResponse(testProject)
	if unfeaturedResponse.IsFeatured {
		t.Errorf("expected IsFeatured to be false, got %v", unfeaturedResponse.IsFeatured)
	}
	unfeaturedJSON, _ := json.Marshal(unfeaturedResponse)
	if !strings.Contains(string(unfeaturedJSON), `"is_featured":false`) {
		t.Errorf("expected JSON to contain '\"is_featured\":false', got: %s", string(unfeaturedJSON))
	}
}

func TestProjectsResponse_MultipleProjectsDefaults(t *testing.T) {
	now := time.Now()
	projects := []db.Project{
		{
			ID:          1,
			Title:       "Project A",
			Slug:        "project-a",
			IsFeatured:  pgtype.Bool{Bool: true, Valid: true},
			ColumnOrder: 1,
			PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		},
		{
			ID:          2,
			Title:       "Project B",
			Slug:        "project-b",
			IsFeatured:  pgtype.Bool{Bool: false, Valid: true},
			ColumnOrder: 7, // Non-consecutive gap
			PublishedAt: pgtype.Timestamptz{Time: now, Valid: true},
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		},
	}

	items := projectsResponse(projects)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if !items[0].IsFeatured {
		t.Errorf("expected item[0].IsFeatured true, got false")
	}
	if items[0].Tags == nil {
		t.Errorf("expected item[0].Tags non-nil, got nil")
	}

	if items[1].IsFeatured {
		t.Errorf("expected item[1].IsFeatured false, got true")
	}
	if items[1].Tags == nil {
		t.Errorf("expected item[1].Tags non-nil, got nil")
	}

	jsonData, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("failed to marshal projectsResponse: %v", err)
	}
	jsonString := string(jsonData)
	if !strings.Contains(jsonString, `"is_featured":true`) {
		t.Errorf("expected JSON to contain '\"is_featured\":true', got: %s", jsonString)
	}
	if !strings.Contains(jsonString, `"is_featured":false`) {
		t.Errorf("expected JSON to contain '\"is_featured\":false', got: %s", jsonString)
	}
	if !strings.Contains(jsonString, `"tags":[]`) {
		t.Errorf("expected JSON to contain '\"tags\":[]', got: %s", jsonString)
	}
}

func TestReorderProjectRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/projects/:id/reorder", func(c *gin.Context) {
		var req reorderProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"direction": req.Direction})
	})

	// Valid direction: up
	reqValidUp := httptest.NewRequest("POST", "/projects/1/reorder", strings.NewReader(`{"direction":"up"}`))
	reqValidUp.Header.Set("Content-Type", "application/json")
	wValidUp := httptest.NewRecorder()
	router.ServeHTTP(wValidUp, reqValidUp)
	if wValidUp.Code != http.StatusOK {
		t.Errorf("expected status 200 for 'up', got %d", wValidUp.Code)
	}

	// Valid direction: down
	reqValidDown := httptest.NewRequest("POST", "/projects/1/reorder", strings.NewReader(`{"direction":"down"}`))
	reqValidDown.Header.Set("Content-Type", "application/json")
	wValidDown := httptest.NewRecorder()
	router.ServeHTTP(wValidDown, reqValidDown)
	if wValidDown.Code != http.StatusOK {
		t.Errorf("expected status 200 for 'down', got %d", wValidDown.Code)
	}

	// Invalid direction: left
	reqInvalid := httptest.NewRequest("POST", "/projects/1/reorder", strings.NewReader(`{"direction":"left"}`))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()
	router.ServeHTTP(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for 'left', got %d", wInvalid.Code)
	}
}
