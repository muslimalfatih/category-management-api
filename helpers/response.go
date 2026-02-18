package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response envelope
type Response struct {
	Status  bool        `json:"status" example:"true"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data,omitempty" swaggertype:"object"`
}

// PaginatedResponse is the standard paginated API response envelope
type PaginatedResponse struct {
	Status  bool           `json:"status" example:"true"`
	Message string         `json:"message" example:"Success"`
	Data    interface{}    `json:"data,omitempty" swaggertype:"object"`
	Meta    PaginationMeta `json:"meta"`
}

// ErrorResponse is the standard error response envelope
type ErrorResponse struct {
	Status  bool   `json:"status" example:"false"`
	Message string `json:"message" example:"Error occurred"`
	Error   string `json:"error,omitempty" example:"validation detail"`
}

// PaginationMeta holds pagination metadata
type PaginationMeta struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"20"`
	Total      int `json:"total" example:"150"`
	TotalPages int `json:"total_pages" example:"8"`
}

// Success sends a standard success response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Status:  true,
		Message: message,
		Data:    data,
	})
}

// Error sends a standard error response
func Error(c *gin.Context, statusCode int, message string, err ...string) {
	resp := ErrorResponse{
		Status:  false,
		Message: message,
	}
	if len(err) > 0 && err[0] != "" {
		resp.Error = err[0]
	}
	c.JSON(statusCode, resp)
}

// Created sends a 201 success response
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// OK sends a 200 success response
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// BadRequest sends a 400 error response
func BadRequest(c *gin.Context, message string, err ...string) {
	Error(c, http.StatusBadRequest, message, err...)
}

// NotFound sends a 404 error response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError sends a 500 error response
func InternalError(c *gin.Context, message string, err ...string) {
	Error(c, http.StatusInternalServerError, message, err...)
}

// Unauthorized sends a 401 error response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 error response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// Paginated sends a standard paginated response
func Paginated(c *gin.Context, message string, data interface{}, meta PaginationMeta) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Status:  true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}
