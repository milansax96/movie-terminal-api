package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/pkg/tmdb"
)

// AuthServiceInterface defines the contract for authentication operations.
type AuthServiceInterface interface {
	GoogleLogin(ctx context.Context, idToken string) (*AuthResult, error)
}

// UserServiceInterface defines the contract for user profile operations.
type UserServiceInterface interface {
	GetProfile(userID uuid.UUID) (*models.User, error)
	UpdateStreamingServices(userID uuid.UUID, serviceIDs []int) error
}

// MovieServiceInterface defines the contract for movie and watchlist operations.
type MovieServiceInterface interface {
	Discover(genre string, page int) ([]tmdb.Movie, error)
	Search(query string, page int) ([]tmdb.Movie, error)
	GetDetail(mediaType string, id int) (*tmdb.MovieDetail, error)
	GetVideos(mediaType string, id int) ([]tmdb.Video, error)
	GetCredits(mediaType string, id int) (*tmdb.CreditsResponse, error)
	GetProviders(mediaType string, id int) (json.RawMessage, error)
	AddToWatchlist(userID uuid.UUID, req AddWatchlistRequest) (*models.Watchlist, error)
	GetWatchlist(userID uuid.UUID) ([]models.Watchlist, error)
	RemoveFromWatchlist(userID uuid.UUID, movieID int) error
	CheckWatchlist(userID uuid.UUID, movieID int) (bool, error)
}

// SocialServiceInterface defines the contract for social/friend operations.
type SocialServiceInterface interface {
	GetFriends(userID uuid.UUID) ([]models.Friendship, error)
	SearchUsers(query string) ([]models.User, error)
	SendFriendRequest(userID uuid.UUID, friendID uuid.UUID) (*models.Friendship, error)
	AcceptFriendRequest(requestID uuid.UUID, friendID uuid.UUID) (*models.Friendship, error)
	GetFriendsFeed(userID uuid.UUID) ([]models.Post, error)
	CreatePost(userID uuid.UUID, tmdbID int, mediaType string, blurb string) (*models.Post, error)
}
