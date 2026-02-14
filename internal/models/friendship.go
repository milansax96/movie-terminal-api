package models

import (
	"time"

	"github.com/google/uuid"
)

type Friendship struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	FriendID  uuid.UUID `gorm:"type:uuid;not null" json:"friend_id"`
	Status    string    `gorm:"default:'pending'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Watchlist struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	TMDBId    int       `gorm:"not null" json:"tmdb_id"`
	MediaType string    `gorm:"not null" json:"media_type"`
	AddedAt   time.Time `json:"added_at"`
}
