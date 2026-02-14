package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

func InitDB() *gorm.DB {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	if host == "" || user == "" || password == "" || dbname == "" || port == "" {
		log.Fatal("Database environment variables are not fully set")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		host, user, password, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")
	return db
}

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
