package tmdb

import (
	"encoding/json"
	"fmt"
	"time"

	cache "github.com/patrickmn/go-cache"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

// TTL constants for different endpoint types.
const (
	ttlTrending  = 1 * time.Hour
	ttlTopRated  = 6 * time.Hour
	ttlGenre     = 3 * time.Hour
	ttlSearch    = 30 * time.Minute
	ttlDetail    = 24 * time.Hour
	ttlVideos    = 24 * time.Hour
	ttlCredits   = 24 * time.Hour
	ttlProviders = 6 * time.Hour

	cleanupInterval = 10 * time.Minute
)

// CachedClient wraps a tmdb.API and caches responses in memory.
type CachedClient struct {
	inner API
	store *cache.Cache
}

// NewCachedClient returns a caching decorator around any tmdb.API implementation.
func NewCachedClient(inner API) *CachedClient {

	return &CachedClient{
		inner: inner,
		store: cache.New(ttlDetail, cleanupInterval),
	}
}

// cacheGet is a generic helper: return from cache on hit, otherwise fetch, store, and return.
func cacheGet[T any](c *CachedClient, key string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	if cached, found := c.store.Get(key); found {

		return cached.(T), nil
	}

	result, err := fetch()
	if err != nil {
		var zero T

		return zero, err
	}

	c.store.Set(key, result, ttl)

	return result, nil
}

// GetTrending returns trending titles, cached for 1 hour.
func (c *CachedClient) GetTrending(mediaType string, timeWindow string) ([]models.Movie, error) {
	key := fmt.Sprintf("trending:%s:%s", mediaType, timeWindow)

	return cacheGet(c, key, ttlTrending, func() ([]models.Movie, error) {
		return c.inner.GetTrending(mediaType, timeWindow)
	})
}

// GetTopRated returns top-rated movies, cached for 6 hours.
func (c *CachedClient) GetTopRated(page int) ([]models.Movie, error) {
	key := fmt.Sprintf("top_rated:%d", page)

	return cacheGet(c, key, ttlTopRated, func() ([]models.Movie, error) {
		return c.inner.GetTopRated(page)
	})
}

// DiscoverByGenre returns movies by genre, cached for 3 hours.
func (c *CachedClient) DiscoverByGenre(genreID int, page int) ([]models.Movie, error) {
	key := fmt.Sprintf("genre:%d:%d", genreID, page)

	return cacheGet(c, key, ttlGenre, func() ([]models.Movie, error) {
		return c.inner.DiscoverByGenre(genreID, page)
	})
}

// SearchMovies returns search results, cached for 30 minutes.
func (c *CachedClient) SearchMovies(query string, page int) ([]models.Movie, error) {
	key := fmt.Sprintf("search:%s:%d", query, page)

	return cacheGet(c, key, ttlSearch, func() ([]models.Movie, error) {
		return c.inner.SearchMovies(query, page)
	})
}

// GetMovieDetails returns movie details, cached for 24 hours.
func (c *CachedClient) GetMovieDetails(mediaType string, id int) (*MovieDetail, error) {
	key := fmt.Sprintf("detail:%s:%d", mediaType, id)

	return cacheGet(c, key, ttlDetail, func() (*MovieDetail, error) {
		return c.inner.GetMovieDetails(mediaType, id)
	})
}

// GetVideos returns videos for a title, cached for 24 hours.
func (c *CachedClient) GetVideos(mediaType string, id int) ([]Video, error) {
	key := fmt.Sprintf("videos:%s:%d", mediaType, id)

	return cacheGet(c, key, ttlVideos, func() ([]Video, error) {
		return c.inner.GetVideos(mediaType, id)
	})
}

// GetCredits returns credits for a title, cached for 24 hours.
func (c *CachedClient) GetCredits(mediaType string, id int) (*CreditsResponse, error) {
	key := fmt.Sprintf("credits:%s:%d", mediaType, id)

	return cacheGet(c, key, ttlCredits, func() (*CreditsResponse, error) {
		return c.inner.GetCredits(mediaType, id)
	})
}

// GetProviders returns streaming providers, cached for 6 hours.
func (c *CachedClient) GetProviders(mediaType string, id int) (json.RawMessage, error) {
	key := fmt.Sprintf("providers:%s:%d", mediaType, id)

	return cacheGet(c, key, ttlProviders, func() (json.RawMessage, error) {
		return c.inner.GetProviders(mediaType, id)
	})
}
