package ui

import (
	"github.com/metwurcht/torrent-all-in-one/internal/media"
	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
)

// SilentPrompter implémente Prompter pour une utilisation non-interactive (API/automation)
type SilentPrompter struct {
	defaultMediaIndex int
	defaultInput      string
	defaultConfirm    bool
	defaultSourceType mediainfo.SourceType
}

// NewSilentPrompter crée un nouveau prompter silencieux
func NewSilentPrompter() *SilentPrompter {
	return &SilentPrompter{
		defaultMediaIndex: 0,
		defaultConfirm:    true,
	}
}

// SetDefaultMediaIndex définit l'index par défaut pour la sélection
func (p *SilentPrompter) SetDefaultMediaIndex(index int) {
	p.defaultMediaIndex = index
}

// SetDefaultInput définit l'entrée par défaut
func (p *SilentPrompter) SetDefaultInput(input string) {
	p.defaultInput = input
}

// SetDefaultConfirm définit la confirmation par défaut
func (p *SilentPrompter) SetDefaultConfirm(confirm bool) {
	p.defaultConfirm = confirm
}

// SelectMedia retourne automatiquement le premier résultat (ou l'index configuré)
func (p *SilentPrompter) SelectMedia(results []media.SearchResult) (*media.SearchResult, error) {
	if len(results) == 0 {
		return nil, nil
	}

	index := p.defaultMediaIndex
	if index >= len(results) {
		index = 0
	}

	return &results[index], nil
}

// SelectSourceType retourne le type de source par défaut
func (p *SilentPrompter) SelectSourceType() (mediainfo.SourceType, error) {
	if p.defaultSourceType == "" {
		return mediainfo.SourceWEB, nil
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
