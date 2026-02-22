// Package config handles application configuration from environment variables.
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values.
type Config struct {
	DBHost              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBPort              string
	TMDBAPIKey          string
	JWTSecret           string
	GoogleClientID      string
	Port                string
	Environment         string
	CloudinaryCloudName string
}

// Load reads configuration from environment variables, optionally loading from .env files.
func Load(paths ...string) *Config {
	if err := godotenv.Load(paths...); err != nil {
		log.Println("No .env file found")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DBHost:              os.Getenv("DB_HOST"),
		DBUser:              os.Getenv("DB_USER"),
		DBPassword:          os.Getenv("DB_PASSWORD"),
		DBName:              os.Getenv("DB_NAME"),
		DBPort:              os.Getenv("DB_PORT"),
		TMDBAPIKey:          os.Getenv("TMDB_API_KEY"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		GoogleClientID:      os.Getenv("GOOGLE_CLIENT_ID"),
		Port:                port,
		Environment:         os.Getenv("ENVIRONMENT"),
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
	}
}
