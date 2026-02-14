package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

func GetFriends(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var friendships []models.Friendship
		if err := db.Where("(user_id = ? OR friend_id = ?) AND status = ?", userID, userID, "accepted").Find(&friendships).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch friends"})
			return
		}

		c.JSON(http.StatusOK, friendships)
	}
}

func SearchUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
			return
		}

		var users []models.User
		if err := db.Where("username ILIKE ?", "%"+query+"%").Limit(20).Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func SendFriendRequest(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req struct {
			FriendID string `json:"friend_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		friendship := models.Friendship{
			UserID:   uuid.MustParse(userID),
			FriendID: uuid.MustParse(req.FriendID),
			Status:   "pending",
		}

		if err := db.Create(&friendship).Error; err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Friend request already exists"})
			return
		}

		c.JSON(http.StatusCreated, friendship)
	}
}

func AcceptFriendRequest(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Param("id")
		userID := c.GetString("user_id")

		var friendship models.Friendship
		if err := db.First(&friendship, "id = ? AND friend_id = ?", requestID, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found"})
			return
		}

		friendship.Status = "accepted"
		if err := db.Save(&friendship).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept request"})
			return
		}

		c.JSON(http.StatusOK, friendship)
	}
}
