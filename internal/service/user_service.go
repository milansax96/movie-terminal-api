package service

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
	"github.com/milansax96/movie-terminal-api/internal/repository"
)

// UserService handles user profile operations.
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetProfile returns the user's profile with streaming services.
func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByIDWithStreaming(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}

	return user, err
}

// UpdateStreamingServices replaces the user's streaming service preferences.
func (s *UserService) UpdateStreamingServices(userID uuid.UUID, serviceIDs []int) error {
	services, err := s.userRepo.FindStreamingServicesByIDs(serviceIDs)
	if err != nil {
		return err
	}

	return s.userRepo.ReplaceStreamingServices(userID, services)
}
