package mediainfo

import (
	"fmt"
	"strings"
)

// MediaInfo contient toutes les métadonnées d'un fichier vidéo
type MediaInfo struct {
	FileName           string         `json:"file_name"`
	FilePath           string         `json:"file_path"`
	FileSize           int64          `json:"file_size"`
	Container          string         `json:"container"`
	ContainerVersion   string         `json:"container_version"`
	Duration           int            `json:"duration"` // en secondes
	OverallBitrate     int            `json:"overall_bitrate"`
	MovieName          string         `json:"movie_name"`
	EncodedDate        string         `json:"encoded_date"`
	WritingApplication string         `json:"writing_application"`
	WritingLibrary     string         `json:"writing_library"`
	Video              VideoInfo      `json:"video"`
	Audio              []AudioInfo    `json:"audio"`
	Subtitles          []SubtitleInfo `json:"subtitles"`
}

// VideoInfo contient les informations de la piste vidéo
type VideoInfo struct {
	Codec                   string  `json:"codec"`
	CodecInfo               string  `json:"codec_info"`
	CodecProfile            string  `json:"codec_profile"`
	CodecID                 string  `json:"codec_id"`
	Width                   int     `json:"width"`
	Height                  int     `json:"height"`
	Resolution              string  `json:"resolution"` // 1080p, 720p, etc.
	Bitrate                 int     `json:"bitrate"`
	FrameRate               float64 `json:"frame_rate"`
	FrameRateMode           string  `json:"frame_rate_mode"`
	AspectRatio             string  `json:"aspect_ratio"`
	BitDepth                int     `json:"bit_depth"`
	HDR                     string  `json:"hdr"` // HDR10, DV, HDR10+, etc.
	ColorSpace              string  `json:"color_space"`
	ChromaSubsampling       string  `json:"chroma_subsampling"`
	ColorRange              string  `json:"color_range"`
	ColorPrimaries          string  `json:"color_primaries"`
	TransferCharacteristics string  `json:"transfer_characteristics"`
	MatrixCoefficients      string  `json:"matrix_coefficients"`
	StreamSize              int     `json:"stream_size"`
}

// AudioInfo contient les informations d'une piste audio
type AudioInfo struct {
	Codec          string `json:"codec"`
	CodecInfo      string `json:"codec_info"`
	CommercialName string `json:"commercial_name"`
	CodecID        string `json:"codec_id"`
	Channels       int    `json:"channels"`
	ChannelLayout  string `json:"channel_layout"`
	SampleRate     int    `json:"sample_rate"`
	Bitrate        int    `json:"bitrate"`
	BitrateMode    string `json:"bitrate_mode"`
	BitDepth       int    `json:"bit_depth"`
	Compression    string `json:"compression"`
	StreamSize     int    `json:"stream_size"`
	Language       string `json:"language"`
	Title          string `json:"title"`
	ServiceKind    string `json:"service_kind"`
	Default        bool   `json:"default"`
	Forced         bool   `json:"forced"`
}

// SubtitleInfo contient les informations d'une piste de sous-titres
type SubtitleInfo struct {
	Format   string `json:"format"`
	CodecID  string `json:"codec_id"`
	Language string `json:"language"`
	Title    string `json:"title"`
	Default  bool   `json:"default"`
	Forced   bool   `json:"forced"`
}

// FileSizeFormatted retourne la taille du fichier formatée
func (m *MediaInfo) FileSizeFormatted() string {
	const unit = 1024
	if m.FileSize < unit {
		return fmt.Sprintf("%d B", m.FileSize)
	}
	div, exp := int64(unit), 0
	for n := m.FileSize / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(m.FileSize)/float64(div), "KMGTPE"[exp])
}

// DurationFormatted retourne la durée formatée
func (m *MediaInfo) DurationFormatted() string {
	hours := m.Duration / 3600
	minutes := (m.Duration % 3600) / 60

	if hours > 0 {
		return fmt.Sprintf("%d h %d min", hours, minutes)
	}
	return fmt.Sprintf("%d min", minutes)
}

