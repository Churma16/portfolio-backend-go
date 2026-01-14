package service

import (
	"context"

	db "go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/dto"
	"go-portfolio-api/internal/util"
)

// CategoryService defines the interface for category business logic
type CategoryService interface {
	CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (db.Category, error)
	GetCategories(ctx context.Context) ([]db.Category, error)
	GetCategory(ctx context.Context, id int64) (db.Category, error)
	UpdateCategory(ctx context.Context, id int64, req dto.UpdateCategoryRequest) (db.Category, error)
	DeleteCategory(ctx context.Context, id int64) (db.Category, error)
}

// categoryService implements CategoryService interface
type categoryService struct {
	store *db.Store
}

// NewCategoryService creates a new instance of CategoryService
func NewCategoryService(store *db.Store) CategoryService {
	return &categoryService{store: store}
}

// CreateCategory creates a new category
func (s *categoryService) CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (db.Category, error) {
	arg := db.CreateCategoryParams{
		Name:  req.Name,
		Slug:  util.GenerateSlug(req.Name),
		Color: util.ConvertToNullString(req.Color),
	}

	return s.store.CreateCategory(ctx, arg)
}

// GetCategories retrieves all categories
func (s *categoryService) GetCategories(ctx context.Context) ([]db.Category, error) {
	return s.store.GetCategories(ctx)
}

// GetCategory retrieves a single category by ID
func (s *categoryService) GetCategory(ctx context.Context, id int64) (db.Category, error) {
	return s.store.GetCategory(ctx, id)
}

// UpdateCategory updates an existing category
func (s *categoryService) UpdateCategory(ctx context.Context, id int64, req dto.UpdateCategoryRequest) (db.Category, error) {
	existingCategory, err := s.store.GetCategory(ctx, id)
	if err != nil {
		return db.Category{}, err
	}

	arg := db.UpdateCategoryParams{
		ID:    id,
		Color: util.ConvertToNullString(req.Color),
	}

	if req.Name != existingCategory.Name {
		arg.Name = req.Name
		arg.Slug = util.GenerateSlug(req.Name)
	} else {
		arg.Name = existingCategory.Name
		arg.Slug = existingCategory.Slug
	}

	return s.store.UpdateCategory(ctx, arg)
}

// DeleteCategory deletes a category by ID
func (s *categoryService) DeleteCategory(ctx context.Context, id int64) (db.Category, error) {
	return s.store.DeleteCategory(ctx, id)
}
