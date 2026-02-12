package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	apiBaseURL = "https://musicbrainz.org/ws/2"
	userAgent  = "TorrentAllInOne/1.0 (https://github.com/metwurcht/torrent-all-in-one)"
)

// Client représente un client pour l'API MusicBrainz
type Client struct {
	httpClient *http.Client
}

// NewClient crée un nouveau client MusicBrainz
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// doRequest effectue une requête GET avec les headers appropriés
func (c *Client) doRequest(ctx context.Context, urlStr string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	// MusicBrainz requiert un User-Agent conforme
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// searchResponse représente la réponse JSON de recherche MusicBrainz
type searchResponse struct {
	Releases []releaseResult `json:"releases"`
}

type releaseResult struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Date         string `json:"date"`
	Country      string `json:"country"`
	Status       string `json:"status"`
	Score        int    `json:"score"`
	ArtistCredit []struct {
		Name   string `json:"name"`
		Artist struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"artist"`
	} `json:"artist-credit"`
	ReleaseGroup struct {
		PrimaryType string `json:"primary-type"`
	} `json:"release-group"`
	LabelInfo []struct {
		CatalogNumber string `json:"catalog-number"`
		Label         struct {
			Name string `json:"name"`
		} `json:"label"`
	} `json:"label-info"`
}

// SearchAlbum recherche des albums par mots-clés
func (c *Client) SearchAlbum(ctx context.Context, query string) ([]Album, error) {
	searchURL := fmt.Sprintf("%s/release?query=%s&fmt=json&limit=15",
		apiBaseURL, url.QueryEscape(query))

	resp, err := c.doRequest(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("erreur requête MusicBrainz: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MusicBrainz erreur: %s", resp.Status)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	var albums []Album
	for _, r := range result.Releases {
		album := Album{
			MBID:    r.ID,
			Title:   r.Title,
			Date:    r.Date,
			Country: r.Country,
			Status:  r.Status,
		}

		// Artiste
		if len(r.ArtistCredit) > 0 {
			var artists []string
			for _, ac := range r.ArtistCredit {
				artists = append(artists, ac.Name)
			}
			album.Artist = strings.Join(artists, ", ")
		}

		// Label
		if len(r.LabelInfo) > 0 {
			album.Label = r.LabelInfo[0].Label.Name
			album.CatalogNumber = r.LabelInfo[0].CatalogNumber
		}

		if album.MBID != "" && album.Title != "" {
			albums = append(albums, album)
		}
	}

	return albums, nil
}

// releaseResponse représente la réponse JSON détaillée d'un release
type releaseResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Date         string `json:"date"`
	Country      string `json:"country"`
	Status       string `json:"status"`
	Barcode      string `json:"barcode"`
	ArtistCredit []struct {
		Name   string `json:"name"`
		Artist struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"artist"`
	} `json:"artist-credit"`
	ReleaseGroup struct {
		PrimaryType string `json:"primary-type"`
		Genres      []struct {
			Name string `json:"name"`
		} `json:"genres"`
	} `json:"release-group"`
	LabelInfo []struct {
		CatalogNumber string `json:"catalog-number"`
		Label         struct {
			Name string `json:"name"`
		} `json:"label"`
	} `json:"label-info"`
	Media []struct {
		Position int `json:"position"`
		Tracks   []struct {
			Number   string `json:"number"`
			Title    string `json:"title"`
			Position int    `json:"position"`
			Length   int    `json:"length"` // millisecondes
		} `json:"tracks"`
		TrackCount int `json:"track-count"`
	} `json:"media"`
	CoverArtArchive struct {
		Front bool `json:"front"`
	} `json:"cover-art-archive"`
}

