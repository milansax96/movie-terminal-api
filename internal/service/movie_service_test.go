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
		setup  func(*TestEnv, uuid.UUID)
		movies []models.Movie
		err    error
	}{
		"trending": {
			"trending",
			1,
			func(env *TestEnv, userID uuid.UUID) {
				env.TMDB.ReturnsTrending(
					[]models.Movie{
						{
							ID:    1,
							Title: "Trending Movie",
						},
					},
				)
				env.Watchlist.ReturnsWatchlist(userID, nil)
			},
			[]models.Movie{
				{ID: 1, Title: "Trending Movie"},
			},
			nil,
		},
		"top rated": {"top_rated", 1, func(env *TestEnv, userID uuid.UUID) {
			env.TMDB.ReturnsTopRated(1, []models.Movie{{ID: 2, Title: "Top Rated"}})
			env.Watchlist.ReturnsWatchlist(userID, nil)
		}, []models.Movie{{ID: 2, Title: "Top Rated"}}, nil},
		"action genre": {"action", 1, func(env *TestEnv, userID uuid.UUID) {
			env.TMDB.ReturnsGenre(28, 1, []models.Movie{{ID: 3, Title: "Action Movie"}})
			env.Watchlist.ReturnsWatchlist(userID, nil)
		}, []models.Movie{{ID: 3, Title: "Action Movie"}}, nil},
		"comedy genre page 2": {"comedy", 2, func(env *TestEnv, userID uuid.UUID) {
			env.TMDB.ReturnsGenre(35, 2, []models.Movie{{ID: 4, Title: "Comedy Movie"}})
			env.Watchlist.ReturnsWatchlist(userID, nil)
		}, []models.Movie{{ID: 4, Title: "Comedy Movie"}}, nil},
		"unknown genre": {"nonexistent", 1, func(_ *TestEnv, _ uuid.UUID) {}, nil, ErrUnknownGenre},
		"tmdb error": {"trending", 1, func(env *TestEnv, _ uuid.UUID) {
			env.TMDB.TrendingFails(errors.New("network error"))
		}, nil, errors.New("network error")},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			userID := uuid.New()
			tt.setup(env, userID)

			movies, err := env.MovieService().Discover(userID, tt.genre, tt.page)

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

func TestDiscover_NewCategories(t *testing.T) {
	tests := map[string]struct {
		genre string
		setup func(*TestEnv, uuid.UUID)
		title string
	}{
		"now_playing": {"now_playing", func(env *TestEnv, userID uuid.UUID) {
			env.TMDB.ReturnsNowPlaying(1, []models.Movie{{ID: 10, Title: "Now Playing"}})
			env.Watchlist.ReturnsWatchlist(userID, nil)
		}, "Now Playing"},
		"popular": {"popular", func(env *TestEnv, userID uuid.UUID) {
			env.TMDB.ReturnsPopular(1, []models.Movie{{ID: 11, Title: "Popular"}})
			env.Watchlist.ReturnsWatchlist(userID, nil)
		}, "Popular"},
		"upcoming": {"upcoming", func(env *TestEnv, userID uuid.UUID) {
			env.TMDB.ReturnsUpcoming(1, []models.Movie{{ID: 12, Title: "Upcoming"}})
			env.Watchlist.ReturnsWatchlist(userID, nil)
		}, "Upcoming"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			userID := uuid.New()
			tt.setup(env, userID)

			movies, err := env.MovieService().Discover(userID, tt.genre, 1)
			require.NoError(t, err)
			require.Len(t, movies, 1)
			assert.Equal(t, tt.title, movies[0].Title)
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

// --- DiscoverAll ---

func TestDiscoverAll(t *testing.T) {
	trailer := []tmdb.Video{{Key: "yt123", Name: "Trailer", Site: "YouTube", Type: "Trailer"}}

	t.Run("fetches categories and filters to movies with trailers", func(t *testing.T) {
		env := newTestEnv(t)
		userID := uuid.New()

		env.TMDB.ReturnsTrending([]models.Movie{{ID: 1, Title: "Trending", MediaType: "movie"}})
		env.TMDB.ReturnsNowPlaying(1, []models.Movie{{ID: 2, Title: "Now Playing", MediaType: "movie"}})
		env.TMDB.ReturnsUpcoming(1, []models.Movie{{ID: 3, Title: "Upcoming", MediaType: "movie"}})
		env.TMDB.ReturnsVideos("movie", 1, trailer)
		env.TMDB.ReturnsVideos("movie", 2, trailer)
		env.TMDB.ReturnsVideos("movie", 3, trailer)
		env.Watchlist.ReturnsWatchlist(userID, nil)

		movies, err := env.MovieService().DiscoverAll(userID)
		require.NoError(t, err)
		assert.Len(t, movies, 3)
	})

	t.Run("excludes movies without trailers", func(t *testing.T) {
		env := newTestEnv(t)

		env.TMDB.ReturnsTrending([]models.Movie{{ID: 1, Title: "Has Trailer", MediaType: "movie"}})
		env.TMDB.ReturnsNowPlaying(1, []models.Movie{{ID: 2, Title: "No Trailer", MediaType: "movie"}})
		env.TMDB.ReturnsUpcoming(1, []models.Movie{{ID: 3, Title: "Also Has Trailer", MediaType: "movie"}})
		env.TMDB.ReturnsVideos("movie", 1, trailer)
		env.TMDB.ReturnsVideos("movie", 2, []tmdb.Video{})
		env.TMDB.ReturnsVideos("movie", 3, trailer)

		movies, err := env.MovieService().DiscoverAll(uuid.Nil)
		require.NoError(t, err)
		assert.Len(t, movies, 2)

		ids := map[int]bool{}
		for _, m := range movies {
			ids[m.ID] = true
		}
		assert.True(t, ids[1])
		assert.True(t, ids[3])
		assert.False(t, ids[2])
	})

	t.Run("enriches with watchlist", func(t *testing.T) {
		env := newTestEnv(t)
		userID := uuid.New()

		env.TMDB.ReturnsTrending([]models.Movie{{ID: 1, Title: "Trending", MediaType: "movie"}})
		env.TMDB.ReturnsNowPlaying(1, []models.Movie{{ID: 2, Title: "Now Playing", MediaType: "movie"}})
		env.TMDB.ReturnsUpcoming(1, []models.Movie{{ID: 3, Title: "Upcoming", MediaType: "movie"}})
		env.TMDB.ReturnsVideos("movie", 1, trailer)
		env.TMDB.ReturnsVideos("movie", 2, trailer)
		env.TMDB.ReturnsVideos("movie", 3, trailer)
		env.Watchlist.ReturnsWatchlist(userID, []models.Watchlist{{TMDBId: 1}})

		movies, err := env.MovieService().DiscoverAll(userID)
		require.NoError(t, err)

		watchlisted := map[int]bool{}
		for _, m := range movies {
			watchlisted[m.ID] = m.IsWatchlisted
		}
		assert.True(t, watchlisted[1])
		assert.False(t, watchlisted[2])
		assert.False(t, watchlisted[3])
	})

	t.Run("partial failure still returns other categories", func(t *testing.T) {
		env := newTestEnv(t)

		env.TMDB.ReturnsTrending([]models.Movie{{ID: 1, Title: "Trending", MediaType: "movie"}})
		env.TMDB.NowPlayingFails(1, errors.New("timeout"))
		env.TMDB.ReturnsUpcoming(1, []models.Movie{{ID: 3, Title: "Upcoming", MediaType: "movie"}})
		env.TMDB.ReturnsVideos("movie", 1, trailer)
		env.TMDB.ReturnsVideos("movie", 3, trailer)

		movies, err := env.MovieService().DiscoverAll(uuid.Nil)
		require.NoError(t, err)
		assert.Len(t, movies, 2)
	})
}

// --- EnrichWithWatchlist ---

func TestEnrichWithWatchlist_NilUser(t *testing.T) {
	env := newTestEnv(t)
	env.TMDB.ReturnsTrending([]models.Movie{{ID: 1, Title: "Movie"}})

	movies, err := env.MovieService().Discover(uuid.Nil, "trending", 1)
	require.NoError(t, err)
	assert.False(t, movies[0].IsWatchlisted)
}

func TestEnrichWithWatchlist_MarksSavedMovies(t *testing.T) {
	env := newTestEnv(t)
	userID := uuid.New()
	env.TMDB.ReturnsTrending([]models.Movie{
		{ID: 1, Title: "Saved"},
		{ID: 2, Title: "Not Saved"},
	})
	env.Watchlist.ReturnsWatchlist(userID, []models.Watchlist{{TMDBId: 1}})

	movies, err := env.MovieService().Discover(userID, "trending", 1)
	require.NoError(t, err)
	assert.True(t, movies[0].IsWatchlisted)
	assert.False(t, movies[1].IsWatchlisted)
}

// --- Search ---

func TestSearch(t *testing.T) {
	tests := map[string]struct {
		query  string
		page   int
		setup  func(*TestEnv, uuid.UUID)
		movies []models.Movie
		hasErr bool
	}{
		"success": {"fight club", 1, func(env *TestEnv, userID uuid.UUID) {
			env.TMDB.SearchReturns("fight club", 1, []models.Movie{{ID: 550, Title: "Fight Club"}})
			env.Watchlist.ReturnsWatchlist(userID, nil)
		}, []models.Movie{{ID: 550, Title: "Fight Club"}}, false},
		"tmdb error": {"query", 1, func(env *TestEnv, _ uuid.UUID) {
			env.TMDB.SearchFails("query", 1, errors.New("timeout"))
		}, nil, true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			userID := uuid.New()
			tt.setup(env, userID)

			movies, err := env.MovieService().Search(userID, tt.query, tt.page)

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

			item, err := env.MovieService().AddToWatchlist(uuid.New(), models.Movie{
				ID: 550, Title: "Fight Club", MediaType: "movie",
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
