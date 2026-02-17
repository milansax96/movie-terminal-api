// Package repository provides database access for all domain models.
package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

// UserRepository defines database operations for users.
type UserRepository interface {
	FindByGoogleID(googleID string) (*models.User, error)
	Create(user *models.User) error
	UpdateProfilePicture(userID uuid.UUID, picture string) error
	FindByIDWithStreaming(userID uuid.UUID) (*models.User, error)
	FindByID(userID uuid.UUID) (*models.User, error)
	ReplaceStreamingServices(userID uuid.UUID, services []models.StreamingService) error
	SearchByUsername(query string, limit int) ([]models.User, error)
	FindStreamingServicesByIDs(ids []int) ([]models.StreamingService, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository backed by GORM.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) FindByGoogleID(googleID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *gormUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *gormUserRepository) UpdateProfilePicture(userID uuid.UUID, picture string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("profile_picture", picture).Error
}

func (r *gormUserRepository) FindByIDWithStreaming(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("StreamingServices").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *gormUserRepository) FindByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *gormUserRepository) ReplaceStreamingServices(userID uuid.UUID, services []models.StreamingService) error {
	var user models.User
	if err := r.db.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	return r.db.Model(&user).Association("StreamingServices").Replace(services)
}

func (r *gormUserRepository) SearchByUsername(query string, limit int) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("username ILIKE ?", "%"+query+"%").Limit(limit).Find(&users).Error

	return users, err
}

func (r *gormUserRepository) FindStreamingServicesByIDs(ids []int) ([]models.StreamingService, error) {
	var services []models.StreamingService
	err := r.db.Where("id IN ?", ids).Find(&services).Error

	return services, err
}