// GetAlbumDetails récupère les détails complets d'un album par MBID
func (c *Client) GetAlbumDetails(ctx context.Context, mbid string) (*Album, error) {
	detailURL := fmt.Sprintf("%s/release/%s?inc=artists+recordings+labels+release-groups+genres&fmt=json",
		apiBaseURL, url.PathEscape(mbid))

	resp, err := c.doRequest(ctx, detailURL)
	if err != nil {
		return nil, fmt.Errorf("erreur requête MusicBrainz: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MusicBrainz erreur: %s", resp.Status)
	}

	var r releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("erreur parsing JSON: %w", err)
	}

	album := &Album{
		MBID:    r.ID,
		Title:   r.Title,
		Date:    r.Date,
		Country: r.Country,
		Status:  r.Status,
		Barcode: r.Barcode,
	}

	// Artiste
	if len(r.ArtistCredit) > 0 {
		var artists []string
		for _, ac := range r.ArtistCredit {
			artists = append(artists, ac.Name)
		}
		album.Artist = strings.Join(artists, ", ")
	}

	// Label
	if len(r.LabelInfo) > 0 {
		album.Label = r.LabelInfo[0].Label.Name
		album.CatalogNumber = r.LabelInfo[0].CatalogNumber
	}

	// Genres
	for _, g := range r.ReleaseGroup.Genres {
		if g.Name != "" {
			album.Genres = append(album.Genres, g.Name)
		}
	}

	// Tracks
	totalTracks := 0
	for _, m := range r.Media {
		for _, t := range m.Tracks {
			album.Tracks = append(album.Tracks, Track{
				Number:   t.Position,
				Title:    t.Title,
				Duration: t.Length,
			})
		}
		totalTracks += m.TrackCount
	}
	album.TotalTracks = totalTracks

	// Cover art
	if r.CoverArtArchive.Front {
		album.CoverArtURL = fmt.Sprintf("https://coverartarchive.org/release/%s/front-500", album.MBID)
	}

	return album, nil
}

// ExtractKeywords extrait les mots-clés d'un nom de dossier d'album
func ExtractKeywords(dirname string) string {
	// Patterns courants à supprimer dans les noms de dossiers musicaux
	patterns := []string{
		`(?i)\b(flac|mp3|aac|opus|ogg|alac|wav|wma)\b`,
		`(?i)\b(16bit|24bit|32bit|16\-bit|24\-bit|32\-bit)\b`,
		`(?i)\b(44\.1kHz|48kHz|96kHz|192kHz|44100|48000|96000|192000)\b`,
		`(?i)\b(320kbps|256kbps|192kbps|128kbps|vbr|cbr|v0|v2)\b`,
		`(?i)\b(web|cd|vinyl|hdtracks|qobuz|tidal|deezer)\b`,
		`\[(.*?)\]`,
		`\{(.*?)\}`,
		`\((.*?)\)`,
		`[-_.]`,
	}

	result := dirname
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		result = re.ReplaceAllString(result, " ")
	}

	// Nettoyer les espaces multiples
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)

	// Extraire potentiellement l'année pour ne garder que ce qui précède
	yearRe := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	if match := yearRe.FindString(dirname); match != "" {
		idx := strings.Index(strings.ToLower(result), strings.ToLower(match))
		if idx > 0 {
			result = strings.TrimSpace(result[:idx])
		}
	}

	// Prendre les 6 premiers mots max (artiste + album peut être long)
	words := strings.Fields(result)
	if len(words) > 6 {
		words = words[:6]
	}

	return strings.Join(words, " ")
}

// ParseDirectMBID parse un MBID direct depuis l'entrée utilisateur
func ParseDirectMBID(input string) (string, bool) {
	input = strings.TrimSpace(input)

	// Format: id:mbid ou mbid:xxx
	prefixes := []string{"id:", "mbid:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(input), prefix) {
			mbid := strings.TrimSpace(input[len(prefix):])
			if isValidMBID(mbid) {
				return mbid, true
			}
		}
	}

	// Essayer de parser directement comme MBID
	if isValidMBID(input) {
		return input, true
	}

	return "", false
}

// isValidMBID vérifie si une chaîne est un UUID valide (format MBID)
func isValidMBID(s string) bool {
	// Format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars)
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}
