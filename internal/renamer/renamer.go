package renamer

import (
	"fmt"
	"strings"

	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
)

// lightBitrateThresholds définit, par résolution puis par codec vidéo, le seuil de
// débit vidéo (en bits/s) en dessous duquel le tag HDLight (1080p) ou 4KLight (2160p)
// est obligatoire. Au-delà ou égal au seuil, aucun tag n'est ajouté.
var lightBitrateThresholds = map[string]map[string]int{
	"1080p": {
		"x264": 8000000, // 8 000 kb/s
		"x265": 6000000, // 6 000 kb/s
		"AV1":  4000000, // 4 000 kb/s
	},
	"2160p": {
		"x265": 12000000, // 12 000 kb/s
		"AV1":  8000000,  // 8 000 kb/s
	},
}

// lightTag retourne le tag HDLight/4KLight à ajouter selon la résolution, le codec
// et le débit vidéo, ou une chaîne vide si aucun tag n'est requis.
func lightTag(video *mediainfo.VideoInfo) string {
	if video.Bitrate <= 0 {
		return ""
	}
	threshold, ok := lightBitrateThresholds[video.Resolution][video.VideoCodecTag()]
	if !ok || video.Bitrate >= threshold {
		return ""
	}
	switch video.Resolution {
	case "1080p":
		return "HDLight"
	case "2160p":
		return "4KLight"
	}
	return ""
}

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
	title := CleanTitle(movie.OriginalTitle)
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
			// Ajouter HDLight / 4KLight selon le seuil de débit propre au codec
			if tag := lightTag(&media.Video); tag != "" {
				parts = append(parts, tag)
			}
		}
	}

	// HDR si présent
	if media.Video.HDR != "" {
		parts = append(parts, media.Video.HDR)
	}

	// Codec audio (premier track principal)
	if len(media.Audio) > 0 {
		audioTag := media.Audio[0].AudioCodecTag()
		channelLayout := media.Audio[0].ChannelLayoutShort()
		parts = append(parts, fmt.Sprintf("%s.%s", audioTag, channelLayout))
	}

	// Codec vidéo
	if codec := media.Video.VideoCodecTag(); codec != "" {
		parts = append(parts, codec)
	}

	// Bit depth si 10-bit
	if media.Video.BitDepth == 10 {
		parts = append(parts, "10bit")
	}

	// Joindre avec des points
	releaseName := strings.Join(parts, ".")

	// Ajouter le groupe
	releaseName = fmt.Sprintf("%s-%s", releaseName, r.groupName)

	return releaseName
}

// detectLanguages détecte les langues des pistes audio
func (r *Renamer) detectLanguages(media *mediainfo.MediaInfo) string {
	if len(media.Audio) == 0 {
		return ""
	}

	uniqueLangs := make(map[mediainfo.Language]bool)
	hasFrench := false
	hasQuebecois := false
	audioDescriptionSuffix := ""

	for _, audio := range media.Audio {
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

	suffix += audioDescriptionSuffix

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
