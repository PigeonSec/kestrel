package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pigeonsec/kestrel/internal/auth"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(keyStore *auth.KeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := extractAPIKey(c)
		if apiKey == "" || !keyStore.IsValid(apiKey) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

// extractAPIKey extracts the API key from various sources
func extractAPIKey(c *gin.Context) string {
	// Check X-API-Key header
	if v := c.GetHeader("X-API-Key"); v != "" {
		return v
	}

	// Check Authorization header (Bearer token)
	if auth := c.GetHeader("Authorization"); len(auth) > 7 && strings.EqualFold(auth[:7], "Bearer ") {
		return strings.TrimSpace(auth[7:])
	}

	// Check query parameter
	if v := c.Query("apikey"); v != "" {
		return v
	}

	return ""
}
