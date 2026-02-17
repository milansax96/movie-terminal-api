package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

func TestGetProfile(t *testing.T) {
	tests := map[string]struct {
		setup func(*TestEnv, uuid.UUID)
		err   error
		check func(*testing.T, *models.User)
	}{
		"success": {func(env *TestEnv, userID uuid.UUID) {
			env.Users.FindsUser(userID, &models.User{ID: userID, Username: "testuser", Email: "test@example.com"})
		}, nil, func(t *testing.T, user *models.User) {
			assert.Equal(t, "testuser", user.Username)
		}},
		"not found": {func(env *TestEnv, userID uuid.UUID) {
			env.Users.UserNotFound(userID)
		}, ErrNotFound, nil},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			userID := uuid.New()
			tt.setup(env, userID)

			user, err := env.UserService().GetProfile(userID)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)

				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, user)
			}
		})
	}
}

func TestUpdateStreamingServices(t *testing.T) {
	env := newTestEnv(t)
	userID := uuid.New()
	serviceIDs := []int{1, 2, 3}
	services := []models.StreamingService{{ID: 1}, {ID: 2}, {ID: 3}}

	env.Users.FindsStreamingServices(serviceIDs, services)
	env.Users.On("ReplaceStreamingServices", userID, services).Return(nil)

	err := env.UserService().UpdateStreamingServices(userID, serviceIDs)
	require.NoError(t, err)
}
