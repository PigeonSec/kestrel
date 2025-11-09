package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateAPIKey generates a new URL-safe API key with the given prefix
func GenerateAPIKey(prefix string) (string, error) {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Encode to URL-safe base64
	encoded := base64.URLEncoding.EncodeToString(randomBytes)

	// Remove padding characters
	encoded = stripPadding(encoded)

	return fmt.Sprintf("%s%s", prefix, encoded), nil
}

func stripPadding(s string) string {
	for len(s) > 0 && s[len(s)-1] == '=' {
		s = s[:len(s)-1]
	}
	return s
}
