package gui

import (
	"context"
	"fmt"
	"sync"

	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GUIPrompter gère les interactions utilisateur via le GUI
type GUIPrompter struct {
	ctx             context.Context
	movieResponse   chan *tmdb.Movie
	sourceResponse  chan mediainfo.SourceType
	confirmResponse chan bool
	currentMovies   []tmdb.Movie
	mu              sync.Mutex
}

// NewGUIPrompter crée un nouveau prompter pour le GUI
func NewGUIPrompter(ctx context.Context) *GUIPrompter {
	return &GUIPrompter{
		ctx:             ctx,
		movieResponse:   make(chan *tmdb.Movie, 1),
		sourceResponse:  make(chan mediainfo.SourceType, 1),
		confirmResponse: make(chan bool, 1),
	}
}

// SelectMovie affiche les résultats de recherche et retourne le choix de l'utilisateur
func (g *GUIPrompter) SelectMovie(results []tmdb.Movie) (*tmdb.Movie, error) {
	// Stocker les films pour la réponse
	g.currentMovies = results

	// Préparer les données pour le frontend
	movies := make([]map[string]interface{}, len(results))
	for i, movie := range results {
		movies[i] = map[string]interface{}{
			"id":            movie.ID,
			"title":         movie.Title,
			"originalTitle": movie.OriginalTitle,
			"releaseDate":   movie.ReleaseDate,
			"overview":      movie.Overview,
			"posterPath":    movie.PosterURL("w185"),
		}
	}

	// Envoyer les films au frontend
	runtime.EventsEmit(g.ctx, "movie-selection-request", movies)

	// Attendre la réponse du frontend
	selectedMovie := <-g.movieResponse
	if selectedMovie == nil {
		return nil, fmt.Errorf("aucune sélection")
	}

	return selectedMovie, nil
}

// OnMovieSelected est appelé par le frontend quand l'utilisateur sélectionne un film
func (g *GUIPrompter) OnMovieSelected(movieID int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i := range g.currentMovies {
		if g.currentMovies[i].ID == movieID {
			g.movieResponse <- &g.currentMovies[i]
			return
		}
	}
	g.movieResponse <- nil
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
