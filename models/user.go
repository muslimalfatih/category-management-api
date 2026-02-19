package models

import "time"

// User represents a user entity
// @Description User information with identity and role
type User struct {
	ID        int       `json:"id" example:"1"`
	Name      string    `json:"name" example:"John Doe"`
	Email     string    `json:"email" example:"john@example.com"`
	Password  string    `json:"-"` // never exposed in JSON
	Role      string    `json:"role" example:"owner" enums:"owner,cashier"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2026-01-30T12:00:00Z"`
}

// UserInput represents the input for creating/updating a user
// @Description Input model for creating or updating a user
type UserInput struct {
	Name     string `json:"name" example:"John Doe" binding:"required"`
	Email    string `json:"email" example:"john@example.com" binding:"required,email"`
	Password string `json:"password" example:"secret123" binding:"required,min=6"`
	Role     string `json:"role" example:"cashier" binding:"required,oneof=owner cashier"`
}

// LoginInput represents the login request body
// @Description Login credentials
type LoginInput struct {
	Email    string `json:"email" example:"admin@retail.com" binding:"required,email"`
	Password string `json:"password" example:"password123" binding:"required"`
}

// LoginResponse represents the login response
// @Description Login response with JWT token and user info
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User  User   `json:"user"`
}
