package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pigeonsec/kestrel/internal/misp"
	"github.com/pigeonsec/kestrel/internal/stix"
	"github.com/pigeonsec/kestrel/internal/storage"
	"github.com/pigeonsec/kestrel/internal/validation"
)

// IOCRequest represents the request to submit a new IOC
type IOCRequest struct {
	// IOC Value (one required)
	Domain   string `json:"domain"`
	IP       string `json:"ip"`
	URL      string `json:"url"`
	Hash     string `json:"hash"`     // MD5, SHA1, or SHA256
	HashType string `json:"hash_type"` // "md5", "sha1", "sha256"
	Email    string `json:"email"`

	// Metadata (required)
	Category    string `json:"category" binding:"required"`
	Comment     string `json:"comment"`
	Feed        string `json:"feed" binding:"required"`
	AccessLevel string `json:"access_level"` // "free" or "paid" (default: "paid")

	// Optional enrichment
	ThreatActor string   `json:"threat_actor,omitempty"`
	Malware     string   `json:"malware,omitempty"`
	Campaign    string   `json:"campaign,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// IOCHandler handles IOC ingestion
type IOCHandler struct {
	storage       storage.Storage
	misp          *misp.Handler
	validator     *validation.Validator
	stixConverter *stix.Converter
}

// NewIOCHandler creates a new IOC handler
func NewIOCHandler(storage storage.Storage, mispHandler *misp.Handler, validator *validation.Validator) *IOCHandler {
	return &IOCHandler{
		storage:       storage,
		misp:          mispHandler,
		validator:     validator,
		stixConverter: stix.NewConverter(),
	}
}

// PostIOC handles POST /api/ioc
func (h *IOCHandler) PostIOC(c *gin.Context) {
	var req IOCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Validate that at least one IOC type is provided
	iocValue, iocType := h.extractIOCValue(req)
	if iocValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one IOC value required (domain, ip, url, hash, or email)",
		})
		return
	}

	// Check for validation parameters (only for domains)
	if iocType == "domain" {
		validateMode := c.Query("validate")
		autoValidate := c.Query("autovalidate")

		// Determine validation mode
		var validationMode validation.ValidationMode
		if autoValidate != "" {
			validationMode = validation.ParseValidationMode(autoValidate)
		} else if validateMode != "" {
			validationMode = validation.ParseValidationMode(validateMode)
		}

		// Perform validation if requested or if globally enabled
		shouldValidate := validationMode != validation.ValidationNone
		if !shouldValidate && h.validator != nil {
			// Check if validation is globally enabled
			shouldValidate = true
			validationMode = validation.ValidationDNS // Default to DNS validation
		}

		if shouldValidate && h.validator != nil {
			result, err := h.validator.Validate(ctx, iocValue, validationMode)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "validation failed",
					"details": err.Error(),
				})
				return
			}

			if !result.Valid {
				// Domain failed validation - do NOT add to feed
				c.JSON(http.StatusBadRequest, gin.H{
					"error":             "domain validation failed - not added to feed",
					"domain":            result.Domain,
					"validation_result": result,
				})
				return
			}
		}
	}

	// Generate event ID
	eventID := uuid.New().String()

	// Create MISP event
	if err := h.misp.CreateEvent(ctx, eventID, iocValue, req.Category, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	// Determine access level (default to paid if not specified)
	accessLevel := req.AccessLevel
	if accessLevel == "" {
		accessLevel = "paid"
	}

	// Add IOC to feed (for now, only domains go to domain feeds)
	if iocType == "domain" {
		if err := h.storage.AddDomain(ctx, req.Feed, iocValue); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add domain to feed"})
			return
		}
	}

	// Set feed access level metadata
	if err := h.storage.SetFeedMeta(ctx, req.Feed, "access_level", accessLevel); err != nil {
		// Log but don't fail - metadata is not critical
		c.Header("X-Warning", "Failed to set feed metadata")
	}

	// Generate STIX indicator
	stixID, err := h.generateStixIndicator(ctx, req)
	if err != nil {
		// Log but don't fail - STIX generation is optional enhancement
		c.Header("X-Warning", "Failed to generate STIX indicator: "+err.Error())
	}

	response := gin.H{
		"status":       "stored",
		"event_id":     eventID,
		"ioc_type":     iocType,
		"ioc_value":    iocValue,
		"feed":         req.Feed,
		"access_level": accessLevel,
	}

	if stixID != "" {
		response["stix_id"] = stixID
	}

	c.JSON(http.StatusOK, response)
}

// extractIOCValue extracts the IOC value and determines its type
func (h *IOCHandler) extractIOCValue(req IOCRequest) (string, string) {
	if req.Domain != "" {
		return req.Domain, "domain"
	}
	if req.IP != "" {
		return req.IP, "ip"
	}
	if req.URL != "" {
		return req.URL, "url"
	}
	if req.Hash != "" {
		if req.HashType == "" {
			// Auto-detect hash type by length
			switch len(req.Hash) {
			case 32:
				return req.Hash, "md5"
			case 40:
				return req.Hash, "sha1"
			case 64:
				return req.Hash, "sha256"
			default:
				return req.Hash, "hash"
			}
		}
		return req.Hash, req.HashType
	}
	if req.Email != "" {
		return req.Email, "email"
	}
	return "", ""
}

// generateStixIndicator creates and stores a STIX indicator for the IOC
func (h *IOCHandler) generateStixIndicator(ctx context.Context, req IOCRequest) (string, error) {
	now := time.Now().UTC()

	// Check if domain already has a STIX ID
	existingStixID, err := h.storage.GetDomainStixID(ctx, req.Domain)
	if err == nil && existingStixID != "" {
		// Reuse existing STIX ID
		return existingStixID, nil
	}

	// Generate new STIX ID
	stixID := stix.GenerateStixID()

	// Prepare IOC data for conversion
	iocData := stix.IOCData{
		Domain:       req.Domain,
		Category:     req.Category,
		Comment:      req.Comment,
		Feed:         req.Feed,
		Source:       "Kestrel",
		Confidence:   80, // Default confidence
		TLP:          "TLP:AMBER",
		FirstSeen:    now,
		LastSeen:     now,
		Organization: "Kestrel CTI",
	}

	// Convert to STIX bundle
	opts := stix.ConversionOptions{
		IncludeIdentity:      true,
		IncludeObservedData:  false,
		IncludeRelationships: false,
		StixID:               stixID,
	}
	bundle := h.stixConverter.ConvertDomainToSTIX(iocData, opts)

	// Extract the indicator from the bundle
	if len(bundle.Objects) == 0 {
		return "", nil
	}

	// Marshal the indicator
	indicatorData, err := json.Marshal(bundle.Objects[0])
	if err != nil {
		return "", err
	}

	// Store STIX object
	if err := h.storage.SetSTIXObject(ctx, stixID, indicatorData); err != nil {
		return "", err
	}

	// Store domain -> STIX ID mapping
	if err := h.storage.SetDomainStixID(ctx, req.Domain, stixID); err != nil {
		return "", err
	}

	return stixID, nil
}

// ListIOCs handles GET /api/iocs
func (h *IOCHandler) ListIOCs(c *gin.Context) {
	ctx := context.Background()
	feed := c.Query("feed")

	// Build IOC list with feed mapping
	type iocWithFeed struct {
		value string
		feed  string
	}

	var iocsToProcess []iocWithFeed

	if feed != "" {
		// Get domains for specific feed
		domains, err := h.storage.GetDomains(ctx, feed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve IOCs"})
			return
		}
		for _, d := range domains {
			iocsToProcess = append(iocsToProcess, iocWithFeed{value: d, feed: feed})
		}
	} else {
		// Get all feeds and collect all domains with their feed
		feeds, err := h.storage.ListFeeds(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list feeds"})
			return
		}

		for _, f := range feeds {
			feedDomains, err := h.storage.GetDomains(ctx, f)
			if err != nil {
				continue
			}
			for _, d := range feedDomains {
				iocsToProcess = append(iocsToProcess, iocWithFeed{value: d, feed: f})
			}
		}
	}

	// Build response with IOC details
	iocs := make([]gin.H, 0, len(iocsToProcess))
	for _, iocData := range iocsToProcess {
		// Try to get STIX ID and MISP event ID
		stixID, _ := h.storage.GetDomainStixID(ctx, iocData.value)

		// Try to get MISP event ID from storage
		mispEventID, _ := h.storage.Get(ctx, "misp:domain:"+iocData.value)

		ioc := gin.H{
			"value": iocData.value,
			"type":  "domain",
			"feed":  iocData.feed,
		}
		if stixID != "" {
			ioc["stix_id"] = stixID
		}
		if len(mispEventID) > 0 {
			ioc["misp_event_id"] = string(mispEventID)
		}

		iocs = append(iocs, ioc)
	}

	c.JSON(http.StatusOK, gin.H{
		"iocs":  iocs,
		"count": len(iocs),
	})
}

// DeleteIOC handles DELETE /api/ioc/:id
func (h *IOCHandler) DeleteIOC(c *gin.Context) {
	ctx := context.Background()
	iocValue := c.Param("id")
	feed := c.Query("feed")

	if feed == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "feed parameter required"})
		return
	}

	// Remove domain from feed
	if err := h.storage.RemoveDomain(ctx, feed, iocValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove IOC"})
		return
	}

	// Optionally delete STIX object
	stixID, _ := h.storage.GetDomainStixID(ctx, iocValue)
	if stixID != "" {
		h.storage.DeleteSTIXObject(ctx, stixID)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "deleted",
		"ioc":     iocValue,
		"feed":    feed,
	})
}

// UpdateIOC handles PUT /api/ioc/:id
func (h *IOCHandler) UpdateIOC(c *gin.Context) {
	ctx := context.Background()
	oldValue := c.Param("id")

	var req struct {
		NewValue    string `json:"new_value"`
		Feed        string `json:"feed" binding:"required"`
		Category    string `json:"category"`
		Comment     string `json:"comment"`
		AccessLevel string `json:"access_level"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If value is being changed, remove old and add new
	if req.NewValue != "" && req.NewValue != oldValue {
		// Remove old domain
		if err := h.storage.RemoveDomain(ctx, req.Feed, oldValue); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove old IOC"})
			return
		}

		// Add new domain
		if err := h.storage.AddDomain(ctx, req.Feed, req.NewValue); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add new IOC"})
			return
		}

		// Delete old STIX mapping
		h.storage.DeleteSTIXObject(ctx, "stix:domain:"+oldValue)
	}

	// Update metadata if provided
	if req.AccessLevel != "" {
		h.storage.SetFeedMeta(ctx, req.Feed, "access_level", req.AccessLevel)
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "updated",
		"ioc":    req.NewValue,
		"feed":   req.Feed,
	})
}
