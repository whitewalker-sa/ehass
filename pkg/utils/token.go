package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
)

// GenerateRandomToken generates a random token of the specified length in bytes
func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// StringToUint converts a string to uint, used for JWT subject claims
func StringToUint(s string) (uint, error) {
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint: %w", err)
	}
	return uint(value), nil
}
