package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pigeonsec/kestrel/internal/auth"
)

type AuthHandler struct {
	jwtManager *auth.JWTManager
	// In production, use a proper user store (database, LDAP, etc.)
	adminUsername string
	adminPassword string
}

func NewAuthHandler(jwtManager *auth.JWTManager, adminUsername, adminPassword string) *AuthHandler {
	return &AuthHandler{
		jwtManager:    jwtManager,
		adminUsername: adminUsername,
		adminPassword: adminPassword,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Simple authentication - in production, use proper password hashing and user store
	if req.Username != h.adminUsername || req.Password != h.adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.Generate(req.Username, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User: User{
			Username: req.Username,
			IsAdmin:  true,
		},
	})
}

func (h *AuthHandler) Verify(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	jwtClaims := claims.(*auth.Claims)
	c.JSON(http.StatusOK, User{
		Username: jwtClaims.Username,
		IsAdmin:  jwtClaims.IsAdmin,
	})
}
