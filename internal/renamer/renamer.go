package renamer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
)

// Renamer gère le renommage des fichiers selon les conventions warez
type Renamer struct {
	groupName string
}

// NewRenamer crée un nouveau renamer
func NewRenamer(groupName string) *Renamer {
	return &Renamer{
		groupName: groupName,
	}
}

// GenerateName génère le nom de release selon les conventions warez
// Format: Titre.Annee.Resolution.Source.VideoCodec.AudioCodec-GROUP
func (r *Renamer) GenerateName(movie *tmdb.Movie, media *mediainfo.MediaInfo, sourceType string) string {
	parts := []string{}

	// Titre (remplacer les espaces par des points, nettoyer les caractères spéciaux)
	title := r.cleanTitle(movie.OriginalTitle)
	parts = append(parts, title)

	// Année
	if year := movie.Year(); year != "" {
		parts = append(parts, year)
	}

	// Résolution
	if media.Video.Resolution != "" {
		parts = append(parts, media.Video.Resolution)
	}

	// Source (fournie par l'utilisateur)
	if sourceType != "" {
		parts = append(parts, sourceType)
	}

	// HDR si présent
	if media.Video.HDR != "" {
		parts = append(parts, media.Video.HDR)
	}

	// Codec vidéo
	if codec := media.Video.VideoCodecTag(); codec != "" {
		parts = append(parts, codec)
	}

	// Bit depth si 10-bit
	if media.Video.BitDepth == 10 {
		parts = append(parts, "10bit")
	}

	// Codec audio (premier track principal)
	if len(media.Audio) > 0 {
		audioTag := media.Audio[0].AudioCodecTag()
		channelLayout := media.Audio[0].ChannelLayoutShort()
		parts = append(parts, fmt.Sprintf("%s.%s", audioTag, channelLayout))
	}

	// Langue(s) détectée(s)
	langs := r.detectLanguages(media)
	if langs != "" {
		parts = append(parts, langs)
	}

	// Joindre avec des points
	releaseName := strings.Join(parts, ".")

	// Ajouter le groupe
	releaseName = fmt.Sprintf("%s-%s", releaseName, r.groupName)

	return releaseName
}

// cleanTitle nettoie le titre pour le format warez
func (r *Renamer) cleanTitle(title string) string {
	// Remplacer les caractères spéciaux
	replacements := map[string]string{
		" ":  ".",
		":":  "",
		"'":  "",
		"\"": "",
		"/":  ".",
		"\\": ".",
		"?":  "",
		"!":  "",
		"*":  "",
		"<":  "",
		">":  "",
		"|":  "",
		"&":  "and",
		"#":  "",
		"%":  "",
		"@":  "at",
		"(":  "",
		")":  "",
		"[":  "",
		"]":  "",
		"{":  "",
		"}":  "",
		",":  "",
		";":  "",
		"-":  "",
	}

	result := title
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Supprimer les points multiples
	re := regexp.MustCompile(`\.+`)
	result = re.ReplaceAllString(result, ".")

	// Supprimer les points au début et à la fin
	result = strings.Trim(result, ".")

	return result
}

// detectSource détecte la source probable du fichier
func (r *Renamer) detectSource(media *mediainfo.MediaInfo) string {
	fileName := strings.ToLower(media.FileName)
	container := strings.ToLower(media.Container)

	// Chercher des indices dans le nom de fichier
	sources := map[string]string{
		"bluray":  "BluRay",
		"blu-ray": "BluRay",
		"bdrip":   "BDRip",
		"brrip":   "BRRip",
		"webrip":  "WEBRip",
		"web-rip": "WEBRip",
		"webdl":   "WEB-DL",
		"web-dl":  "WEB-DL",
		"web":     "WEB",
		"hdtv":    "HDTV",
		"hdrip":   "HDRip",
		"dvdrip":  "DVDRip",
		"dvd":     "DVD",
		"uhd":     "UHD.BluRay",
	}

	for pattern, source := range sources {
		if strings.Contains(fileName, pattern) {
			return source
		}
	}

	// Deviner à partir des caractéristiques techniques
	if media.Video.Resolution == "2160p" {
		if container == "mkv" {
			return "UHD.BluRay"
		}
		return "WEB-DL"
	}
	if media.Video.Resolution == "1080p" {
		if container == "mkv" && media.Video.Bitrate > 10000000 {
			return "BluRay"
		}
		return "WEB-DL"
	}

	return ""
}

// detectLanguages détecte les langues des pistes audio
func (r *Renamer) detectLanguages(media *mediainfo.MediaInfo) string {
	if len(media.Audio) == 0 {
		return ""
	}

	langMap := map[string]string{
		"fr":      "FRENCH",
		"fre":     "FRENCH",
		"fra":     "FRENCH",
		"french":  "FRENCH",
		"eng":     "", // Anglais est par défaut
		"english": "",
		"ger":     "GERMAN",
		"deu":     "GERMAN",
		"german":  "GERMAN",
		"spa":     "SPANISH",
		"ita":     "ITALIAN",
		"jpn":     "JAPANESE",
		"kor":     "KOREAN",
		"chi":     "CHINESE",
		"zho":     "CHINESE",
		"rus":     "RUSSIAN",
		"por":     "PORTUGUESE",
		"ara":     "ARABIC",
	}

	langs := []string{}
	hasFrench := false
	hasEnglish := false

	for _, audio := range media.Audio {
		lang := strings.ToLower(audio.Language)
		if mapped, ok := langMap[lang]; ok {
			if mapped == "FRENCH" {
				hasFrench = true
			}
			if lang == "eng" || lang == "english" {
				hasEnglish = true
			}
			if mapped != "" && !contains(langs, mapped) {
				langs = append(langs, mapped)
			}
		}
	}

	// Si français + anglais, c'est MULTI
	if hasFrench && hasEnglish {
		return "MULTI"
	}

	// Si seulement français avec VO originale
	if hasFrench {
		//TODO : vérifier si c'est la VO pour mettre VOF
		return "VF"
	}

	if len(langs) > 0 {
		return strings.Join(langs, ".")
	}

	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
