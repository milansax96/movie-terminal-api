package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/pkg/tmdb"
)

// --- Discover ---

func TestDiscover(t *testing.T) {
	tests := map[string]struct {
		genre  string
		page   int
		setup  func(*TestEnv)
		movies []tmdb.Movie
		err    error
	}{
		"trending": {
			"trending",
			1,
			func(env *TestEnv) {
				env.TMDB.ReturnsTrending(
					[]tmdb.Movie{
						{
							ID:    1,
							Title: "Trending Movie",
						},
					},
				)
			},
			[]tmdb.Movie{
				{ID: 1, Title: "Trending Movie"},
			},
			nil,
		},
		"top rated": {"top_rated", 1, func(env *TestEnv) {
			env.TMDB.ReturnsTopRated(1, []tmdb.Movie{{ID: 2, Title: "Top Rated"}})
		}, []tmdb.Movie{{ID: 2, Title: "Top Rated"}}, nil},
		"action genre": {"action", 1, func(env *TestEnv) {
			env.TMDB.ReturnsGenre(28, 1, []tmdb.Movie{{ID: 3, Title: "Action Movie"}})
		}, []tmdb.Movie{{ID: 3, Title: "Action Movie"}}, nil},
		"comedy genre page 2": {"comedy", 2, func(env *TestEnv) {
			env.TMDB.ReturnsGenre(35, 2, []tmdb.Movie{{ID: 4, Title: "Comedy Movie"}})
		}, []tmdb.Movie{{ID: 4, Title: "Comedy Movie"}}, nil},
		"unknown genre": {"nonexistent", 1, func(_ *TestEnv) {}, nil, ErrUnknownGenre},
		"tmdb error": {"trending", 1, func(env *TestEnv) {
			env.TMDB.TrendingFails(errors.New("network error"))
		}, nil, errors.New("network error")},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			tt.setup(env)

			movies, err := env.MovieService().Discover(tt.genre, tt.page)

			if tt.err != nil {
				assert.Error(t, err)
				if errors.Is(tt.err, ErrUnknownGenre) {
					assert.ErrorIs(t, err, ErrUnknownGenre)
				} else {
					assert.Contains(t, err.Error(), tt.err.Error())
				}

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.movies, movies)
		})
	}
}

func TestDiscover_GenreMapContainsExpectedKeys(t *testing.T) {
	expected := []string{"action", "comedy", "horror", "romance", "mystery", "sci_fi", "western", "animation", "tv_movie"}
	for _, key := range expected {
		_, ok := genreMap[key]
		assert.True(t, ok, "genreMap missing key: %s", key)
	}
}

// --- Search ---

func TestSearch(t *testing.T) {
	tests := map[string]struct {
		query  string
		page   int
		setup  func(*TestEnv)
		movies []tmdb.Movie
		hasErr bool
	}{
		"success": {"fight club", 1, func(env *TestEnv) {
			env.TMDB.SearchReturns("fight club", 1, []tmdb.Movie{{ID: 550, Title: "Fight Club"}})
		}, []tmdb.Movie{{ID: 550, Title: "Fight Club"}}, false},
		"tmdb error": {"query", 1, func(env *TestEnv) {
			env.TMDB.SearchFails("query", 1, errors.New("timeout"))
		}, nil, true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			tt.setup(env)

			movies, err := env.MovieService().Search(tt.query, tt.page)

			if tt.hasErr {
				assert.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.movies, movies)
		})
	}
}

// --- GetDetail ---

func TestGetDetail(t *testing.T) {
	env := newTestEnv(t)
	expected := &tmdb.MovieDetail{ID: 550, Title: "Fight Club"}
	env.TMDB.ReturnsDetails("movie", 550, expected)

	detail, err := env.MovieService().GetDetail("movie", 550)
	require.NoError(t, err)
	assert.Equal(t, expected, detail)
}

// --- GetVideos ---

func TestGetVideos(t *testing.T) {
	env := newTestEnv(t)
	expected := []tmdb.Video{{Key: "abc", Name: "Trailer", Site: "YouTube", Type: "Trailer"}}
	env.TMDB.ReturnsVideos("movie", 550, expected)

	videos, err := env.MovieService().GetVideos("movie", 550)
	require.NoError(t, err)
	assert.Equal(t, expected, videos)
}

// --- GetCredits ---

func TestGetCredits(t *testing.T) {
	env := newTestEnv(t)
	expected := &tmdb.CreditsResponse{Cast: []tmdb.CastMember{{ID: 1, Name: "Brad Pitt"}}}
	env.TMDB.ReturnsCredits("movie", 550, expected)

	credits, err := env.MovieService().GetCredits("movie", 550)
	require.NoError(t, err)
	assert.Equal(t, expected, credits)
}

// --- Watchlist ---

func TestAddToWatchlist(t *testing.T) {
	tests := map[string]struct {
		setup func(*TestEnv)
		err   error
		check func(*testing.T, *models.Watchlist)
	}{
		"success": {func(env *TestEnv) {
			env.Watchlist.AddsItem()
		}, nil, func(t *testing.T, item *models.Watchlist) {
			assert.Equal(t, 550, item.TMDBId)
			assert.Equal(t, "Fight Club", item.Title)
		}},
		"duplicate": {func(env *TestEnv) {
			env.Watchlist.AddFails(errors.New("unique constraint violation"))
		}, ErrAlreadyExists, nil},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			tt.setup(env)

			item, err := env.MovieService().AddToWatchlist(uuid.New(), AddWatchlistRequest{
				MovieID: 550, Title: "Fight Club", MediaType: "movie",
			})

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)

				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, item)
			}
		})
	}
}

func TestRemoveFromWatchlist(t *testing.T) {
	tests := map[string]struct {
		movieID int
		setup   func(*TestEnv, uuid.UUID)
		err     error
	}{
		"success": {550, func(env *TestEnv, userID uuid.UUID) {
			env.Watchlist.RemovesItem(userID, 550)
		}, nil},
		"not found": {999, func(env *TestEnv, userID uuid.UUID) {
			env.Watchlist.ItemNotFound(userID, 999)
		}, ErrNotFound},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			userID := uuid.New()
			tt.setup(env, userID)

			err := env.MovieService().RemoveFromWatchlist(userID, tt.movieID)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)

				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCheckWatchlist(t *testing.T) {
	tests := map[string]struct {
		movieID int
		exists  bool
	}{
		"saved":     {550, true},
		"not saved": {999, false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			userID := uuid.New()
			env.Watchlist.ItemExists(userID, tt.movieID, tt.exists)

			saved, err := env.MovieService().CheckWatchlist(userID, tt.movieID)
			require.NoError(t, err)
			assert.Equal(t, tt.exists, saved)
		})
	}
}

func TestGetWatchlist(t *testing.T) {
	env := newTestEnv(t)
	userID := uuid.New()
	expected := []models.Watchlist{
		{TMDBId: 550, Title: "Fight Club"},
		{TMDBId: 680, Title: "Pulp Fiction"},
	}
	env.Watchlist.ReturnsWatchlist(userID, expected)

	items, err := env.MovieService().GetWatchlist(userID)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "Fight Club", items[0].Title)
}
