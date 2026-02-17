package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/internal/service"
	svcMocks "github.com/milansax96/movie-terminal-api/internal/service/mocks"
	"github.com/milansax96/movie-terminal-api/pkg/tmdb"
)

const testUserID = "550e8400-e29b-41d4-a716-446655440000"

// --- TestServer ---

type TestServer struct {
	Router *gin.Engine
	Auth   *AuthSvcHelper
	Users  *UserSvcHelper
	Movies *MovieSvcHelper
	Social *SocialSvcHelper
}

func newTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)

	ts := &TestServer{
		Auth:   &AuthSvcHelper{svcMocks.NewMockAuthServiceInterface(t)},
		Users:  &UserSvcHelper{svcMocks.NewMockUserServiceInterface(t)},
		Movies: &MovieSvcHelper{svcMocks.NewMockMovieServiceInterface(t)},
		Social: &SocialSvcHelper{svcMocks.NewMockSocialServiceInterface(t)},
	}

	authH := NewAuthHandler(ts.Auth.MockAuthServiceInterface)
	userH := NewUserHandler(ts.Users.MockUserServiceInterface)
	movieH := NewMovieHandler(ts.Movies.MockMovieServiceInterface)
	socialH := NewSocialHandler(ts.Social.MockSocialServiceInterface)

	r := gin.New()

	// Auth routes (no user_id middleware)
	r.POST("/auth/google", authH.GoogleLogin)

	// Protected routes (inject test user_id)
	protected := r.Group("/")
	protected.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})

	// User
	protected.GET("/user/profile", userH.GetProfile)
	protected.PUT("/user/streaming-services", userH.UpdateStreamingServices)

	// Movies
	protected.GET("/discover", movieH.GetDiscoverFeed)
	protected.GET("/search", movieH.SearchMovies)
	protected.GET("/movies/:id", movieH.GetMovieDetail)
	protected.GET("/movies/:id/videos", movieH.GetMovieVideos)
	protected.GET("/movies/:id/credits", movieH.GetMovieCredits)
	protected.GET("/movies/:id/providers", movieH.GetMovieProviders)

	// Watchlist
	protected.GET("/watchlist", movieH.GetWatchlist)
	protected.POST("/watchlist", movieH.AddToWatchlist)
	protected.DELETE("/watchlist/:movie_id", movieH.RemoveFromWatchlist)
	protected.GET("/watchlist/:movie_id/check", movieH.CheckWatchlist)

	// Social
	protected.GET("/friends", socialH.GetFriends)
	protected.POST("/friends/request", socialH.SendFriendRequest)
	protected.PUT("/friends/accept/:id", socialH.AcceptFriendRequest)
	protected.GET("/friends/search", socialH.SearchUsers)
	protected.GET("/feed", socialH.GetFriendsFeed)
	protected.POST("/posts", socialH.CreatePost)

	ts.Router = r

	return ts
}

func (ts *TestServer) Do(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	ts.Router.ServeHTTP(w, req)

	return w
}

// --- AuthSvcHelper ---

type AuthSvcHelper struct {
	*svcMocks.MockAuthServiceInterface
}

func (h *AuthSvcHelper) LogsIn(token string, result *service.AuthResult) {
	h.On("GoogleLogin", mock.Anything, token).Return(result, nil)
}

func (h *AuthSvcHelper) LoginFails(token string, err error) {
	h.On("GoogleLogin", mock.Anything, token).Return((*service.AuthResult)(nil), err)
}

// --- UserSvcHelper ---

type UserSvcHelper struct {
	*svcMocks.MockUserServiceInterface
}

func (h *UserSvcHelper) GetsProfile(user *models.User) {
	h.On("GetProfile", mock.AnythingOfType("uuid.UUID")).Return(user, nil)
}

func (h *UserSvcHelper) ProfileNotFound() {
	h.On("GetProfile", mock.AnythingOfType("uuid.UUID")).Return((*models.User)(nil), service.ErrNotFound)
}

func (h *UserSvcHelper) UpdatesStreamingServices(serviceIDs []int) {
	h.On("UpdateStreamingServices", mock.AnythingOfType("uuid.UUID"), serviceIDs).Return(nil)
}

func (h *UserSvcHelper) UpdateFails(err error) {
	h.On("UpdateStreamingServices", mock.AnythingOfType("uuid.UUID"), mock.Anything).Return(err)
}

// --- MovieSvcHelper ---

type MovieSvcHelper struct {
	*svcMocks.MockMovieServiceInterface
}

