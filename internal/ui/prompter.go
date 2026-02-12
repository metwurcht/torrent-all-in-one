package ui

import (
	"github.com/metwurcht/torrent-all-in-one/internal/media"
	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
)

// Prompter définit l'interface pour les interactions utilisateur.
// Cette interface permet de facilement remplacer l'implémentation CLI
// par une autre (WebUI, API, tests, etc.)
type Prompter interface {
	// SelectMedia affiche une liste de résultats de recherche et retourne le choix de l'utilisateur
	SelectMedia(results []media.SearchResult) (*media.SearchResult, error)

	// SelectSourceType demande à l'utilisateur de choisir le type de source
	SelectSourceType() (mediainfo.SourceType, error)

	// AskForInput demande une entrée texte à l'utilisateur
	AskForInput(prompt string) (string, error)

	// Confirm demande une confirmation oui/non
	Confirm(prompt string) (bool, error)

	// ShowProgress affiche une barre de progression
	ShowProgress(current, total int, message string)

	// ShowMessage affiche un message
	ShowMessage(message string)

	// ShowError affiche une erreur
	ShowError(message string)
}
