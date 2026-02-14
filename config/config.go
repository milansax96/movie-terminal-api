package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	TMDBAPIKey  string
	JWTSecret   string
	Port        string
	Environment string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		TMDBAPIKey:  os.Getenv("TMDB_API_KEY"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        port,
		Environment: os.Getenv("ENVIRONMENT"),
	}
}
