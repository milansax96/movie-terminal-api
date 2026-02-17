package tmdb

import (
	"github.com/milansax96/movie-terminal-api/internal/models"
)

// ToDomain converts the raw TMDB Movie struct into our clean internal Model.
func (m Movie) ToDomain() models.Movie {
	// TMDB uses "Name" for TV shows and "Title" for Movies.
	displayTitle := m.Title
	if displayTitle == "" {
		displayTitle = m.Name
	}

	return models.Movie{
		ID:           m.ID,
		Title:        displayTitle,
		Overview:     m.Overview,
		PosterPath:   m.PosterPath,
		BackdropPath: m.BackdropPath,
		ReleaseDate:  m.ReleaseDate,
		VoteAverage:  m.VoteAverage,
		MediaType:    m.MediaType,
	}
}

func toDomainList(tmdbMovies []Movie) []models.Movie {
	domain := make([]models.Movie, len(tmdbMovies))
	for i, m := range tmdbMovies {
		domain[i] = m.ToDomain()
	}

	return domain
}
