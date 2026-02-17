package models

// Movie is our clean internal domain model.
// No TMDB-specific tags here; just what our app needs.
type Movie struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	ReleaseDate   string  `json:"release_date"`
	VoteAverage   float64 `json:"vote_average"`
	MediaType     string  `json:"media_type"`
	IsWatchlisted bool    `json:"is_watchlisted"`
}
