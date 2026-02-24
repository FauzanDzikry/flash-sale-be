package handler

import (
	"errors"
	"net/http"
	"time"

	"flash-sale-be/internal/dto"
	"flash-sale-be/internal/service"
	"flash-sale-be/internal/store"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
	blacklist   store.TokenBlacklist
}

func NewAuthHandler(authService service.AuthService, blacklist store.TokenBlacklist) *AuthHandler {
	return &AuthHandler{authService: authService, blacklist: blacklist}
}

// Register endpoint
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request", "error": err.Error()})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"message": "Email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to register"})
		return
	}
	c.JSON(http.StatusCreated, user)
}

// Login endpoint
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request", "error": err.Error()})
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to login"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Logout endpoint
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	if h.blacklist != nil {
		if raw, ok := c.Get("token_raw"); ok {
			if tokenStr, ok := raw.(string); ok {
				var expiresAt time.Time
				if exp, ok := c.Get("token_exp"); ok {
					if t, ok := exp.(time.Time); ok {
						expiresAt = t
					}
				}
				if expiresAt.IsZero() {
					expiresAt = time.Now().Add(24 * time.Hour)
				}
				h.blacklist.Add(tokenStr, expiresAt)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Me endpoint
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	idStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Invalid User Context"})
		return
	}

	user, err := h.authService.GetProfile(idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
