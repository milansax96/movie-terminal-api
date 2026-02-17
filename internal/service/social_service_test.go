package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

func TestGetFriendsFeed(t *testing.T) {
	tests := map[string]struct {
		setup func(*TestEnv, uuid.UUID)
	}{
		"with friends": {func(env *TestEnv, userID uuid.UUID) {
			friendA := uuid.New()
			friendB := uuid.New()
			env.Friends.ReturnsFriendships(userID, []models.Friendship{
				{UserID: userID, FriendID: friendA, Status: "accepted"},
				{UserID: friendB, FriendID: userID, Status: "accepted"},
			})
			env.Posts.ReturnsPosts([]models.Post{})
		}},
		"no friends": {func(env *TestEnv, userID uuid.UUID) {
			env.Friends.ReturnsFriendships(userID, []models.Friendship{})
		}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			userID := uuid.New()
			tt.setup(env, userID)

			posts, err := env.SocialService().GetFriendsFeed(userID)
			require.NoError(t, err)
			assert.Empty(t, posts)
		})
	}
}

func TestSendFriendRequest(t *testing.T) {
	tests := map[string]struct {
		setup func(*TestEnv)
		err   error
		check func(*testing.T, *models.Friendship)
	}{
		"success": {func(env *TestEnv) {
			env.Friends.CreatesRequest()
		}, nil, func(t *testing.T, f *models.Friendship) {
			assert.Equal(t, "pending", f.Status)
		}},
		"duplicate": {func(env *TestEnv) {
			env.Friends.CreateFails(errors.New("duplicate"))
		}, ErrAlreadyExists, nil},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			tt.setup(env)

			friendship, err := env.SocialService().SendFriendRequest(uuid.New(), uuid.New())

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)

				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, friendship)
			}
		})
	}
}

func TestAcceptFriendRequest(t *testing.T) {
	tests := map[string]struct {
		setup func(*TestEnv, uuid.UUID, uuid.UUID)
		err   error
		check func(*testing.T, *models.Friendship)
	}{
		"success": {func(env *TestEnv, requestID, friendID uuid.UUID) {
			expected := &models.Friendship{UserID: uuid.New(), FriendID: friendID, Status: "accepted"}
			env.Friends.AcceptsRequest(requestID, friendID, expected)
		}, nil, func(t *testing.T, f *models.Friendship) {
			assert.Equal(t, "accepted", f.Status)
		}},
		"not found": {func(env *TestEnv, _, _ uuid.UUID) {
			env.Friends.AcceptFails(errors.New("not found"))
		}, ErrNotFound, nil},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			requestID := uuid.New()
			friendID := uuid.New()
			tt.setup(env, requestID, friendID)

			friendship, err := env.SocialService().AcceptFriendRequest(requestID, friendID)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)

				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, friendship)
			}
		})
	}
}

func TestCreatePost(t *testing.T) {
	env := newTestEnv(t)
	env.Posts.CreatesPost()

	post, err := env.SocialService().CreatePost(uuid.New(), 550, "movie", "Great film!")
	require.NoError(t, err)
	assert.Equal(t, "Great film!", post.Blurb)
}

func TestSearchUsers(t *testing.T) {
	env := newTestEnv(t)
	env.Users.SearchReturns("test", []models.User{{Username: "testuser"}})

	users, err := env.SocialService().SearchUsers("test")
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "testuser", users[0].Username)
}
