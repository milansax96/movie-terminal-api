package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// parseUserID extracts and validates the user_id from the Gin context.
// Returns uuid.Nil and false (after writing a 401 response) if invalid.
func parseUserID(c *gin.Context) (uuid.UUID, bool) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})

		return uuid.Nil, false
	}

	return uid, true
}
