package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/pkg/tmdb"
)

func GetDiscoverFeed(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var user models.User
		if err := db.Preload("StreamingServices").First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		}

		tmdbClient := tmdb.NewClient()
		movies, err := tmdbClient.GetTrending("all", "week")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trending"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": movies,
		})
	}
}

func AddToWatchlist(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req struct {
			TMDBId    int    `json:"tmdb_id" binding:"required"`
			MediaType string `json:"media_type" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		watchlist := models.Watchlist{
			UserID:    uuid.MustParse(userID),
			TMDBId:    req.TMDBId,
			MediaType: req.MediaType,
		}

		if err := db.Create(&watchlist).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to watchlist"})
			return
		}

		c.JSON(http.StatusCreated, watchlist)
	}
}
