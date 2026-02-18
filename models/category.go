package models

import "time"

// Category represents a category entity
// @Description Category information with ID, name and description
type Category struct {
	ID          int       `json:"id" example:"1"`
	Name        string    `json:"name" example:"Electronics" binding:"required"`
	Description string    `json:"description" example:"Electronic devices and gadgets"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-30T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-30T12:00:00Z"`
}

// CategoryInput represents the input for creating/updating a category
// @Description Input model for creating or updating a category (ID is auto-generated)
type CategoryInput struct {
	Name        string `json:"name" example:"Electronics" binding:"required"`
	Description string `json:"description" example:"Electronic devices and gadgets"`
}