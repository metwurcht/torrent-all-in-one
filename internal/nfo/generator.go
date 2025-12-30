package nfo

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
)

// Generator génère des fichiers NFO
type Generator struct {
	groupName string
}

// NewGenerator crée un nouveau générateur NFO
func NewGenerator(groupName string) *Generator {
	return &Generator{
		groupName: groupName,
	}
}

// Generate génère le contenu du fichier NFO au format MediaInfo
func (g *Generator) Generate(movie *tmdb.Movie, media *mediainfo.MediaInfo, newFileName string) string {
	width := 80
	border := strings.Repeat("═", width)
	thinBorder := strings.Repeat("─", width)

	var sb strings.Builder

	// En-tête
	sb.WriteString(fmt.Sprintf("╔%s╗\n", border))
	sb.WriteString(g.centerLine(g.groupName+" presents", width))
	sb.WriteString(fmt.Sprintf("╠%s╣\n", border))
	sb.WriteString(g.centerLine(movie.Title, width))
	if movie.OriginalTitle != "" && movie.OriginalTitle != movie.Title {
		sb.WriteString(g.centerLine(fmt.Sprintf("(%s)", movie.OriginalTitle), width))
	}
	sb.WriteString(fmt.Sprintf("╠%s╣\n", border))

	// Section General
	sb.WriteString(g.centerLine("General", width))
	sb.WriteString(fmt.Sprintf("╠%s╣\n", thinBorder))
	sb.WriteString(g.formatLine("Complete name", newFileName, width))
	sb.WriteString(g.formatLine("Format", media.Container, width))
	if media.ContainerVersion != "" {
		sb.WriteString(g.formatLine("Format version", media.ContainerVersion, width))
	}
	sb.WriteString(g.formatLine("File size", media.FileSizeFormatted(), width))
	sb.WriteString(g.formatLine("Duration", media.DurationFormatted(), width))
	if media.OverallBitrate > 0 {
		sb.WriteString(g.formatLine("Overall bit rate", fmt.Sprintf("%d kb/s", media.OverallBitrate/1000), width))
	}
	if media.MovieName != "" {
		sb.WriteString(g.formatLine("Movie name", media.MovieName, width))
	}
	if media.EncodedDate != "" {
		sb.WriteString(g.formatLine("Encoded date", media.EncodedDate, width))
	}
	if media.WritingApplication != "" {
		sb.WriteString(g.formatLine("Writing application", media.WritingApplication, width))
	}
	if media.WritingLibrary != "" {
		sb.WriteString(g.formatLine("Writing library", media.WritingLibrary, width))
	}

	sb.WriteString(fmt.Sprintf("╠%s╣\n", border))

	// Section Video
	sb.WriteString(g.centerLine("Video", width))
	sb.WriteString(fmt.Sprintf("╠%s╣\n", thinBorder))
	sb.WriteString(g.formatLine("Format", media.Video.Codec, width))
	if media.Video.CodecInfo != "" {
		sb.WriteString(g.formatLine("Format/Info", media.Video.CodecInfo, width))
	}
	if media.Video.CodecProfile != "" {
		sb.WriteString(g.formatLine("Format profile", media.Video.CodecProfile, width))
	}
	if media.Video.CodecID != "" {
		sb.WriteString(g.formatLine("Codec ID", media.Video.CodecID, width))
	}
	sb.WriteString(g.formatLine("Duration", media.DurationFormatted(), width))
	if media.Video.Bitrate > 0 {
		sb.WriteString(g.formatLine("Bit rate", fmt.Sprintf("%d kb/s", media.Video.Bitrate/1000), width))
	}
	sb.WriteString(g.formatLine("Width", fmt.Sprintf("%d pixels", media.Video.Width), width))
	sb.WriteString(g.formatLine("Height", fmt.Sprintf("%d pixels", media.Video.Height), width))
	if media.Video.AspectRatio != "" {
		sb.WriteString(g.formatLine("Display aspect ratio", media.Video.AspectRatio, width))
	}
	if media.Video.FrameRateMode != "" {
		sb.WriteString(g.formatLine("Frame rate mode", media.Video.FrameRateMode, width))
	}
	sb.WriteString(g.formatLine("Frame rate", fmt.Sprintf("%.3f FPS", media.Video.FrameRate), width))
	if media.Video.ColorSpace != "" {
		sb.WriteString(g.formatLine("Color space", media.Video.ColorSpace, width))
	}
	if media.Video.ChromaSubsampling != "" {
		sb.WriteString(g.formatLine("Chroma subsampling", media.Video.ChromaSubsampling, width))
	}
	if media.Video.BitDepth > 0 {
		sb.WriteString(g.formatLine("Bit depth", fmt.Sprintf("%d bits", media.Video.BitDepth), width))
	}
	if media.Video.StreamSize > 0 {
		percentage := float64(media.Video.StreamSize) / float64(media.FileSize) * 100
		sb.WriteString(g.formatLine("Stream size", fmt.Sprintf("%.2f GiB (%.0f%%)", float64(media.Video.StreamSize)/(1024*1024*1024), percentage), width))
	}
	if media.Video.ColorRange != "" {
		sb.WriteString(g.formatLine("Color range", media.Video.ColorRange, width))
	}
	if media.Video.ColorPrimaries != "" {
		sb.WriteString(g.formatLine("Color primaries", media.Video.ColorPrimaries, width))
	}
	if media.Video.TransferCharacteristics != "" {
		sb.WriteString(g.formatLine("Transfer characteristics", media.Video.TransferCharacteristics, width))
	}
	if media.Video.MatrixCoefficients != "" {
		sb.WriteString(g.formatLine("Matrix coefficients", media.Video.MatrixCoefficients, width))
	}
	if media.Video.HDR != "" {
		sb.WriteString(g.formatLine("HDR format", media.Video.HDR, width))
	}

	// Sections Audio
	for i, audio := range media.Audio {
		sb.WriteString(fmt.Sprintf("╠%s╣\n", border))
		sb.WriteString(g.centerLine(fmt.Sprintf("Audio #%d", i+1), width))
		sb.WriteString(fmt.Sprintf("╠%s╣\n", thinBorder))
		sb.WriteString(g.formatLine("Format", audio.Codec, width))
		if audio.CodecInfo != "" {
			sb.WriteString(g.formatLine("Format/Info", audio.CodecInfo, width))
		}
		if audio.CommercialName != "" {
			sb.WriteString(g.formatLine("Commercial name", audio.CommercialName, width))
		}
		if audio.CodecID != "" {
			sb.WriteString(g.formatLine("Codec ID", audio.CodecID, width))
		}
		sb.WriteString(g.formatLine("Duration", media.DurationFormatted(), width))
		if audio.BitrateMode != "" {
			sb.WriteString(g.formatLine("Bit rate mode", audio.BitrateMode, width))
		}
		if audio.Bitrate > 0 {
			sb.WriteString(g.formatLine("Bit rate", fmt.Sprintf("%d kb/s", audio.Bitrate/1000), width))
		}
		sb.WriteString(g.formatLine("Channel(s)", fmt.Sprintf("%d channels", audio.Channels), width))
		if audio.ChannelLayoutFormatted() != "" {
			sb.WriteString(g.formatLine("Channel layout", audio.ChannelLayoutFormatted(), width))
		}
		if audio.SampleRate > 0 {
			sb.WriteString(g.formatLine("Sampling rate", fmt.Sprintf("%.1f kHz", float64(audio.SampleRate)/1000), width))
		}
		if audio.BitDepth > 0 {
			sb.WriteString(g.formatLine("Bit depth", fmt.Sprintf("%d bits", audio.BitDepth), width))
		}
		if audio.Compression != "" {
			sb.WriteString(g.formatLine("Compression mode", audio.Compression, width))
		}
		if audio.StreamSize > 0 {
			percentage := float64(audio.StreamSize) / float64(media.FileSize) * 100
			sb.WriteString(g.formatLine("Stream size", fmt.Sprintf("%.0f MiB (%.0f%%)", float64(audio.StreamSize)/(1024*1024), percentage), width))
		}
		if audio.Title != "" {
			sb.WriteString(g.formatLine("Title", audio.Title, width))
		}
		if audio.Language != "" {
			sb.WriteString(g.formatLine("Language", audio.Language, width))
		}
		if audio.ServiceKind != "" {
			sb.WriteString(g.formatLine("Service kind", audio.ServiceKind, width))
		}
		sb.WriteString(g.formatLine("Default", g.boolToYesNo(audio.Default), width))
		sb.WriteString(g.formatLine("Forced", g.boolToYesNo(audio.Forced), width))
	}

	// Section Subtitles
	if len(media.Subtitles) > 0 {
		for i, sub := range media.Subtitles {
			sb.WriteString(fmt.Sprintf("╠%s╣\n", border))
			sb.WriteString(g.centerLine(fmt.Sprintf("Text #%d", i+1), width))
			sb.WriteString(fmt.Sprintf("╠%s╣\n", thinBorder))
			sb.WriteString(g.formatLine("Format", sub.Format, width))
			if sub.CodecID != "" {
				sb.WriteString(g.formatLine("Codec ID", sub.CodecID, width))
			}
			if sub.Title != "" {
				sb.WriteString(g.formatLine("Title", sub.Title, width))
			}
			if sub.Language != "" {
				sb.WriteString(g.formatLine("Language", sub.Language, width))
			}
			sb.WriteString(g.formatLine("Default", g.boolToYesNo(sub.Default), width))
			sb.WriteString(g.formatLine("Forced", g.boolToYesNo(sub.Forced), width))
		}
	}

	sb.WriteString(fmt.Sprintf("╠%s╣\n", border))

	// Section Movie Info
	sb.WriteString(g.centerLine("Movie Info", width))
	sb.WriteString(fmt.Sprintf("╠%s╣\n", thinBorder))
	if movie.ReleaseDate != "" {
		sb.WriteString(g.formatLine("Release Date", movie.ReleaseDate, width))
	}
	if len(movie.Genres) > 0 {
		sb.WriteString(g.formatLine("Genre", strings.Join(movie.Genres, ", "), width))
	}
	if movie.Runtime > 0 {
		sb.WriteString(g.formatLine("Runtime", fmt.Sprintf("%d min", movie.Runtime), width))
	}
	if movie.VoteAverage > 0 {
		sb.WriteString(g.formatLine("Rating", fmt.Sprintf("%.1f/10", movie.VoteAverage), width))
	}
	if movie.IMDbID != "" {
		sb.WriteString(g.formatLine("IMDb", movie.IMDbURL(), width))
	}
	sb.WriteString(g.formatLine("TMDB", movie.TMDbURL(), width))

	if len(movie.Directors) > 0 {
		sb.WriteString(g.formatLine("Director", strings.Join(movie.Directors, ", "), width))
	}

	if len(movie.Cast) > 0 {
		actors := make([]string, 0, 5)
		for i, c := range movie.Cast {
			if i >= 5 {
				break
			}
			actors = append(actors, c.Name)
		}
		sb.WriteString(g.formatLine("Cast", strings.Join(actors, ", "), width))
	}

	// Synopsis
	if movie.Overview != "" {
		sb.WriteString(fmt.Sprintf("╠%s╣\n", border))
		sb.WriteString(g.centerLine("Synopsis", width))
		sb.WriteString(fmt.Sprintf("╠%s╣\n", thinBorder))
		sb.WriteString(g.wrapText(movie.Overview, width))
	}

	sb.WriteString(fmt.Sprintf("╠%s╣\n", border))

	// Footer
	sb.WriteString(g.centerLine("Generated by torrent AIO", width))
	sb.WriteString(g.centerLine(time.Now().Format("2006-01-02"), width))
	sb.WriteString(fmt.Sprintf("╚%s╝\n", border))

	return sb.String()
}

