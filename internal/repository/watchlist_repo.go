package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

// WatchlistRepository defines database operations for watchlists.
type WatchlistRepository interface {
	Add(item *models.Watchlist) error
	GetByUserID(userID uuid.UUID) ([]models.Watchlist, error)
	Remove(userID uuid.UUID, tmdbID int) (int64, error)
	Exists(userID uuid.UUID, tmdbID int) (bool, error)
}

type gormWatchlistRepository struct {
	db *gorm.DB
}

// NewWatchlistRepository creates a new WatchlistRepository backed by GORM.
func NewWatchlistRepository(db *gorm.DB) WatchlistRepository {
	return &gormWatchlistRepository{db: db}
}

func (r *gormWatchlistRepository) Add(item *models.Watchlist) error {
	return r.db.Create(item).Error
}

func (r *gormWatchlistRepository) GetByUserID(userID uuid.UUID) ([]models.Watchlist, error) {
	var items []models.Watchlist
	err := r.db.Where("user_id = ?", userID).Order("added_at DESC").Find(&items).Error

	return items, err
}

func (r *gormWatchlistRepository) Remove(userID uuid.UUID, tmdbID int) (int64, error) {
	result := r.db.Where("user_id = ? AND tmdb_id = ?", userID, tmdbID).Delete(&models.Watchlist{})

	return result.RowsAffected, result.Error
}

func (r *gormWatchlistRepository) Exists(userID uuid.UUID, tmdbID int) (bool, error) {
	var count int64
	err := r.db.Model(&models.Watchlist{}).Where("user_id = ? AND tmdb_id = ?", userID, tmdbID).Count(&count).Error

	return count > 0, err
}
