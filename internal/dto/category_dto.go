package dto

// CreateCategoryRequest represents the request body for creating a category
type CreateCategoryRequest struct {
	Name  string `form:"name" binding:"required"`
	Slug  string `form:"slug"`
	Color string `form:"color"`
}

// UpdateCategoryRequest represents the request body for updating a category
type UpdateCategoryRequest struct {
	Name  string `form:"name" binding:"required"`
	Slug  string `form:"slug"`
	Color string `form:"color"`
}

// CategoryData represents the response data for a category
type CategoryData struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
