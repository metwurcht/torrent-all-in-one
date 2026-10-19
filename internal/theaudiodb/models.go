package theaudiodb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/metwurcht/torrent-all-in-one/internal/media"
)

// Album contains the album metadata used by the music pipeline.
type Album struct {
	ID            string
	Title         string
	Artist        string
	Date          string
	Country       string
	Label         string
	CatalogNumber string
	Barcode       string
	Status        string
	Genres        []string
	Tracks        []Track
	TotalTracks   int
	CoverArtURL   string
}

type Track struct {
	Number   int
	Title    string
	Duration int
}

func (a *Album) Year() string {
	if len(a.Date) >= 4 {
		return a.Date[:4]
	}
	return ""
}

func (a *Album) CoverURL() string { return a.CoverArtURL }

func (t *Track) DurationFormatted() string {
	if t.Duration <= 0 {
		return ""
	}
	return fmt.Sprintf("%d:%02d", t.Duration/60, t.Duration%60)
}

func (a *Album) MediaType() media.Type    { return media.TypeMusic }
func (a *Album) GetTitle() string         { return a.Title }
func (a *Album) GetOriginalTitle() string { return a.Title }
func (a *Album) GetYear() string          { return a.Year() }

func parseNumber(value string) int {
	number, _ := strconv.Atoi(strings.TrimSpace(value))
	return number
}

func parseDuration(value, seconds string) int {
	if parsed := parseNumber(seconds); parsed > 0 {
		return parsed
	}
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) == 2 {
		return parseNumber(parts[0])*60 + parseNumber(parts[1])
	}
	return parseNumber(value)
}
