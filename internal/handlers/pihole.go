package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pigeonsec/kestrel/internal/auth"
	"github.com/pigeonsec/kestrel/internal/storage"
)

// FeedHandler handles blocklist/feed endpoints for various systems (Pi-hole, AdGuard, firewalls, etc.)
// Access control is data-driven based on feed metadata from IOC ingestion
type FeedHandler struct {
	storage  storage.Storage
	keyStore *auth.KeyStore
}

// NewFeedHandler creates a new feed handler
func NewFeedHandler(storage storage.Storage, keyStore *auth.KeyStore) *FeedHandler {
	return &FeedHandler{
		storage:  storage,
		keyStore: keyStore,
	}
}

// GetDynamicFeed handles fully dynamic paths
// Feed name is extracted from the last segment of the path
// Access control is determined by feed metadata set during IOC ingestion
func (h *FeedHandler) GetDynamicFeed(c *gin.Context) {
	// Get the full request path
	requestPath := c.Request.URL.Path

	// Extract feed name from the last segment
	// e.g., /list/pihole/community.txt -> "community"
	// e.g., /list/adguard/premium.txt -> "premium"
	parts := strings.Split(strings.Trim(requestPath, "/"), "/")
	if len(parts) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	feedParam := parts[len(parts)-1]
	// Remove .txt extension if present
	feed := feedParam
	if len(feedParam) > 4 && feedParam[len(feedParam)-4:] == ".txt" {
		feed = feedParam[:len(feedParam)-4]
	}

	// Empty feed name
	if feed == "" {
		c.Status(http.StatusNotFound)
		return
	}

	ctx := c.Request.Context()

	// Check feed access level from storage metadata
	accessLevel, err := h.storage.GetFeedMeta(ctx, feed, "access_level")
	if err != nil {
		// Feed not found or no metadata - default to requiring auth
		accessLevel = "paid"
	}

	// Check authentication based on access level
	if accessLevel != "free" {
		apiKey := extractAPIKey(c)
		if apiKey == "" || !h.keyStore.IsValid(apiKey) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}

	// Fetch domains from storage
	domains, err := h.storage.GetDomains(ctx, feed)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Build response
	var builder strings.Builder
	for _, domain := range domains {
		builder.WriteString(domain)
		builder.WriteByte('\n')
	}

	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(builder.String()))
}
