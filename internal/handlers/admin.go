package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pigeonsec/kestrel/internal/auth"
)

// AdminHandler handles administrative endpoints
type AdminHandler struct {
	keyStore *auth.KeyStore
	prefix   string
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(keyStore *auth.KeyStore, apiKeyPrefix string) *AdminHandler {
	return &AdminHandler{
		keyStore: keyStore,
		prefix:   apiKeyPrefix,
	}
}

// GenerateAPIKey handles POST /api/admin/generate-key
func (h *AdminHandler) GenerateAPIKey(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
		Plan  string `json:"plan"`
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
		Email:  req.Email,
		Plan:   req.Plan,
		Active: true,
	}

	if err := h.keyStore.AddAccount(account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": apiKey,
		"email":   req.Email,
		"plan":    req.Plan,
	})
}

// ListAccounts handles GET /api/admin/accounts
func (h *AdminHandler) ListAccounts(c *gin.Context) {
	accounts := h.keyStore.ListAccounts()
	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

// AddAccount handles POST /api/admin/accounts
func (h *AdminHandler) AddAccount(c *gin.Context) {
	var account auth.Account
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if account.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "api_key is required"})
		return
	}

	if err := h.keyStore.AddAccount(&account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "account added"})
}

// RemoveAccount handles DELETE /api/admin/accounts/:apikey
func (h *AdminHandler) RemoveAccount(c *gin.Context) {
	apiKey := c.Param("apikey")

	if err := h.keyStore.RemoveAccount(apiKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "account removed"})
}

// GetAccount handles GET /api/admin/accounts/:apikey
func (h *AdminHandler) GetAccount(c *gin.Context) {
	apiKey := c.Param("apikey")

	account, ok := h.keyStore.GetAccount(apiKey)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}
