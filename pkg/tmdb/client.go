// Package tmdb provides a client for the TMDB (The Movie Database) API.
package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

// Client is an HTTP client for the TMDB API.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// Movie represents a movie or TV show summary from TMDB.
type Movie struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`
	FirstAirDate string  `json:"first_air_date"`
	VoteAverage  float64 `json:"vote_average"`
	VoteCount    int     `json:"vote_count"`
	GenreIDs     []int   `json:"genre_ids"`
	MediaType    string  `json:"media_type"`
}

// Genre represents a TMDB genre.
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MovieDetail represents full details for a movie or TV show.
type MovieDetail struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`
	FirstAirDate string  `json:"first_air_date"`
	VoteAverage  float64 `json:"vote_average"`
	VoteCount    int     `json:"vote_count"`
	Genres       []Genre `json:"genres"`
	Tagline      string  `json:"tagline"`
	Runtime      int     `json:"runtime"`
	MediaType    string  `json:"media_type,omitempty"`
}

// Video represents a video (trailer, teaser, etc.) associated with a title.
type Video struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Site string `json:"site"`
	Type string `json:"type"`
}

// CastMember represents an actor in a movie's credits.
type CastMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
}

// MovieListResponse wraps a paginated list of movies from TMDB.
type MovieListResponse struct {
	Results []Movie `json:"results"`
	Page    int     `json:"page"`
}

// VideosResponse wraps a list of videos from TMDB.
type VideosResponse struct {
	Results []Video `json:"results"`
}

// CreditsResponse wraps cast members from TMDB.
type CreditsResponse struct {
	Cast []CastMember `json:"cast"`
}

// API defines the interface for TMDB API operations.
type API interface {
	GetTrending(mediaType string, timeWindow string) ([]Movie, error)
	GetTopRated(page int) ([]Movie, error)
	DiscoverByGenre(genreID int, page int) ([]Movie, error)
	SearchMovies(query string, page int) ([]Movie, error)
	GetMovieDetails(mediaType string, id int) (*MovieDetail, error)
	GetVideos(mediaType string, id int) ([]Video, error)
	GetCredits(mediaType string, id int) (*CreditsResponse, error)
	GetProviders(mediaType string, id int) (json.RawMessage, error)
}

// NewClient creates a new TMDB client using the TMDB_API_KEY environment variable.
func NewClient() *Client {
	return &Client{
		APIKey:     os.Getenv("TMDB_API_KEY"),
		BaseURL:    "https://api.themoviedb.org/3",
		HTTPClient: &http.Client{},
	}
}

// NewClientWithKey creates a new TMDB client with the given API key.
func NewClientWithKey(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    "https://api.themoviedb.org/3",
		HTTPClient: &http.Client{},
	}
}

func (c *Client) get(path string) (*http.Response, error) {
	return c.HTTPClient.Get(fmt.Sprintf("%s%s", c.BaseURL, path))
}

// GetTrending returns trending movies or TV shows for the given time window.
func (c *Client) GetTrending(mediaType string, timeWindow string) (_ []Movie, err error) {
	resp, err := c.get(fmt.Sprintf("/trending/%s/%s?api_key=%s", mediaType, timeWindow, c.APIKey))
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var result MovieListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// GetTopRated returns top-rated movies for the given page.
func (c *Client) GetTopRated(page int) (_ []Movie, err error) {
	resp, err := c.get(fmt.Sprintf("/movie/top_rated?api_key=%s&page=%d", c.APIKey, page))
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var result MovieListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// DiscoverByGenre returns movies matching the given genre ID.
func (c *Client) DiscoverByGenre(genreID int, page int) (_ []Movie, err error) {
	resp, err := c.get(fmt.Sprintf("/discover/movie?api_key=%s&with_genres=%d&page=%d", c.APIKey, genreID, page))
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var result MovieListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// SearchMovies searches for movies and TV shows matching the query.
func (c *Client) SearchMovies(query string, page int) (_ []Movie, err error) {
	resp, err := c.get(fmt.Sprintf("/search/multi?api_key=%s&query=%s&page=%d", c.APIKey, url.QueryEscape(query), page))
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var result MovieListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// GetMovieDetails returns full details for a movie or TV show.
func (c *Client) GetMovieDetails(mediaType string, id int) (_ *MovieDetail, err error) {
	resp, err := c.get(fmt.Sprintf("/%s/%d?api_key=%s", mediaType, id, c.APIKey))
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var detail MovieDetail
	if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

// GetVideos returns videos (trailers, teasers, etc.) for a title.
func (c *Client) GetVideos(mediaType string, id int) (_ []Video, err error) {
	resp, err := c.get(fmt.Sprintf("/%s/%d/videos?api_key=%s", mediaType, id, c.APIKey))
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var result VideosResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// GetCredits returns cast and crew credits for a title.
func (c *Client) GetCredits(mediaType string, id int) (_ *CreditsResponse, err error) {
	resp, err := c.get(fmt.Sprintf("/%s/%d/credits?api_key=%s", mediaType, id, c.APIKey))
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var result CreditsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetProviders returns streaming provider information for a title.
func (c *Client) GetProviders(mediaType string, id int) (_ json.RawMessage, err error) {
	resp, err := c.get(fmt.Sprintf("/%s/%d/watch/providers?api_key=%s", mediaType, id, c.APIKey))
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var result json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
