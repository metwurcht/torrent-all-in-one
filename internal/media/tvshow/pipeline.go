// Package tvshow provides the media.Pipeline implementation for TV show processing.
// It wraps the tmdb client for TV show search/details and provides directory-based
// renaming, NFO generation, and BBCode presentation for multi-file TV series.
package tvshow

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/metwurcht/torrent-all-in-one/internal/media"
	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
)

// NewPipeline creates a complete TV show processing pipeline with all components.
func NewPipeline(groupName string) *media.Pipeline {
	renamer := NewRenamer(groupName)
	nfoGen := NewNFOGenerator(groupName)
	presenter := NewPresenter()

	return &media.Pipeline{
		Type:     media.TypeTVShow,
		Provider: NewProvider(),
		// Single-file interfaces (not used for TV shows, but required by Pipeline)
		Renamer:      renamer,
		NFOGenerator: nfoGen,
		Presenter:    presenter,
		// Directory-aware interfaces
		DirectoryRenamer:      renamer,
		DirectoryNFOGenerator: nfoGen,
		DirectoryPresenter:    presenter,
	}
}

// ---------------------------------------------------------------------------
// Provider — wraps tmdb.Client for TV show search/details
// ---------------------------------------------------------------------------

// Provider wraps the TMDB client to implement media.Provider for TV shows.
type Provider struct {
	client *tmdb.Client
}

// NewProvider creates a new TV show metadata provider backed by TMDB scraping.
func NewProvider() *Provider {
	return &Provider{client: tmdb.NewClient()}
}

// Search searches for TV shows via TMDB and returns generic search results.
func (p *Provider) Search(ctx context.Context, query string) ([]media.SearchResult, error) {
	shows, err := p.client.SearchTVShow(ctx, query)
	if err != nil {
		return nil, err
	}

	results := make([]media.SearchResult, len(shows))
	for i, s := range shows {
		results[i] = media.SearchResult{
			ID:            s.ID,
			Title:         s.Name,
			OriginalTitle: s.OriginalName,
			Year:          s.Year(),
			Overview:      s.Overview,
			PosterURL:     s.PosterURL("w185"),
			VoteAverage:   s.VoteAverage,
		}
	}
	return results, nil
}

// GetDetails retrieves complete TV show metadata by TMDB ID.
func (p *Provider) GetDetails(ctx context.Context, id int) (media.Metadata, error) {
	return p.client.GetTVShowDetails(ctx, id)
}

// ExtractKeywords extracts search keywords from a directory name.
func (p *Provider) ExtractKeywords(filename string) string {
	return tmdb.ExtractKeywords(filename)
}

// ParseDirectID parses a direct TMDB ID from user input.
func (p *Provider) ParseDirectID(input string) (int, bool) {
	return tmdb.ParseDirectID(input)
}

// ---------------------------------------------------------------------------
// TVShowRenamer — implements both media.Renamer and media.DirectoryRenamer
// ---------------------------------------------------------------------------

// TVShowRenamer generates release names for TV shows.
type TVShowRenamer struct {
	groupName string
}

// NewRenamer creates a new TV show renamer.
func NewRenamer(groupName string) *TVShowRenamer {
	return &TVShowRenamer{groupName: groupName}
}

// GenerateName generates a release name for a single episode (fallback).
func (r *TVShowRenamer) GenerateName(metadata media.Metadata, info *mediainfo.MediaInfo) string {
	return r.GenerateDirectoryName(metadata, info)
}

