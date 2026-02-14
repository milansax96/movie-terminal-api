package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	StreamingServices []StreamingService `gorm:"many2many:user_streaming_services" json:"streaming_services,omitempty"`
}

type StreamingService struct {
	ID   int    `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`
	Slug string `gorm:"uniqueIndex" json:"slug"`
}

type UserStreamingService struct {
	UserID    uuid.UUID `gorm:"type:uuid"`
	ServiceID int
}
