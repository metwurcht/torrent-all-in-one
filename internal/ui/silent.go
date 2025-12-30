package ui

import (
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
)

// SilentPrompter implémente Prompter pour une utilisation non-interactive (API/automation)
type SilentPrompter struct {
	defaultMovieIndex int
	defaultInput      string
	defaultConfirm    bool
	defaultSourceType string
}

// NewSilentPrompter crée un nouveau prompter silencieux
func NewSilentPrompter() *SilentPrompter {
	return &SilentPrompter{
		defaultMovieIndex: 0,
		defaultConfirm:    true,
	}
}

// SetDefaultMovieIndex définit l'index par défaut pour la sélection de films
func (p *SilentPrompter) SetDefaultMovieIndex(index int) {
	p.defaultMovieIndex = index
}

// SetDefaultInput définit l'entrée par défaut
func (p *SilentPrompter) SetDefaultInput(input string) {
	p.defaultInput = input
}

// SetDefaultConfirm définit la confirmation par défaut
func (p *SilentPrompter) SetDefaultConfirm(confirm bool) {
	p.defaultConfirm = confirm
}

// SelectMovie retourne automatiquement le premier film (ou l'index configuré)
func (p *SilentPrompter) SelectMovie(movies []tmdb.Movie) (*tmdb.Movie, error) {
	if len(movies) == 0 {
		return nil, nil
	}

	index := p.defaultMovieIndex
	if index >= len(movies) {
		index = 0
	}

	return &movies[index], nil
}

// SelectSourceType retourne le type de source par défaut
func (p *SilentPrompter) SelectSourceType() (string, error) {
	if p.defaultSourceType == "" {
		return "WEB-DL", nil
	}
	return p.defaultSourceType, nil
}

// AskForInput retourne l'entrée par défaut
func (p *SilentPrompter) AskForInput(prompt string) (string, error) {
	return p.defaultInput, nil
}

// Confirm retourne la confirmation par défaut
func (p *SilentPrompter) Confirm(prompt string) (bool, error) {
	return p.defaultConfirm, nil
}

// ShowProgress ne fait rien en mode silencieux
func (p *SilentPrompter) ShowProgress(current, total int, message string) {
	// Silencieux
}

// ShowMessage ne fait rien en mode silencieux
func (p *SilentPrompter) ShowMessage(message string) {
	// Silencieux
}

// ShowError ne fait rien en mode silencieux
func (p *SilentPrompter) ShowError(message string) {
	// Silencieux
}