// GenerateDirectoryName generates the release directory name.
// Format: Series.Name.S01.1080p.BluRay.x265-GROUP or Series.Name.INTEGRALE.1080p.BluRay.x265-GROUP
func (r *TVShowRenamer) GenerateDirectoryName(metadata media.Metadata, info *mediainfo.MediaInfo) string {
	show := metadata.(*tmdb.TVShow)
	parts := []string{}

	// Titre nettoyé
	title := cleanTitle(show.OriginalName)
	parts = append(parts, title)

	// Tag de saison/intégrale
	parts = append(parts, show.SeasonTag())

	// Langue(s)
	if langs := detectLanguages(info); langs != "" {
		parts = append(parts, langs)
	}

	// Résolution
	if info.Video.Resolution != "" {
		parts = append(parts, info.Video.Resolution)
	}

	// Source
	if info.SourceType != "" {
		parts = append(parts, string(info.SourceType))
	}

	// HDR
	if info.Video.HDR != "" {
		parts = append(parts, info.Video.HDR)
	}

	// Codec vidéo
	if codec := info.Video.VideoCodecTag(); codec != "" {
		parts = append(parts, codec)
	}

	// 10-bit
	if info.Video.BitDepth == 10 {
		parts = append(parts, "10bit")
	}

	// Codec audio (premier track)
	if len(info.Audio) > 0 {
		audioTag := info.Audio[0].AudioCodecTag()
		channelLayout := info.Audio[0].ChannelLayoutShort()
		parts = append(parts, fmt.Sprintf("%s.%s", audioTag, channelLayout))
	}

	releaseName := strings.Join(parts, ".")
	return fmt.Sprintf("%s-%s", releaseName, r.groupName)
}

// GenerateFileName generates a release file name for an individual episode.
// Format: Series.Name.S01E01.MULTI.VFF.1080p.BluRay.HDR.x265.10bit.EAC3.5.1-GROUP
func (r *TVShowRenamer) GenerateFileName(metadata media.Metadata, info *mediainfo.MediaInfo, episodeNumber int) string {
	show := metadata.(*tmdb.TVShow)
	parts := []string{}

	title := cleanTitle(show.OriginalName)
	parts = append(parts, title)

	// Épisode tag: S01E01
	if show.IsCompleteSeries {
		parts = append(parts, fmt.Sprintf("E%02d", episodeNumber))
	} else {
		parts = append(parts, fmt.Sprintf("S%02dE%02d", show.Season, episodeNumber))
	}

	// Langue(s)
	if langs := detectLanguages(info); langs != "" {
		parts = append(parts, langs)
	}

	// Résolution
	if info.Video.Resolution != "" {
		parts = append(parts, info.Video.Resolution)
	}

	// Source
	if info.SourceType != "" {
		parts = append(parts, string(info.SourceType))
	}

	// HDR
	if info.Video.HDR != "" {
		parts = append(parts, info.Video.HDR)
	}

	// Codec vidéo
	if codec := info.Video.VideoCodecTag(); codec != "" {
		parts = append(parts, codec)
	}

	// 10-bit
	if info.Video.BitDepth == 10 {
		parts = append(parts, "10bit")
	}

	// Codec audio (premier track)
	if len(info.Audio) > 0 {
		audioTag := info.Audio[0].AudioCodecTag()
		channelLayout := info.Audio[0].ChannelLayoutShort()
		parts = append(parts, fmt.Sprintf("%s.%s", audioTag, channelLayout))
	}

	releaseName := strings.Join(parts, ".")
	return fmt.Sprintf("%s-%s", releaseName, r.groupName)
}

// ---------------------------------------------------------------------------
// TVShowNFOGenerator — implements both media.NFOGenerator and media.DirectoryNFOGenerator
// ---------------------------------------------------------------------------

const nfoWidth = 120

// TVShowNFOGenerator generates NFO content for TV shows.
type TVShowNFOGenerator struct {
	groupName string
}

// NewNFOGenerator creates a new TV show NFO generator.
func NewNFOGenerator(groupName string) *TVShowNFOGenerator {
	return &TVShowNFOGenerator{groupName: groupName}
}

// Generate generates NFO content for a single file (fallback).
func (g *TVShowNFOGenerator) Generate(metadata media.Metadata, info *mediainfo.MediaInfo, fileName string) string {
	return g.GenerateDirectory(metadata, []*mediainfo.MediaInfo{info}, fileName)
}