// StreamSizeFormatted retourne la taille du stream formatée
func formatStreamSize(size int) string {
	if size == 0 {
		return ""
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := unit, 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f MiB", float64(size)/float64(div))
}

// VideoCodecTag retourne le tag du codec vidéo pour le nom de release
func (v *VideoInfo) VideoCodecTag() string {
	codec := strings.ToLower(v.Codec)

	switch {
	case strings.Contains(codec, "hevc") || strings.Contains(codec, "h265") || strings.Contains(codec, "x265"):
		return "x265"
	case strings.Contains(codec, "avc") || strings.Contains(codec, "h264") || strings.Contains(codec, "x264"):
		return "x264"
	case strings.Contains(codec, "av1"):
		return "AV1"
	case strings.Contains(codec, "vp9"):
		return "VP9"
	default:
		return strings.ToUpper(v.Codec)
	}
}

// AudioCodecTag retourne le tag du codec audio pour le nom de release
func (a *AudioInfo) AudioCodecTag() string {
	codec := strings.ToLower(a.Codec)

	switch {
	case strings.Contains(codec, "truehd") && strings.Contains(strings.ToLower(a.Title), "atmos"):
		return "TrueHD.Atmos"
	case strings.Contains(codec, "truehd"):
		return "TrueHD"
	case strings.Contains(codec, "dts") && strings.Contains(strings.ToLower(a.Title), "x"):
		return "DTS-X"
	case strings.Contains(codec, "dts-hd") || strings.Contains(codec, "dts") && strings.Contains(strings.ToLower(a.Title), "ma"):
		return "DTS-HD.MA"
	case strings.Contains(codec, "dts"):
		return "DTS"
	case strings.Contains(codec, "e-ac-3") || strings.Contains(codec, "eac3"):
		if strings.Contains(strings.ToLower(a.Title), "atmos") {
			return "EAC3.Atmos"
		}
		return "EAC3"
	case strings.Contains(codec, "ac-3") || strings.Contains(codec, "ac3"):
		return "AC3"
	case strings.Contains(codec, "aac"):
		return "AAC"
	case strings.Contains(codec, "flac"):
		return "FLAC"
	case strings.Contains(codec, "opus"):
		return "Opus"
	default:
		return strings.ToUpper(a.Codec)
	}
}

// ChannelLayoutFormatted retourne le layout des canaux audio formaté
func (a *AudioInfo) ChannelLayoutFormatted() string {
	// Si on a un layout explicite, l'utiliser
	if a.ChannelLayout != "" {
		return a.ChannelLayout
	}
	// Sinon, générer à partir du nombre de canaux
	switch a.Channels {
	case 1:
		return "C"
	case 2:
		return "L R"
	case 6:
		return "L R C LFE Ls Rs"
	case 8:
		return "L R C LFE Ls Rs Lb Rb"
	default:
		return ""
	}
}

// ChannelLayoutShort retourne le layout court (2.0, 5.1, etc.)
func (a *AudioInfo) ChannelLayoutShort() string {
	switch a.Channels {
	case 1:
		return "1.0"
	case 2:
		return "2.0"
	case 6:
		return "5.1"
	case 8:
		return "7.1"
	default:
		return fmt.Sprintf("%d.0", a.Channels)
	}
}

// Structures pour le parsing JSON de mediainfo

type mediaInfoJSON struct {
	Media struct {
		Tracks []mediaInfoTrack `json:"track"`
	} `json:"media"`
}

type mediaInfoTrack struct {
	Type                    string `json:"@type"`
	Format                  string `json:"Format"`
	FormatInfo              string `json:"Format_Info"`
	FormatProfile           string `json:"Format_Profile"`
	FormatCommercialIfAny   string `json:"Format_Commercial_IfAny"`
	CodecID                 string `json:"CodecID"`
	FormatVersion           string `json:"Format_Version"`
	Duration                string `json:"Duration"`
	OverallBitRate          string `json:"OverallBitRate"`
	MovieName               string `json:"Movie"`
	EncodedDate             string `json:"Encoded_Date"`
	WritingApplication      string `json:"Writing_Application"`
	WritingLibrary          string `json:"Writing_Library"`
	Width                   string `json:"Width"`
	Height                  string `json:"Height"`
	BitRate                 string `json:"BitRate"`
	BitRateMode             string `json:"BitRate_Mode"`
	FrameRate               string `json:"FrameRate"`
	FrameRateMode           string `json:"FrameRate_Mode"`
	DisplayAspectRatio      string `json:"DisplayAspectRatio"`
	BitDepth                string `json:"BitDepth"`
	HDRFormat               string `json:"HDR_Format"`
	TransferCharacteristics string `json:"transfer_characteristics"`
	ColorSpace              string `json:"ColorSpace"`
	ChromaSubsampling       string `json:"ChromaSubsampling"`
	ColorRange              string `json:"colour_range"`
	ColorPrimaries          string `json:"colour_primaries"`
	MatrixCoefficients      string `json:"matrix_coefficients"`
	StreamSize              string `json:"StreamSize"`
	Channels                string `json:"Channels"`
	ChannelLayout           string `json:"ChannelLayout"`
	SamplingRate            string `json:"SamplingRate"`
	CompressionMode         string `json:"Compression_Mode"`
	ServiceKind             string `json:"ServiceKind"`
	Language                string `json:"Language"`
	Title                   string `json:"Title"`
	Default                 string `json:"Default"`
	Forced                  string `json:"Forced"`
}
