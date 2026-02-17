// Package database handles database initialization and migrations.
package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/config"
	"github.com/milansax96/movie-terminal-api/internal/models"
)

// InitDB opens a PostgreSQL connection using the provided config.
func InitDB(cfg *config.Config) *gorm.DB {
	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" || cfg.DBPort == "" {
		log.Fatal("Database environment variables are not fully set")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	return db
}

// Migrate runs auto-migrations and seeds initial data.
func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.StreamingService{},
		&models.Friendship{},
		&models.Post{},
		&models.Watchlist{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	SeedStreamingServices(db)

	log.Println("Database migration completed")
}

// SeedStreamingServices inserts default streaming services if they don't exist.
func SeedStreamingServices(db *gorm.DB) {
	services := []models.StreamingService{
		{Name: "Netflix", Slug: "netflix"},
		{Name: "Hulu", Slug: "hulu"},
		{Name: "Disney+", Slug: "disney_plus"},
		{Name: "HBO Max", Slug: "hbo_max"},
		{Name: "Amazon Prime Video", Slug: "prime_video"},
		{Name: "Apple TV+", Slug: "apple_tv_plus"},
		{Name: "Paramount+", Slug: "paramount_plus"},
		{Name: "Peacock", Slug: "peacock"},
	}

	for _, s := range services {
		db.Where("slug = ?", s.Slug).FirstOrCreate(&s)
	}
}