// GenerateDirectory generates NFO content for a TV show directory.
func (g *TVShowNFOGenerator) GenerateDirectory(metadata media.Metadata, infos []*mediainfo.MediaInfo, dirName string) string {
	show := metadata.(*tmdb.TVShow)
	var sb strings.Builder

	border := strings.Repeat("=", nfoWidth)
	thinBorder := strings.Repeat("-", nfoWidth)

	// Header
	sb.WriteString(border + "\n")
	sb.WriteString(centerText(g.groupName+" presents", nfoWidth) + "\n")
	sb.WriteString(border + "\n")
	sb.WriteString(centerText(show.Name, nfoWidth) + "\n")
	if show.OriginalName != "" && show.OriginalName != show.Name {
		sb.WriteString(centerText(fmt.Sprintf("(%s)", show.OriginalName), nfoWidth) + "\n")
	}
	sb.WriteString(thinBorder + "\n")
	sb.WriteString(fmt.Sprintf("Release Name: %s\n", dirName))

	if show.FirstAirDate != "" {
		sb.WriteString(fmt.Sprintf("First Air Date: %s\n", show.FirstAirDate))
	}
	if len(show.Genres) > 0 {
		sb.WriteString(fmt.Sprintf("Genre: %s\n", strings.Join(show.Genres, ", ")))
	}
	if show.NumberOfSeasons > 0 {
		sb.WriteString(fmt.Sprintf("Seasons: %d\n", show.NumberOfSeasons))
	}
	if show.NumberOfEpisodes > 0 {
		sb.WriteString(fmt.Sprintf("Episodes: %d\n", show.NumberOfEpisodes))
	}
	if show.VoteAverage > 0 {
		sb.WriteString(fmt.Sprintf("Rating: %.1f/10\n", show.VoteAverage))
	}
	if show.IMDbID != "" {
		sb.WriteString(fmt.Sprintf("IMDb: %s\n", show.IMDbURL()))
	}
	sb.WriteString(fmt.Sprintf("TMDB: %s\n", show.TMDbURL()))

	if len(show.Creators) > 0 {
		sb.WriteString(fmt.Sprintf("Creators: %s\n", strings.Join(show.Creators, ", ")))
	}

	if len(show.Cast) > 0 {
		actors := make([]string, 0, 5)
		for i, c := range show.Cast {
			if i >= 5 {
				break
			}
			actors = append(actors, c.Name)
		}
		sb.WriteString(fmt.Sprintf("Cast: %s\n", strings.Join(actors, ", ")))
	}

	sb.WriteString(border + "\n")
	sb.WriteString(centerText("SYNOPSIS", nfoWidth) + "\n")
	sb.WriteString(thinBorder + "\n\n")
	if show.Overview != "" {
		sb.WriteString(wrapText(show.Overview, nfoWidth))
	}
	sb.WriteString("\n")
	sb.WriteString(border + "\n")
	sb.WriteString(centerText("MEDIA INFORMATION", nfoWidth) + "\n")
	sb.WriteString(thinBorder + "\n")

	// MediaInfo output for each file
	for i, info := range infos {
		if i > 0 {
			sb.WriteString("\n" + thinBorder + "\n")
		}
		mediaInfoOutput, err := getMediaInfoOutput(info.FilePath)
		if err != nil {
			sb.WriteString(fmt.Sprintf("File %d: %s\n", i+1, info.FileName))
			sb.WriteString(fmt.Sprintf("Error getting MediaInfo: %v\n", err))
		} else {
			sb.WriteString(mediaInfoOutput)
		}
	}

	sb.WriteString("\n" + border + "\n")
	sb.WriteString(centerText("Generated by Torrent-AIO", nfoWidth) + "\n")
	sb.WriteString(centerText(time.Now().Format("2006-01-02 15:04:05"), nfoWidth) + "\n")
	sb.WriteString(border + "\n")

	return sb.String()
}

// ---------------------------------------------------------------------------
// TVShowPresenter — implements both media.Presenter and media.DirectoryPresenter
// ---------------------------------------------------------------------------

