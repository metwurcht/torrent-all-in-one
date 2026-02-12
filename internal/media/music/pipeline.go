// Package music provides the media.Pipeline implementation for music album processing.
// It wraps the MusicBrainz client for album search/details and provides directory-based
// renaming, NFO generation, and BBCode presentation for music albums.
package music

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/metwurcht/torrent-all-in-one/internal/media"
	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/musicbrainz"
)

// NewPipeline creates a complete music album processing pipeline with all components.
func NewPipeline(groupName string) *media.Pipeline {
	renamer := NewRenamer(groupName)
	nfoGen := NewNFOGenerator(groupName)
	presenter := NewPresenter()

	return &media.Pipeline{
		Type:     media.TypeMusic,
		Provider: NewProvider(),
		// Single-file interfaces (fallback)
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
// Provider — wraps musicbrainz.Client, maps int index ↔ MBID
// ---------------------------------------------------------------------------

// Provider wraps the MusicBrainz client to implement media.Provider.
// Since media.Provider uses int IDs but MusicBrainz uses string MBIDs,
// we maintain an internal mapping from sequential int indices to MBIDs.
type Provider struct {
	client   *musicbrainz.Client
	mbidMap  map[int]string // int index → MBID
	albumMap map[int]*musicbrainz.Album
}

// NewProvider creates a new music metadata provider backed by MusicBrainz.
func NewProvider() *Provider {
	return &Provider{
		client:   musicbrainz.NewClient(),
		mbidMap:  make(map[int]string),
		albumMap: make(map[int]*musicbrainz.Album),
	}
}

// Search searches for albums via MusicBrainz and returns generic search results.
func (p *Provider) Search(ctx context.Context, query string) ([]media.SearchResult, error) {
	albums, err := p.client.SearchAlbum(ctx, query)
	if err != nil {
		return nil, err
	}

	// Reset the MBID mapping for each new search
	p.mbidMap = make(map[int]string)
	p.albumMap = make(map[int]*musicbrainz.Album)

	results := make([]media.SearchResult, len(albums))
	for i, a := range albums {
		id := i + 1 // 1-based indices
		p.mbidMap[id] = a.MBID
		albumCopy := a
		p.albumMap[id] = &albumCopy

		// Formater le titre d'affichage : "Artist - Album"
		displayTitle := a.Title
		if a.Artist != "" {
			displayTitle = fmt.Sprintf("%s - %s", a.Artist, a.Title)
		}

		overview := ""
		if a.Label != "" {
			overview = fmt.Sprintf("Label: %s", a.Label)
		}
		if a.Country != "" {
			if overview != "" {
				overview += " | "
			}
			overview += fmt.Sprintf("Pays: %s", a.Country)
		}
		if a.Status != "" {
			if overview != "" {
				overview += " | "
			}
			overview += a.Status
		}

		results[i] = media.SearchResult{
			ID:            id,
			Title:         displayTitle,
			OriginalTitle: a.Title,
			Year:          a.Year(),
			Overview:      overview,
		}
	}
	return results, nil
}

// GetDetails retrieves complete album metadata by mapped int ID.
func (p *Provider) GetDetails(ctx context.Context, id int) (media.Metadata, error) {
	mbid, ok := p.mbidMap[id]
	if !ok {
		return nil, fmt.Errorf("ID %d non trouvé dans le cache de recherche", id)
	}

	album, err := p.client.GetAlbumDetails(ctx, mbid)
	if err != nil {
		return nil, err
	}

	return album, nil
}

// ExtractKeywords extracts search keywords from a directory name.
func (p *Provider) ExtractKeywords(filename string) string {
	return musicbrainz.ExtractKeywords(filename)
}

// ParseDirectID parses a direct MBID from user input (e.g., "id:xxx" or "mbid:xxx").
// If a valid MBID is found, it's stored in the map with a synthetic int key.
func (p *Provider) ParseDirectID(input string) (int, bool) {
	mbid, ok := musicbrainz.ParseDirectMBID(input)
	if !ok {
		return 0, false
	}
	// Use a large synthetic ID to avoid collisions with search indices
	syntheticID := 99999
	p.mbidMap[syntheticID] = mbid
	return syntheticID, true
}

// ---------------------------------------------------------------------------
// MusicRenamer — implements both media.Renamer and media.DirectoryRenamer
// ---------------------------------------------------------------------------

// MusicRenamer generates release names for music albums.
type MusicRenamer struct {
	groupName string
}

// NewRenamer creates a new music renamer.
func NewRenamer(groupName string) *MusicRenamer {
	return &MusicRenamer{groupName: groupName}
}

// GenerateName generates a release name (fallback for single file).
func (r *MusicRenamer) GenerateName(metadata media.Metadata, info *mediainfo.MediaInfo) string {
	return r.GenerateDirectoryName(metadata, info)
}

// GenerateDirectoryName generates the release directory name.
// Format: {Artist} - {Album} - {Year} [{Format} {Bitrate}]-GROUP
func (r *MusicRenamer) GenerateDirectoryName(metadata media.Metadata, info *mediainfo.MediaInfo) string {
	album := metadata.(*musicbrainz.Album)

	artist := sanitizeName(album.Artist)
	title := sanitizeName(album.Title)
	year := album.Year()

	// Déterminer le format audio et le bitrate
	format := detectAudioFormat(info)
	bitrate := detectAudioBitrate(info)

	var techTag string
	if format != "" && bitrate != "" {
		techTag = fmt.Sprintf("%s %s", format, bitrate)
	} else if format != "" {
		techTag = format
	}

	var parts []string
	parts = append(parts, artist)
	parts = append(parts, title)
	if year != "" {
		parts = append(parts, year)
	}

	name := strings.Join(parts, " - ")
	if techTag != "" {
		name = fmt.Sprintf("%s [%s]", name, techTag)
	}

	return fmt.Sprintf("%s-%s", name, r.groupName)
}

// GenerateFileName — music tracks are NOT renamed, so this returns the original name.
func (r *MusicRenamer) GenerateFileName(metadata media.Metadata, info *mediainfo.MediaInfo, trackNumber int) string {
	// Les pistes ne sont pas renommées
	if info != nil && info.FileName != "" {
		ext := ""
		if idx := strings.LastIndex(info.FileName, "."); idx >= 0 {
			ext = info.FileName[idx:]
			return info.FileName[:idx]
		}
		_ = ext
		return info.FileName
	}
	return fmt.Sprintf("track_%02d", trackNumber)
}

// ---------------------------------------------------------------------------
// MusicNFOGenerator — implements both media.NFOGenerator and media.DirectoryNFOGenerator
// ---------------------------------------------------------------------------

const nfoWidth = 81

// MusicNFOGenerator generates NFO content for music albums.
type MusicNFOGenerator struct {
	groupName string
}

// NewNFOGenerator creates a new music NFO generator.
func NewNFOGenerator(groupName string) *MusicNFOGenerator {
	return &MusicNFOGenerator{groupName: groupName}
}

// Generate generates NFO content for a single file (fallback).
func (g *MusicNFOGenerator) Generate(metadata media.Metadata, info *mediainfo.MediaInfo, fileName string) string {
	return g.GenerateDirectory(metadata, []*mediainfo.MediaInfo{info}, fileName)
}

// GenerateDirectory generates NFO content for a music album directory.
func (g *MusicNFOGenerator) GenerateDirectory(metadata media.Metadata, infos []*mediainfo.MediaInfo, dirName string) string {
	album := metadata.(*musicbrainz.Album)
	var sb strings.Builder

	border := strings.Repeat("=", nfoWidth)
	thinBorder := strings.Repeat("-", nfoWidth)

	// Header
	sb.WriteString(border + "\n")
	sb.WriteString(centerText(g.groupName+" presents", nfoWidth) + "\n")
	sb.WriteString(border + "\n\n")

	// Album info with dotted fields
	sb.WriteString(dottedField("Artiste", album.Artist) + "\n")
	sb.WriteString(dottedField("Album", album.Title) + "\n")
	if len(album.Genres) > 0 {
		sb.WriteString(dottedField("Genre(s)", strings.Join(album.Genres, ", ")) + "\n")
	}
	if album.Date != "" {
		sb.WriteString(dottedField("Year", album.Date) + "\n")
	}
	sb.WriteString("\n")

	// Technical info from first file via mediainfo --Inform
	if len(infos) > 0 {
		fields := extractAudioNFOFields(infos[0].FilePath)
		if fields.Codec != "" {
			sb.WriteString(dottedField("Codec", fields.Codec) + "\n")
		}
		if fields.Format != "" {
			sb.WriteString(dottedField("Format", fields.Format) + "\n")
		}
		if fields.OverallBitRate != "" {
			sb.WriteString(dottedField("Overall bit rate", fields.OverallBitRate) + "\n")
		}
		if fields.BitRateMode != "" {
			sb.WriteString(dottedField("Bit rate mode", fields.BitRateMode) + "\n")
		}
		if fields.Channel != "" {
			sb.WriteString(dottedField("Channel", fields.Channel) + "\n")
		}
		if fields.Quality != "" {
			sb.WriteString(dottedField("Quality", fields.Quality) + "\n")
		}
		if fields.WritingLibrary != "" {
			sb.WriteString(dottedField("Writing library", fields.WritingLibrary) + "\n")
		}
		encodingSettings := fields.EncodingSettings
		if encodingSettings == "" {
			encodingSettings = "undefined"
		}
		sb.WriteString(dottedField("Encoding settings", encodingSettings) + "\n")
	}

	sb.WriteString("\n\n")

	// Tracklist
	sb.WriteString(thinBorder + "\n")
	sb.WriteString(centerText("Tracklist", nfoWidth) + "\n")
	sb.WriteString(thinBorder + "\n\n")

	var totalDuration int
	var totalSize int64

	for i, mi := range infos {
		trackNum := i + 1
		trackTitle := resolveTrackTitle(mi.FileName, album, i)

		totalDuration += mi.Duration
		totalSize += mi.FileSize

		left := fmt.Sprintf("%02d. %s - %s", trackNum, album.Artist, trackTitle)
		sizeStr := formatSizeMiB(mi.FileSize)
		durStr := formatDurationMMSS(mi.Duration)
		right := fmt.Sprintf("[%s] [%s]", sizeStr, durStr)

		padding := nfoWidth - len(left) - len(right)
		if padding < 1 {
			padding = 1
		}
		sb.WriteString(left + strings.Repeat(" ", padding) + right + "\n")
	}

	sb.WriteString("\n\n")
	sb.WriteString(dottedField("Playing Time", formatDurationMMSS(totalDuration)) + "\n")
	sb.WriteString(dottedField("Total Size", formatTotalSizeMO(totalSize)) + "\n\n")

	// Footer
	sb.WriteString(border + "\n")
	sb.WriteString(centerText("Generated by Torrent-AIO", nfoWidth) + "\n")
	sb.WriteString(centerText(time.Now().Format("2006-01-02 15:04:05"), nfoWidth) + "\n")
	sb.WriteString(border + "\n")

	return sb.String()
}

// ---------------------------------------------------------------------------
// MusicPresenter — implements both media.Presenter and media.DirectoryPresenter
// ---------------------------------------------------------------------------

// MusicPresenter generates BBCode presentation for music albums.
type MusicPresenter struct{}

// NewPresenter creates a new music presenter.
func NewPresenter() *MusicPresenter {
	return &MusicPresenter{}
}

// GenerateBBCode generates BBCode for a single file (fallback).
func (p *MusicPresenter) GenerateBBCode(metadata media.Metadata, info *mediainfo.MediaInfo) string {
	return p.GenerateDirectoryBBCode(metadata, []*mediainfo.MediaInfo{info})
}

// GenerateDirectoryBBCode generates BBCode presentation for a music album.
func (p *MusicPresenter) GenerateDirectoryBBCode(metadata media.Metadata, infos []*mediainfo.MediaInfo) string {
	album := metadata.(*musicbrainz.Album)
	if len(infos) == 0 {
		return ""
	}

	ref := infos[0]
	var sb strings.Builder

	sb.WriteString("[center]")

	// Titre de l'album
	sb.WriteString(fmt.Sprintf("[font=Verdana][size=200][color=#aa0000][b]%s[/b][/color][/size][/font]\n", album.Title))
	sb.WriteString(fmt.Sprintf("[font=Verdana][size=150][color=#aa0000]%s[/color][/size][/font]\n", album.Artist))
	if album.Year() != "" {
		sb.WriteString(fmt.Sprintf("[font=Verdana][size=120][color=#aa0000](%s)[/color][/size][/font]\n", album.Year()))
	}
	sb.WriteString("\n\n")

	// Pochette de l'album
	coverURL := album.CoverURL()
	if coverURL != "" {
		sb.WriteString(fmt.Sprintf("[img]%s[/img]\n\n", coverURL))
	}

	// Section Informations
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Informations[/b][/size][/color][/font]\n \n[font=Verdana]")

	sb.WriteString(fmt.Sprintf("[b]Artiste :[/b] %s\n", album.Artist))
	sb.WriteString(fmt.Sprintf("[b]Album :[/b] %s\n", album.Title))

	if album.Year() != "" {
		sb.WriteString(fmt.Sprintf("[b]Année :[/b] %s\n", album.Year()))
	}
	if album.Label != "" {
		sb.WriteString(fmt.Sprintf("[b]Label :[/b] %s\n", album.Label))
	}
	if album.CatalogNumber != "" {
		sb.WriteString(fmt.Sprintf("[b]Cat. No :[/b] %s\n", album.CatalogNumber))
	}
	if album.Country != "" {
		sb.WriteString(fmt.Sprintf("[b]Pays :[/b] %s\n", album.Country))
	}
	if len(album.Genres) > 0 {
		sb.WriteString(fmt.Sprintf("[b]Genre(s) :[/b] %s\n", strings.Join(album.Genres, ", ")))
	}

	sb.WriteString(fmt.Sprintf("[b]MusicBrainz :[/b] [url=https://musicbrainz.org/release/%s]Fiche de l'album[/url]\n", album.MBID))
	sb.WriteString("[/font]\n \n")

	// Tracklist
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Tracklist[/b][/size][/color][/font]\n \n[font=Verdana]")

	for _, track := range album.Tracks {
		dur := track.DurationFormatted()
		if dur != "" {
			sb.WriteString(fmt.Sprintf("[b]%02d.[/b] %s [i](%s)[/i]\n", track.Number, track.Title, dur))
		} else {
			sb.WriteString(fmt.Sprintf("[b]%02d.[/b] %s\n", track.Number, track.Title))
		}
	}

	// Durée totale
	totalDuration := 0
	for _, track := range album.Tracks {
		totalDuration += track.Duration
	}
	if totalDuration > 0 {
		totalSec := totalDuration / 1000
		minutes := totalSec / 60
		seconds := totalSec % 60
		sb.WriteString(fmt.Sprintf("\n[b]Durée totale :[/b] %d:%02d\n", minutes, seconds))
	}

	sb.WriteString("[/font]\n \n")

	// Détails techniques
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Détails techniques[/b][/size][/color][/font]\n \n[font=Verdana]")

	// Format conteneur
	if ref.Container != "" {
		sb.WriteString(fmt.Sprintf("[b]Format :[/b] %s\n", strings.ToUpper(ref.Container)))
	}

	// Codec audio
	if len(ref.Audio) > 0 {
		audio := ref.Audio[0]
		sb.WriteString(fmt.Sprintf("[b]Codec :[/b] %s\n", audio.AudioCodecTag()))
		if audio.SampleRate > 0 {
			sb.WriteString(fmt.Sprintf("[b]Fréquence d'échantillonnage :[/b] %s\n", formatSampleRate(audio.SampleRate)))
		}
		if audio.BitDepth > 0 {
			sb.WriteString(fmt.Sprintf("[b]Profondeur :[/b] %d bits\n", audio.BitDepth))
		}
		if audio.Bitrate > 0 {
			sb.WriteString(fmt.Sprintf("[b]Débit :[/b] %d kb/s\n", audio.Bitrate/1000))
		}
		if audio.BitrateMode != "" {
			sb.WriteString(fmt.Sprintf("[b]Mode débit :[/b] %s\n", audio.BitrateMode))
		}
		if audio.Channels > 0 {
			sb.WriteString(fmt.Sprintf("[b]Canaux :[/b] %s\n", audio.ChannelLayoutShort()))
		}
	}

	sb.WriteString(" \n")

	// Nombre de fichiers et poids
	sb.WriteString(fmt.Sprintf("[b]Nombre de fichiers :[/b] %d\n", len(infos)))

	var totalSize int64
	for _, info := range infos {
		totalSize += info.FileSize
	}
	sb.WriteString(fmt.Sprintf("[b]Poids Total :[/b] %s", formatFileSize(totalSize)))

	sb.WriteString("[/font]\n \n")
	sb.WriteString("[/center] \n")

	return sb.String()
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func sanitizeName(name string) string {
	// Remplacer les caractères invalides pour les noms de fichiers
	replacements := map[string]string{
		"/": "-", "\\": "-", ":": "-", "*": "", "?": "", "\"": "",
		"<": "", ">": "", "|": "",
	}
	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	return strings.TrimSpace(result)
}

func detectAudioFormat(info *mediainfo.MediaInfo) string {
	if info == nil {
		return ""
	}

	// Prioriser le format conteneur pour les fichiers audio
	switch strings.ToLower(info.Container) {
	case "mpeg audio", "mp3":
		return "MP3"
	case "flac":
		return "FLAC"
	case "aac":
		return "AAC"
	case "ogg", "vorbis":
		return "OGG"
	case "opus":
		return "Opus"
	case "wav", "wave":
		return "WAV"
	case "wavpack":
		return "WavPack"
	case "alac":
		return "ALAC"
	}

	// Fallback: codec de la piste audio
	if len(info.Audio) > 0 {
		tag := info.Audio[0].AudioCodecTag()
		if strings.EqualFold(tag, "MPEG AUDIO") {
			return "MP3"
		}
		return tag
	}
	return ""
}

func detectAudioBitrate(info *mediainfo.MediaInfo) string {
	if info == nil || len(info.Audio) == 0 {
		return ""
	}

	audio := info.Audio[0]
	codec := strings.ToLower(audio.Codec)

	// Pour les formats lossless, indiquer la profondeur/sample rate
	if strings.Contains(codec, "flac") || strings.Contains(codec, "alac") ||
		strings.Contains(codec, "wav") || strings.Contains(codec, "pcm") {
		parts := []string{}
		if audio.BitDepth > 0 {
			parts = append(parts, fmt.Sprintf("%dbit", audio.BitDepth))
		}
		if audio.SampleRate > 0 {
			parts = append(parts, formatSampleRateShort(audio.SampleRate))
		}
		if len(parts) > 0 {
			return strings.Join(parts, " ")
		}
		return "Lossless"
	}

	// Pour les formats lossy, indiquer le bitrate
	if audio.Bitrate > 0 {
		return fmt.Sprintf("%dkbps", audio.Bitrate/1000)
	}

	// Fallback: bitrate global
	if info.OverallBitrate > 0 {
		return fmt.Sprintf("%dkbps", info.OverallBitrate/1000)
	}

	return ""
}

func formatSampleRate(rate int) string {
	if rate%1000 == 0 {
		return fmt.Sprintf("%d kHz", rate/1000)
	}
	return fmt.Sprintf("%.1f kHz", float64(rate)/1000)
}

func formatSampleRateShort(rate int) string {
	if rate%1000 == 0 {
		return fmt.Sprintf("%dkHz", rate/1000)
	}
	return fmt.Sprintf("%.1fkHz", float64(rate)/1000)
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

// ---------------------------------------------------------------------------
// NFO helper functions
// ---------------------------------------------------------------------------

// dottedField formats a label/value pair with dot padding.
// Format: "Label............................: Value" (35 chars before the space)
func dottedField(label, value string) string {
	const fieldWidth = 35
	dotsNeeded := fieldWidth - len(label) - 1 // -1 for the colon
	if dotsNeeded < 1 {
		dotsNeeded = 1
	}
	return label + strings.Repeat(".", dotsNeeded) + ": " + value
}

// audioNFOFields contains extracted audio info for NFO generation
type audioNFOFields struct {
	Codec            string
	Format           string
	OverallBitRate   string
	BitRateMode      string
	Channel          string
	Quality          string
	WritingLibrary   string
	EncodingSettings string
}

// extractAudioNFOFields runs mediainfo --Inform to extract precise audio info
func extractAudioNFOFields(filePath string) *audioNFOFields {
	fields := &audioNFOFields{}

	// General section: overall bitrate
	if out, err := exec.Command("mediainfo", "--Inform=General;%OverallBitRate/String%", filePath).Output(); err == nil {
		fields.OverallBitRate = strings.TrimSpace(string(out))
	}

	// Audio section: detailed fields
	template := `Audio;%Format%||%Format_Version%||%Format_Profile%||%BitRate_Mode/String%||%Channel(s)/String%||%SamplingRate/String%||%BitDepth/String%||%Compression_Mode%||%Encoded_Library%||%Encoded_Library_Settings%` + "\n"
	if out, err := exec.Command("mediainfo", "--Inform="+template, filePath).Output(); err == nil {
		// Take only the first line (first audio track)
		lines := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)
		if len(lines) > 0 {
			parts := strings.Split(lines[0], "||")
			if len(parts) >= 10 {
				format := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				profile := strings.TrimSpace(parts[2])

				fields.Codec = deriveCodecName(format, version, profile)

				// Build format string: "MPEG Audio / Version 1 / Layer 3"
				var formatParts []string
				if format != "" {
					formatParts = append(formatParts, format)
				}
				if version != "" {
					formatParts = append(formatParts, version)
				}
				if profile != "" {
					formatParts = append(formatParts, profile)
				}
				fields.Format = strings.Join(formatParts, " / ")

				fields.BitRateMode = strings.TrimSpace(parts[3])

				// Channel: "2 channels / 44.1 kHz" or "2 channels / 44.1 kHz / 16 bits"
				var channelParts []string
				if p := strings.TrimSpace(parts[4]); p != "" {
					channelParts = append(channelParts, p)
				}
				if p := strings.TrimSpace(parts[5]); p != "" {
					channelParts = append(channelParts, p)
				}
				if p := strings.TrimSpace(parts[6]); p != "" {
					if !strings.Contains(strings.ToLower(p), "bit") {
						p += " bits"
					}
					channelParts = append(channelParts, p)
				}
				fields.Channel = strings.Join(channelParts, " / ")

				fields.Quality = strings.TrimSpace(parts[7])
				fields.WritingLibrary = strings.TrimSpace(parts[8])
				fields.EncodingSettings = strings.TrimSpace(parts[9])
			}
		}
	}

	return fields
}

// deriveCodecName returns a human-friendly codec name from mediainfo format fields
func deriveCodecName(format, version, profile string) string {
	fl := strings.ToLower(format)
	pl := strings.ToLower(profile)

	switch {
	case strings.Contains(fl, "mpeg audio"):
		if strings.Contains(pl, "layer 3") {
			return "MP3"
		}
		if strings.Contains(pl, "layer 2") {
			return "MP2"
		}
		return "MPEG Audio"
	case strings.Contains(fl, "flac"):
		return "FLAC"
	case strings.Contains(fl, "aac"):
		return "AAC"
	case strings.Contains(fl, "opus"):
		return "Opus"
	case strings.Contains(fl, "vorbis"):
		return "Vorbis"
	case strings.Contains(fl, "pcm"):
		return "PCM"
	case strings.Contains(fl, "alac"):
		return "ALAC"
	case strings.Contains(fl, "wavpack"):
		return "WavPack"
	default:
		if format != "" {
			return strings.ToUpper(format)
		}
		return "Unknown"
	}
}

// resolveTrackTitle returns the track title, using MusicBrainz data if available,
// otherwise parsing it from the filename.
func resolveTrackTitle(filename string, album *musicbrainz.Album, index int) string {
	if index < len(album.Tracks) && album.Tracks[index].Title != "" {
		return album.Tracks[index].Title
	}

	// Fallback: extract from filename
	name := filename
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[:idx]
	}
	// Remove leading track number patterns: "01 - ", "01. "
	name = strings.TrimLeft(name, "0123456789")
	name = strings.TrimLeft(name, " .-_")
	// Remove artist prefix if present
	if album.Artist != "" {
		prefix := album.Artist + " - "
		if strings.HasPrefix(name, prefix) {
			name = name[len(prefix):]
		}
	}
	return strings.TrimSpace(name)
}

func formatSizeMiB(size int64) string {
	mib := float64(size) / (1024 * 1024)
	return fmt.Sprintf("%.2f MiB", mib)
}

func formatDurationMMSS(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func formatTotalSizeMO(size int64) string {
	mo := float64(size) / (1024 * 1024)
	return fmt.Sprintf("%.2f MO", mo)
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text
}
