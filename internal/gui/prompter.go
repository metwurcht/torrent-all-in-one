package gui

import (
	"context"
	"fmt"
	"sync"

	"github.com/metwurcht/torrent-all-in-one/internal/media"
	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GUIPrompter gère les interactions utilisateur via le GUI
type GUIPrompter struct {
	ctx             context.Context
	mediaResponse   chan *media.SearchResult
	sourceResponse  chan mediainfo.SourceType
	confirmResponse chan bool
	currentResults  []media.SearchResult
	mu              sync.Mutex
}

// NewGUIPrompter crée un nouveau prompter pour le GUI
func NewGUIPrompter(ctx context.Context) *GUIPrompter {
	return &GUIPrompter{
		ctx:             ctx,
		mediaResponse:   make(chan *media.SearchResult, 1),
		sourceResponse:  make(chan mediainfo.SourceType, 1),
		confirmResponse: make(chan bool, 1),
	}
}

// SelectMedia affiche les résultats de recherche et retourne le choix de l'utilisateur
func (g *GUIPrompter) SelectMedia(results []media.SearchResult) (*media.SearchResult, error) {
	// Stocker les résultats pour la réponse
	g.currentResults = results

	// Préparer les données pour le frontend
	items := make([]map[string]interface{}, len(results))
	for i, result := range results {
		items[i] = map[string]interface{}{
			"id":            result.ID,
			"title":         result.Title,
			"originalTitle": result.OriginalTitle,
			"year":          result.Year,
			"overview":      result.Overview,
			"posterPath":    result.PosterURL,
		}
	}

	// Envoyer les résultats au frontend
	runtime.EventsEmit(g.ctx, "media-selection-request", items)

	// Attendre la réponse du frontend
	selectedResult := <-g.mediaResponse
	if selectedResult == nil {
		return nil, fmt.Errorf("aucune sélection")
	}

	return selectedResult, nil
}

// OnMediaSelected est appelé par le frontend quand l'utilisateur sélectionne un média
func (g *GUIPrompter) OnMediaSelected(mediaID int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i := range g.currentResults {
		if g.currentResults[i].ID == mediaID {
			g.mediaResponse <- &g.currentResults[i]
			return
		}
	}
	g.mediaResponse <- nil
}

// SelectSourceType demande à l'utilisateur de sélectionner un type de source
func (g *GUIPrompter) SelectSourceType() (mediainfo.SourceType, error) {
	// Envoyer les options au frontend
	options := mediainfo.AllSourceTypes()
	sourceTypes := make([]map[string]interface{}, len(options))
	for i, opt := range options {
		sourceTypes[i] = map[string]interface{}{
			"value": opt.Value,
			"label": opt.Display,
		}
	}

	runtime.EventsEmit(g.ctx, "source-type-selection-request", sourceTypes)

	// Attendre la réponse du frontend
	selectedType := <-g.sourceResponse
	if selectedType == "" {
		return "", fmt.Errorf("aucune sélection")
	}

	return selectedType, nil
}

// OnSourceTypeSelected est appelé par le frontend quand l'utilisateur sélectionne un type
func (g *GUIPrompter) OnSourceTypeSelected(sourceType string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sourceResponse <- mediainfo.SourceType(sourceType)
}

// AskForInput demande une saisie utilisateur
func (g *GUIPrompter) AskForInput(prompt string) (string, error) {
	// TODO: Implémenter via événements personnalisés
	return "", fmt.Errorf("non implémenté dans le GUI")
}

// Confirm demande une confirmation oui/non à l'utilisateur
func (g *GUIPrompter) Confirm(message string) (bool, error) {
	runtime.EventsEmit(g.ctx, "confirm-request", message)
	result := <-g.confirmResponse
	return result, nil
}

// OnConfirmResponse est appelé par le frontend avec la réponse de confirmation
func (g *GUIPrompter) OnConfirmResponse(response bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.confirmResponse <- response
}

// ShowError affiche un message d'erreur à l'utilisateur
func (g *GUIPrompter) ShowError(message string) {
	runtime.MessageDialog(g.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   "Erreur",
		Message: message,
	})
}

// ShowMessage affiche un message d'information à l'utilisateur
func (g *GUIPrompter) ShowMessage(message string) {
	runtime.MessageDialog(g.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   "Information",
		Message: message,
	})
}

// ShowProgress affiche un message de progression (dans le GUI, utilisé via les événements)
func (g *GUIPrompter) ShowProgress(current, total int, message string) {
	// Dans le GUI, la progression est gérée via le GUIReporter et les événements
	// Cette méthode émet un événement personnalisé
	runtime.EventsEmit(g.ctx, "progress", map[string]interface{}{
		"current": current,
		"total":   total,
		"message": message,
	})
}