// TVShowPresenter generates BBCode presentation for TV shows.
type TVShowPresenter struct{}

// NewPresenter creates a new TV show presenter.
func NewPresenter() *TVShowPresenter {
	return &TVShowPresenter{}
}

// GenerateBBCode generates BBCode for a single file (fallback).
func (p *TVShowPresenter) GenerateBBCode(metadata media.Metadata, info *mediainfo.MediaInfo) string {
	return p.GenerateDirectoryBBCode(metadata, []*mediainfo.MediaInfo{info})
}

// GenerateDirectoryBBCode generates BBCode presentation for a TV show directory.
// Uses average bitrates and audio/video info from the first file.
func (p *TVShowPresenter) GenerateDirectoryBBCode(metadata media.Metadata, infos []*mediainfo.MediaInfo) string {
	show := metadata.(*tmdb.TVShow)
	if len(infos) == 0 {
		return ""
	}

	// Utiliser le premier fichier comme référence pour audio/video
	ref := infos[0]

	// Calculer les bitrates moyens
	avgOverallBitrate := averageBitrate(infos, func(mi *mediainfo.MediaInfo) int { return mi.OverallBitrate })
	avgVideoBitrate := averageBitrate(infos, func(mi *mediainfo.MediaInfo) int { return mi.Video.Bitrate })

	var sb strings.Builder
	sb.WriteString("[center]")

	// Titre
	sb.WriteString(fmt.Sprintf("[font=Verdana][size=200][color=#aa0000][b]%s[/b][/color][/size][/font]\n", show.Name))
	if show.Year() != "" {
		sb.WriteString(fmt.Sprintf("[font=Verdana][size=150][color=#aa0000](%s)[/color][/size][/font]\n", show.Year()))
	}
	sb.WriteString("\n\n")

	// Poster
	if show.PosterPath != "" {
		sb.WriteString(fmt.Sprintf("[img]%s[/img]\n\n", show.PosterURL("w500")))
	}

	// Tagline
	if show.Tagline != "" {
		sb.WriteString(fmt.Sprintf("[font=Verdana][size=100][color=#aa0000][i]« %s »[/i][/color][/size][/font]\n", show.Tagline))
		sb.WriteString(" \n \n")
	}

	// Section Informations
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Informations[/b][/size][/color][/font]\n \n[font=Verdana]")

	if show.OriginalName != "" && show.OriginalName != show.Name {
		sb.WriteString(fmt.Sprintf("[b]Titre original :[/b] %s\n", show.OriginalName))
	}
	if show.FirstAirDate != "" {
		sb.WriteString(fmt.Sprintf("[b]Première diffusion :[/b] %s\n", show.FirstAirDate))
	}
	if show.NumberOfSeasons > 0 {
		sb.WriteString(fmt.Sprintf("[b]Saisons :[/b] %d\n", show.NumberOfSeasons))
	}
	if show.NumberOfEpisodes > 0 {
		sb.WriteString(fmt.Sprintf("[b]Épisodes :[/b] %d\n", show.NumberOfEpisodes))
	}
	if show.Status != "" {
		sb.WriteString(fmt.Sprintf("[b]Statut :[/b] %s\n", show.Status))
	}
	sb.WriteString(fmt.Sprintf("[b]Épisodes dans cette release :[/b] %d\n", len(infos)))
	sb.WriteString(" \n")

	// Créateurs
	if len(show.Creators) > 0 {
		sb.WriteString(fmt.Sprintf("[b]Créateur(s) :[/b] %s\n \n", strings.Join(show.Creators, ", ")))
	}

	// Acteurs
	if len(show.Cast) > 0 {
		sb.WriteString("[b]Acteurs :[/b]\n")
		for i, actor := range show.Cast {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("%s, ", actor.Name))
		}
		sb.WriteString("\n \n")
	}

	// Genres
	if len(show.Genres) > 0 {
		sb.WriteString(fmt.Sprintf("[b]Genres :[/b]\n%s\n \n", strings.Join(show.Genres, ", ")))
	}

	// Networks
	if len(show.Networks) > 0 {
		sb.WriteString(fmt.Sprintf("[b]Réseau(x) :[/b] %s\n \n", strings.Join(show.Networks, ", ")))
	}

	// Note
	if show.VoteAverage > 0 {
		sb.WriteString(fmt.Sprintf("[img]https://zupimages.net/up/21/02/xro7.png[/img] %.2f\n \n", show.VoteAverage))
	}

	// Liens
	sb.WriteString(fmt.Sprintf("[img]https://zupimages.net/up/21/03/mxao.png[/img] [url=%s]Fiche de la série[/url]\n", show.TMDbURL()))
	if show.IMDbID != "" {
		sb.WriteString(fmt.Sprintf("[img]https://zupimages.net/up/21/03/od5a.png[/img] [url=%s]%s[/url]\n", show.IMDbURL(), show.IMDbID))
	}
	sb.WriteString("[/font]\n \n")

	// Synopsis
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Synopsis[/b][/size][/color][/font]\n \n[font=Verdana]\n")
	if show.Overview != "" {
		sb.WriteString(show.Overview)
	}
	sb.WriteString("\n \n \n[/font]\n")

	// Images du casting
	if len(show.Cast) >= 2 {
		for i := 0; i < 2 && i < len(show.Cast); i++ {
			if show.Cast[i].ProfilePath != "" {
				sb.WriteString(fmt.Sprintf(" [img]https://image.tmdb.org/t/p/w138_and_h175_face%s[/img] ", show.Cast[i].ProfilePath))
			}
		}
		sb.WriteString("\n \n \n")
	}

	// Détails techniques (basés sur le premier fichier + bitrates moyens)
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Détails techniques[/b][/size][/color][/font]\n \n[font=Verdana]")

	sb.WriteString(fmt.Sprintf("[b]Format :[/b] %s\n", strings.ToUpper(ref.Container)))
	sb.WriteString(fmt.Sprintf("[b]Codec Vidéo :[/b] %s", ref.Video.VideoCodecTag()))
	if ref.Video.BitDepth == 10 {
		sb.WriteString(" 10-bit")
	}
	sb.WriteString("\n")

	if avgVideoBitrate > 0 {
		sb.WriteString(fmt.Sprintf("[b]Débit Vidéo (moyen) :[/b] ~%d kb/s\n", avgVideoBitrate/1000))
	}
	sb.WriteString(fmt.Sprintf("[b]Résolution :[/b] %s\n \n", ref.Video.Resolution))

	// Pistes audio
	if len(ref.Audio) > 0 {
		sb.WriteString("[b]Langue(s) :[/b]\n")
		for _, audio := range ref.Audio {
			flag := fmt.Sprintf("[img]%s[/img]", audio.Language.FlagURL())
			langName := audio.Language.ShortCode()
			description := fmt.Sprintf("%s %s [%s] | %s", flag, langName, audio.ChannelLayoutShort(), audio.AudioCodecTag())
			if audio.Bitrate > 0 {
				description += fmt.Sprintf(" à %d kb/s", audio.Bitrate/1000)
			}
			if audio.AudioDescription {
				description += " (Audiodescription)"
			}
			sb.WriteString(description + "\n")
		}
		sb.WriteString("\n \n")
	}

	// Sous-titres
	if len(ref.Subtitles) > 0 {
		sb.WriteString("[b]Sous-titres :[/b]\n")
		for _, sub := range ref.Subtitles {
			flag := fmt.Sprintf("[img]%s[/img]", sub.Language.FlagURL())
			langName := sub.Language.ShortCode()
			subType := "full"
			if sub.Forced {
				subType = "forced"
			}
			sb.WriteString(fmt.Sprintf("%s %s | %s (%s)\n", flag, langName, sub.Format, subType))
		}
		sb.WriteString("\n \n")
	}

	// Débit global moyen
	if avgOverallBitrate > 0 {
		sb.WriteString(fmt.Sprintf("[b]Débit Global (moyen) :[/b] ~%d kb/s\n", avgOverallBitrate/1000))
	}

	// Taille totale
	var totalSize int64
	for _, info := range infos {
		totalSize += info.FileSize
	}
	sb.WriteString(fmt.Sprintf("[b]Nombre de fichiers :[/b] %d\n", len(infos)))
	sb.WriteString(fmt.Sprintf("[b]Poids Total :[/b] %s", formatFileSize(totalSize)))

	sb.WriteString("[/font]\n \n")
	sb.WriteString("[/center] \n")

	return sb.String()
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func cleanTitle(title string) string {
	replacements := map[string]string{
		" ": ".", ":": "", "'": "", "\"": "", "/": ".", "\\": ".",
		"?": "", "!": "", "*": "", "<": "", ">": "", "|": "",
		"&": "and", "#": "", "%": "", "@": "at", "(": "", ")": "",
		"[": "", "]": "", "{": "", "}": "", ",": "", ";": "", "-": "",
	}
	result := title
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	// Supprimer les points multiples
	for strings.Contains(result, "..") {
		result = strings.ReplaceAll(result, "..", ".")
	}
	return strings.Trim(result, ".")
}

