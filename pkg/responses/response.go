package responses

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Status  string      `json:"status"`  // "success"
	Message string      `json:"message"` // Optional success message
	Data    interface{} `json:"data"`    // The actual data payload
}

type ErrorResponse struct {
	Status  string `json:"status"`  // "error" or "fail"
	Message string `json:"message"` // Error message
	Code    int    `json:"code"`    // HTTP status code
}

// PaginatedResponse represents a success response for lists with pagination details.
type PaginatedResponse struct {
	Status     string      `json:"status"`  // "success"
	Message    string      `json:"message"` // Optional success message
	Data       interface{} `json:"data"`    // The list of items
	Pagination Pagination  `json:"pagination"`
}

// Pagination holds pagination information.
type Pagination struct {
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	HasNextPage  bool  `json:"has_next_page"`
	HasPrevPage  bool  `json:"has_prev_page"`
	NextPage     *int  `json:"next_page,omitempty"`
	PreviousPage *int  `json:"previous_page,omitempty"`
}

// SendSuccess sends a standardized success response.
func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	if message == "" {
		message = "Operation completed successfully"
	}
	c.JSON(statusCode, SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// SendError sends a standardized error response.
func SendError(c *gin.Context, statusCode int, message string) {
	statusText := "error"
	if statusCode >= http.StatusInternalServerError {
		statusText = "fail" // Differentiate client errors from server failures
	}
	c.AbortWithStatusJSON(statusCode, ErrorResponse{
		Status:  statusText,
		Message: message,
		Code:    statusCode,
	})
}

func SendPaginated(c *gin.Context, statusCode int, message string, data interface{}, totalItems int64, currentPage int, pageSize int) {
	if message == "" {
		message = "Data retrieved successfully"
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	if totalPages == 0 && totalItems > 0 { // Ensure at least one page if there are items
		totalPages = 1
	}

	hasNextPage := currentPage < totalPages
	hasPrevPage := currentPage > 1

	var nextPage *int
	if hasNextPage {
		val := currentPage + 1
		nextPage = &val
	}

	var prevPage *int
	if hasPrevPage {
		val := currentPage - 1
		prevPage = &val
	}

	c.JSON(statusCode, PaginatedResponse{
		Status:  "success",
		Message: message,
		Data:    data,
		Pagination: Pagination{
			TotalItems:   totalItems,
			TotalPages:   totalPages,
			CurrentPage:  currentPage,
			PageSize:     pageSize,
			HasNextPage:  hasNextPage,
			HasPrevPage:  hasPrevPage,
			NextPage:     nextPage,
			PreviousPage: prevPage,
		},
	})
}
