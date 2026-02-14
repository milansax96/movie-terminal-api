package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/milansax96/movie-terminal-api/internal/database"
	"github.com/milansax96/movie-terminal-api/internal/handlers"
	"github.com/milansax96/movie-terminal-api/internal/middleware"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found")
	}

	db := database.InitDB()

	database.Migrate(db)

	r := gin.Default()
	r.Use(middleware.CORS())

	// Public routes
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/signup", handlers.Signup(db))
		auth.POST("/login", handlers.Login(db))
	}

	// Protected routes
	api := r.Group("/api/v1")
	api.Use(middleware.AuthRequired())
	{
		// User profile
		api.GET("/user/profile", handlers.GetProfile(db))
		api.PUT("/user/streaming-services", handlers.UpdateStreamingServices(db))

		// Discovery
		api.GET("/discover", handlers.GetDiscoverFeed(db))
		api.POST("/watchlist", handlers.AddToWatchlist(db))

		// Friends
		api.GET("/friends", handlers.GetFriends(db))
		api.POST("/friends/request", handlers.SendFriendRequest(db))
		api.PUT("/friends/accept/:id", handlers.AcceptFriendRequest(db))
		api.GET("/friends/search", handlers.SearchUsers(db))

		// Feed
		api.GET("/feed", handlers.GetFriendsFeed(db))
		api.POST("/posts", handlers.CreatePost(db))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}
