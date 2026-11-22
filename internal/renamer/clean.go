package renamer

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Lettres qui ne se décomposent pas en « lettre + accent » via Unicode.
var specialLetters = strings.NewReplacer(
	"œ", "oe", "Œ", "OE",
	"æ", "ae", "Æ", "AE",
	"ß", "ss",
	"ø", "o", "Ø", "O",
	"đ", "d", "Đ", "D",
	"ł", "l", "Ł", "L",
	"&", "and",
)

// removeAccents décompose les caractères (é → e + ´) puis retire les accents.
var removeAccents = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// isSeparator indique si le caractère doit être remplacé par un point :
// espaces, apostrophes, tirets (et les points déjà présents).
func isSeparator(r rune) bool {
	switch r {
	case '.', '\'', '’', '‘', 'ʼ', '`', '-', '‐', '‑', '–', '—':
		return true
	}
	return unicode.IsSpace(r)
}

// CleanTitle convertit un titre au format des noms de release :
//   - les accents sont retirés (é → e, ç → c, œ → oe...)
//   - les espaces, apostrophes et tirets deviennent des points
//   - « & » devient « and »
//   - tous les autres caractères non alphanumériques sont supprimés
//   - les points consécutifs sont fusionnés, et retirés en début/fin
//
// Exemple : "La fête est finie" → "La.fete.est.finie"
func CleanTitle(title string) string {
	s := specialLetters.Replace(title)
	if out, _, err := transform.String(removeAccents, s); err == nil {
		s = out
	}

	var sb strings.Builder
	lastDot := true // évite un point en début de chaîne
	for _, r := range s {
		switch {
		case isSeparator(r):
			if !lastDot {
				sb.WriteRune('.')
				lastDot = true
			}
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			sb.WriteRune(r)
			lastDot = false
		}
	}
	return strings.TrimRight(sb.String(), ".")
}
