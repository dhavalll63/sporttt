package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("critical security failure: could not generate random token: %v", err))
	}
	return hex.EncodeToString(b)
}

func EnsureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

func IsValidFileType(ext string, allowedTypes []string) bool {
	for _, t := range allowedTypes {
		if ext == t {
			return true
		}
	}
	return false
}
