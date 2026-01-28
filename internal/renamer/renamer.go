package renamer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
)

const (
	// HDLightBitrateThreshold définit le seuil de bitrate (en bits/s) pour le tag HDLight
	HDLightBitrateThreshold = 2000000 // 2000 kb/s
	// FourKLightBitrateThreshold définit le seuil de bitrate (en bits/s) pour le tag 4KLight
	FourKLightBitrateThreshold = 4000000 // 4000 kb/s
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
func (r *Renamer) GenerateName(movie *tmdb.Movie, media *mediainfo.MediaInfo) string {
	parts := []string{}

	// Titre (remplacer les espaces par des points, nettoyer les caractères spéciaux)
	title := r.cleanTitle(movie.OriginalTitle)
	parts = append(parts, title)

	// Année
	if year := movie.Year(); year != "" {
		parts = append(parts, year)
	}

	// Langue(s) détectée(s)
	langs := r.detectLanguages(media)
	if langs != "" {
		parts = append(parts, langs)
	}

	// Résolution
	if media.Video.Resolution != "" {
		parts = append(parts, media.Video.Resolution)
	}

	// Source
	if media.SourceType != "" {
		parts = append(parts, string(media.SourceType))

		if media.SourceType == mediainfo.SourceBluRayRip {
			// Ajouter HDLight si BluRay 1080p avec débit < HDLightBitrateThreshold
			if media.Video.Resolution == "1080p" && media.Video.Bitrate < HDLightBitrateThreshold {
				parts = append(parts, "HDLight")
			}

			// Ajouter 4KLight si Bluray 2160p avec débit < FourKLightBitrateThreshold
			if media.Video.Resolution == "2160p" && media.Video.Bitrate < FourKLightBitrateThreshold {
				parts = append(parts, "4KLight")
			}
		}
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
		"en":      "ENGLISH",
		"eng":     "ENGLISH",
		"english": "ENGLISH",
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

	uniqueLangs := make(map[string]bool)
	hasFrench := false
	hasQuebecois := false

	for _, audio := range media.Audio {
		lang := strings.ToLower(audio.Language)
		title := strings.ToLower(audio.Title)

		// Détecter le québécois via le titre de la piste ou la langue (CA)
		if strings.Contains(title, "vfq") ||
			strings.Contains(title, "quebec") || strings.Contains(title, "québec") ||
			strings.Contains(title, "quebecois") || strings.Contains(title, "québécois") ||
			strings.Contains(lang, "(ca)") || strings.Contains(lang, "french (ca)") {
			hasQuebecois = true
			uniqueLangs["FRENCH"] = true
			continue
		}

		if mapped, ok := langMap[lang]; ok {
			if mapped == "FRENCH" {
				hasFrench = true
			}
			uniqueLangs[mapped] = true
		}
	}

	// Déterminer le préfixe MULTI si au moins 2 langues différentes
	isMulti := len(uniqueLangs) >= 2

	// Déterminer le suffixe selon français/québécois
	var suffix string
	if hasFrench && hasQuebecois {
		suffix = "VF2"
	} else if hasFrench {
		suffix = "VFF"
	} else if hasQuebecois {
		suffix = "VFQ"
	}

	// Construire le résultat
	if isMulti && suffix != "" {
		return "MULTI." + suffix
	} else if suffix != "" {
		return suffix
	}

	// Si aucun français/québécois, retourner les langues trouvées
	if len(uniqueLangs) > 0 {
		langs := []string{}
		for lang := range uniqueLangs {
			if lang != "QUEBECOIS" {
				langs = append(langs, lang)
			}
		}
		return strings.Join(langs, ".")
	}

	return ""
}
