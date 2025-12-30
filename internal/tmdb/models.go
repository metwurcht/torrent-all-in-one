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
