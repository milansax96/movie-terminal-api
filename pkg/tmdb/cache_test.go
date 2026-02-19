package tmdb_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/pkg/tmdb"
	tmdbMocks "github.com/milansax96/movie-terminal-api/pkg/tmdb/mocks"
)

func newCachedClient(t *testing.T) (*tmdb.CachedClient, *tmdbMocks.MockAPI) {
	t.Helper()

	inner := tmdbMocks.NewMockAPI(t)
	cached := tmdb.NewCachedClient(inner)

	return cached, inner
}

func TestGetTrending_CacheHit(t *testing.T) {
	client, inner := newCachedClient(t)
	movies := []models.Movie{{ID: 1, Title: "Trending"}}
	inner.On("GetTrending", "all", "week").Return(movies, nil).Once()

	first, err := client.GetTrending("all", "week")
	assert.NoError(t, err)
	assert.Equal(t, movies, first)

	second, err := client.GetTrending("all", "week")
	assert.NoError(t, err)
	assert.Equal(t, movies, second)

	inner.AssertNumberOfCalls(t, "GetTrending", 1)
}

func TestGetTrending_ErrorNotCached(t *testing.T) {
	client, inner := newCachedClient(t)
	inner.On("GetTrending", "all", "week").Return([]models.Movie(nil), errors.New("network error")).Once()
	inner.On("GetTrending", "all", "week").Return([]models.Movie{{ID: 1}}, nil).Once()

	_, err := client.GetTrending("all", "week")
	assert.Error(t, err)

	movies, err := client.GetTrending("all", "week")
	assert.NoError(t, err)
	assert.Len(t, movies, 1)

	inner.AssertNumberOfCalls(t, "GetTrending", 2)
}

func TestGetTopRated_CacheHit(t *testing.T) {
	client, inner := newCachedClient(t)
	movies := []models.Movie{{ID: 2, Title: "Top Rated"}}
	inner.On("GetTopRated", 1).Return(movies, nil).Once()

	first, _ := client.GetTopRated(1)
	second, _ := client.GetTopRated(1)
	assert.Equal(t, first, second)

	inner.AssertNumberOfCalls(t, "GetTopRated", 1)
}

func TestDiscoverByGenre_CacheHit(t *testing.T) {
	client, inner := newCachedClient(t)
	movies := []models.Movie{{ID: 3, Title: "Action Movie"}}
	inner.On("DiscoverByGenre", 28, 1).Return(movies, nil).Once()

	first, _ := client.DiscoverByGenre(28, 1)
	second, _ := client.DiscoverByGenre(28, 1)
	assert.Equal(t, first, second)

	inner.AssertNumberOfCalls(t, "DiscoverByGenre", 1)
}

func TestSearchMovies_CacheHit(t *testing.T) {
	client, inner := newCachedClient(t)
	movies := []models.Movie{{ID: 4, Title: "Inception"}}
	inner.On("SearchMovies", "inception", 1).Return(movies, nil).Once()

	first, _ := client.SearchMovies("inception", 1)
	second, _ := client.SearchMovies("inception", 1)
	assert.Equal(t, first, second)

	inner.AssertNumberOfCalls(t, "SearchMovies", 1)
}

func TestGetMovieDetails_CacheHit(t *testing.T) {
	client, inner := newCachedClient(t)
	detail := &tmdb.MovieDetail{ID: 550, Title: "Fight Club"}
	inner.On("GetMovieDetails", "movie", 550).Return(detail, nil).Once()

	first, _ := client.GetMovieDetails("movie", 550)
	second, _ := client.GetMovieDetails("movie", 550)
	assert.Equal(t, first, second)

	inner.AssertNumberOfCalls(t, "GetMovieDetails", 1)
}

func TestGetVideos_CacheHit(t *testing.T) {
	client, inner := newCachedClient(t)
	videos := []tmdb.Video{{Key: "abc123", Name: "Trailer", Site: "YouTube", Type: "Trailer"}}
	inner.On("GetVideos", "movie", 550).Return(videos, nil).Once()

	first, _ := client.GetVideos("movie", 550)
	second, _ := client.GetVideos("movie", 550)
	assert.Equal(t, first, second)

	inner.AssertNumberOfCalls(t, "GetVideos", 1)
}

func TestGetCredits_CacheHit(t *testing.T) {
	client, inner := newCachedClient(t)
	credits := &tmdb.CreditsResponse{Cast: []tmdb.CastMember{{ID: 1, Name: "Brad Pitt"}}}
	inner.On("GetCredits", "movie", 550).Return(credits, nil).Once()

	first, _ := client.GetCredits("movie", 550)
	second, _ := client.GetCredits("movie", 550)
	assert.Equal(t, first, second)

	inner.AssertNumberOfCalls(t, "GetCredits", 1)
}

func TestGetProviders_CacheHit(t *testing.T) {
	client, inner := newCachedClient(t)
	providers := json.RawMessage(`{"US":{"flatrate":[{"provider_name":"Netflix"}]}}`)
	inner.On("GetProviders", "movie", 550).Return(providers, nil).Once()

	first, _ := client.GetProviders("movie", 550)
	second, _ := client.GetProviders("movie", 550)
	assert.Equal(t, first, second)

	inner.AssertNumberOfCalls(t, "GetProviders", 1)
}

func TestDifferentKeysDontCollide(t *testing.T) {
	client, inner := newCachedClient(t)
	inner.On("GetMovieDetails", "movie", 550).Return(&tmdb.MovieDetail{ID: 550, Title: "Fight Club"}, nil).Once()
	inner.On("GetMovieDetails", "tv", 550).Return(&tmdb.MovieDetail{ID: 550, Title: "TV Show"}, nil).Once()

	movie, _ := client.GetMovieDetails("movie", 550)
	tv, _ := client.GetMovieDetails("tv", 550)
	assert.Equal(t, "Fight Club", movie.Title)
	assert.Equal(t, "TV Show", tv.Title)

	inner.AssertExpectations(t)
}

func TestSearchDifferentQueries(t *testing.T) {
	client, inner := newCachedClient(t)
	inner.On("SearchMovies", "inception", 1).Return([]models.Movie{{ID: 1}}, nil).Once()
	inner.On("SearchMovies", "matrix", 1).Return([]models.Movie{{ID: 2}}, nil).Once()

	r1, _ := client.SearchMovies("inception", 1)
	r2, _ := client.SearchMovies("matrix", 1)
	assert.Equal(t, 1, r1[0].ID)
	assert.Equal(t, 2, r2[0].ID)

	_ = mock.Anything
	inner.AssertExpectations(t)
}
