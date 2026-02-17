package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/internal/service"
)

func TestGetFriends(t *testing.T) {
	tests := map[string]struct {
		setup  func(*TestServer)
		status int
	}{
		"success": {func(ts *TestServer) {
			ts.Social.ReturnsFriends([]models.Friendship{{UserID: uuid.New(), FriendID: uuid.New(), Status: "accepted"}})
		}, http.StatusOK},
		"db error": {func(ts *TestServer) {
			ts.Social.GetFriendsFails(errors.New("db error"))
		}, http.StatusInternalServerError},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)
			w := ts.Do(httptest.NewRequest("GET", "/friends", nil))
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestSearchUsers(t *testing.T) {
	tests := map[string]struct {
		path   string
		setup  func(*TestServer)
		status int
	}{
		"success": {"/friends/search?q=test", func(ts *TestServer) {
			ts.Social.SearchReturns("test", []models.User{{Username: "testuser"}})
		}, http.StatusOK},
		"missing query": {"/friends/search", func(_ *TestServer) {}, http.StatusBadRequest},
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

func TestSendFriendRequest(t *testing.T) {
	friendID := uuid.New().String()
	tests := map[string]struct {
		body   string
		setup  func(*TestServer)
		status int
	}{
		"success": {`{"friend_id": "` + friendID + `"}`, func(ts *TestServer) {
			ts.Social.SendsRequest(&models.Friendship{Status: "pending"})
		}, http.StatusCreated},
		"duplicate": {`{"friend_id": "` + friendID + `"}`, func(ts *TestServer) {
			ts.Social.SendRequestFails(service.ErrAlreadyExists)
		}, http.StatusConflict},
		"missing body": {`{}`, func(_ *TestServer) {}, http.StatusBadRequest},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)

			req := httptest.NewRequest("POST", "/friends/request", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := ts.Do(req)

			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	tests := map[string]struct {
		setup  func(*TestServer)
		status int
	}{
		"success": {func(ts *TestServer) {
			ts.Social.AcceptsRequest(&models.Friendship{Status: "accepted"})
		}, http.StatusOK},
		"not found": {func(ts *TestServer) {
			ts.Social.AcceptFails(service.ErrNotFound)
		}, http.StatusNotFound},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)
			w := ts.Do(httptest.NewRequest("PUT", "/friends/accept/"+uuid.New().String(), nil))
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestGetFriendsFeed(t *testing.T) {
	ts := newTestServer(t)
	ts.Social.ReturnsFeed([]models.Post{{TMDBId: 550, Blurb: "Great film!"}})

	w := ts.Do(httptest.NewRequest("GET", "/feed", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreatePost(t *testing.T) {
	tests := map[string]struct {
		body   string
		setup  func(*TestServer)
		status int
	}{
		"success": {`{"tmdb_id": 550, "media_type": "movie", "blurb": "Great film!"}`, func(ts *TestServer) {
			ts.Social.CreatesPost(&models.Post{TMDBId: 550, MediaType: "movie", Blurb: "Great film!"})
		}, http.StatusCreated},
		"missing body": {`{}`, func(_ *TestServer) {}, http.StatusBadRequest},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)

			req := httptest.NewRequest("POST", "/posts", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := ts.Do(req)

			assert.Equal(t, tt.status, w.Code)
		})
	}
}
