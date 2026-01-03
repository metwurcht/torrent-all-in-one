package tmdb

import "fmt"

// Movie représente un film avec ses métadonnées TMDB
type Movie struct {
	ID                  int          `json:"id"`
	Title               string       `json:"title"`
	OriginalTitle       string       `json:"original_title"`
	Overview            string       `json:"overview"`
	ReleaseDate         string       `json:"release_date"`
	PosterPath          string       `json:"poster_path"`
	BackdropPath        string       `json:"backdrop_path"`
	VoteAverage         float64      `json:"vote_average"`
	VoteCount           int          `json:"vote_count"`
	Runtime             int          `json:"runtime"`
	Budget              int64        `json:"budget"`
	Revenue             int64        `json:"revenue"`
	Tagline             string       `json:"tagline"`
	IMDbID              string       `json:"imdb_id"`
	Genres              []string     `json:"genres"`
	ProductionCompanies []string     `json:"production_companies"`
	Directors           []string     `json:"directors"`
	Cast                []CastMember `json:"cast"`
}

// CastMember représente un membre du casting
type CastMember struct {
	Name        string `json:"name"`
	Character   string `json:"character"`
	Order       int    `json:"order"`
	ProfilePath string `json:"profile_path"`
}

// Year retourne l'année de sortie du film
func (m *Movie) Year() string {
	if m.ReleaseDate == "" {
		return ""
	}

	// Format API: "2009-08-19" ou "2009"
	if len(m.ReleaseDate) >= 4 && m.ReleaseDate[4] == '-' {
		return m.ReleaseDate[:4]
	}

	// Format scraping français: "19/08/2009 (FR)" ou "27/11/2024 (FR)"
	// Chercher une année à 4 chiffres dans la chaîne
	for i := 0; i <= len(m.ReleaseDate)-4; i++ {
		candidate := m.ReleaseDate[i : i+4]
		// Vérifier que ce sont 4 chiffres commençant par 1 ou 2
		if len(candidate) == 4 && (candidate[0] == '1' || candidate[0] == '2') {
			isYear := true
			for _, c := range candidate {
				if c < '0' || c > '9' {
					isYear = false
					break
				}
			}
			if isYear {
				return candidate
			}
		}
	}

	// Fallback: si au moins 4 caractères, prendre les 4 premiers
	if len(m.ReleaseDate) >= 4 {
		return m.ReleaseDate[:4]
	}

	return ""
}

// PosterURL retourne l'URL complète du poster
func (m *Movie) PosterURL(size string) string {
	if m.PosterPath == "" {
		return ""
	}
	if size == "" {
		size = "w500"
	}
	return "https://image.tmdb.org/t/p/" + size + m.PosterPath
}

// BackdropURL retourne l'URL complète du backdrop
func (m *Movie) BackdropURL(size string) string {
	if m.BackdropPath == "" {
		return ""
	}
	if size == "" {
		size = "w1280"
	}
	return "https://image.tmdb.org/t/p/" + size + m.BackdropPath
}

// IMDbURL retourne l'URL IMDb du film
func (m *Movie) IMDbURL() string {
	if m.IMDbID == "" {
		return ""
	}
	return "https://www.imdb.com/title/" + m.IMDbID
}

// TMDbURL retourne l'URL TMDB du film
func (m *Movie) TMDbURL() string {
	return fmt.Sprintf("https://www.themoviedb.org/movie/%d", m.ID)
}
