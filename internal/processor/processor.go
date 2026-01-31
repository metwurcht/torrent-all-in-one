package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/metwurcht/torrent-all-in-one/internal/mediainfo"
	"github.com/metwurcht/torrent-all-in-one/internal/nfo"
	"github.com/metwurcht/torrent-all-in-one/internal/presenter"
	"github.com/metwurcht/torrent-all-in-one/internal/renamer"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
	"github.com/metwurcht/torrent-all-in-one/internal/torrent"
	"github.com/metwurcht/torrent-all-in-one/internal/ui"
)

// Result contient les résultats du traitement
type Result struct {
	Movie              *tmdb.Movie
	MediaInfo          *mediainfo.MediaInfo
	ReleaseName        string
	NewFilePath        string
	NFOPath            string
	PresentationPath   string
	TorrentPath        string
	SourceTypeSelected mediainfo.SourceType
}

// Processor gère le workflow complet de traitement d'un fichier vidéo
type Processor struct {
	tmdbClient *tmdb.Client
	analyzer   *mediainfo.Analyzer
	prompter   ui.Prompter
}

// NewProcessor crée un nouveau processeur avec les dépendances nécessaires
func NewProcessor(prompter ui.Prompter) *Processor {
	return &Processor{
		tmdbClient: tmdb.NewClient(),
		analyzer:   mediainfo.NewAnalyzer(),
		prompter:   prompter,
	}
}

// Process traite un fichier vidéo selon les options fournies
func (p *Processor) Process(ctx context.Context, inputFile string, opts *Options) (*Result, error) {
	// Utiliser un reporter par défaut si non fourni
	reporter := opts.ProgressReporter
	if reporter == nil {
		reporter = &SilentReporter{}
	}

	// Vérifier que le fichier existe
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("fichier introuvable: %s", inputFile)
	}

	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		return nil, fmt.Errorf("erreur chemin absolu: %w", err)
	}

	// Lancer l'analyse du fichier en parallèle
	var mediaInfo *mediainfo.MediaInfo
	var mediaErr error
	var wg sync.WaitGroup

	reporter.OnProgress("🔍 Analyse du fichier en cours...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		mediaInfo, mediaErr = p.analyzer.Analyze(absPath)
	}()

	// Identification TMDB
	reporter.OnProgress("🎬 Identification du film...")
	filename := filepath.Base(inputFile)
	movie, err := p.identifyMovie(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("erreur identification: %w", err)
	}

	reporter.OnComplete("✅ Film identifié: " + movie.OriginalTitle)

	// Attendre la fin de l'analyse
	wg.Wait()
	if mediaErr != nil {
		return nil, fmt.Errorf("erreur analyse fichier: %w", mediaErr)
	}

	reporter.OnComplete("✅ Analyse terminée")

	// Déterminer le dossier de sortie
	outDir := opts.OutputDir
	if outDir == "" {
		outDir = filepath.Dir(absPath)
	}

	var newName string
	var newPath string
	var sourceType mediainfo.SourceType

	if opts.NoRename {
		// Utiliser le nom de fichier actuel sans renommer
		newName = filepath.Base(absPath)
		newName = newName[:len(newName)-len(filepath.Ext(absPath))] // Retirer l'extension
		newPath = absPath
		// Essayer de détecter le type de source depuis le nom
		sourceType = mediaInfo.SourceType
		reporter.OnProgress("📝 Utilisation du nom actuel: " + newName)
	} else {
		// Demander le type de source si non fourni
		if opts.SourceType == nil {
			selectedSourceType, err := p.prompter.SelectSourceType()
			if err != nil {
				return nil, fmt.Errorf("erreur sélection source: %w", err)
			}
			sourceType = selectedSourceType
		} else {
			sourceType = *opts.SourceType
		}

		// Définir le type de source dans mediaInfo
		mediaInfo.SourceType = sourceType

		// Générer un nouveau nom et renommer
		ren := renamer.NewRenamer(opts.GroupName)
		newName = ren.GenerateName(movie, mediaInfo)
		newPath = filepath.Join(outDir, newName+filepath.Ext(absPath))

		reporter.OnProgress("📝 Renommage: " + newName)
		if err := os.Rename(absPath, newPath); err != nil {
			return nil, fmt.Errorf("erreur renommage: %w", err)
		}

		// Mettre à jour le chemin dans mediaInfo après le renommage
		mediaInfo.FilePath = newPath
	}

	result := &Result{
		Movie:              movie,
		MediaInfo:          mediaInfo,
		ReleaseName:        newName,
		NewFilePath:        newPath,
		SourceTypeSelected: sourceType,
	}

	// Générer le NFO
	reporter.OnProgress("📄 Génération du NFO...")
	nfoGen := nfo.NewGenerator(opts.GroupName)
	nfoContent := nfoGen.Generate(movie, mediaInfo, newName+filepath.Ext(absPath))
	nfoPath := filepath.Join(outDir, newName+".nfo")
	if err := os.WriteFile(nfoPath, []byte(nfoContent), 0644); err != nil {
		return nil, fmt.Errorf("erreur écriture NFO: %w", err)
	}
	result.NFOPath = nfoPath
	reporter.OnComplete("✅ NFO créé: " + nfoPath)

	// Générer la présentation BBCode
	reporter.OnProgress("📋 Génération de la présentation...")
	presentationContent := presenter.GenerateBBcode(movie, mediaInfo)
	presentationPath := filepath.Join(outDir, newName+".bbcode")
	if err := os.WriteFile(presentationPath, []byte(presentationContent), 0644); err != nil {
		return nil, fmt.Errorf("erreur écriture présentation: %w", err)
	}
	result.PresentationPath = presentationPath
	reporter.OnComplete("📋 Présentation créée: " + presentationPath)

	// Générer le torrent
	if !opts.SkipTorrent {
		reporter.OnProgress("🧲 Génération du torrent...")
		torrentGen := torrent.NewGenerator()
		torrentPath := filepath.Join(outDir, newName+".torrent")
		if err := torrentGen.Create(newPath, torrentPath); err != nil {
			return nil, fmt.Errorf("erreur génération torrent: %w", err)
		}
		result.TorrentPath = torrentPath
		reporter.OnComplete("✅ Torrent créé: " + torrentPath)
	}

	reporter.OnComplete("\n🎉 Traitement terminé avec succès!")

	return result, nil
}

// identifyMovie identifie un film via TMDB en utilisant le prompter pour l'interaction
func (p *Processor) identifyMovie(ctx context.Context, filename string) (*tmdb.Movie, error) {
	// Extraire les mots-clés du nom de fichier
	keywords := tmdb.ExtractKeywords(filename)

	for {
		// Rechercher sur TMDB
		results, err := p.tmdbClient.SearchMovie(ctx, keywords)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			// Aucun résultat, demander une nouvelle recherche
		} else {
			// Afficher les résultats
			choice, err := p.prompter.SelectMovie(results)
			if err == nil {
				// Récupérer les détails complets du film
				return p.tmdbClient.GetMovieDetails(ctx, choice.ID)
			}
		}

		// Demander une nouvelle recherche ou un ID direct
		input, err := p.prompter.AskForInput("Entrez un nouveau terme de recherche ou un ID TMDB (ex: id:12345):")
		if err != nil {
			return nil, err
		}

		// Vérifier si c'est un ID direct
		if id, ok := tmdb.ParseDirectID(input); ok {
			return p.tmdbClient.GetMovieDetails(ctx, id)
		}

		// Nouvelle recherche avec les termes fournis
		keywords = input
	}
}