func detectLanguages(info *mediainfo.MediaInfo) string {
	if len(info.Audio) == 0 {
		return ""
	}
	uniqueLangs := make(map[mediainfo.Language]bool)
	hasFrench := false
	hasQuebecois := false
	audioDescriptionSuffix := ""

	for _, audio := range info.Audio {
		if audio.AudioDescription {
			audioDescriptionSuffix = ".AD"
		}
		uniqueLangs[audio.Language] = true
		if audio.Language == mediainfo.LanguageFrench {
			hasFrench = true
		}
		if audio.Language == mediainfo.LanguageQuebecois {
			hasQuebecois = true
		}
	}

	isMulti := len(uniqueLangs) >= 2
	var suffix string
	if hasFrench && hasQuebecois {
		suffix = "VF2"
	} else if hasFrench {
		suffix = "VFF"
	} else if hasQuebecois {
		suffix = "VFQ"
	}
	suffix += audioDescriptionSuffix

	if isMulti && suffix != "" {
		return "MULTI." + suffix
	} else if suffix != "" {
		return suffix
	}

	if len(uniqueLangs) > 0 {
		langs := []string{}
		for lang := range uniqueLangs {
			if lang != mediainfo.LanguageUnknown {
				langs = append(langs, string(lang))
			}
		}
		if len(langs) > 0 {
			return strings.Join(langs, ".")
		}
	}
	return ""
}

