package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pigeonsec/kestrel/internal/storage"
)

type FeedsHandler struct {
	storage storage.Storage
}

func NewFeedsHandler(stor storage.Storage) *FeedsHandler {
	return &FeedsHandler{
		storage: stor,
	}
}

type FeedInfo struct {
	Name        string `json:"name"`
	Count       int    `json:"count"`
	AccessLevel string `json:"access_level,omitempty"`
}

func (h *FeedsHandler) ListFeeds(c *gin.Context) {
	ctx := context.Background()

	feedNames, err := h.storage.ListFeeds(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list feeds"})
		return
	}

	feeds := make([]FeedInfo, 0, len(feedNames))
	for _, name := range feedNames {
		domains, err := h.storage.GetDomains(ctx, name)
		if err != nil {
			continue
		}

		// Get access level from feed metadata
		accessLevel, _ := h.storage.GetFeedMeta(ctx, name, "access_level")

		feeds = append(feeds, FeedInfo{
			Name:        name,
			Count:       len(domains),
			AccessLevel: accessLevel,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"feeds": feeds,
		"count": len(feeds),
	})
}

// UpdateFeedPermissions handles PUT /api/feeds/:name/permissions
func (h *FeedsHandler) UpdateFeedPermissions(c *gin.Context) {
	ctx := context.Background()
	feedName := c.Param("name")

	var req struct {
		AccessLevel string `json:"access_level" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate access level
	if req.AccessLevel != "free" && req.AccessLevel != "paid" && req.AccessLevel != "private" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "access_level must be 'free', 'paid', or 'private'"})
		return
	}

	// Update feed metadata
	if err := h.storage.SetFeedMeta(ctx, feedName, "access_level", req.AccessLevel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update feed permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "updated",
		"feed":         feedName,
		"access_level": req.AccessLevel,
	})
}
