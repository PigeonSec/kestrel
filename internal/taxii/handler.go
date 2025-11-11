package taxii

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pigeonsec/kestrel/internal/storage"
)

// Handler handles TAXII 2.1 endpoints
type Handler struct {
	storage     storage.Storage
	baseURL     string
	collections map[string]*Collection
}

// NewHandler creates a new TAXII handler
func NewHandler(stor storage.Storage, baseURL string) *Handler {
	h := &Handler{
		storage:     stor,
		baseURL:     baseURL,
		collections: make(map[string]*Collection),
	}

	// Initialize default collection
	h.collections["kestrel-indicators"] = &Collection{
		ID:          "kestrel-indicators",
		Title:       "Kestrel Threat Indicators",
		Description: "Domain-based threat indicators from Kestrel CTI",
		CanRead:     true,
		CanWrite:    false,
		MediaTypes:  []string{MediaTypeSTIX},
	}

	return h
}

// GetDiscovery handles GET /taxii2/
func (h *Handler) GetDiscovery(c *gin.Context) {
	discovery := Discovery{
		Title:       "Kestrel TAXII 2.1 Server",
		Description: "High-performance threat intelligence distribution via TAXII 2.1",
		Default:     h.baseURL + "/taxii2/api1/",
		APIRoots:    []string{h.baseURL + "/taxii2/api1/"},
	}

	c.Header("Content-Type", MediaTypeTAXII)
	c.JSON(http.StatusOK, discovery)
}

// GetAPIRoot handles GET /taxii2/:root/
func (h *Handler) GetAPIRoot(c *gin.Context) {
	apiRoot := APIRoot{
		Title:            "Kestrel API Root",
		Description:      "Primary TAXII API root for Kestrel threat intelligence",
		Versions:         []string{TAXIIVersion},
		MaxContentLength: MaxContentLength,
	}

	c.Header("Content-Type", MediaTypeTAXII)
	c.JSON(http.StatusOK, apiRoot)
}

// GetCollections handles GET /taxii2/:root/collections/
func (h *Handler) GetCollections(c *gin.Context) {
	collections := make([]Collection, 0, len(h.collections))
	for _, col := range h.collections {
		collections = append(collections, *col)
	}

	response := Collections{
		Collections: collections,
	}

	c.Header("Content-Type", MediaTypeTAXII)
	c.JSON(http.StatusOK, response)
}

// GetCollection handles GET /taxii2/:root/collections/:id/
func (h *Handler) GetCollection(c *gin.Context) {
	collectionID := c.Param("id")

	col, exists := h.collections[collectionID]
	if !exists {
		h.sendError(c, http.StatusNotFound, "Collection not found", "")
		return
	}

	c.Header("Content-Type", MediaTypeTAXII)
	c.JSON(http.StatusOK, col)
}

// GetObjects handles GET /taxii2/:root/collections/:id/objects/
func (h *Handler) GetObjects(c *gin.Context) {
	ctx := c.Request.Context()
	collectionID := c.Param("id")

	// Verify collection exists
	if _, exists := h.collections[collectionID]; !exists {
		h.sendError(c, http.StatusNotFound, "Collection not found", "")
		return
	}

	// Parse filter parameters
	filters := h.parseFilters(c)

	// Get all STIX objects
	stixIDs, err := h.storage.ListSTIXObjects(ctx)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to list objects", "")
		return
	}

	// Apply filtering
	filteredIDs := h.applyFilters(stixIDs, filters)

	// Apply pagination
	start := 0
	if filters.Next != "" {
		// Parse pagination token (simple offset for now)
		if offset, err := strconv.Atoi(filters.Next); err == nil {
			start = offset
		}
	}

	limit := filters.Limit
	if limit == 0 || limit > MaxLimit {
		limit = DefaultLimit
	}

	end := start + limit
	hasMore := end < len(filteredIDs)
	if end > len(filteredIDs) {
		end = len(filteredIDs)
	}

	pageIDs := filteredIDs[start:end]

	// Fetch objects
	objects := []any{}
	for _, stixID := range pageIDs {
		data, err := h.storage.GetSTIXObject(ctx, stixID)
		if err != nil {
			continue
		}

		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			continue
		}

		// Apply type filtering
		if len(filters.Type) > 0 {
			objType, ok := obj["type"].(string)
			if !ok || !contains(filters.Type, objType) {
				continue
			}
		}

		objects = append(objects, obj)
	}

	// Build response envelope
	envelope := Envelope{
		More:    hasMore,
		Objects: objects,
	}

	if hasMore {
		envelope.Next = strconv.Itoa(end)
	}

	c.Header("Content-Type", MediaTypeSTIX)
	c.JSON(http.StatusOK, envelope)
}

