package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pigeonsec/kestrel/internal/misp"
	"github.com/pigeonsec/kestrel/internal/storage"
	"github.com/pigeonsec/kestrel/internal/validation"
)

// IOCRequest represents the request to submit a new IOC
type IOCRequest struct {
	Domain   string `json:"domain" binding:"required"`
	Category string `json:"category" binding:"required"`
	Comment  string `json:"comment"`
	Feed     string `json:"feed" binding:"required"`
}

// IOCHandler handles IOC ingestion
type IOCHandler struct {
	storage   storage.Storage
	misp      *misp.Handler
	validator *validation.Validator
}

// NewIOCHandler creates a new IOC handler
func NewIOCHandler(storage storage.Storage, mispHandler *misp.Handler, validator *validation.Validator) *IOCHandler {
	return &IOCHandler{
		storage:   storage,
		misp:      mispHandler,
		validator: validator,
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

	// Check for validation parameters
	validateMode := c.Query("validate")
	autoValidate := c.Query("autovalidate")

	// Determine validation mode
	var validationMode validation.ValidationMode
	if autoValidate != "" {
		validationMode = validation.ParseValidationMode(autoValidate)
	} else if validateMode != "" {
		validationMode = validation.ParseValidationMode(validateMode)
	}

	// Perform validation if requested
	if validationMode != validation.ValidationNone {
		result, err := h.validator.Validate(ctx, req.Domain, validationMode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "validation failed",
				"details": err.Error(),
			})
			return
		}

		if !result.Valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "domain validation failed",
				"domain":            result.Domain,
				"validation_result": result,
			})
			return
		}
	}

	// Generate event ID
	eventID := uuid.New().String()

	// Create MISP event
	if err := h.misp.CreateEvent(ctx, eventID, req.Domain, req.Category, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	// Add domain to feed
	if err := h.storage.AddDomain(ctx, req.Feed, req.Domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add domain to feed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "stored",
		"event_id": eventID,
		"domain":   req.Domain,
		"feed":     req.Feed,
	})
}
