package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/internal/service"
)

func TestGoogleLogin(t *testing.T) {
	tests := map[string]struct {
		body   string
		setup  func(*TestServer)
		status int
	}{
		"new user": {`{"access_token": "valid-google-token"}`, func(ts *TestServer) {
			ts.Auth.LogsIn("valid-google-token", &service.AuthResult{
				User:  &models.User{ID: uuid.New(), Username: "newuser", Email: "new@example.com"},
				Token: "jwt-token-123",
				IsNew: true,
			})
		}, http.StatusCreated},
		"existing user": {`{"access_token": "valid-google-token"}`, func(ts *TestServer) {
			ts.Auth.LogsIn("valid-google-token", &service.AuthResult{
				User:  &models.User{ID: uuid.New(), Username: "existing", Email: "existing@example.com"},
				Token: "jwt-token-456",
				IsNew: false,
			})
		}, http.StatusOK},
		"invalid token": {`{"access_token": "bad-token"}`, func(ts *TestServer) {
			ts.Auth.LoginFails("bad-token", service.ErrInvalidToken)
		}, http.StatusUnauthorized},
		"missing claims": {`{"access_token": "missing-claims-token"}`, func(ts *TestServer) {
			ts.Auth.LoginFails("missing-claims-token", service.ErrMissingClaims)
		}, http.StatusBadRequest},
		"missing body": {`{}`, func(_ *TestServer) {}, http.StatusBadRequest},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ts := newTestServer(t)
			tt.setup(ts)

			req := httptest.NewRequest("POST", "/auth/google", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := ts.Do(req)

			assert.Equal(t, tt.status, w.Code)
		})
	}
}
