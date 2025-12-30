package tmdb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL = "https://www.themoviedb.org"
)

// Client représente un client pour le scraping TMDB
type Client struct {
	httpClient *http.Client
	language   string
	userAgent  string
}

// NewClient crée un nouveau client TMDB (scraping)
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		language:  "fr-FR",
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// SetLanguage définit la langue pour les requêtes
func (c *Client) SetLanguage(lang string) {
	c.language = lang
}

// doRequest effectue une requête HTTP avec les headers appropriés
func (c *Client) doRequest(ctx context.Context, urlStr string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", c.language+",en;q=0.5")

	return c.httpClient.Do(req)
}

// SearchMovie recherche des films par mots-clés via scraping
func (c *Client) SearchMovie(ctx context.Context, query string) ([]Movie, error) {
	searchURL := fmt.Sprintf("%s/search/movie?query=%s&language=%s",
		baseURL, url.QueryEscape(query), c.language)

	resp, err := c.doRequest(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("erreur requête TMDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB erreur: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur parsing HTML: %w", err)
	}

	var movies []Movie

	// Parser les résultats de recherche de films
	doc.Find("div.search_results.movie div.card").Each(func(i int, s *goquery.Selection) {
		movie := Movie{}

		// Extraire le lien et l'ID
		link := s.Find("a.result")
		if href, exists := link.Attr("href"); exists {
			// Format: /movie/12345-slug
			movie.ID = extractIDFromURL(href)
		}

		// Titre
		movie.Title = cleanText(s.Find("h2").First().Text())

		// Titre original (dans le span.title à l'intérieur du h2)
		if origTitle := s.Find("h2 span.title").Text(); origTitle != "" {
			// Nettoyer les parenthèses
			origTitle = strings.TrimPrefix(origTitle, "(")
			origTitle = strings.TrimSuffix(origTitle, ")")
			movie.OriginalTitle = cleanText(origTitle)
		}
		if movie.OriginalTitle == "" {
			movie.OriginalTitle = movie.Title
		}

		// Date de sortie
		movie.ReleaseDate = cleanText(s.Find("span.release_date").Text())

		// Synopsis
		movie.Overview = cleanText(s.Find("div.overview p").Text())

		// Poster
		if img := s.Find("img.poster"); img.Length() > 0 {
			if src, exists := img.Attr("src"); exists {
				movie.PosterPath = extractPosterPath(src)
			}
		}

		if movie.ID > 0 && movie.Title != "" {
			movies = append(movies, movie)
		}
	})

	return movies, nil
}

// GetMovieDetails récupère les détails complets d'un film via scraping
func (c *Client) GetMovieDetails(ctx context.Context, id int) (*Movie, error) {
	movieURL := fmt.Sprintf("%s/movie/%d?language=%s", baseURL, id, c.language)

	resp, err := c.doRequest(ctx, movieURL)
	if err != nil {
		return nil, fmt.Errorf("erreur requête TMDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB erreur: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur parsing HTML: %w", err)
	}

	movie := &Movie{
		ID: id,
	}

	// Titre principal
	movie.Title = cleanText(doc.Find("div.title h2").First().Text())

	// Titre original
	doc.Find("section.facts.left_column p.wrap").Each(func(i int, s *goquery.Selection) {
		strong := cleanText(s.Find("strong").Text())
		if strings.Contains(strings.ToLower(strong), "titre") && strings.Contains(strings.ToLower(strong), "origine") {
			fullText := cleanText(s.Text())
			movie.OriginalTitle = strings.TrimSpace(strings.TrimPrefix(fullText, strong))
		}
	})
	if movie.OriginalTitle == "" {
		movie.OriginalTitle = movie.Title
	}

	// Tagline
	movie.Tagline = cleanText(doc.Find("section.content_wrapper h3.tagline").Text())

	// Synopsis
	movie.Overview = cleanText(doc.Find("section.content_wrapper div.overview p").Text())

	// Date de sortie et runtime depuis les facts
	doc.Find("section.content_wrapper span.release").Each(func(i int, s *goquery.Selection) {
		text := cleanText(s.Text())
		if movie.ReleaseDate == "" && len(text) > 0 {
			movie.ReleaseDate = text
		}
	})

	// Runtime
	doc.Find("section.content_wrapper span.runtime").Each(func(i int, s *goquery.Selection) {
		text := cleanText(s.Text())
		movie.Runtime = parseRuntime(text)
	})

	// Genres
	doc.Find("section.content_wrapper span.genres a").Each(func(i int, s *goquery.Selection) {
		genre := cleanText(s.Text())
		if genre != "" {
			movie.Genres = append(movie.Genres, genre)
		}
	})

	// Note
	doc.Find("div.user_score_chart").Each(func(i int, s *goquery.Selection) {
		if percent, exists := s.Attr("data-percent"); exists {
			if val, err := strconv.ParseFloat(percent, 64); err == nil {
				movie.VoteAverage = val / 10.0
			}
		}
	})

	// Poster
	if img := doc.Find("section.content_wrapper img.poster"); img.Length() > 0 {
		if src, exists := img.Attr("src"); exists {
			movie.PosterPath = extractPosterPath(src)
		}
	}

	// Backdrop
	doc.Find("div.image_content.backdrop img").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			movie.BackdropPath = extractPosterPath(src)
		}
	})

	// Réalisateurs et cast - récupérer depuis la page des credits
	c.getCredits(ctx, id, movie)

	// IMDb ID - récupérer depuis les liens externes
	doc.Find("a.social_link").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			if strings.Contains(href, "imdb.com") {
				// Extraire l'ID IMDb
				re := regexp.MustCompile(`(tt\d+)`)
				if match := re.FindString(href); match != "" {
					movie.IMDbID = match
				}
			}
		}
	})

	return movie, nil
}

