package gui

import (
	"context"
	"fmt"
	"os"

	"github.com/metwurcht/torrent-all-in-one/internal/config"
	"github.com/metwurcht/torrent-all-in-one/internal/media"
	"github.com/metwurcht/torrent-all-in-one/internal/media/movie"
	"github.com/metwurcht/torrent-all-in-one/internal/media/music"
	"github.com/metwurcht/torrent-all-in-one/internal/media/tvshow"
	"github.com/metwurcht/torrent-all-in-one/internal/processor"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	currentPrompter *GUIPrompter
}

// NewApp crée une nouvelle instance de l'application
func NewApp() *App {
	return &App{}
}

// Startup est appelé au démarrage de l'application
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// ProcessFileRequest représente une requête de traitement
type ProcessFileRequest struct {
	FilePath    string  `json:"filePath"`
	GroupName   string  `json:"groupName"`
	OutputDir   string  `json:"outputDir"`
	SkipTorrent bool    `json:"skipTorrent"`
	NoRename    bool    `json:"noRename"`
	SourceType  *string `json:"sourceType"`
	MediaType   string  `json:"mediaType"` // "movie", "tvshow", "music" — defaults to "movie"
}

// ProcessFileResponse représente la réponse du traitement
type ProcessFileResponse struct {
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
	ReleaseName      string `json:"releaseName,omitempty"`
	Title            string `json:"title,omitempty"`
	NFOPath          string `json:"nfoPath,omitempty"`
	PresentationPath string `json:"presentationPath,omitempty"`
	TorrentPath      string `json:"torrentPath,omitempty"`
}

// createPipeline creates the appropriate media pipeline based on media type
func createPipeline(mediaType string, groupName string) *media.Pipeline {
	switch media.Type(mediaType) {
	case media.TypeTVShow:
		return tvshow.NewPipeline(groupName)
	case media.TypeMusic:
		return music.NewPipeline(groupName)
	default:
		return movie.NewPipeline(groupName)
	}
}

// ProcessFile traite un fichier vidéo
func (a *App) ProcessFile(req ProcessFileRequest) ProcessFileResponse {
	// Créer un prompter GUI
	a.currentPrompter = NewGUIPrompter(a.ctx)

	// Déterminer le type de pipeline
	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = string(media.TypeMovie)
	}

	// Détecter automatiquement si c'est un dossier → série TV ou musique
	info, err := os.Stat(req.FilePath)
	if err != nil {
		return ProcessFileResponse{
			Success: false,
			Error:   fmt.Sprintf("chemin introuvable: %s", req.FilePath),
		}
	}
	if info.IsDir() {
		dirType := processor.DetectDirectoryType(req.FilePath)
		switch dirType {
		case "music":
			mediaType = string(media.TypeMusic)
		default:
			mediaType = string(media.TypeTVShow)
		}
	}

	// Créer le pipeline et le processor
	pipeline := createPipeline(mediaType, req.GroupName)
	proc := processor.NewProcessor(pipeline, a.currentPrompter)

	// Créer un reporter GUI
	reporter := NewGUIReporter(a.ctx)

	// Préparer les options
	opts := &processor.Options{
		GroupName:        req.GroupName,
		OutputDir:        req.OutputDir,
		SkipTorrent:      req.SkipTorrent,
		NoRename:         req.NoRename,
		ProgressReporter: reporter,
	}

	// Convertir le sourceType si fourni
	if req.SourceType != nil && *req.SourceType != "" {
		// TODO: Convertir la string en mediainfo.SourceType
	}

	// Exécuter le traitement
	var result *processor.Result
	if info.IsDir() {
		if media.Type(mediaType) == media.TypeMusic {
			result, err = proc.ProcessMusicDirectory(context.Background(), req.FilePath, opts)
		} else {
			result, err = proc.ProcessDirectory(context.Background(), req.FilePath, opts)
		}
	} else {
		result, err = proc.Process(context.Background(), req.FilePath, opts)
	}
	if err != nil {
		return ProcessFileResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	return ProcessFileResponse{
		Success:          true,
		ReleaseName:      result.ReleaseName,
		Title:            result.Metadata.GetTitle(),
		NFOPath:          result.NFOPath,
		PresentationPath: result.PresentationPath,
		TorrentPath:      result.TorrentPath,
	}
}

// SelectFile ouvre un dialogue de sélection de fichier
func (a *App) SelectFile() string {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Sélectionner un fichier vidéo",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Fichiers vidéo",
				Pattern:     "*.mkv;*.mp4;*.avi;*.mov;*.m4v",
			},
			{
				DisplayName: "Tous les fichiers",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return ""
	}

	return filePath
}

// SelectDirectory ouvre un dialogue de sélection de dossier
func (a *App) SelectDirectory() string {
	dirPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Sélectionner un dossier de sortie",
	})

	if err != nil {
		return ""
	}

	return dirPath
}

// GetDefaultConfig retourne la configuration par défaut
func (a *App) GetDefaultConfig() map[string]interface{} {
	cfg := config.Default()
	return map[string]interface{}{
		"groupName":   cfg.GroupName,
		"skipTorrent": cfg.SkipTorrent,
		"noRename":    cfg.NoRename,
	}
}

// Greet retourne un message de bienvenue (pour tester)
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Bonjour %s! Bienvenue dans Torrent All-In-One", name)
}

// SelectMedia est appelé par le frontend quand l'utilisateur sélectionne un média
func (a *App) SelectMedia(mediaID int) {
	if a.currentPrompter != nil {
		a.currentPrompter.OnMediaSelected(mediaID)
	}
}

// SelectSourceType est appelé par le frontend quand l'utilisateur sélectionne un type de source
func (a *App) SelectSourceType(sourceType string) {
	if a.currentPrompter != nil {
		a.currentPrompter.OnSourceTypeSelected(sourceType)
	}
}

// RespondConfirm est appelé par le frontend en réponse à une demande de confirmation
func (a *App) RespondConfirm(response bool) {
	if a.currentPrompter != nil {
		a.currentPrompter.OnConfirmResponse(response)
	}
}