func averageBitrate(infos []*mediainfo.MediaInfo, fn func(*mediainfo.MediaInfo) int) int {
	if len(infos) == 0 {
		return 0
	}
	total := 0
	count := 0
	for _, info := range infos {
		val := fn(info)
		if val > 0 {
			total += val
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f Go", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f Mo", float64(size)/float64(MB))
	default:
		return fmt.Sprintf("%.2f Ko", float64(size)/float64(KB))
	}
}

func getMediaInfoOutput(filePath string) (string, error) {
	cmd := exec.Command("mediainfo", filePath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run mediainfo: %w", err)
	}
	result := string(output)
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if strings.Contains(line, "Complete name") && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fileName := strings.TrimSpace(parts[1])
				if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
					fileName = fileName[strings.LastIndex(fileName, "/")+1:]
				}
				lines[i] = parts[0] + ": " + fileName
			}
		}
	}
	return strings.Join(lines, "\n"), nil
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text
}

func wrapText(text string, width int) string {
	words := strings.Fields(text)
	var lines []string
	var cur strings.Builder
	for _, w := range words {
		if cur.Len()+len(w)+1 > width {
			if cur.Len() > 0 {
				lines = append(lines, cur.String())
				cur.Reset()
			}
		}
		if cur.Len() > 0 {
			cur.WriteString(" ")
		}
		cur.WriteString(w)
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return strings.Join(lines, "\n") + "\n"
}
