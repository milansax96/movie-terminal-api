package service

import (
	"encoding/json"
	"sync"
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
	tmdb                tmdb.API
	watchlistRepo       repository.WatchlistRepository
	cloudinaryCloudName string
}

// NewMovieService creates and returns a new MovieService instance.
func NewMovieService(tmdbClient tmdb.API, watchlistRepo repository.WatchlistRepository, cloudinaryCloudName string) *MovieService {

	return &MovieService{
		tmdb:                tmdbClient,
		watchlistRepo:       watchlistRepo,
		cloudinaryCloudName: cloudinaryCloudName,
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
	case "now_playing":
		movies, err = s.tmdb.GetNowPlaying(page)
	case "popular":
		movies, err = s.tmdb.GetPopular(page)
	case "upcoming":
		movies, err = s.tmdb.GetUpcoming(page)
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
func (s *MovieService) DiscoverAll(userID uuid.UUID) ([]models.Movie, error) {
	categories := []string{"trending", "now_playing", "upcoming"}
	uniqueMovies := make(map[int]models.Movie)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, cat := range categories {
		wg.Add(1)
		go func(category string) {
			defer wg.Done()

			movies, err := s.Discover(userID, category, 1)
			if err != nil {
				return
			}

			var videoWg sync.WaitGroup
			for i := range movies {
				videoWg.Add(1)
				go func(idx int) {
					defer videoWg.Done()

					mType := movies[idx].MediaType
					if mType == "" {
						mType = "movie"
					}

					videos, err := s.GetVideos(mType, movies[idx].ID)
					if err == nil && len(videos) > 0 {
						for _, v := range videos {
							// Look for official YouTube trailers
							if v.Site == "YouTube" && v.Type == "Trailer" {
								movies[idx].TrailerKey = v.Key
								break
							}
						}
					}
				}(i)
			}
			videoWg.Wait()

			// Final Filter: Only keep movies with a valid trailer key
			mu.Lock()
			for _, m := range movies {
				if m.TrailerKey != "" {
					uniqueMovies[m.ID] = m
				}
			}
			mu.Unlock()
		}(cat)
	}

	wg.Wait()

	finalResults := make([]models.Movie, 0, len(uniqueMovies))
	for _, movie := range uniqueMovies {
		finalResults = append(finalResults, movie)
	}

	// s.enrichWithSmartCrop(finalResults)

	return s.enrichWithWatchlist(userID, finalResults)
}

// enrichAllWithWatchlist marks IsWatchlisted across all categories using a single DB query.
func (s *MovieService) enrichAllWithWatchlist(userID uuid.UUID, feed map[string][]models.Movie) (map[string][]models.Movie, error) {
	if userID == uuid.Nil {
		return feed, nil
	}

	watchlist, err := s.watchlistRepo.GetByUserID(userID)
	if err != nil {
		return feed, nil
	}

	saved := make(map[int]struct{})
	for _, item := range watchlist {
		saved[item.TMDBId] = struct{}{}
	}

	for cat := range feed {
		for i := range feed[cat] {
			if _, exists := saved[feed[cat][i].ID]; exists {
				feed[cat][i].IsWatchlisted = true
			}
		}
	}

	return feed, nil
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
	println("Request trailer", req.TrailerKey) // Debug log to verify trailer key presence

	item := &models.Watchlist{
		UserID:       userID,
		TMDBId:       req.ID,
		Title:        req.Title,
		PosterPath:   req.PosterPath,
		BackdropPath: req.BackdropPath,
		MediaType:    req.MediaType,
		TrailerKey:   req.TrailerKey,
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

// enrichWithSmartCrop generates Cloudinary smart-crop URLs and fires warm-up requests.
// func (s *MovieService) enrichWithSmartCrop(movies []models.Movie) {
// 	if s.cloudinaryCloudName == "" {
// 		return
// 	}

// 	for i := range movies {
// 		if movies[i].TrailerKey == "" {
// 			continue
// 		}

// 		url := cloudinary.GenerateSmartCropURL(s.cloudinaryCloudName, movies[i].TrailerKey)
// 		movies[i].ProcessedVideoURL = url
// 		cloudinary.WarmUp(url)
// 	}
// }

// enrichWithWatchlist marks movies as 'IsWatchlisted' based on the user's data.
func (s *MovieService) enrichWithWatchlist(userID uuid.UUID, movies []models.Movie) ([]models.Movie, error) {
	if userID == uuid.Nil {
		return movies, nil
	}

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
