// Package movie provides the media.Pipeline implementation for movie processing.
// It wraps the existing tmdb, renamer, nfo, and presenter packages as adapters
// that implement the media interfaces, allowing the processor to handle movies
// through the generic pipeline pattern.
package movie

import (
	"context"

	"github.com/metwurcht/torrent-all-in-one/internal/media"
	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/nfo"
	"github.com/metwurcht/torrent-all-in-one/internal/presenter"
	"github.com/metwurcht/torrent-all-in-one/internal/renamer"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
)

// NewPipeline creates a complete movie processing pipeline with all components.
func NewPipeline(groupName string) *media.Pipeline {
	return &media.Pipeline{
		Type:         media.TypeMovie,
		Provider:     NewProvider(),
		Renamer:      NewRenamer(groupName),
		NFOGenerator: NewNFOGenerator(groupName),
		Presenter:    NewPresenter(),
	}
}

// ---------------------------------------------------------------------------
// Provider — wraps tmdb.Client to implement media.Provider
// ---------------------------------------------------------------------------

// Provider wraps the TMDB client to implement media.Provider for movies.
type Provider struct {
	client *tmdb.Client
}

// NewProvider creates a new movie metadata provider backed by TMDB scraping.
func NewProvider() *Provider {
	return &Provider{client: tmdb.NewClient()}
}

// Search searches for movies via TMDB and returns generic search results.
func (p *Provider) Search(ctx context.Context, query string) ([]media.SearchResult, error) {
	movies, err := p.client.SearchMovie(ctx, query)
	if err != nil {
		return nil, err
	}

	results := make([]media.SearchResult, len(movies))
	for i, m := range movies {
		results[i] = media.SearchResult{
			ID:            m.ID,
			Title:         m.Title,
			OriginalTitle: m.OriginalTitle,
			Year:          m.Year(),
			Overview:      m.Overview,
			PosterURL:     m.PosterURL("w185"),
			VoteAverage:   m.VoteAverage,
		}
	}
	return results, nil
}

// GetDetails retrieves complete movie metadata by TMDB ID.
func (p *Provider) GetDetails(ctx context.Context, id int) (media.Metadata, error) {
	return p.client.GetMovieDetails(ctx, id)
}

// ExtractKeywords extracts search keywords from a movie filename.
func (p *Provider) ExtractKeywords(filename string) string {
	return tmdb.ExtractKeywords(filename)
}

// ParseDirectID parses a direct TMDB ID from user input.
func (p *Provider) ParseDirectID(input string) (int, bool) {
	return tmdb.ParseDirectID(input)
}

// ---------------------------------------------------------------------------
// Renamer — wraps renamer.Renamer to implement media.Renamer
// ---------------------------------------------------------------------------

// MovieRenamer wraps the existing renamer package for movie-specific naming.
type MovieRenamer struct {
	renamer *renamer.Renamer
}

// NewRenamer creates a new movie renamer.
func NewRenamer(groupName string) *MovieRenamer {
	return &MovieRenamer{renamer: renamer.NewRenamer(groupName)}
}

// GenerateName generates a release name for a movie.
func (r *MovieRenamer) GenerateName(metadata media.Metadata, info *mediainfo.MediaInfo) string {
	movie := metadata.(*tmdb.Movie)
	return r.renamer.GenerateName(movie, info)
}

// ---------------------------------------------------------------------------
// NFOGenerator — wraps nfo.Generator to implement media.NFOGenerator
// ---------------------------------------------------------------------------

// MovieNFOGenerator wraps the existing nfo package for movie-specific NFO generation.
type MovieNFOGenerator struct {
	generator *nfo.Generator
}

// NewNFOGenerator creates a new movie NFO generator.
func NewNFOGenerator(groupName string) *MovieNFOGenerator {
	return &MovieNFOGenerator{generator: nfo.NewGenerator(groupName)}
}

// Generate creates NFO content for a movie.
func (g *MovieNFOGenerator) Generate(metadata media.Metadata, info *mediainfo.MediaInfo, fileName string) string {
	movie := metadata.(*tmdb.Movie)
	return g.generator.Generate(movie, info, fileName)
}

// ---------------------------------------------------------------------------
// Presenter — wraps presenter.GenerateBBcode to implement media.Presenter
// ---------------------------------------------------------------------------

// MoviePresenter wraps the existing presenter package for movie-specific BBCode.
type MoviePresenter struct{}

// NewPresenter creates a new movie presenter.
func NewPresenter() *MoviePresenter {
	return &MoviePresenter{}
}

// GenerateBBCode creates BBCode presentation content for a movie.
func (p *MoviePresenter) GenerateBBCode(metadata media.Metadata, info *mediainfo.MediaInfo) string {
	movie := metadata.(*tmdb.Movie)
	return presenter.GenerateBBcode(movie, info)
}
