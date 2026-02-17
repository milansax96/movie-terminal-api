package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/milansax96/movie-terminal-api/internal/models"
)

// API defines the interface for TMDB API operations returning Domain Models.
type API interface {
	GetTrending(mediaType string, timeWindow string) ([]models.Movie, error)
	GetTopRated(page int) ([]models.Movie, error)
	DiscoverByGenre(genreID int, page int) ([]models.Movie, error)
	SearchMovies(query string, page int) ([]models.Movie, error)
	GetMovieDetails(mediaType string, id int) (*MovieDetail, error)
	GetVideos(mediaType string, id int) ([]Video, error)
	GetCredits(mediaType string, id int) (*CreditsResponse, error)
	GetProviders(mediaType string, id int) (json.RawMessage, error)
}

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// This must exist in the same package for the mapper to work!
type Movie struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`
	VoteAverage  float64 `json:"vote_average"`
	MediaType    string  `json:"media_type"`
}

type MovieDetail struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`
	VoteAverage  float64 `json:"vote_average"`
	Genres       []Genre `json:"genres"`
	Tagline      string  `json:"tagline"`
	Runtime      int     `json:"runtime"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Video struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Site string `json:"site"`
	Type string `json:"type"`
}

type CastMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
}

type MovieListResponse struct {
	Results []Movie `json:"results"`
}

type VideosResponse struct {
	Results []Video `json:"results"`
}

type CreditsResponse struct {
	Cast []CastMember `json:"cast"`
}

func NewClient() *Client {
	return &Client{
		APIKey:     os.Getenv("TMDB_API_KEY"),
		BaseURL:    "https://api.themoviedb.org/3",
		HTTPClient: &http.Client{},
	}
}

func (c *Client) fetch(path string, target interface{}) error {
	fullURL := fmt.Sprintf("%s%s", c.BaseURL, path)

	// Handle API Key query param injection
	u, err := url.Parse(fullURL)
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("api_key", c.APIKey)
	u.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tmdb api error: status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// --- API Implementations ---

func (c *Client) GetTrending(mediaType string, timeWindow string) ([]models.Movie, error) {
	var res MovieListResponse
	path := fmt.Sprintf("/trending/%s/%s", mediaType, timeWindow)
	if err := c.fetch(path, &res); err != nil {
		return nil, err
	}
	return toDomainList(res.Results), nil
}

func (c *Client) GetTopRated(page int) ([]models.Movie, error) {
	var res MovieListResponse
	path := fmt.Sprintf("/movie/top_rated?page=%d", page)
	if err := c.fetch(path, &res); err != nil {
		return nil, err
	}
	return toDomainList(res.Results), nil
}

func (c *Client) DiscoverByGenre(genreID int, page int) ([]models.Movie, error) {
	var res MovieListResponse
	path := fmt.Sprintf("/discover/movie?with_genres=%d&page=%d", genreID, page)
	if err := c.fetch(path, &res); err != nil {
		return nil, err
	}
	return toDomainList(res.Results), nil
}

func (c *Client) SearchMovies(query string, page int) ([]models.Movie, error) {
	var res MovieListResponse
	path := fmt.Sprintf("/search/multi?query=%s&page=%d", url.QueryEscape(query), page)
	if err := c.fetch(path, &res); err != nil {
		return nil, err
	}
	return toDomainList(res.Results), nil
}

// Detailed views remain mostly the same but use the fetch helper
func (c *Client) GetMovieDetails(mediaType string, id int) (*MovieDetail, error) {
	var detail MovieDetail
	if err := c.fetch(fmt.Sprintf("/%s/%d", mediaType, id), &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

// GetVideos returns videos (trailers, teasers, etc.) for a title.
func (c *Client) GetVideos(mediaType string, id int) ([]Video, error) {
	var res VideosResponse
	path := fmt.Sprintf("/%s/%d/videos", mediaType, id)

	if err := c.fetch(path, &res); err != nil {
		return nil, err
	}

	return res.Results, nil
}

// GetCredits returns cast and crew credits for a title.
func (c *Client) GetCredits(mediaType string, id int) (*CreditsResponse, error) {
	var res CreditsResponse
	path := fmt.Sprintf("/%s/%d/credits", mediaType, id)

	if err := c.fetch(path, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetProviders returns streaming provider information for a title.
func (c *Client) GetProviders(mediaType string, id int) (json.RawMessage, error) {
	var res json.RawMessage
	path := fmt.Sprintf("/%s/%d/watch/providers", mediaType, id)

	if err := c.fetch(path, &res); err != nil {
		return nil, err
	}

	return res, nil
}
