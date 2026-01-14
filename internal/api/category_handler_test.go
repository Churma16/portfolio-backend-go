package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/dto"

	"github.com/gin-gonic/gin"
)

// MockCategoryService untuk testing handler tanpa logic service
type MockCategoryService struct {
	CreateCategoryFunc func(ctx context.Context, req dto.CreateCategoryRequest) (db.Category, error)
	GetCategoryFunc    func(ctx context.Context, id int64) (db.Category, error)
	GetCategoriesFunc  func(ctx context.Context) ([]db.Category, error)
	UpdateCategoryFunc func(ctx context.Context, id int64, req dto.UpdateCategoryRequest) (db.Category, error)
	DeleteCategoryFunc func(ctx context.Context, id int64) (db.Category, error)
}

func (m *MockCategoryService) CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (db.Category, error) {
	if m.CreateCategoryFunc != nil {
		return m.CreateCategoryFunc(ctx, req)
	}
	return db.Category{}, nil
}

func (m *MockCategoryService) GetCategory(ctx context.Context, id int64) (db.Category, error) {
	if m.GetCategoryFunc != nil {
		return m.GetCategoryFunc(ctx, id)
	}
	return db.Category{}, nil
}

func (m *MockCategoryService) GetCategories(ctx context.Context) ([]db.Category, error) {
	if m.GetCategoriesFunc != nil {
		return m.GetCategoriesFunc(ctx)
	}
	return []db.Category{}, nil
}

func (m *MockCategoryService) UpdateCategory(ctx context.Context, id int64, req dto.UpdateCategoryRequest) (db.Category, error) {
	if m.UpdateCategoryFunc != nil {
		return m.UpdateCategoryFunc(ctx, id, req)
	}
	return db.Category{}, nil
}

func (m *MockCategoryService) DeleteCategory(ctx context.Context, id int64) (db.Category, error) {
	if m.DeleteCategoryFunc != nil {
		return m.DeleteCategoryFunc(ctx, id)
	}
	return db.Category{}, nil
}

// Example Test
func TestShowCategoriesHandler(t *testing.T) {
	mockService := &MockCategoryService{
		GetCategoriesFunc: func(ctx context.Context) ([]db.Category, error) {
			return []db.Category{
				{ID: 1, Name: "Backend"},
				{ID: 2, Name: "Frontend"},
			}, nil
		},
	}

	server := &Server{
		categoryService: mockService,
	}

	router := gin.Default()
	router.GET("/categories", server.showCategories)

	req := httptest.NewRequest("GET", "/categories", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCreateCategoryHandler(t *testing.T) {
	mockService := &MockCategoryService{
		CreateCategoryFunc: func(ctx context.Context, req dto.CreateCategoryRequest) (db.Category, error) {
			return db.Category{
				ID:   1,
				Name: req.Name,
				Slug: "backend",
			}, nil
		},
	}

	server := &Server{
		categoryService: mockService,
	}

	router := gin.Default()
	router.POST("/categories", server.createCategory)

	body := bytes.NewBufferString("name=Backend&color=%23FF0000")
	req := httptest.NewRequest("POST", "/categories", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
