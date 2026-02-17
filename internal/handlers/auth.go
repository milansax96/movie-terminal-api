// Package handlers implements HTTP request handlers for the API.
package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/milansax96/movie-terminal-api/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	svc service.AuthServiceInterface
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// GoogleLogin handles Google OAuth login requests.
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req struct {
		AccessToken string `json:"access_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	result, err := h.svc.GoogleLogin(c.Request.Context(), req.AccessToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google access token"})
		case errors.Is(err, service.ErrMissingClaims):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Google token missing required claims"})
		case errors.Is(err, service.ErrAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Account with this email already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login"})
		}

		return
	}

	status := http.StatusOK
	if result.IsNew {
		status = http.StatusCreated
	}

	c.JSON(status, gin.H{
		"user":  result.User,
		"token": result.Token,
	})
}
