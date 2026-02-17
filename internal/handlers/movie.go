package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/internal/service"
)

// MovieHandler handles movie, discovery, and watchlist endpoints.
type MovieHandler struct {
	svc service.MovieServiceInterface
}

// NewMovieHandler creates a new MovieHandler.
func NewMovieHandler(svc service.MovieServiceInterface) *MovieHandler {
	return &MovieHandler{svc: svc}
}

// GetDiscoverFeed returns movies for a genre or trending feed.
func (h *MovieHandler) GetDiscoverFeed(c *gin.Context) {
	genre := c.DefaultQuery("genre", "trending")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	val, _ := c.Get("user_id")
	uid, ok := val.(uuid.UUID)
	if !ok {
		uid = uuid.Nil // Handle guest mode
	}

	movies, err := h.svc.Discover(uid, genre, page)
	if err != nil {
		if errors.Is(err, service.ErrUnknownGenre) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown genre: " + genre})

			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"results": movies})
}

// SearchMovies searches for movies and TV shows.
func (h *MovieHandler) SearchMovies(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q parameter is required"})

		return
	}

	val, _ := c.Get("user_id")
	uid, ok := val.(uuid.UUID)
	if !ok {
		uid = uuid.Nil // Handle guest mode
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	movies, err := h.svc.Search(uid, query, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search movies"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"results": movies})
}

// GetMovieDetail returns full details for a movie or TV show.
func (h *MovieHandler) GetMovieDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})

		return
	}

	mediaType := c.DefaultQuery("media_type", "movie")

	detail, err := h.svc.GetDetail(mediaType, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movie details"})

		return
	}

	detail.MediaType = mediaType
	c.JSON(http.StatusOK, detail)
}

// GetMovieVideos returns videos for a movie or TV show.
func (h *MovieHandler) GetMovieVideos(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})

		return
	}

	mediaType := c.DefaultQuery("media_type", "movie")

	videos, err := h.svc.GetVideos(mediaType, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch videos"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"results": videos})
}

// GetMovieCredits returns credits for a movie or TV show.
func (h *MovieHandler) GetMovieCredits(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})

		return
	}

	mediaType := c.DefaultQuery("media_type", "movie")

	credits, err := h.svc.GetCredits(mediaType, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch credits"})

		return
	}

	c.JSON(http.StatusOK, credits)
}

// GetMovieProviders returns streaming providers for a title.
func (h *MovieHandler) GetMovieProviders(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})

		return
	}

	mediaType := c.DefaultQuery("media_type", "movie")

	providers, err := h.svc.GetProviders(mediaType, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})

		return
	}

	c.JSON(http.StatusOK, providers)
}

// GetWatchlist returns the user's watchlist.
func (h *MovieHandler) GetWatchlist(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))

	items, err := h.svc.GetWatchlist(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch watchlist"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"results": items})
}

// AddToWatchlist adds a movie to the user's watchlist.
func (h *MovieHandler) AddToWatchlist(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))

	var req struct {
		MovieID      int    `json:"movie_id" binding:"required"`
		Title        string `json:"title"`
		PosterPath   string `json:"poster_path"`
		BackdropPath string `json:"backdrop_path"`
		MediaType    string `json:"media_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	item, err := h.svc.AddToWatchlist(userID, models.Movie{
		ID:           req.MovieID,
		Title:        req.Title,
		PosterPath:   req.PosterPath,
		BackdropPath: req.BackdropPath,
		MediaType:    req.MediaType,
	})
	if err != nil {
		if errors.Is(err, service.ErrAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Movie already in watchlist"})

			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to watchlist"})

		return
	}

	c.JSON(http.StatusCreated, item)
}

// RemoveFromWatchlist removes a movie from the user's watchlist.
func (h *MovieHandler) RemoveFromWatchlist(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))
	movieID, err := strconv.Atoi(c.Param("movie_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})

		return
	}

	if err := h.svc.RemoveFromWatchlist(userID, movieID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found in watchlist"})

			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from watchlist"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from watchlist"})
}

// CheckWatchlist checks if a movie is in the user's watchlist.
func (h *MovieHandler) CheckWatchlist(c *gin.Context) {
	userID := uuid.MustParse(c.GetString("user_id"))
	movieID, err := strconv.Atoi(c.Param("movie_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})

		return
	}

	saved, err := h.svc.CheckWatchlist(userID, movieID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check watchlist"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"saved": saved})
}
