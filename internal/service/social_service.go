package service

import (
	"github.com/google/uuid"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/internal/repository"
)

// SocialService handles friend and social feed operations.
type SocialService struct {
	friendRepo repository.FriendshipRepository
	postRepo   repository.PostRepository
	userRepo   repository.UserRepository
}

// NewSocialService creates a new SocialService.
func NewSocialService(friendRepo repository.FriendshipRepository, postRepo repository.PostRepository, userRepo repository.UserRepository) *SocialService {
	return &SocialService{friendRepo: friendRepo, postRepo: postRepo, userRepo: userRepo}
}

// GetFriends returns accepted friendships for the user.
func (s *SocialService) GetFriends(userID uuid.UUID) ([]models.Friendship, error) {
	return s.friendRepo.GetAcceptedFriendships(userID)
}

// SearchUsers searches for users by username.
func (s *SocialService) SearchUsers(query string) ([]models.User, error) {
	return s.userRepo.SearchByUsername(query, 20)
}

// SendFriendRequest creates a pending friend request.
func (s *SocialService) SendFriendRequest(userID uuid.UUID, friendID uuid.UUID) (*models.Friendship, error) {
	friendship := &models.Friendship{
		UserID:   userID,
		FriendID: friendID,
		Status:   "pending",
	}

	if err := s.friendRepo.Create(friendship); err != nil {
		return nil, ErrAlreadyExists
	}

	return friendship, nil
}

// AcceptFriendRequest accepts a pending friend request.
func (s *SocialService) AcceptFriendRequest(requestID uuid.UUID, friendID uuid.UUID) (*models.Friendship, error) {
	friendship, err := s.friendRepo.AcceptRequest(requestID, friendID)
	if err != nil {
		return nil, ErrNotFound
	}

	return friendship, nil
}

// GetFriendsFeed returns recent posts from the user's friends.
func (s *SocialService) GetFriendsFeed(userID uuid.UUID) ([]models.Post, error) {
	friendships, err := s.friendRepo.GetAcceptedFriendships(userID)
	if err != nil {
		return nil, err
	}

	friendIDs := make([]uuid.UUID, 0, len(friendships))
	for _, f := range friendships {
		if f.UserID == userID {
			friendIDs = append(friendIDs, f.FriendID)
		} else {
			friendIDs = append(friendIDs, f.UserID)
		}
	}

	if len(friendIDs) == 0 {
		return []models.Post{}, nil
	}

	return s.postRepo.GetByUserIDs(friendIDs, 50)
}

// CreatePost creates a new post about a movie or TV show.
func (s *SocialService) CreatePost(userID uuid.UUID, tmdbID int, mediaType string, blurb string) (*models.Post, error) {
	post := &models.Post{
		UserID:    userID,
		TMDBId:    tmdbID,
		MediaType: mediaType,
		Blurb:     blurb,
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	return post, nil
}
