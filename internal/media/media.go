// Package media defines the core types and interfaces for media processing.
// It provides the abstraction layer that allows the processor to work with
// different media types (movies, TV shows, music albums) through a common pipeline.
package media

import (
	"context"

	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
)

// Type represents the type of media being processed
type Type string

const (
	TypeMovie  Type = "movie"
	TypeTVShow Type = "tvshow"
	TypeMusic  Type = "music"
)

// SearchResult represents a generic search result for any media type.
// It contains the minimal information needed to display a list of results
// and let the user pick one.
type SearchResult struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Year          string  `json:"year"`
	Overview      string  `json:"overview"`
	PosterURL     string  `json:"poster_url"`
	VoteAverage   float64 `json:"vote_average"`
}

// Metadata is the common interface for all media metadata.
// Concrete types (e.g., tmdb.Movie) implement this interface.
// Type-specific generators can type-assert to the concrete type
// to access detailed fields.
type Metadata interface {
	// MediaType returns the type of media (movie, tvshow, music)
	MediaType() Type
	// GetTitle returns the localized title
	GetTitle() string
	// GetOriginalTitle returns the original title
	GetOriginalTitle() string
	// GetYear returns the release year as a string
	GetYear() string
}

// Provider handles searching and fetching detailed metadata for a media type.
// Each media type (movie, TV show, music) has its own Provider implementation.
type Provider interface {
	// Search searches for media by query string and returns generic search results
	Search(ctx context.Context, query string) ([]SearchResult, error)
	// GetDetails retrieves complete metadata by ID, returning the concrete Metadata type
	GetDetails(ctx context.Context, id int) (Metadata, error)
	// ExtractKeywords extracts search keywords from a filename
	ExtractKeywords(filename string) string
	// ParseDirectID parses a direct ID from user input (e.g., "id:12345")
	ParseDirectID(input string) (int, bool)
}

// Renamer generates release names for a specific media type.
type Renamer interface {
	GenerateName(metadata Metadata, info *mediainfo.MediaInfo) string
}

// DirectoryRenamer generates release names for directory-based media (TV shows).
// It provides both the directory name and individual file names.
type DirectoryRenamer interface {
	// GenerateDirectoryName generates the release directory name (e.g., Series.S01.1080p.BluRay.x265-GROUP)
	GenerateDirectoryName(metadata Metadata, info *mediainfo.MediaInfo) string
	// GenerateFileName generates a release file name for an individual episode
	GenerateFileName(metadata Metadata, info *mediainfo.MediaInfo, episodeNumber int) string
}

// NFOGenerator creates NFO file content for a specific media type.
type NFOGenerator interface {
	Generate(metadata Metadata, info *mediainfo.MediaInfo, fileName string) string
}

// DirectoryNFOGenerator creates NFO file content for directory-based media.
// It receives all mediainfo results for a comprehensive NFO.
type DirectoryNFOGenerator interface {
	GenerateDirectory(metadata Metadata, infos []*mediainfo.MediaInfo, dirName string) string
}

// Presenter creates presentation content (e.g., BBCode) for a specific media type.
type Presenter interface {
	GenerateBBCode(metadata Metadata, info *mediainfo.MediaInfo) string
}

// DirectoryPresenter creates presentation content for directory-based media.
type DirectoryPresenter interface {
	GenerateDirectoryBBCode(metadata Metadata, infos []*mediainfo.MediaInfo) string
}

// Pipeline groups all media-type-specific components needed to process a file.
// The processor uses a Pipeline to delegate type-specific operations while
// keeping the orchestration logic generic.
type Pipeline struct {
	Type         Type
	Provider     Provider
	Renamer      Renamer
	NFOGenerator NFOGenerator
	Presenter    Presenter

	// Directory-aware components for multi-file media (TV shows).
	// These are optional and only used by ProcessDirectory.
	DirectoryRenamer      DirectoryRenamer
	DirectoryNFOGenerator DirectoryNFOGenerator
	DirectoryPresenter    DirectoryPresenter
}