// GetManifest handles GET /taxii2/:root/collections/:id/manifest/
func (h *Handler) GetManifest(c *gin.Context) {
	ctx := c.Request.Context()
	collectionID := c.Param("id")

	// Verify collection exists
	if _, exists := h.collections[collectionID]; !exists {
		h.sendError(c, http.StatusNotFound, "Collection not found", "")
		return
	}

	// Parse filter parameters
	filters := h.parseFilters(c)

	// Get all STIX objects
	stixIDs, err := h.storage.ListSTIXObjects(ctx)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to list objects", "")
		return
	}

	// Apply filtering
	filteredIDs := h.applyFilters(stixIDs, filters)

	// Build manifest records
	records := []ManifestRecord{}
	for _, stixID := range filteredIDs {
		data, err := h.storage.GetSTIXObject(ctx, stixID)
		if err != nil {
			continue
		}

		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			continue
		}

		created, _ := obj["created"].(string)
		modified, _ := obj["modified"].(string)

		record := ManifestRecord{
			ID:        stixID,
			DateAdded: created,
			Version:   modified,
			MediaType: MediaTypeSTIX,
		}
		records = append(records, record)
	}

	// Apply pagination
	start := 0
	if filters.Next != "" {
		if offset, err := strconv.Atoi(filters.Next); err == nil {
			start = offset
		}
	}

	limit := filters.Limit
	if limit == 0 || limit > MaxLimit {
		limit = DefaultLimit
	}

	end := start + limit
	hasMore := end < len(records)
	if end > len(records) {
		end = len(records)
	}

	pageRecords := records[start:end]

	manifest := Manifest{
		More:    hasMore,
		Objects: pageRecords,
	}

	if hasMore {
		manifest.Next = strconv.Itoa(end)
	}

	c.Header("Content-Type", MediaTypeTAXII)
	c.JSON(http.StatusOK, manifest)
}

// GetObjectByID handles GET /taxii2/:root/collections/:id/objects/:object_id/
func (h *Handler) GetObjectByID(c *gin.Context) {
	ctx := c.Request.Context()
	collectionID := c.Param("id")
	objectID := c.Param("object_id")

	// Verify collection exists
	if _, exists := h.collections[collectionID]; !exists {
		h.sendError(c, http.StatusNotFound, "Collection not found", "")
		return
	}

	// Fetch object
	data, err := h.storage.GetSTIXObject(ctx, objectID)
	if err == storage.ErrNotFound {
		h.sendError(c, http.StatusNotFound, "Object not found", "")
		return
	}
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Failed to fetch object", "")
		return
	}

	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		h.sendError(c, http.StatusInternalServerError, "Invalid object data", "")
		return
	}

	// Return as envelope with single object
	envelope := Envelope{
		More:    false,
		Objects: []any{obj},
	}

	c.Header("Content-Type", MediaTypeSTIX)
	c.JSON(http.StatusOK, envelope)
}

// GetVersions handles GET /taxii2/:root/collections/:id/objects/:object_id/versions/
func (h *Handler) GetVersions(c *gin.Context) {
	// For now, we only support single versions
	// Future: implement versioning in storage layer
	h.sendError(c, http.StatusNotImplemented, "Versioning not yet implemented", "")
}

// Helper functions

func (h *Handler) parseFilters(c *gin.Context) FilterParams {
	filters := FilterParams{
		AddedAfter: c.Query("added_after"),
		Next:       c.Query("next"),
		Version:    c.Query("version"),
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if matchStr := c.Query("match[id]"); matchStr != "" {
		filters.Match = strings.Split(matchStr, ",")
	}

	if typeStr := c.Query("match[type]"); typeStr != "" {
		filters.Type = strings.Split(typeStr, ",")
	}

	return filters
}

func (h *Handler) applyFilters(stixIDs []string, filters FilterParams) []string {
	filtered := stixIDs

	// Filter by match IDs
	if len(filters.Match) > 0 {
		matchSet := make(map[string]bool)
		for _, id := range filters.Match {
			matchSet[id] = true
		}

		result := []string{}
		for _, id := range filtered {
			if matchSet[id] {
				result = append(result, id)
			}
		}
		filtered = result
	}

	// TODO: Implement added_after filtering
	// Requires storing timestamps in storage layer

	return filtered
}

func (h *Handler) sendError(c *gin.Context, status int, title, description string) {
	errorResponse := Error{
		Title:       title,
		Description: description,
		ErrorID:     uuid.New().String(),
		HTTPStatus:  status,
	}

	c.Header("Content-Type", MediaTypeTAXII)
	c.JSON(status, errorResponse)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