// getCredits récupère le casting et l'équipe technique
func (c *Client) getCredits(ctx context.Context, id int, movie *Movie) {
	creditsURL := fmt.Sprintf("%s/movie/%d/cast?language=%s", baseURL, id, c.language)

	resp, err := c.doRequest(ctx, creditsURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	// Cast
	doc.Find("section.panel.top_billed ol.people li").Each(func(i int, s *goquery.Selection) {
		if i >= 10 {
			return
		}
		name := cleanText(s.Find("p a").First().Text())
		character := cleanText(s.Find("p.character").Text())

		// Récupérer la photo de profil
		profilePath := ""
		if img := s.Find("img.profile"); img.Length() > 0 {
			if src, exists := img.Attr("src"); exists {
				profilePath = extractPosterPath(src)
			}
		}

		if name != "" {
			movie.Cast = append(movie.Cast, CastMember{
				Name:        name,
				Character:   character,
				Order:       i,
				ProfilePath: profilePath,
			})
		}
	})

	// Réalisateurs
	doc.Find("section.panel ol.people.crew li").Each(func(i int, s *goquery.Selection) {
		job := cleanText(s.Find("p.character").Text())
		if strings.Contains(strings.ToLower(job), "director") || strings.Contains(strings.ToLower(job), "réalisateur") {
			name := cleanText(s.Find("p a").First().Text())
			if name != "" {
				movie.Directors = append(movie.Directors, name)
			}
		}
	})
}

// extractIDFromURL extrait l'ID depuis une URL TMDB
func extractIDFromURL(urlPath string) int {
	// Format: /movie/12345-slug ou /movie/12345
	re := regexp.MustCompile(`/movie/(\d+)`)
	matches := re.FindStringSubmatch(urlPath)
	if len(matches) >= 2 {
		id, _ := strconv.Atoi(matches[1])
		return id
	}
	return 0
}

// extractPosterPath extrait le chemin du poster depuis l'URL complète
func extractPosterPath(src string) string {
	// Format: https://media.themoviedb.org/t/p/w94_and_h141_face/xxx.jpg
	re := regexp.MustCompile(`/t/p/[^/]+(/[^"]+)`)
	matches := re.FindStringSubmatch(src)
	if len(matches) >= 2 {
		return matches[1]
	}
	// Essayer un autre format
	re = regexp.MustCompile(`/p/[^/]+(/[^"]+)`)
	matches = re.FindStringSubmatch(src)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// parseRuntime convertit une durée texte en minutes
func parseRuntime(text string) int {
	// Format: "2h 15m" ou "135m" ou "2 h 15 min"
	text = strings.ToLower(text)

	hours := 0
	minutes := 0

	// Chercher les heures
	reHours := regexp.MustCompile(`(\d+)\s*h`)
	if matches := reHours.FindStringSubmatch(text); len(matches) >= 2 {
		hours, _ = strconv.Atoi(matches[1])
	}

	// Chercher les minutes
	reMinutes := regexp.MustCompile(`(\d+)\s*m`)
	if matches := reMinutes.FindStringSubmatch(text); len(matches) >= 2 {
		minutes, _ = strconv.Atoi(matches[1])
	}

	return hours*60 + minutes
}

// ExtractKeywords extrait les mots-clés pertinents d'un nom de fichier
func ExtractKeywords(filename string) string {
	// Supprimer l'extension
	name := strings.TrimSuffix(filename, "."+getExtension(filename))

	// Patterns courants à supprimer
	patterns := []string{
		`\b(1080p|720p|2160p|4k|uhd|hdr|bluray|brrip|webrip|web-dl|hdtv|dvdrip)\b`,
		`\b(x264|x265|h264|h265|hevc|avc|xvid)\b`,
		`\b(dts|dd5\.1|ac3|aac|flac|truehd|atmos)\b`,
		`\b(multi|french|vff|vfi|vostfr|truefrench|english)\b`,
		`\b(proper|repack|internal|limited|extended|unrated|directors\.cut)\b`,
		`\[(.*?)\]`,
		`\{(.*?)\}`,
		`[-_.]`,
	}

	result := strings.ToLower(name)
	for _, p := range patterns {
		re := regexp.MustCompile("(?i)" + p)
		result = re.ReplaceAllString(result, " ")
	}

	// Nettoyer les espaces multiples
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	result = strings.TrimSpace(result)

	// Extraire potentiellement l'année
	yearRe := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	if match := yearRe.FindString(name); match != "" {
		// Garder seulement ce qui précède l'année
		idx := strings.Index(strings.ToLower(name), match)
		if idx > 0 {
			result = strings.TrimSpace(result[:min(idx, len(result))])
		}
	}

	// Prendre les 4 premiers mots max
	words := strings.Fields(result)
	if len(words) > 4 {
		words = words[:4]
	}

	return strings.Join(words, " ")
}

// ParseDirectID parse un ID TMDB direct depuis une entrée utilisateur
func ParseDirectID(input string) (int, bool) {
	input = strings.TrimSpace(strings.ToLower(input))

	// Format: id:12345 ou tmdb:12345
	prefixes := []string{"id:", "tmdb:"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(input, prefix) {
			idStr := strings.TrimPrefix(input, prefix)
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				return id, true
			}
		}
	}

	// Essayer de parser directement comme un nombre
	if id, err := strconv.Atoi(input); err == nil && id > 0 {
		return id, true
	}

	return 0, false
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

// cleanText nettoie une chaîne en supprimant les retours à la ligne et espaces multiples
func cleanText(s string) string {
	// Remplacer les retours à la ligne et tabulations par des espaces
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	// Supprimer les espaces multiples
	spaceRegex := regexp.MustCompile(`\s+`)
	s = spaceRegex.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
