package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/milansax96/movie-terminal-api/internal/middleware"
	"github.com/milansax96/movie-terminal-api/internal/service"
)

// RegisterAuthRoutes registers public authentication routes.
func RegisterAuthRoutes(r *gin.Engine, authSvc service.AuthServiceInterface) {
	authH := NewAuthHandler(authSvc)

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/google", authH.GoogleLogin)
	}
}

// RegisterProtectedRoutes registers JWT-protected API routes.
func RegisterProtectedRoutes(r *gin.Engine, jwtSecret string, userSvc service.UserServiceInterface, movieSvc service.MovieServiceInterface, socialSvc service.SocialServiceInterface) {
	userH := NewUserHandler(userSvc)
	movieH := NewMovieHandler(movieSvc)
	socialH := NewSocialHandler(socialSvc)

	api := r.Group("/api/v1")
	api.Use(middleware.AuthRequired(jwtSecret))
	{
		// User profile
		api.GET("/user/profile", userH.GetProfile)
		api.PUT("/user/streaming-services", userH.UpdateStreamingServices)

		// Discovery & Search
		api.GET("/discover", movieH.GetDiscoverFeed)
		api.GET("/search", movieH.SearchMovies)

		// Movie Detail (TMDB proxy)
		api.GET("/movies/:id", movieH.GetMovieDetail)
		api.GET("/movies/:id/videos", movieH.GetMovieVideos)
		api.GET("/movies/:id/credits", movieH.GetMovieCredits)
		api.GET("/movies/:id/providers", movieH.GetMovieProviders)

		// Watchlist
		api.GET("/watchlist", movieH.GetWatchlist)
		api.POST("/watchlist", movieH.AddToWatchlist)
		api.DELETE("/watchlist/:movie_id", movieH.RemoveFromWatchlist)
		api.GET("/watchlist/:movie_id/check", movieH.CheckWatchlist)

		// Friends
		api.GET("/friends", socialH.GetFriends)
		api.POST("/friends/request", socialH.SendFriendRequest)
		api.PUT("/friends/accept/:id", socialH.AcceptFriendRequest)
		api.GET("/friends/search", socialH.SearchUsers)

		// Feed
		api.GET("/feed", socialH.GetFriendsFeed)
		api.POST("/posts", socialH.CreatePost)
	}
}
