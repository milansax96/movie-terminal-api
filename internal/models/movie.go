package models

// Movie is our clean internal domain model.
// No TMDB-specific tags here; just what our app needs.
type Movie struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_url"` // Renamed for clarity
	BackdropPath  string  `json:"backdrop_url"`
	ReleaseDate   string  `json:"release_date"`
	VoteAverage   float64 `json:"rating"`         // Renamed from VoteAverage
	MediaType     string  `json:"type"`           // "movie" or "tv"
	IsWatchlisted bool    `json:"is_watchlisted"` // We can compute this in the service!
}
