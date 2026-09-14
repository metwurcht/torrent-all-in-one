package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.themoviedb.org/3"

type Client struct {
	httpClient *http.Client
	language   string
	apiKey     string
	baseURL    string
}

// NewClient crée un client TMDB configuré avec TMDB_API_KEY.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		language:   "fr-FR",
		apiKey:     strings.TrimSpace(os.Getenv("TMDB_API_KEY")),
		baseURL:    defaultBaseURL,
	}
}

func (c *Client) SetLanguage(lang string) { c.language = lang }

func (c *Client) doRequest(ctx context.Context, endpoint string, params url.Values, target any) error {
	if c.apiKey == "" {
		return fmt.Errorf("la variable d'environnement TMDB_API_KEY est obligatoire")
	}
	params.Set("api_key", c.apiKey)
	params.Set("language", c.language)
	requestURL := strings.TrimRight(c.baseURL, "/") + endpoint + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("erreur création requête TMDB: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erreur requête TMDB: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr struct {
			StatusMessage string `json:"status_message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.StatusMessage != "" {
			return fmt.Errorf("TMDB erreur (%s): %s", resp.Status, apiErr.StatusMessage)
		}
		return fmt.Errorf("TMDB erreur: %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("erreur décodage réponse TMDB: %w", err)
	}
	return nil
}

type apiNamed struct {
	Name string `json:"name"`
}
type apiCast struct {
	Name        string `json:"name"`
	Character   string `json:"character"`
	Order       int    `json:"order"`
	ProfilePath string `json:"profile_path"`
}
type apiCrew struct {
	Name string `json:"name"`
	Job  string `json:"job"`
}
type apiCredits struct {
	Cast []apiCast `json:"cast"`
	Crew []apiCrew `json:"crew"`
}
type apiExternalIDs struct {
	IMDbID string `json:"imdb_id"`
}

type apiMovie struct {
	ID                  int            `json:"id"`
	Title               string         `json:"title"`
	OriginalTitle       string         `json:"original_title"`
	Overview            string         `json:"overview"`
	ReleaseDate         string         `json:"release_date"`
	PosterPath          string         `json:"poster_path"`
	BackdropPath        string         `json:"backdrop_path"`
	VoteAverage         float64        `json:"vote_average"`
	VoteCount           int            `json:"vote_count"`
	Runtime             int            `json:"runtime"`
	Budget              int64          `json:"budget"`
	Revenue             int64          `json:"revenue"`
	Tagline             string         `json:"tagline"`
	Genres              []apiNamed     `json:"genres"`
	ProductionCompanies []apiNamed     `json:"production_companies"`
	Credits             apiCredits     `json:"credits"`
	ExternalIDs         apiExternalIDs `json:"external_ids"`
}
type apiTVShow struct {
	ID               int            `json:"id"`
	Name             string         `json:"name"`
	OriginalName     string         `json:"original_name"`
	Overview         string         `json:"overview"`
	FirstAirDate     string         `json:"first_air_date"`
	PosterPath       string         `json:"poster_path"`
	BackdropPath     string         `json:"backdrop_path"`
	VoteAverage      float64        `json:"vote_average"`
	VoteCount        int            `json:"vote_count"`
	NumberOfSeasons  int            `json:"number_of_seasons"`
	NumberOfEpisodes int            `json:"number_of_episodes"`
	Tagline          string         `json:"tagline"`
	Genres           []apiNamed     `json:"genres"`
	Networks         []apiNamed     `json:"networks"`
	CreatedBy        []apiNamed     `json:"created_by"`
	Status           string         `json:"status"`
	Credits          apiCredits     `json:"credits"`
	ExternalIDs      apiExternalIDs `json:"external_ids"`
}

func movieFromAPI(value apiMovie) Movie {
	movie := Movie{ID: value.ID, Title: value.Title, OriginalTitle: value.OriginalTitle, Overview: value.Overview, ReleaseDate: value.ReleaseDate, PosterPath: value.PosterPath, BackdropPath: value.BackdropPath, VoteAverage: value.VoteAverage, VoteCount: value.VoteCount, Runtime: value.Runtime, Budget: value.Budget, Revenue: value.Revenue, Tagline: value.Tagline, IMDbID: value.ExternalIDs.IMDbID}
	for _, item := range value.Genres {
		movie.Genres = append(movie.Genres, item.Name)
	}
	for _, item := range value.ProductionCompanies {
		movie.ProductionCompanies = append(movie.ProductionCompanies, item.Name)
	}
	for _, item := range value.Credits.Cast {
		movie.Cast = append(movie.Cast, CastMember{Name: item.Name, Character: item.Character, Order: item.Order, ProfilePath: item.ProfilePath})
	}
	for _, item := range value.Credits.Crew {
		if item.Job == "Director" {
			movie.Directors = append(movie.Directors, item.Name)
		}
	}
	if movie.OriginalTitle == "" {
		movie.OriginalTitle = movie.Title
	}
	return movie
}

func showFromAPI(value apiTVShow) TVShow {
	show := TVShow{ID: value.ID, Name: value.Name, OriginalName: value.OriginalName, Overview: value.Overview, FirstAirDate: value.FirstAirDate, PosterPath: value.PosterPath, BackdropPath: value.BackdropPath, VoteAverage: value.VoteAverage, VoteCount: value.VoteCount, NumberOfSeasons: value.NumberOfSeasons, NumberOfEpisodes: value.NumberOfEpisodes, Tagline: value.Tagline, Status: value.Status, IMDbID: value.ExternalIDs.IMDbID}
	for _, item := range value.Genres {
		show.Genres = append(show.Genres, item.Name)
	}
	for _, item := range value.Networks {
		show.Networks = append(show.Networks, item.Name)
	}
	for _, item := range value.CreatedBy {
		show.Creators = append(show.Creators, item.Name)
	}
	for _, item := range value.Credits.Cast {
		show.Cast = append(show.Cast, CastMember{Name: item.Name, Character: item.Character, Order: item.Order, ProfilePath: item.ProfilePath})
	}
	if show.OriginalName == "" {
		show.OriginalName = show.Name
	}
	return show
}

func (c *Client) SearchMovie(ctx context.Context, query string) ([]Movie, error) {
	var response struct {
		Results []apiMovie `json:"results"`
	}
	if err := c.doRequest(ctx, "/search/movie", url.Values{"query": {query}}, &response); err != nil {
		return nil, err
	}
	movies := make([]Movie, 0, len(response.Results))
	for _, item := range response.Results {
		movies = append(movies, movieFromAPI(item))
	}
	return movies, nil
}

func (c *Client) GetMovieDetails(ctx context.Context, id int) (*Movie, error) {
	var response apiMovie
	if err := c.doRequest(ctx, "/movie/"+strconv.Itoa(id), url.Values{"append_to_response": {"credits,external_ids"}}, &response); err != nil {
		return nil, err
	}
	movie := movieFromAPI(response)
	return &movie, nil
}

func (c *Client) SearchTVShow(ctx context.Context, query string) ([]TVShow, error) {
	var response struct {
		Results []apiTVShow `json:"results"`
	}
	if err := c.doRequest(ctx, "/search/tv", url.Values{"query": {query}}, &response); err != nil {
		return nil, err
	}
	shows := make([]TVShow, 0, len(response.Results))
	for _, item := range response.Results {
		shows = append(shows, showFromAPI(item))
	}
	return shows, nil
}

func (c *Client) GetTVShowDetails(ctx context.Context, id int) (*TVShow, error) {
	var response apiTVShow
	if err := c.doRequest(ctx, "/tv/"+strconv.Itoa(id), url.Values{"append_to_response": {"credits,external_ids"}}, &response); err != nil {
		return nil, err
	}
	show := showFromAPI(response)
	return &show, nil
}

func ExtractKeywords(filename string) string {
	name := strings.TrimSuffix(filename, "."+getExtension(filename))
	patterns := []string{`\b(1080p|720p|2160p|4k|uhd|hdr|bluray|brrip|webrip|web-dl|hdtv|dvdrip)\b`, `\b(x264|x265|h264|h265|hevc|avc|xvid)\b`, `\b(dts|dd5\.1|ac3|aac|flac|truehd|atmos)\b`, `\b(multi|french|vff|vfi|vostfr|truefrench|english)\b`, `\b(proper|repack|internal|limited|extended|unrated|directors\.cut)\b`, `\[(.*?)\]`, `\{(.*?)\}`, `[-_.]`}
	result := strings.ToLower(name)
	for _, pattern := range patterns {
		result = regexp.MustCompile("(?i)"+pattern).ReplaceAllString(result, " ")
	}
	result = strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(result, " "))
	if match := regexp.MustCompile(`\b(19|20)\d{2}\b`).FindString(name); match != "" {
		if index := strings.Index(strings.ToLower(name), match); index > 0 {
			result = strings.TrimSpace(result[:min(index, len(result))])
		}
	}
	words := strings.Fields(result)
	if len(words) > 4 {
		words = words[:4]
	}
	return strings.Join(words, " ")
}

func ParseDirectID(input string) (int, bool) {
	input = strings.TrimSpace(strings.ToLower(input))
	for _, prefix := range []string{"id:", "tmdb:"} {
		if strings.HasPrefix(input, prefix) {
			id, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(input, prefix)))
			return id, err == nil && id > 0
		}
	}
	id, err := strconv.Atoi(input)
	return id, err == nil && id > 0
}

func getExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
