package responses

import (
	"math"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

type PaginatedResponse struct {
	Status     string      `json:"status"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func SendError(c *gin.Context, statusCode int, message string, errors interface{}) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{ // Abort to stop further processing
		Status:  "error",
		Message: message,
		Errors:  errors,
	})
}

func SendPaginated(c *gin.Context, statusCode int, message string, data interface{}, total int64, page, pageSize int) {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}
	c.JSON(statusCode, PaginatedResponse{
		Status:     "success",
		Message:    message,
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}
