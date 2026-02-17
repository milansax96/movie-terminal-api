package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

// PostRepository defines database operations for posts.
type PostRepository interface {
	Create(post *models.Post) error
	GetByUserIDs(userIDs []uuid.UUID, limit int) ([]models.Post, error)
}

type gormPostRepository struct {
	db *gorm.DB
}

// NewPostRepository creates a new PostRepository backed by GORM.
func NewPostRepository(db *gorm.DB) PostRepository {
	return &gormPostRepository{db: db}
}

func (r *gormPostRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *gormPostRepository) GetByUserIDs(userIDs []uuid.UUID, limit int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Preload("User").Where("user_id IN ?", userIDs).Order("created_at DESC").Limit(limit).Find(&posts).Error

	return posts, err
}