func (h *MovieSvcHelper) Discovers(genre string, page int, movies []tmdb.Movie) {
	h.On("Discover", genre, page).Return(movies, nil)
}

func (h *MovieSvcHelper) DiscoverFails(genre string, err error) {
	h.On("Discover", genre, 1).Return([]tmdb.Movie(nil), err)
}

func (h *MovieSvcHelper) Searches(query string, page int, movies []tmdb.Movie) {
	h.On("Search", query, page).Return(movies, nil)
}

func (h *MovieSvcHelper) ReturnsDetail(mediaType string, id int, detail *tmdb.MovieDetail) {
	h.On("GetDetail", mediaType, id).Return(detail, nil)
}

func (h *MovieSvcHelper) DetailFails(mediaType string, id int, err error) {
	h.On("GetDetail", mediaType, id).Return((*tmdb.MovieDetail)(nil), err)
}

func (h *MovieSvcHelper) ReturnsVideos(mediaType string, id int, videos []tmdb.Video) {
	h.On("GetVideos", mediaType, id).Return(videos, nil)
}

func (h *MovieSvcHelper) ReturnsCredits(mediaType string, id int, credits *tmdb.CreditsResponse) {
	h.On("GetCredits", mediaType, id).Return(credits, nil)
}

func (h *MovieSvcHelper) ReturnsProviders(mediaType string, id int, providers json.RawMessage) {
	h.On("GetProviders", mediaType, id).Return(providers, nil)
}

func (h *MovieSvcHelper) ReturnsWatchlist(items []models.Watchlist) {
	h.On("GetWatchlist", mock.AnythingOfType("uuid.UUID")).Return(items, nil)
}

func (h *MovieSvcHelper) AddsToWatchlist(item *models.Watchlist) {
	h.On("AddToWatchlist", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("service.AddWatchlistRequest")).
		Return(item, nil)
}

func (h *MovieSvcHelper) AddToWatchlistFails(err error) {
	h.On("AddToWatchlist", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("service.AddWatchlistRequest")).
		Return((*models.Watchlist)(nil), err)
}

func (h *MovieSvcHelper) RemovesFromWatchlist(movieID int) {
	h.On("RemoveFromWatchlist", mock.AnythingOfType("uuid.UUID"), movieID).Return(nil)
}

func (h *MovieSvcHelper) RemoveFails(movieID int, err error) {
	h.On("RemoveFromWatchlist", mock.AnythingOfType("uuid.UUID"), movieID).Return(err)
}

func (h *MovieSvcHelper) ChecksWatchlist(movieID int, saved bool) {
	h.On("CheckWatchlist", mock.AnythingOfType("uuid.UUID"), movieID).Return(saved, nil)
}

// --- SocialSvcHelper ---

type SocialSvcHelper struct {
	*svcMocks.MockSocialServiceInterface
}

func (h *SocialSvcHelper) ReturnsFriends(friendships []models.Friendship) {
	h.On("GetFriends", mock.AnythingOfType("uuid.UUID")).Return(friendships, nil)
}

func (h *SocialSvcHelper) GetFriendsFails(err error) {
	h.On("GetFriends", mock.AnythingOfType("uuid.UUID")).Return([]models.Friendship(nil), err)
}

func (h *SocialSvcHelper) SearchReturns(query string, users []models.User) {
	h.On("SearchUsers", query).Return(users, nil)
}

func (h *SocialSvcHelper) SendsRequest(friendship *models.Friendship) {
	h.On("SendFriendRequest", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
		Return(friendship, nil)
}

func (h *SocialSvcHelper) SendRequestFails(err error) {
	h.On("SendFriendRequest", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
		Return((*models.Friendship)(nil), err)
}

func (h *SocialSvcHelper) AcceptsRequest(friendship *models.Friendship) {
	h.On("AcceptFriendRequest", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
		Return(friendship, nil)
}

func (h *SocialSvcHelper) AcceptFails(err error) {
	h.On("AcceptFriendRequest", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
		Return((*models.Friendship)(nil), err)
}

func (h *SocialSvcHelper) ReturnsFeed(posts []models.Post) {
	h.On("GetFriendsFeed", mock.AnythingOfType("uuid.UUID")).Return(posts, nil)
}

func (h *SocialSvcHelper) CreatesPost(post *models.Post) {
	h.On("CreatePost", mock.AnythingOfType("uuid.UUID"), post.TMDBId, post.MediaType, post.Blurb).
		Return(post, nil)
}
