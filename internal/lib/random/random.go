package random

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

func NewRandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}

	// Calculate the number of bytes needed to generate the desired string length.
	// Base64 encoding uses 4 characters to represent 3 bytes, so we round up.
	byteLength := (length * 6) / 8
	if (length*6)%8 != 0 {
		byteLength++
	}

	// Generate random bytes.
	randomBytes := make([]byte, byteLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode the random bytes to a base64 string.
	randomString := base64.URLEncoding.EncodeToString(randomBytes)

	// Trim the string to the desired length.
	return randomString[:length], nil
}
