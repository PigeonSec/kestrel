package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pigeonsec/kestrel/internal/auth"
)

type APIKeysHandler struct {
	keyStore *auth.KeyStore
	prefix   string
}

func NewAPIKeysHandler(keyStore *auth.KeyStore, prefix string) *APIKeysHandler {
	return &APIKeysHandler{
		keyStore: keyStore,
		prefix:   prefix,
	}
}

type APIKeyResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

// ListAPIKeys handles GET /api/keys
func (h *APIKeysHandler) ListAPIKeys(c *gin.Context) {
	// Get all accounts from key store
	accounts := h.keyStore.ListAccounts()

	keys := make([]APIKeyResponse, 0, len(accounts))
	for _, account := range accounts {
		// Skip admin account
		if account.Plan == "admin" {
			continue
		}

		keys = append(keys, APIKeyResponse{
			ID:        account.APIKey[:8], // Use first 8 chars as ID
			Name:      account.Email,      // Using email as name for now
			Key:       account.APIKey,
			Role:      account.Plan,
			CreatedAt: time.Now().Format(time.RFC3339), // TODO: Add created_at to Account struct
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":  keys,
		"count": len(keys),
	})
}

// CreateAPIKey handles POST /api/keys
func (h *APIKeysHandler) CreateAPIKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate new API key
	apiKey, err := auth.GenerateAPIKey(h.prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate API key"})
		return
	}

	// Create account
	account := &auth.Account{
		APIKey: apiKey,
		Email:  req.Name,
		Plan:   req.Role,
		Active: true,
	}

	if err := h.keyStore.AddAccount(account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "created",
		"key": APIKeyResponse{
			ID:        apiKey[:8],
			Name:      req.Name,
			Key:       apiKey,
			Role:      req.Role,
			CreatedAt: time.Now().Format(time.RFC3339),
		},
	})
}

// DeleteAPIKey handles DELETE /api/keys/:id
func (h *APIKeysHandler) DeleteAPIKey(c *gin.Context) {
	keyID := c.Param("id")

	// Find the account with this key prefix
	accounts := h.keyStore.ListAccounts()
	var targetKey string

	for _, account := range accounts {
		if len(account.APIKey) >= len(keyID) && account.APIKey[:len(keyID)] == keyID {
			targetKey = account.APIKey
			break
		}
	}

	if targetKey == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	if err := h.keyStore.RemoveAccount(targetKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "deleted",
		"id":     keyID,
	})
}
