package mapper

import (
	"go-portfolio-api/db/sqlc"
	"go-portfolio-api/internal/dto"
)

// MapCategoryToData converts a database Category model to CategoryData DTO
func MapCategoryToData(category db.Category) dto.CategoryData {
	return dto.CategoryData{
		ID:        category.ID,
		Name:      category.Name,
		Slug:      category.Slug,
		Color:     category.Color.String,
		CreatedAt: category.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		UpdatedAt: category.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
	}
}

// MapCategoriesToData converts a slice of database Category models to CategoryData DTOs
func MapCategoriesToData(categories []db.Category) []dto.CategoryData {
	data := make([]dto.CategoryData, len(categories))
	for i, category := range categories {
		data[i] = MapCategoryToData(category)
	}
	return data
}
