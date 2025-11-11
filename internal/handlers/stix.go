package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pigeonsec/kestrel/internal/storage"
	"github.com/pigeonsec/kestrel/internal/stix"
)

// STIXHandler handles STIX 2.1 endpoints
type STIXHandler struct {
	storage   storage.Storage
	converter *stix.Converter
}

// NewSTIXHandler creates a new STIX handler
func NewSTIXHandler(stor storage.Storage) *STIXHandler {
	return &STIXHandler{
		storage:   stor,
		converter: stix.NewConverter(),
	}
}

// GetBundle returns a STIX bundle containing all indicators
// GET /stix/bundle
func (h *STIXHandler) GetBundle(c *gin.Context) {
	ctx := c.Request.Context()

	// Get all STIX object IDs
	stixIDs, err := h.storage.ListSTIXObjects(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list STIX objects"})
		return
	}

	// Fetch all STIX objects
	objects := []any{}
	for _, stixID := range stixIDs {
		data, err := h.storage.GetSTIXObject(ctx, stixID)
		if err != nil {
			continue // Skip missing objects
		}

		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			continue // Skip invalid objects
		}
		objects = append(objects, obj)
	}

	// Create bundle
	bundle := stix.Bundle{
		Type:    "bundle",
		ID:      "bundle--" + generateUUID(),
		Objects: objects,
	}

	c.Header("Content-Type", "application/stix+json;version=2.1")
	c.JSON(http.StatusOK, bundle)
}

// GetIndicator returns a specific STIX indicator by ID
// GET /stix/indicators/:id
func (h *STIXHandler) GetIndicator(c *gin.Context) {
	ctx := c.Request.Context()
	indicatorID := c.Param("id")

	// Ensure proper indicator ID format
	if !strings.HasPrefix(indicatorID, "indicator--") {
		indicatorID = "indicator--" + indicatorID
	}

	// Fetch STIX object
	data, err := h.storage.GetSTIXObject(ctx, indicatorID)
	if err == storage.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Indicator not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch indicator"})
		return
	}

	// Parse as indicator
	var indicator stix.Indicator
	if err := json.Unmarshal(data, &indicator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid indicator data"})
		return
	}

	c.Header("Content-Type", "application/stix+json;version=2.1")
	c.JSON(http.StatusOK, indicator)
}

// GetIndicatorBundle returns a STIX bundle for a specific indicator
// GET /stix/indicators/:id/bundle
func (h *STIXHandler) GetIndicatorBundle(c *gin.Context) {
	ctx := c.Request.Context()
	indicatorID := c.Param("id")

	// Ensure proper indicator ID format
	if !strings.HasPrefix(indicatorID, "indicator--") {
		indicatorID = "indicator--" + indicatorID
	}

	// Fetch STIX object
	data, err := h.storage.GetSTIXObject(ctx, indicatorID)
	if err == storage.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Indicator not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch indicator"})
		return
	}

	// Parse as indicator
	var indicator stix.Indicator
	if err := json.Unmarshal(data, &indicator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid indicator data"})
		return
	}

	// Wrap in bundle
	bundle := stix.Bundle{
		Type:    "bundle",
		ID:      "bundle--" + generateUUID(),
		Objects: []any{indicator},
	}

	c.Header("Content-Type", "application/stix+json;version=2.1")
	c.JSON(http.StatusOK, bundle)
}

// ListIndicators returns a list of all STIX indicator IDs
// GET /stix/indicators
func (h *STIXHandler) ListIndicators(c *gin.Context) {
	ctx := c.Request.Context()

	// Get all STIX object IDs
	stixIDs, err := h.storage.ListSTIXObjects(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list STIX objects"})
		return
	}

	// Filter for indicators only
	indicators := []string{}
	for _, id := range stixIDs {
		if strings.HasPrefix(id, "indicator--") {
			indicators = append(indicators, id)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"count":      len(indicators),
		"indicators": indicators,
	})
}

// GetObjects returns all STIX objects (TAXII-like endpoint)
// GET /stix/objects
func (h *STIXHandler) GetObjects(c *gin.Context) {
	ctx := c.Request.Context()

	// Get all STIX object IDs
	stixIDs, err := h.storage.ListSTIXObjects(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list STIX objects"})
		return
	}

	// Fetch all STIX objects
	objects := []any{}
	for _, stixID := range stixIDs {
		data, err := h.storage.GetSTIXObject(ctx, stixID)
		if err != nil {
			continue // Skip missing objects
		}

		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			continue // Skip invalid objects
		}
		objects = append(objects, obj)
	}

	// Return as TAXII-like envelope
	envelope := stix.Envelope{
		More:    false,
		Objects: objects,
	}

	c.Header("Content-Type", "application/stix+json;version=2.1")
	c.JSON(http.StatusOK, envelope)
}

// Helper function to generate UUID
func generateUUID() string {
	return stix.GenerateStixID()[len("indicator--"):]
}
