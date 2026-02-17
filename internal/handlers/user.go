package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/milansax96/movie-terminal-api/internal/service"
)

// UserHandler handles user profile endpoints.
type UserHandler struct {
	svc service.UserServiceInterface
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc service.UserServiceInterface) *UserHandler {
	return &UserHandler{svc: svc}
}

// GetProfile returns the authenticated user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))

	user, err := h.svc.GetProfile(userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})

			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})

		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateStreamingServices updates the user's streaming service preferences.
func (h *UserHandler) UpdateStreamingServices(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))

	var req struct {
		ServiceIDs []int `json:"service_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	if err := h.svc.UpdateStreamingServices(userID, req.ServiceIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update services"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Streaming services updated"})
}
