// Add utility functions here
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken generates a random token of the specified length.
func GenerateRandomToken(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)[:length]
}
