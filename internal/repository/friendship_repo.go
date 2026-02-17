package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

// FriendshipRepository defines database operations for friendships.
type FriendshipRepository interface {
	GetAcceptedFriendships(userID uuid.UUID) ([]models.Friendship, error)
	Create(friendship *models.Friendship) error
	AcceptRequest(requestID uuid.UUID, friendID uuid.UUID) (*models.Friendship, error)
}

type gormFriendshipRepository struct {
	db *gorm.DB
}

// NewFriendshipRepository creates a new FriendshipRepository backed by GORM.
func NewFriendshipRepository(db *gorm.DB) FriendshipRepository {
	return &gormFriendshipRepository{db: db}
}

func (r *gormFriendshipRepository) GetAcceptedFriendships(userID uuid.UUID) ([]models.Friendship, error) {
	var friendships []models.Friendship
	err := r.db.Where("(user_id = ? OR friend_id = ?) AND status = ?", userID, userID, "accepted").Find(&friendships).Error

	return friendships, err
}

func (r *gormFriendshipRepository) Create(friendship *models.Friendship) error {
	return r.db.Create(friendship).Error
}

func (r *gormFriendshipRepository) AcceptRequest(requestID uuid.UUID, friendID uuid.UUID) (*models.Friendship, error) {
	var friendship models.Friendship
	if err := r.db.First(&friendship, "id = ? AND friend_id = ?", requestID, friendID).Error; err != nil {
		return nil, err
	}

	friendship.Status = "accepted"
	if err := r.db.Save(&friendship).Error; err != nil {
		return nil, err
	}

	return &friendship, nil
}
