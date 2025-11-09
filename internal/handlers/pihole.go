package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pigeonsec/kestrel/internal/storage"
)

// PiHoleHandler handles PiHole/AdGuard blocklist endpoints
type PiHoleHandler struct {
	storage   storage.Storage
	freeFeeds map[string]bool
}

// NewPiHoleHandler creates a new PiHole handler
func NewPiHoleHandler(storage storage.Storage) *PiHoleHandler {
	return &PiHoleHandler{
		storage: storage,
		freeFeeds: map[string]bool{
			"public": true,
		},
	}
}

// GetFeed handles GET /pihole/:feed
func (h *PiHoleHandler) GetFeed(c *gin.Context) {
	feedParam := c.Param("feed")
	// Remove .txt extension if present
	feed := feedParam
	if len(feedParam) > 4 && feedParam[len(feedParam)-4:] == ".txt" {
		feed = feedParam[:len(feedParam)-4]
	}

	// Check if feed requires authentication
	requiresAuth := !h.freeFeeds[feed]
	if requiresAuth {
		// Authentication will be handled by middleware
		// This is just for documentation
	}

	// Fetch domains from storage
	domains, err := h.storage.GetDomains(c.Request.Context(), feed)
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
