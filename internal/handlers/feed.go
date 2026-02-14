package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

func GetFriendsFeed(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var friendships []models.Friendship
		db.Where("(user_id = ? OR friend_id = ?) AND status = ?", userID, userID, "accepted").Find(&friendships)

		friendIDs := make([]uuid.UUID, 0)
		for _, f := range friendships {
			if f.UserID.String() == userID {
				friendIDs = append(friendIDs, f.FriendID)
			} else {
				friendIDs = append(friendIDs, f.UserID)
			}
		}

		var posts []models.Post
		if err := db.Preload("User").Where("user_id IN ?", friendIDs).Order("created_at DESC").Limit(50).Find(&posts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feed"})
			return
		}

		c.JSON(http.StatusOK, posts)
	}
}

func CreatePost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req struct {
			TMDBId    int    `json:"tmdb_id" binding:"required"`
			MediaType string `json:"media_type" binding:"required"`
			Blurb     string `json:"blurb"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		post := models.Post{
			UserID:    uuid.MustParse(userID),
			TMDBId:    req.TMDBId,
			MediaType: req.MediaType,
			Blurb:     req.Blurb,
		}

		if err := db.Create(&post).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
			return
		}

		c.JSON(http.StatusCreated, post)
	}
}
