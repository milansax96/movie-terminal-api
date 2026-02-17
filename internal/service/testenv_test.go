package service

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
	repoMocks "github.com/milansax96/movie-terminal-api/internal/repository/mocks"
	"github.com/milansax96/movie-terminal-api/pkg/tmdb"
	tmdbMocks "github.com/milansax96/movie-terminal-api/pkg/tmdb/mocks"
)

// --- TestEnv ---

type TestEnv struct {
	TMDB      *TMDBHelper
	Users     *UserRepoHelper
	Watchlist *WatchlistRepoHelper
	Friends   *FriendRepoHelper
	Posts     *PostRepoHelper
}

func newTestEnv(t *testing.T) *TestEnv {
	return &TestEnv{
		TMDB:      &TMDBHelper{tmdbMocks.NewMockTMDBClient(t)},
		Users:     &UserRepoHelper{repoMocks.NewMockUserRepository(t)},
		Watchlist: &WatchlistRepoHelper{repoMocks.NewMockWatchlistRepository(t)},
		Friends:   &FriendRepoHelper{repoMocks.NewMockFriendshipRepository(t)},
		Posts:     &PostRepoHelper{repoMocks.NewMockPostRepository(t)},
	}
}

func (e *TestEnv) MovieService() *MovieService {
	return NewMovieService(e.TMDB.MockTMDBClient, e.Watchlist.MockWatchlistRepository)
}

func (e *TestEnv) UserService() *UserService {
	return NewUserService(e.Users.MockUserRepository)
}

func (e *TestEnv) SocialService() *SocialService {
	return NewSocialService(e.Friends.MockFriendshipRepository, e.Posts.MockPostRepository, e.Users.MockUserRepository)
}

// --- TMDBHelper ---

type TMDBHelper struct {
	*tmdbMocks.MockTMDBClient
}

func (h *TMDBHelper) ReturnsTrending(movies []tmdb.Movie) {
	h.On("GetTrending", "all", "week").Return(movies, nil)
}

func (h *TMDBHelper) ReturnsTopRated(page int, movies []tmdb.Movie) {
	h.On("GetTopRated", page).Return(movies, nil)
}

func (h *TMDBHelper) ReturnsGenre(genreID, page int, movies []tmdb.Movie) {
	h.On("DiscoverByGenre", genreID, page).Return(movies, nil)
}

func (h *TMDBHelper) SearchReturns(query string, page int, movies []tmdb.Movie) {
	h.On("SearchMovies", query, page).Return(movies, nil)
}

func (h *TMDBHelper) ReturnsDetails(mediaType string, id int, detail *tmdb.MovieDetail) {
	h.On("GetMovieDetails", mediaType, id).Return(detail, nil)
}

func (h *TMDBHelper) ReturnsVideos(mediaType string, id int, videos []tmdb.Video) {
	h.On("GetVideos", mediaType, id).Return(videos, nil)
}

func (h *TMDBHelper) ReturnsCredits(mediaType string, id int, credits *tmdb.CreditsResponse) {
	h.On("GetCredits", mediaType, id).Return(credits, nil)
}

func (h *TMDBHelper) ReturnsProviders(mediaType string, id int, providers json.RawMessage) {
	h.On("GetProviders", mediaType, id).Return(providers, nil)
}

func (h *TMDBHelper) TrendingFails(err error) {
	h.On("GetTrending", "all", "week").Return([]tmdb.Movie(nil), err)
}

func (h *TMDBHelper) SearchFails(query string, page int, err error) {
	h.On("SearchMovies", query, page).Return([]tmdb.Movie(nil), err)
}

// --- WatchlistRepoHelper ---

type WatchlistRepoHelper struct {
	*repoMocks.MockWatchlistRepository
}

func (h *WatchlistRepoHelper) AddsItem() {
	h.On("Add", mock.AnythingOfType("*models.Watchlist")).Return(nil)
}

func (h *WatchlistRepoHelper) AddFails(err error) {
	h.On("Add", mock.AnythingOfType("*models.Watchlist")).Return(err)
}

func (h *WatchlistRepoHelper) ReturnsWatchlist(userID uuid.UUID, items []models.Watchlist) {
	h.On("GetByUserID", userID).Return(items, nil)
}

func (h *WatchlistRepoHelper) RemovesItem(userID uuid.UUID, tmdbID int) {
	h.On("Remove", userID, tmdbID).Return(int64(1), nil)
}

func (h *WatchlistRepoHelper) ItemNotFound(userID uuid.UUID, tmdbID int) {
	h.On("Remove", userID, tmdbID).Return(int64(0), nil)
}

func (h *WatchlistRepoHelper) ItemExists(userID uuid.UUID, tmdbID int, exists bool) {
	h.On("Exists", userID, tmdbID).Return(exists, nil)
}

// --- UserRepoHelper ---

type UserRepoHelper struct {
	*repoMocks.MockUserRepository
}

func (h *UserRepoHelper) FindsUser(userID uuid.UUID, user *models.User) {
	h.On("FindByIDWithStreaming", userID).Return(user, nil)
}

func (h *UserRepoHelper) UserNotFound(userID uuid.UUID) {
	h.On("FindByIDWithStreaming", userID).Return((*models.User)(nil), gorm.ErrRecordNotFound)
}

func (h *UserRepoHelper) FindsStreamingServices(ids []int, services []models.StreamingService) {
	h.On("FindStreamingServicesByIDs", ids).Return(services, nil)
}

func (h *UserRepoHelper) ReplacesStreamingServices(userID uuid.UUID) {
	h.On("ReplaceStreamingServices", userID, mock.AnythingOfType("[]models.StreamingService")).Return(nil)
}

func (h *UserRepoHelper) SearchReturns(query string, users []models.User) {
	h.On("SearchByUsername", query, 20).Return(users, nil)
}

// --- FriendRepoHelper ---

type FriendRepoHelper struct {
	*repoMocks.MockFriendshipRepository
}

func (h *FriendRepoHelper) ReturnsFriendships(userID uuid.UUID, friendships []models.Friendship) {
	h.On("GetAcceptedFriendships", userID).Return(friendships, nil)
}

func (h *FriendRepoHelper) CreatesRequest() {
	h.On("Create", mock.AnythingOfType("*models.Friendship")).Return(nil)
}

func (h *FriendRepoHelper) CreateFails(err error) {
	h.On("Create", mock.AnythingOfType("*models.Friendship")).Return(err)
}

func (h *FriendRepoHelper) AcceptsRequest(requestID, friendID uuid.UUID, friendship *models.Friendship) {
	h.On("AcceptRequest", requestID, friendID).Return(friendship, nil)
}

func (h *FriendRepoHelper) AcceptFails(err error) {
	h.On("AcceptRequest", mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
		Return((*models.Friendship)(nil), err)
}

// --- PostRepoHelper ---

type PostRepoHelper struct {
	*repoMocks.MockPostRepository
}

func (h *PostRepoHelper) CreatesPost() {
	h.On("Create", mock.AnythingOfType("*models.Post")).Return(nil)
}

func (h *PostRepoHelper) ReturnsPosts(posts []models.Post) {
	h.On("GetByUserIDs", mock.Anything, 50).Return(posts, nil)
}
