package mediainfo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Analyzer analyse les fichiers vidéo pour extraire les métadonnées
type Analyzer struct {
	mediaInfoPath string
}

// NewAnalyzer crée un nouvel analyseur
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		mediaInfoPath: "mediainfo",
	}
}

// Analyze analyse un fichier vidéo et retourne ses métadonnées
func (a *Analyzer) Analyze(filePath string) (*MediaInfo, error) {
	// Vérifier que le fichier existe
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("fichier introuvable: %w", err)
	}

	mi := &MediaInfo{
		FileName: filepath.Base(filePath),
		FilePath: filePath,
		FileSize: info.Size(),
	}

	// Analyser avec mediainfo
	if err := a.analyzeWithMediaInfo(filePath, mi); err != nil {
		return nil, fmt.Errorf("impossible d'analyser le fichier avec mediainfo: %w", err)
	}

	return mi, nil
}

func (a *Analyzer) analyzeWithMediaInfo(filePath string, mi *MediaInfo) error {
	cmd := exec.Command(a.mediaInfoPath, "--Output=JSON", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return err
	}

	var result mediaInfoJSON
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return err
	}

	for _, track := range result.Media.Tracks {
		switch track.Type {
		case "General":
			if track.Duration != "" {
				mi.Duration = parseDuration(track.Duration)
			}
			if track.Format != "" {
				mi.Container = track.Format
			}
			if track.FormatVersion != "" {
				mi.ContainerVersion = track.FormatVersion
			}
			if track.OverallBitRate != "" {
				mi.OverallBitrate = parseInt(track.OverallBitRate)
			}
			mi.MovieName = track.MovieName
			mi.EncodedDate = track.EncodedDate
			mi.WritingApplication = track.WritingApplication
			mi.WritingLibrary = track.WritingLibrary
		case "Video":
			mi.Video = VideoInfo{
				Codec:                   track.Format,
				CodecInfo:               track.FormatInfo,
				CodecProfile:            track.FormatProfile,
				CodecID:                 track.CodecID,
				Width:                   parseInt(track.Width),
				Height:                  parseInt(track.Height),
				Bitrate:                 parseInt(track.BitRate),
				FrameRate:               parseFloat(track.FrameRate),
				FrameRateMode:           track.FrameRateMode,
				AspectRatio:             track.DisplayAspectRatio,
				BitDepth:                parseInt(track.BitDepth),
				HDR:                     detectHDR(track),
				ColorSpace:              track.ColorSpace,
				ChromaSubsampling:       track.ChromaSubsampling,
				ColorRange:              track.ColorRange,
				ColorPrimaries:          track.ColorPrimaries,
				TransferCharacteristics: track.TransferCharacteristics,
				MatrixCoefficients:      track.MatrixCoefficients,
				StreamSize:              parseInt(track.StreamSize),
			}
			mi.Video.Resolution = determineResolution(mi.Video.Width, mi.Video.Height)
		case "Audio":
			audio := AudioInfo{
				Codec:          track.Format,
				CodecInfo:      track.FormatInfo,
				CommercialName: track.FormatCommercialIfAny,
				CodecID:        track.CodecID,
				Channels:       parseInt(track.Channels),
				ChannelLayout:  track.ChannelLayout,
				SampleRate:     parseInt(track.SamplingRate),
				Bitrate:        parseInt(track.BitRate),
				BitrateMode:    track.BitRateMode,
				BitDepth:       parseInt(track.BitDepth),
				Compression:    track.CompressionMode,
				StreamSize:     parseInt(track.StreamSize),
				Language:       track.Language,
				Title:          track.Title,
				ServiceKind:    track.ServiceKind,
				Default:        track.Default == "Yes",
				Forced:         track.Forced == "Yes",
			}
			mi.Audio = append(mi.Audio, audio)
		case "Text":
			sub := SubtitleInfo{
				Format:   track.Format,
				CodecID:  track.CodecID,
				Language: track.Language,
				Title:    track.Title,
				Default:  track.Default == "Yes",
				Forced:   track.Forced == "Yes",
			}
			mi.Subtitles = append(mi.Subtitles, sub)
		}
	}

	return nil
}

func detectHDR(track mediaInfoTrack) string {
	hdrFormats := []string{}

	if strings.Contains(strings.ToLower(track.HDRFormat), "dolby vision") {
		hdrFormats = append(hdrFormats, "DV")
	}
	if strings.Contains(strings.ToLower(track.HDRFormat), "hdr10+") {
		hdrFormats = append(hdrFormats, "HDR10+")
	} else if strings.Contains(strings.ToLower(track.HDRFormat), "hdr10") ||
		strings.Contains(strings.ToLower(track.TransferCharacteristics), "pq") {
		hdrFormats = append(hdrFormats, "HDR10")
	}
	if strings.Contains(strings.ToLower(track.HDRFormat), "hlg") {
		hdrFormats = append(hdrFormats, "HLG")
	}

	if len(hdrFormats) > 0 {
		return strings.Join(hdrFormats, "+")
	}
	return ""
}

func determineResolution(width, height int) string {
	if height >= 2160 || width >= 3840 {
		return "2160p"
	}
	if height >= 1080 || width >= 1920 {
		return "1080p"
	}
	if height >= 720 || width >= 1280 {
		return "720p"
	}
	if height >= 480 {
		return "480p"
	}
	return fmt.Sprintf("%dp", height)
}

func parseDuration(s string) int {
	// Essayer de parser en secondes
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(f)
	}
	return 0
}

func parseInt(s string) int {
	// Nettoyer la chaîne
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")

	// Essayer de parser
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(f)
	}
	return 0
}

func parseFloat(s string) float64 {
	if f, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
		return f
	}
	return 0
}
