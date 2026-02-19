// Package main is the entry point for the Movie Terminal API server.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/milansax96/movie-terminal-api/config"
	"github.com/milansax96/movie-terminal-api/internal/database"
	"github.com/milansax96/movie-terminal-api/internal/handlers"
	"github.com/milansax96/movie-terminal-api/internal/middleware"
	"github.com/milansax96/movie-terminal-api/internal/repository"
	"github.com/milansax96/movie-terminal-api/internal/service"
	"github.com/milansax96/movie-terminal-api/pkg/tmdb"
)

func main() {
	testToken := flag.Bool("test-token", false, "Print a valid JWT for testing and exit")
	userID := flag.String("user-id", "test-user", "User ID to embed in the test token")
	flag.Parse()

	cfg := config.Load()

	if *testToken {
		token, err := generateTestToken(*userID, cfg.JWTSecret)
		if err != nil {
			log.Fatalf("Failed to generate token: %v", err)
		}

		fmt.Print(token)
		os.Exit(0)
	}

	db := database.InitDB(cfg)
	database.Migrate(db)

	tmdbClient := tmdb.NewCachedClient(tmdb.NewClient())

	// Repositories
	userRepo := repository.NewUserRepository(db)
	watchlistRepo := repository.NewWatchlistRepository(db)
	friendshipRepo := repository.NewFriendshipRepository(db)
	postRepo := repository.NewPostRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg)
	userSvc := service.NewUserService(userRepo)
	movieSvc := service.NewMovieService(tmdbClient, watchlistRepo)
	socialSvc := service.NewSocialService(friendshipRepo, postRepo, userRepo)

	// Router
	r := gin.Default()
	r.Use(middleware.CORS())

	handlers.RegisterAuthRoutes(r, authSvc)
	handlers.RegisterProtectedRoutes(r, cfg.JWTSecret, userSvc, movieSvc, socialSvc)

	log.Printf("Server starting on port %s", cfg.Port)
	err := r.Run(":" + cfg.Port)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func generateTestToken(userID string, jwtSecret string) (string, error) {
	claims := &middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtSecret))
}
