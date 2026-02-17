package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/internal/service"
	"github.com/milansax96/movie-terminal-api/pkg/tmdb"
)

// --- Discover ---

func TestGetDiscoverFeed(t *testing.T) {
	tests := map[string]struct {
		path   string
		setup  func(*TestServer)
		status int
		check  func(*testing.T, *httptest.ResponseRecorder)
	}{
		"success": {
			"/discover",
			func(ts *TestServer) {
				ts.Movies.Discovers(
					"trending",
					1,
					[]models.Movie{
						{
							ID:    1,
							Title: "Test Movie",
						},
					},
				)
			},
			http.StatusOK,
			func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Len(t, resp["results"].([]interface{}), 1)
			},
		},
		"with genre": {
			"/discover?genre=action",
			func(ts *TestServer) {
				ts.Movies.Discovers("action", 1, []models.Movie{{ID: 2, Title: "Action Movie"}})
			},
			http.StatusOK,
			func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Len(t, resp["results"].([]interface{}), 1)
			},
		},
		"unknown genre": {"/discover?genre=nonexistent", func(ts *TestServer) {
			ts.Movies.DiscoverFails("nonexistent", service.ErrUnknownGenre)
		}, http.StatusBadRequest, nil},
		"internal error": {"/discover", func(ts *TestServer) {
			ts.Movies.DiscoverFails("trending", errors.New("tmdb down"))
		}, http.StatusInternalServerError, nil},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)
			w := ts.Do(httptest.NewRequest("GET", tt.path, nil))
			assert.Equal(t, tt.status, w.Code)
			if tt.check != nil {
				tt.check(t, w)
			}
		})
	}
}

// --- Search ---

func TestSearchMovies(t *testing.T) {
	tests := map[string]struct {
		path   string
		setup  func(*TestServer)
		status int
	}{
		"success": {
			"/search?q=fight+club",
			func(ts *TestServer) {
				ts.Movies.Searches(
					"fight club",
					1,
					[]models.Movie{
						{
							ID:    550,
							Title: "Fight Club",
						},
					},
				)
			}, http.StatusOK,
		},
		"missing query": {
			"/search",
			func(_ *TestServer) {},
			http.StatusBadRequest,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)
			w := ts.Do(httptest.NewRequest("GET", tt.path, nil))
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

// --- Movie Detail ---

func TestGetMovieDetail(t *testing.T) {
	tests := map[string]struct {
		path   string
		setup  func(*TestServer)
		status int
	}{
		"success": {
			"/movies/550",
			func(ts *TestServer) {
				movieDetail := &tmdb.MovieDetail{
					ID:    550,
					Title: "Fight Club",
				}

				ts.Movies.ReturnsDetail(
					"movie",
					movieDetail.ID,
					movieDetail,
				)
			},
			http.StatusOK,
		},
		"invalid id": {
			"/movies/abc",
			func(_ *TestServer) {},
			http.StatusBadRequest,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)
			w := ts.Do(httptest.NewRequest("GET", tt.path, nil))
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

// --- Videos ---

func TestGetMovieVideos(t *testing.T) {
	ts := newTestServer(t)
	ts.Movies.ReturnsVideos("movie", 550, []tmdb.Video{{Key: "abc", Name: "Trailer"}})

	w := ts.Do(httptest.NewRequest("GET", "/movies/550/videos", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Credits ---

func TestGetMovieCredits(t *testing.T) {
	ts := newTestServer(t)
	ts.Movies.ReturnsCredits("movie", 550, &tmdb.CreditsResponse{Cast: []tmdb.CastMember{{Name: "Brad Pitt"}}})

	w := ts.Do(httptest.NewRequest("GET", "/movies/550/credits", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Providers ---

func TestGetMovieProviders(t *testing.T) {
	ts := newTestServer(t)
	ts.Movies.ReturnsProviders("movie", 550, json.RawMessage(`{"US":{"flatrate":[]}}`))

	w := ts.Do(httptest.NewRequest("GET", "/movies/550/providers", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- Watchlist ---

func TestGetWatchlist(t *testing.T) {
	ts := newTestServer(t)
	ts.Movies.ReturnsWatchlist([]models.Watchlist{{TMDBId: 550, Title: "Fight Club"}})

	w := ts.Do(httptest.NewRequest("GET", "/watchlist", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddToWatchlist(t *testing.T) {
	tests := map[string]struct {
		body   string
		setup  func(*TestServer)
		status int
	}{
		"success": {
			`{"movie_id": 550, "title": "Fight Club", "media_type": "movie"}`,
			func(ts *TestServer) {
				ts.Movies.AddsToWatchlist(
					&models.Watchlist{
						TMDBId:    550,
						Title:     "Fight Club",
						MediaType: "movie",
					},
				)
			},
			http.StatusCreated,
		},
		"missing body": {
			`{}`,
			func(_ *TestServer) {},
			http.StatusBadRequest,
		},
		"duplicate": {
			`{"movie_id": 550, "title": "Fight Club", "media_type": "movie"}`,
			func(ts *TestServer) {
				ts.Movies.AddToWatchlistFails(service.ErrAlreadyExists)
			},
			http.StatusConflict,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)

			req := httptest.NewRequest("POST", "/watchlist", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := ts.Do(req)

			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestRemoveFromWatchlist(t *testing.T) {
	tests := map[string]struct {
		path   string
		setup  func(*TestServer)
		status int
	}{
		"success": {
			"/watchlist/550",
			func(ts *TestServer) {
				ts.Movies.RemovesFromWatchlist(550)
			},
			http.StatusOK,
		},
		"not found": {"/watchlist/999", func(ts *TestServer) {
			ts.Movies.RemoveFails(999, service.ErrNotFound)
		}, http.StatusNotFound},
		"invalid id": {"/watchlist/abc", func(_ *TestServer) {}, http.StatusBadRequest},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)
			w := ts.Do(httptest.NewRequest("DELETE", tt.path, nil))
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestCheckWatchlist(t *testing.T) {
	tests := map[string]struct {
		path   string
		setup  func(*TestServer)
		status int
		check  func(*testing.T, *httptest.ResponseRecorder)
	}{
		"saved": {"/watchlist/550/check", func(ts *TestServer) {
			ts.Movies.ChecksWatchlist(550, true)
		}, http.StatusOK, func(t *testing.T, w *httptest.ResponseRecorder) {
			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, true, resp["saved"])
		}},
		"invalid id": {"/watchlist/abc/check", func(_ *TestServer) {}, http.StatusBadRequest, nil},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)
			w := ts.Do(httptest.NewRequest("GET", tt.path, nil))
			assert.Equal(t, tt.status, w.Code)
			if tt.check != nil {
				tt.check(t, w)
			}
		})
	}
}
