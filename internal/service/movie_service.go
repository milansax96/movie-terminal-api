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

// NewMovieService creates and returns a new MovieService instance.
func NewMovieService(tmdbClient tmdb.API, watchlistRepo repository.WatchlistRepository) *MovieService {
	return &MovieService{
		tmdb:          tmdbClient,
		watchlistRepo: watchlistRepo,
	}
}

// Discover handles categorized movie fetching and enriches results with user watchlist status.
func (s *MovieService) Discover(userID uuid.UUID, genre string, page int) ([]models.Movie, error) {
	var movies []models.Movie
	var err error

	switch genre {
	case "trending":
		movies, err = s.tmdb.GetTrending("all", "week")
	case "top_rated":
		movies, err = s.tmdb.GetTopRated(page)
	default:
		genreID, ok := genreMap[genre]
		if !ok {
			return nil, ErrUnknownGenre
		}
		movies, err = s.tmdb.DiscoverByGenre(genreID, page)
	}

	if err != nil {
		return nil, err
	}

	return s.enrichWithWatchlist(userID, movies)
}

// Search queries TMDB and checks which results the user has already saved.
func (s *MovieService) Search(userID uuid.UUID, query string, page int) ([]models.Movie, error) {
	movies, err := s.tmdb.SearchMovies(query, page)
	if err != nil {
		return nil, err
	}

	return s.enrichWithWatchlist(userID, movies)
}

// GetDetail returns full details (No changes needed here unless you create a Detail domain model too).
func (s *MovieService) GetDetail(mediaType string, id int) (*tmdb.MovieDetail, error) {
	return s.tmdb.GetMovieDetails(mediaType, id)
}

// GetVideos returns videos for a movie or TV show.
func (s *MovieService) GetVideos(mediaType string, id int) ([]tmdb.Video, error) {
	return s.tmdb.GetVideos(mediaType, id)
}

// GetCredits returns credits for a movie or TV show.
func (s *MovieService) GetCredits(mediaType string, id int) (*tmdb.CreditsResponse, error) {
	return s.tmdb.GetCredits(mediaType, id)
}

// GetProviders returns streaming providers for a title.
func (s *MovieService) GetProviders(mediaType string, id int) (json.RawMessage, error) {
	return s.tmdb.GetProviders(mediaType, id)
}

// GetWatchlist returns all watchlist items for a user.
func (s *MovieService) GetWatchlist(userID uuid.UUID) ([]models.Watchlist, error) {
	return s.watchlistRepo.GetByUserID(userID)
}

// CheckWatchlist checks if a movie is in the user's watchlist.
func (s *MovieService) CheckWatchlist(userID uuid.UUID, movieID int) (bool, error) {
	return s.watchlistRepo.Exists(userID, movieID)
}

// AddToWatchlist converts a request into a Database Watchlist model.
func (s *MovieService) AddToWatchlist(userID uuid.UUID, req models.Movie) (*models.Watchlist, error) {
	item := &models.Watchlist{
		UserID:       userID,
		TMDBId:       req.ID,
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

// RemoveFromWatchlist removes an item and validates if it existed.
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

// enrichWithWatchlist marks movies as 'IsWatchlisted' based on the user's data.
func (s *MovieService) enrichWithWatchlist(userID uuid.UUID, movies []models.Movie) ([]models.Movie, error) {
	watchlist, err := s.watchlistRepo.GetByUserID(userID)
	if err != nil {
		return movies, nil // Fail silently on enrichment; better to show movies without icons than error.
	}

	// Create map for O(1) lookup
	saved := make(map[int]struct{})
	for _, item := range watchlist {
		saved[item.TMDBId] = struct{}{}
	}

	for i := range movies {
		if _, exists := saved[movies[i].ID]; exists {
			movies[i].IsWatchlisted = true
		}
	}

	return movies, nil
}
