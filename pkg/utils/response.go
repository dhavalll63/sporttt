package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standard error response format
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a standard success response format
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ValidationErrorResponse represents a validation error with field-specific details
type ValidationErrorResponse struct {
	Error  string                 `json:"error"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

// PaginationData represents standard pagination metadata
type PaginationData struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int64 `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginatedResponse represents a paginated response with data and pagination metadata
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination PaginationData
}

// ErrorJSON sends a JSON error response with the specified HTTP status code
func ErrorJSON(ctx *gin.Context, statusCode int, err error) {
	ctx.JSON(statusCode, ErrorResponse{Error: err.Error()})
}

// ValidationErrorJSON sends a validation error response with field details
func ValidationErrorJSON(ctx *gin.Context, message string, fields map[string]interface{}) {
	ctx.JSON(http.StatusBadRequest, ValidationErrorResponse{
		Error:  message,
		Fields: fields,
	})
}

// SuccessJSON sends a JSON success response with optional data
func SuccessJSON(ctx *gin.Context, statusCode int, message string, data interface{}) {
	ctx.JSON(statusCode, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

// PaginatedJSON sends a paginated JSON response
func PaginatedJSON(ctx *gin.Context, data interface{}, page, limit int, total int64) {
	totalPages := (total + int64(limit) - 1) / int64(limit)
	hasNext := int64(page) < totalPages
	hasPrev := page > 1

	ctx.JSON(http.StatusOK, PaginatedResponse{
		Data: data,
		Pagination: PaginationData{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	})
}

// UnauthorizedJSON sends an unauthorized error response
func UnauthorizedJSON(ctx *gin.Context) {
	ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized access"})
}

// ForbiddenJSON sends a forbidden error response
func ForbiddenJSON(ctx *gin.Context) {
	ctx.JSON(http.StatusForbidden, ErrorResponse{Error: "Access forbidden"})
}

// NotFoundJSON sends a not found error response
func NotFoundJSON(ctx *gin.Context, resource string) {
	ctx.JSON(http.StatusNotFound, ErrorResponse{Error: resource + " not found"})
}

// BadRequestJSON sends a bad request error response
func BadRequestJSON(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: message})
}

// InternalErrorJSON sends an internal server error response
func InternalErrorJSON(ctx *gin.Context, err error) {
	// In production, you might want to log the error here
	// but not expose the actual error message to clients
	ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
}

// ConflictJSON sends a conflict error response
func ConflictJSON(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusConflict, ErrorResponse{Error: message})
}
