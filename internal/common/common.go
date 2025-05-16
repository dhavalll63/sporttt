package common

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const (
	// Context keys
	ContextUserKey   = "currentUser" // Key to store user object in context
	ContextUserIDKey = "userID"      // Key to store user ID in context
)

func GetUserIDFromContext(c *gin.Context) (uint, error) {
	userIDInterface, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, errors.New("user ID not found in context")
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		return 0, errors.New("user ID in context is not of type uint")
	}
	return userID, nil
}

func GetCurrentUserInterface(c *gin.Context) (interface{}, bool) {
	userInterface, exists := c.Get(ContextUserKey)
	if !exists {
		return nil, false
	}
	return userInterface, true
}