func (g *Generator) boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func (g *Generator) centerLine(text string, width int) string {
	textLen := utf8.RuneCountInString(text)
	if textLen > width-4 {
		// Tronquer en gardant les bons caractères
		runes := []rune(text)
		text = string(runes[:width-7]) + "..."
		textLen = width - 4
	}
	padding := (width - textLen) / 2
	return fmt.Sprintf("║%s%s%s║\n",
		strings.Repeat(" ", padding),
		text,
		strings.Repeat(" ", width-padding-textLen))
}

func (g *Generator) formatLine(label, value string, width int) string {
	labelWidth := 25
	valueWidth := width - labelWidth - 4 // ║ + espace + : + espace + espace + ║

	valueLen := utf8.RuneCountInString(value)
	if valueLen > valueWidth {
		runes := []rune(value)
		value = string(runes[:valueWidth-3]) + "..."
		valueLen = valueWidth
	}

	labelLen := utf8.RuneCountInString(label)
	labelPadding := labelWidth - labelLen
	valuePadding := valueWidth - valueLen

	return fmt.Sprintf("║ %s%s: %s%s ║\n",
		label,
		strings.Repeat(" ", labelPadding),
		value,
		strings.Repeat(" ", valuePadding))
}

func (g *Generator) wrapText(text string, width int) string {
	words := strings.Fields(text)
	var lines []string
	var currentLine strings.Builder
	lineWidth := width - 4

	for _, word := range words {
		if currentLine.Len()+len(word)+1 > lineWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("║ %-*s ║\n", lineWidth, line))
	}
	return sb.String()
}
