package models

import (
	"time"

	"github.com/google/uuid"
)

// Post represents a user's post about a movie or TV show.
type Post struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	TMDBId    int       `gorm:"not null" json:"tmdb_id"`
	MediaType string    `gorm:"not null" json:"media_type"`
	Blurb     string    `json:"blurb"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
