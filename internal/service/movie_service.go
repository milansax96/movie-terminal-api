package service

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/internal/repository"
	"github.com/milansax96/movie-terminal-api/pkg/tmdb"
)

var genreMap = map[string]int{
	"action":    28,
	"comedy":    35,
	"horror":    27,
	"romance":   10749,
	"mystery":   9648,
	"sci_fi":    878,
	"western":   37,
	"animation": 16,
	"tv_movie":  10770,
}

// MovieService handles movie discovery, search, and watchlist operations.
type MovieService struct {
	tmdb          tmdb.API
	watchlistRepo repository.WatchlistRepository
}

// NewMovieService creates a new MovieService.
func NewMovieService(tmdbClient tmdb.API, watchlistRepo repository.WatchlistRepository) *MovieService {
	return &MovieService{tmdb: tmdbClient, watchlistRepo: watchlistRepo}
}

// Discover returns movies for the given genre or trending feed.
func (s *MovieService) Discover(genre string, page int) ([]tmdb.Movie, error) {
	switch genre {
	case "trending":
		return s.tmdb.GetTrending("all", "week")
	case "top_rated":
		return s.tmdb.GetTopRated(page)
	default:
		genreID, ok := genreMap[genre]
		if !ok {
			return nil, ErrUnknownGenre
		}

		return s.tmdb.DiscoverByGenre(genreID, page)
	}
}

// Search finds movies and TV shows matching the query.
func (s *MovieService) Search(query string, page int) ([]tmdb.Movie, error) {
	return s.tmdb.SearchMovies(query, page)
}

// GetDetail returns full details for a movie or TV show.
func (s *MovieService) GetDetail(mediaType string, id int) (*tmdb.MovieDetail, error) {
	return s.tmdb.GetMovieDetails(mediaType, id)
}

// GetVideos returns videos for a title.
func (s *MovieService) GetVideos(mediaType string, id int) ([]tmdb.Video, error) {
	return s.tmdb.GetVideos(mediaType, id)
}

// GetCredits returns credits for a title.
func (s *MovieService) GetCredits(mediaType string, id int) (*tmdb.CreditsResponse, error) {
	return s.tmdb.GetCredits(mediaType, id)
}

// GetProviders returns streaming provider information for a title.
func (s *MovieService) GetProviders(mediaType string, id int) (json.RawMessage, error) {
	return s.tmdb.GetProviders(mediaType, id)
}

// AddWatchlistRequest holds the data needed to add a movie to the watchlist.
type AddWatchlistRequest struct {
	MovieID      int
	Title        string
	PosterPath   string
	BackdropPath string
	MediaType    string
}

// AddToWatchlist adds a movie to the user's watchlist.
func (s *MovieService) AddToWatchlist(userID uuid.UUID, req AddWatchlistRequest) (*models.Watchlist, error) {
	item := &models.Watchlist{
		UserID:       userID,
		TMDBId:       req.MovieID,
		Title:        req.Title,
		PosterPath:   req.PosterPath,
		BackdropPath: req.BackdropPath,
		MediaType:    req.MediaType,
		AddedAt:      time.Now(),
	}

	if err := s.watchlistRepo.Add(item); err != nil {
		return nil, ErrAlreadyExists
	}

	return item, nil
}

// GetWatchlist returns all items in the user's watchlist.
func (s *MovieService) GetWatchlist(userID uuid.UUID) ([]models.Watchlist, error) {
	return s.watchlistRepo.GetByUserID(userID)
}

// RemoveFromWatchlist removes a movie from the user's watchlist.
func (s *MovieService) RemoveFromWatchlist(userID uuid.UUID, movieID int) error {
	rows, err := s.watchlistRepo.Remove(userID, movieID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// CheckWatchlist checks if a movie is in the user's watchlist.
func (s *MovieService) CheckWatchlist(userID uuid.UUID, movieID int) (bool, error) {
	return s.watchlistRepo.Exists(userID, movieID)
}
