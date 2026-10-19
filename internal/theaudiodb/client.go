package theaudiodb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	api "github.com/metwurcht/torrent-all-in-one/internal/theaudiodb/generated"
)

const defaultBaseURL = "https://www.theaudiodb.com/api/v1/json"

type Client struct {
	client *api.Client
	apiKey string
}

func NewClient() *Client {
	baseURL := strings.TrimRight(os.Getenv("THEAUDIODB_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	generatedClient, err := api.NewClient(baseURL, api.WithHTTPClient(&http.Client{Timeout: 15 * time.Second}))
	if err != nil {
		panic(fmt.Sprintf("initialisation du client TheAudioDB impossible: %v", err))
	}
	key := strings.TrimSpace(os.Getenv("THEAUDIODB_API_KEY"))
	if key == "" {
		key = "123"
	}
	return &Client{client: generatedClient, apiKey: key}
}

func (c *Client) SearchAlbum(ctx context.Context, query string) ([]Album, error) {
	artist, title := splitSearchQuery(query)
	params := &api.SearchAlbumParams{S: artist}
	if title != "" {
		params.A = &title
	}
	response, err := c.client.SearchAlbum(ctx, c.apiKey, params)
	if err != nil {
		return nil, fmt.Errorf("requête TheAudioDB: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("TheAudioDB erreur: %s", response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse albums TheAudioDB: %w", err)
	}
	return decodeAlbums(body)
}

func (c *Client) GetAlbumDetails(ctx context.Context, id string) (*Album, error) {
	albumID := parseNumber(id)
	if albumID == 0 {
		return nil, fmt.Errorf("ID d'album TheAudioDB invalide: %q", id)
	}
	albumResponse, err := c.client.LookupAlbum(ctx, c.apiKey, &api.LookupAlbumParams{M: &albumID})
	if err != nil {
		return nil, fmt.Errorf("requête détail TheAudioDB: %w", err)
	}
	defer albumResponse.Body.Close()
	if albumResponse.StatusCode < 200 || albumResponse.StatusCode >= 300 {
		return nil, fmt.Errorf("TheAudioDB erreur: %s", albumResponse.Status)
	}
	albumBody, err := io.ReadAll(albumResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse détail TheAudioDB: %w", err)
	}
	albums, err := decodeAlbums(albumBody)
	if err != nil {
		return nil, err
	}
	if len(albums) == 0 {
		return nil, fmt.Errorf("album TheAudioDB %q introuvable", id)
	}
	album := &albums[0]

	trackResponse, err := c.client.LookupTrack(ctx, c.apiKey, &api.LookupTrackParams{M: &albumID})
	if err != nil {
		return nil, fmt.Errorf("requête pistes TheAudioDB: %w", err)
	}
	defer trackResponse.Body.Close()
	if trackResponse.StatusCode < 200 || trackResponse.StatusCode >= 300 {
		return nil, fmt.Errorf("TheAudioDB erreur pistes: %s", trackResponse.Status)
	}
	trackBody, err := io.ReadAll(trackResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("lecture réponse pistes TheAudioDB: %w", err)
	}
	album.Tracks, err = decodeTracks(trackBody)
	if err != nil {
		return nil, err
	}
	album.TotalTracks = len(album.Tracks)
	return album, nil
}

type albumResponse struct {
	Albums []albumRecord `json:"album"`
}

type albumRecord struct {
	ID            flexibleString `json:"idAlbum"`
	Artist        flexibleString `json:"strArtist"`
	Title         flexibleString `json:"strAlbum"`
	Year          flexibleString `json:"intYearReleased"`
	Country       flexibleString `json:"strCountry"`
	Label         flexibleString `json:"strLabel"`
	CatalogNumber flexibleString `json:"strCatalogNumber"`
	Barcode       flexibleString `json:"strBarcode"`
	Status        flexibleString `json:"strStatus"`
	Genre         flexibleString `json:"strGenre"`
	Cover         flexibleString `json:"strAlbumThumbHQ"`
	CoverFallback flexibleString `json:"strAlbumThumb"`
}

func decodeAlbums(body []byte) ([]Album, error) {
	var response albumResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("décodage réponse albums TheAudioDB: %w", err)
	}
	albums := make([]Album, 0, len(response.Albums))
	for _, item := range response.Albums {
		if item.ID == "" || item.Title == "" {
			continue
		}
		cover := string(item.Cover)
		if cover == "" {
			cover = string(item.CoverFallback)
		}
		var genres []string
		if item.Genre != "" {
			genres = []string{string(item.Genre)}
		}
		albums = append(albums, Album{
			ID: string(item.ID), Title: string(item.Title), Artist: string(item.Artist), Date: string(item.Year),
			Country: string(item.Country), Label: string(item.Label), CatalogNumber: string(item.CatalogNumber),
			Barcode: string(item.Barcode), Status: string(item.Status), Genres: genres, CoverArtURL: cover,
		})
	}
	return albums, nil
}

type trackResponse struct {
	Tracks []trackRecord `json:"track"`
}

type trackRecord struct {
	Number  flexibleString `json:"intTrackNumber"`
	Title   flexibleString `json:"strTrack"`
	Length  flexibleString `json:"strTrackDuration"`
	Seconds flexibleString `json:"intDurationSeconds"`
}

func decodeTracks(body []byte) ([]Track, error) {
	var response trackResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("décodage réponse pistes TheAudioDB: %w", err)
	}
	tracks := make([]Track, 0, len(response.Tracks))
	for index, item := range response.Tracks {
		number := parseNumber(string(item.Number))
		if number == 0 {
			number = index + 1
		}
		tracks = append(tracks, Track{Number: number, Title: string(item.Title), Duration: parseDuration(string(item.Length), string(item.Seconds))})
	}
	return tracks, nil
}

type flexibleString string

func (s *flexibleString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*s = flexibleString(text)
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return err
	}
	*s = flexibleString(number.String())
	return nil
}

func ExtractKeywords(dirname string) string {
	patterns := []string{
		`(?i)\b(flac|mp3|aac|opus|ogg|alac|wav|wma)\b`,
		`(?i)\b(16bit|24bit|32bit|16\-bit|24\-bit|32\-bit)\b`,
		`(?i)\b(44\.1kHz|48kHz|96kHz|192kHz|44100|48000|96000|192000)\b`,
		`(?i)\b(320kbps|256kbps|192kbps|128kbps|vbr|cbr|v0|v2)\b`,
		`(?i)\b(web|cd|vinyl|hdtracks|qobuz|tidal|deezer)\b`,
		`\[(.*?)\]`, `\{(.*?)\}`, `\((.*?)\)`,
		`[_.]`,
	}
	result := dirname
	for _, pattern := range patterns {
		result = regexp.MustCompile(pattern).ReplaceAllString(result, " ")
	}
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)
	if match := regexp.MustCompile(`\b(19|20)\d{2}\b`).FindString(result); match != "" {
		result = strings.TrimSpace(strings.SplitN(result, match, 2)[0])
	}
	return result
}

func splitSearchQuery(query string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(query), " - ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(query), ""
}

func ParseDirectID(input string) (string, bool) {
	input = strings.TrimSpace(input)
	for _, prefix := range []string{"id:", "album:"} {
		if strings.HasPrefix(strings.ToLower(input), prefix) {
			input = strings.TrimSpace(input[len(prefix):])
			break
		}
	}
	if regexp.MustCompile(`^\d+$`).MatchString(input) {
		return input, true
	}
	return "", false
}

func (c *Client) URLForAlbum(id string) string {
	return "https://www.theaudiodb.com/album/" + url.PathEscape(id)
}
