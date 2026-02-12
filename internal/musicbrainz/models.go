// Package musicbrainz provides types and a client for the MusicBrainz API.
package musicbrainz

import (
	"fmt"

	"github.com/metwurcht/torrent-all-in-one/internal/media"
)

// Album représente un album de musique avec ses métadonnées MusicBrainz
type Album struct {
	MBID          string   `json:"id"`
	Title         string   `json:"title"`
	Artist        string   `json:"artist"`
	Date          string   `json:"date"` // Date de sortie (YYYY-MM-DD ou YYYY)
	Country       string   `json:"country"`
	Label         string   `json:"label"`
	CatalogNumber string   `json:"catalog_number"`
	Barcode       string   `json:"barcode"`
	Status        string   `json:"status"` // Official, Bootleg, etc.
	Genres        []string `json:"genres"`
	Tracks        []Track  `json:"tracks"`
	TotalTracks   int      `json:"total_tracks"`
	CoverArtURL   string   `json:"cover_art_url"`
}

// Track représente une piste d'un album
type Track struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	Duration int    `json:"duration"` // en millisecondes
}

// Year retourne l'année de sortie
func (a *Album) Year() string {
	if len(a.Date) >= 4 {
		return a.Date[:4]
	}
	return ""
}

// CoverURL retourne l'URL de la pochette via Cover Art Archive
func (a *Album) CoverURL() string {
	if a.CoverArtURL != "" {
		return a.CoverArtURL
	}
	if a.MBID != "" {
		return fmt.Sprintf("https://coverartarchive.org/release/%s/front-500", a.MBID)
	}
	return ""
}

// DurationFormatted retourne la durée d'une piste formatée (M:SS)
func (t *Track) DurationFormatted() string {
	if t.Duration <= 0 {
		return ""
	}
	totalSeconds := t.Duration / 1000
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// --- media.Metadata interface implementation ---

func (a *Album) MediaType() media.Type    { return media.TypeMusic }
func (a *Album) GetTitle() string         { return a.Title }
func (a *Album) GetOriginalTitle() string { return a.Title }
func (a *Album) GetYear() string          { return a.Year() }
