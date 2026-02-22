package models

import (
	"time"

	"github.com/google/uuid"
)

// Friendship represents a friend connection between two users.
type Friendship struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	FriendID  uuid.UUID `gorm:"type:uuid;not null" json:"friend_id"`
	Status    string    `gorm:"default:'pending'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Watchlist represents a movie saved to a user's watchlist.
type Watchlist struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_movie" json:"user_id"`
	TMDBId       int       `gorm:"not null;uniqueIndex:idx_user_movie" json:"tmdb_id"`
	Title        string    `json:"title"`
	PosterPath   string    `json:"poster_path"`
	BackdropPath string    `json:"backdrop_path"`
	MediaType    string    `gorm:"not null" json:"media_type"`
	TrailerKey   string    `json:"trailer_key"`
	AddedAt      time.Time `json:"added_at"`
}
