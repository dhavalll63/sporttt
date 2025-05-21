package matchresponse

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10" // For handling validation errors
)

// --- Structs for Standardized JSON Response Bodies ---

// jsonSuccessResponse is the structure for successful responses.
type jsonSuccessResponse struct {
	Status  string      `json:"status"`            // Typically "success"
	Message string      `json:"message,omitempty"` // Optional descriptive message
	Data    interface{} `json:"data,omitempty"`    // The actual data payload
}

// jsonErrorResponse is the structure for error responses.
type jsonErrorResponse struct {
	Status  string      `json:"status"`           // "error" or "fail"
	Message string      `json:"message"`          // Error message
	Code    int         `json:"code"`             // HTTP status code
	Errors  interface{} `json:"errors,omitempty"` // Detailed errors, e.g., for validation
}

// jsonPaginatedResponse is the structure for responses containing paginated data.
type jsonPaginatedResponse struct {
	Status     string      `json:"status"`            // Typically "success"
	Message    string      `json:"message,omitempty"` // Optional descriptive message
	Data       interface{} `json:"data"`              // The list of items
	Pagination pagination  `json:"pagination"`
}

// pagination holds pagination details.
type pagination struct {
	TotalItems   int64 `json:"total_items"`
	TotalPages   int   `json:"total_pages"`
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	HasNextPage  bool  `json:"has_next_page"`
	HasPrevPage  bool  `json:"has_prev_page"`
	NextPage     *int  `json:"next_page,omitempty"`
	PreviousPage *int  `json:"previous_page,omitempty"`
}

// --- Public Response Helper Functions ---

// ErrorResponse sends a standardized error JSON response.
// It's used for general errors.
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	statusText := "error"
	if statusCode >= http.StatusInternalServerError {
		statusText = "fail" // Differentiate client errors from server failures
	}
	c.AbortWithStatusJSON(statusCode, jsonErrorResponse{
		Status:  statusText,
		Message: message,
		Code:    statusCode,
	})
}

// formatValidationErrors converts validator.ValidationErrors into a map.
func formatValidationErrors(errs validator.ValidationErrors) map[string]string {
	formattedErrors := make(map[string]string)
	for _, err := range errs {
		// Use a simple field name (lowercase) for the key.
		// For more complex scenarios, you might want to use JSON tags.
		fieldKey := strings.ToLower(err.Field())
		// Construct a user-friendly message
		var errMsg string
		switch err.Tag() {
		case "required":
			errMsg = fmt.Sprintf("The %s field is required.", err.Field())
		case "min":
			errMsg = fmt.Sprintf("The %s field must be at least %s.", err.Field(), err.Param())
		case "max":
			errMsg = fmt.Sprintf("The %s field must not exceed %s.", err.Field(), err.Param())
		case "oneof":
			errMsg = fmt.Sprintf("The %s field must be one of the following: %s.", err.Field(), strings.ReplaceAll(err.Param(), " ", ", "))
		case "email":
			errMsg = fmt.Sprintf("The %s field must be a valid email address.", err.Field())
		default:
			errMsg = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag.", err.Field(), err.Tag())
		}
		formattedErrors[fieldKey] = errMsg
	}
	return formattedErrors
}

// ValidationErrorResponse sends a structured JSON response for validation errors
// originating from `c.ShouldBindJSON()` or similar.
func ValidationErrorResponse(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		// It's a validator.ValidationErrors, format it.
		c.AbortWithStatusJSON(http.StatusBadRequest, jsonErrorResponse{
			Status:  "error",
			Message: "Validation failed. Please check your input.",
			Code:    http.StatusBadRequest,
			Errors:  formatValidationErrors(ve),
		})
		return
	}
	// For other binding errors (e.g., malformed JSON)
	ErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
}

// SuccessResponse sends a standardized success JSON response.
// The `data` argument provided by the controller is wrapped in the response structure.
// If `data` is `gin.H` and contains a "message" key (string), it's used as the top-level message,
// and the rest of `gin.H` becomes the `data` payload. Otherwise, the whole `data` argument becomes the payload.
func SuccessResponse(c *gin.Context, statusCode int, responseData interface{}) {
	payload := jsonSuccessResponse{
		Status: "success",
	}

	if gh, ok := responseData.(gin.H); ok {
		if msgVal, exists := gh["message"]; exists {
			if msgStr, isStr := msgVal.(string); isStr {
				payload.Message = msgStr
				// If a message was extracted, use the rest of gin.H as data
				dataMap := make(gin.H)
				hasOtherData := false
				for k, v := range gh {
					if k != "message" {
						dataMap[k] = v
						hasOtherData = true
					}
				}
				if hasOtherData {
					payload.Data = dataMap
				}
				// If only message was present in gin.H, payload.Data remains nil (or empty map if preferred)
			} else {
				// "message" key exists but not a string, or other structure. Treat whole gin.H as data.
				payload.Data = responseData
			}
		} else {
			// gin.H without a "message" key. Treat whole gin.H as data.
			payload.Data = responseData
		}
	} else if responseData != nil {
		// responseData is not gin.H (e.g., a struct, slice, primitive).
		payload.Data = responseData
	}
	// If responseData is nil, payload.Data will be nil.

	c.JSON(statusCode, payload)
}

// PaginatedResponse sends a standardized success JSON response for paginated data.
// The `itemsData` argument from the controller is the list of items.
func PaginatedResponse(c *gin.Context, statusCode int, itemsData interface{}, currentPage int, pageSize int, totalItems int64) {
	if pageSize <= 0 {
		pageSize = 10 // Default page size if invalid or not provided
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(pageSize)))
		if totalPages == 0 { // Handles cases like totalItems=5, pageSize=10 -> totalPages should be 1
			totalPages = 1
		}
	}

	hasNextPage := currentPage < totalPages
	hasPrevPage := currentPage > 1 && currentPage <= totalPages // Ensures currentPage is valid

	var nextPageNum *int
	if hasNextPage {
		val := currentPage + 1
		nextPageNum = &val
	}

	var prevPageNum *int
	if hasPrevPage {
		val := currentPage - 1
		prevPageNum = &val
	}

	// You can add an optional message parameter to PaginatedResponse if needed
	// For now, it's omitted from the top-level JSON for paginated responses.
	c.JSON(statusCode, jsonPaginatedResponse{
		Status: "success",
		Data:   itemsData,
		Pagination: pagination{
			TotalItems:   totalItems,
			TotalPages:   totalPages,
			CurrentPage:  currentPage,
			PageSize:     pageSize,
			HasNextPage:  hasNextPage,
			HasPrevPage:  hasPrevPage,
			NextPage:     nextPageNum,
			PreviousPage: prevPageNum,
		},
	})
}
