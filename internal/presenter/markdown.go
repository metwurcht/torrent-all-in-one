package presenter

import (
	"fmt"
	"strings"

	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
)

// GenerateMarkdown génère une présentation BBCode du film pour forums
func GenerateMarkdown(movie *tmdb.Movie, media *mediainfo.MediaInfo) string {
	var sb strings.Builder

	sb.WriteString("[center]")

	// Titre principal en rouge
	sb.WriteString(fmt.Sprintf("[font=Verdana][size=200][color=#aa0000][b]%s[/b][/color][/size][/font]\n", movie.Title))
	if movie.Year() != "" {
		sb.WriteString(fmt.Sprintf("[font=Verdana][size=150][color=#aa0000](%s)[/color][/size][/font]\n", movie.Year()))
	}
	sb.WriteString("\n\n")

	// Poster
	if movie.PosterPath != "" {
		sb.WriteString(fmt.Sprintf("[img]%s[/img]\n\n", movie.PosterURL("w500")))
	}

	// Tagline en italique rouge
	if movie.Tagline != "" {
		sb.WriteString(fmt.Sprintf("[font=Verdana][size=100][color=#aa0000][i]« %s »[/i][/color][/size][/font]\n", movie.Tagline))
		sb.WriteString(" \n \n")
	}

	// Section Informations
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Informations[/b][/size][/color][/font]\n \n[font=Verdana]")

	// Titre original
	if movie.OriginalTitle != "" && movie.OriginalTitle != movie.Title {
		sb.WriteString(fmt.Sprintf("[b]Titre original :[/b] %s\n", movie.OriginalTitle))
	}

	// Date de sortie
	if movie.ReleaseDate != "" {
		sb.WriteString(fmt.Sprintf("[b]Sortie :[/b] %s\n", movie.ReleaseDate))
	}

	// Durée
	if movie.Runtime > 0 {
		sb.WriteString(fmt.Sprintf("[b]Durée :[/b] %d min\n", movie.Runtime))
	}
	sb.WriteString(" \n")

	// Réalisateur
	if len(movie.Directors) > 0 {
		sb.WriteString(fmt.Sprintf("[b]Réalisateur :[/b] %s\n \n", strings.Join(movie.Directors, ", ")))
	}

	// Acteurs (premiers 5)
	if len(movie.Cast) > 0 {
		sb.WriteString("[b]Acteurs :[/b]\n")
		for i, actor := range movie.Cast {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("%s, ", actor.Name))
		}
		sb.WriteString("\n \n")
	}

	// Genres
	if len(movie.Genres) > 0 {
		sb.WriteString(fmt.Sprintf("[b]Genres :[/b]\n%s\n \n", strings.Join(movie.Genres, ", ")))
	}

	// Note TMDB
	if movie.VoteAverage > 0 {
		sb.WriteString(fmt.Sprintf("[img]https://zupimages.net/up/21/02/xro7.png[/img] %.2f\n \n", movie.VoteAverage))
	}

	// Lien TMDB
	sb.WriteString(fmt.Sprintf("[img]https://zupimages.net/up/21/03/mxao.png[/img] [url=https://www.themoviedb.org/movie/%d]Fiche du film[/url]\n", movie.ID))

	// IMDb
	if movie.IMDbID != "" {
		sb.WriteString(fmt.Sprintf("[img]https://zupimages.net/up/21/03/od5a.png[/img] [url=%s]%s[/url]\n", movie.IMDbURL(), movie.IMDbID))
	}

	sb.WriteString("[/font]\n \n")

	// Section Synopsis
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Synopsis[/b][/size][/color][/font]\n \n[font=Verdana]\n")
	if movie.Overview != "" {
		sb.WriteString(movie.Overview)
	}
	sb.WriteString("\n \n \n[/font]\n")

	// Images du casting (2 premiers)
	if len(movie.Cast) >= 2 {
		for i := 0; i < 2 && i < len(movie.Cast); i++ {
			if movie.Cast[i].ProfilePath != "" {
				sb.WriteString(fmt.Sprintf(" [img]https://image.tmdb.org/t/p/w138_and_h175_face%s[/img] ", movie.Cast[i].ProfilePath))
			}
		}
		sb.WriteString("\n \n \n")
	}

	// Section Détails techniques
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Détails techniques[/b][/size][/color][/font]\n \n[font=Verdana]")

	// Format et codecs
	sb.WriteString(fmt.Sprintf("[b]Format :[/b] %s\n", strings.ToUpper(media.Container)))
	sb.WriteString(fmt.Sprintf("[b]Codec Vidéo :[/b] %s", media.Video.VideoCodecTag()))
	if media.Video.BitDepth == 10 {
		sb.WriteString(" 10-bit")
	}
	sb.WriteString("\n")

	if media.Video.Bitrate > 0 {
		sb.WriteString(fmt.Sprintf("[b]Débit Vidéo :[/b] ~%d kb/s\n", media.Video.Bitrate/1000))
	}
	sb.WriteString(fmt.Sprintf("[b]Résolution :[/b] %s\n \n", media.Video.Resolution))

	// Pistes audio avec drapeaux
	if len(media.Audio) > 0 {
		sb.WriteString("[b]Langue(s) :[/b]\n")
		for _, audio := range media.Audio {
			flag := getCountryFlag(audio.Language)
			langName := getLanguageName(audio.Language)
			sb.WriteString(fmt.Sprintf("%s %s [%s] | %s", flag, langName, audio.ChannelLayoutShort(), audio.AudioCodecTag()))
			if audio.Bitrate > 0 {
				sb.WriteString(fmt.Sprintf(" à %d kb/s", audio.Bitrate/1000))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n \n")
	}

	// Sous-titres
	if len(media.Subtitles) > 0 {
		sb.WriteString("[b]Sous-titres :[/b]\n")
		for _, sub := range media.Subtitles {
			flag := getCountryFlag(sub.Language)
			langName := getLanguageName(sub.Language)
			subType := "full"
			if sub.Forced {
				subType = "forced"
			}
			sb.WriteString(fmt.Sprintf("%s %s | %s (%s)\n", flag, langName, sub.Format, subType))
		}
		sb.WriteString("\n \n")
	}

	// Débit global
	if media.OverallBitrate > 0 {
		sb.WriteString(fmt.Sprintf("[b]Débit Global :[/b] ~%d kb/s", media.OverallBitrate/1000))
	}

	sb.WriteString("[/font]\n \n")

	// Section Téléchargements
	sb.WriteString("[font=Verdana][color=#9900ff][size=150][b]Téléchargements[/b][/size][/color][/font]\n \n")
	sb.WriteString(fmt.Sprintf("[b]Fichier :[/b] %s\n", media.FileName))
	sb.WriteString(fmt.Sprintf("[b]Poids Total :[/b] %s", media.FileSizeFormatted()))

	sb.WriteString("[/center]")

	return sb.String()
}

// getCountryFlag retourne l'icône de drapeau pour une langue
func getCountryFlag(lang string) string {
	langMap := map[string]string{
		"fre": "[img]https://flagcdn.com/20x15/fr.png[/img]",
		"fra": "[img]https://flagcdn.com/20x15/fr.png[/img]",
		"eng": "[img]https://flagcdn.com/20x15/gb.png[/img]",
		"jpn": "[img]https://flagcdn.com/20x15/jp.png[/img]",
		"ger": "[img]https://flagcdn.com/20x15/de.png[/img]",
		"deu": "[img]https://flagcdn.com/20x15/de.png[/img]",
		"spa": "[img]https://flagcdn.com/20x15/es.png[/img]",
		"ita": "[img]https://flagcdn.com/20x15/it.png[/img]",
		"por": "[img]https://flagcdn.com/20x15/pt.png[/img]",
		"rus": "[img]https://flagcdn.com/20x15/ru.png[/img]",
		"chi": "[img]https://flagcdn.com/20x15/cn.png[/img]",
		"zho": "[img]https://flagcdn.com/20x15/cn.png[/img]",
		"kor": "[img]https://flagcdn.com/20x15/kr.png[/img]",
		"ara": "[img]https://flagcdn.com/20x15/sa.png[/img]",
	}

	if flag, ok := langMap[strings.ToLower(lang)]; ok {
		return flag
	}
	return "[img]https://flagcdn.com/20x15/un.png[/img]"
}

// getLanguageName retourne le nom de la langue
func getLanguageName(lang string) string {
	langMap := map[string]string{
		"fre": "Français",
		"fra": "Français",
		"eng": "Anglais",
		"jpn": "Japonais",
		"ger": "Allemand",
		"deu": "Allemand",
		"spa": "Espagnol",
		"ita": "Italien",
		"por": "Portugais",
		"rus": "Russe",
		"chi": "Chinois",
		"zho": "Chinois",
		"kor": "Coréen",
		"ara": "Arabe",
	}

	if name, ok := langMap[strings.ToLower(lang)]; ok {
		return name
	}
	return strings.ToUpper(lang)
}
