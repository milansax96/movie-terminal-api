package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

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
	GenreIDs     []int   `json:"genre_ids"`
	MediaType    string  `json:"media_type"`
}

type TrendingResponse struct {
	Results []Movie `json:"results"`
	Page    int     `json:"page"`
}

type VideosResponse struct {
	Results []Video `json:"results"`
}

type Video struct {
	Key  string `json:"key"`
	Site string `json:"site"`
	Type string `json:"type"`
}

func NewClient() *Client {
	return &Client{
		APIKey:     os.Getenv("TMDB_API_KEY"),
		BaseURL:    "https://api.themoviedb.org/3",
		HTTPClient: &http.Client{},
	}
}

func (c *Client) GetTrending(mediaType string, timeWindow string) ([]Movie, error) {
	url := fmt.Sprintf("%s/trending/%s/%s?api_key=%s", c.BaseURL, mediaType, timeWindow, c.APIKey)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TrendingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

func (c *Client) GetVideos(mediaType string, id int) ([]Video, error) {
	url := fmt.Sprintf("%s/%s/%d/videos?api_key=%s", c.BaseURL, mediaType, id, c.APIKey)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result VideosResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

func (c *Client) GetMovieDetails(mediaType string, id int) (*Movie, error) {
	url := fmt.Sprintf("%s/%s/%d?api_key=%s", c.BaseURL, mediaType, id, c.APIKey)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var movie Movie
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		return nil, err
	}

	return &movie, nil
}
