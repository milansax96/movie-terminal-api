package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

func TestGetProfile(t *testing.T) {
	tests := map[string]struct {
		setup  func(*TestServer)
		status int
	}{
		"success": {func(ts *TestServer) {
			ts.Users.GetsProfile(&models.User{Username: "testuser", Email: "test@example.com"})
		}, http.StatusOK},
		"not found": {func(ts *TestServer) {
			ts.Users.ProfileNotFound()
		}, http.StatusNotFound},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)
			w := ts.Do(httptest.NewRequest("GET", "/user/profile", nil))
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func TestUpdateStreamingServices(t *testing.T) {
	tests := map[string]struct {
		body   string
		setup  func(*TestServer)
		status int
	}{
		"success": {`{"service_ids": [1, 2]}`, func(ts *TestServer) {
			ts.Users.UpdatesStreamingServices([]int{1, 2})
		}, http.StatusOK},
		"missing body": {`{}`, func(_ *TestServer) {}, http.StatusBadRequest},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)

			req := httptest.NewRequest("PUT", "/user/streaming-services", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := ts.Do(req)

			assert.Equal(t, tt.status, w.Code)
		})
	}
}
