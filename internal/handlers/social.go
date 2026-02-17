package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/milansax96/movie-terminal-api/internal/service"
)

// SocialHandler handles friend and social feed endpoints.
type SocialHandler struct {
	svc service.SocialServiceInterface
}

// NewSocialHandler creates a new SocialHandler.
func NewSocialHandler(svc service.SocialServiceInterface) *SocialHandler {
	return &SocialHandler{svc: svc}
}

// GetFriends returns the user's accepted friendships.
func (h *SocialHandler) GetFriends(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))

	friendships, err := h.svc.GetFriends(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch friends"})

		return
	}

	c.JSON(http.StatusOK, friendships)
}

// SearchUsers searches for users by username.
func (h *SocialHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})

		return
	}

	users, err := h.svc.SearchUsers(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})

		return
	}

	c.JSON(http.StatusOK, users)
}

// SendFriendRequest sends a friend request to another user.
func (h *SocialHandler) SendFriendRequest(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))

	var req struct {
		FriendID string `json:"friend_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	friendship, err := h.svc.SendFriendRequest(userID, uuid.MustParse(req.FriendID))
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Friend request already exists"})

			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})

		return
	}

	c.JSON(http.StatusCreated, friendship)
}

// AcceptFriendRequest accepts a pending friend request.
func (h *SocialHandler) AcceptFriendRequest(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))
	requestID := uuid.MustParse(c.Param("id"))

	friendship, err := h.svc.AcceptFriendRequest(requestID, userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found"})

			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept request"})

		return
	}

	c.JSON(http.StatusOK, friendship)
}

// GetFriendsFeed returns posts from the user's friends.
func (h *SocialHandler) GetFriendsFeed(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))

	posts, err := h.svc.GetFriendsFeed(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feed"})

		return
	}

	c.JSON(http.StatusOK, posts)
}

// CreatePost creates a new post about a movie or TV show.
func (h *SocialHandler) CreatePost(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))

	var req struct {
		TMDBId    int    `json:"tmdb_id" binding:"required"`
		MediaType string `json:"media_type" binding:"required"`
		Blurb     string `json:"blurb"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	post, err := h.svc.CreatePost(userID, req.TMDBId, req.MediaType, req.Blurb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})

		return
	}

	c.JSON(http.StatusCreated, post)
}
