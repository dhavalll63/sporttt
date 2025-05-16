package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateRandomToken creates a secure random token with the specified byte length
func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// If crypto/rand fails, this is a serious issue
		panic(fmt.Errorf("critical security failure: could not generate random token: %v", err))
	}
	return hex.EncodeToString(b)
}

// EnsureDir makes sure a directory exists, creating it if necessary
func EnsureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// GetFileExtension returns the file extension from a filename
func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// IsValidFileType checks if the provided file extension is in the list of allowed types
func IsValidFileType(ext string, allowedTypes []string) bool {
	for _, t := range allowedTypes {
		if ext == t {
			return true
		}
	}
	return false
}
