package cli

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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runProcess(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	inputFile := args[0]

	// Vérifier que le fichier existe
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("fichier introuvable: %s", inputFile)
	}

	absPath, err := filepath.Abs(inputFile)
	if err != nil {
		return fmt.Errorf("erreur chemin absolu: %w", err)
	}

	// Créer les services (plus besoin de clé API - on fait du scraping)
	tmdbClient := tmdb.NewClient()
	analyzer := mediainfo.NewAnalyzer()
	prompter := ui.NewInteractivePrompter()

	// Lancer l'analyse du fichier en parallèle
	var mediaInfo *mediainfo.MediaInfo
	var mediaErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("🔍 Analyse du fichier en cours...")
		mediaInfo, mediaErr = analyzer.Analyze(absPath)
	}()

	// Identification TMDB
	fmt.Println("🎬 Identification du film...")
	filename := filepath.Base(inputFile)
	movie, err := identifyMovie(ctx, tmdbClient, prompter, filename)
	if err != nil {
		return fmt.Errorf("erreur identification: %w", err)
	}

	// Attendre la fin de l'analyse
	wg.Wait()
	if mediaErr != nil {
		return fmt.Errorf("erreur analyse fichier: %w", mediaErr)
	}

	fmt.Println("✅ Film identifié:", movie.OriginalTitle)
	fmt.Println("✅ Analyse terminée")

	// Déterminer le dossier de sortie
	outDir := outputDir
	if outDir == "" {
		outDir = viper.GetString("output")
	}
	if outDir == "" {
		outDir = filepath.Dir(absPath)
	}

	// Récupérer la configuration avec priorité: flag CLI > config > défaut
	group := groupName
	if group == "" || !cmd.Flags().Changed("group") {
		group = viper.GetString("group_name")
	}

	skipTorrentFlag := skipTorrent
	if !cmd.Flags().Changed("skip-torrent") {
		skipTorrentFlag = viper.GetBool("skip_torrent")
	}

	noRenameFlag := noRename
	if !cmd.Flags().Changed("no-rename") {
		noRenameFlag = viper.GetBool("no_rename")
	}

	var newName string
	var newPath string

	if noRenameFlag {
		// Utiliser le nom de fichier actuel sans renommer
		newName = filepath.Base(absPath)
		newName = newName[:len(newName)-len(filepath.Ext(absPath))] // Retirer l'extension
		newPath = absPath
		fmt.Printf("📝 Utilisation du nom actuel: %s\n", newName)
	} else {
		// Demander le type de source à l'utilisateur
		sourceType, err := prompter.SelectSourceType()
		if err != nil {
			return fmt.Errorf("erreur sélection source: %w", err)
		}
		// Définir le type de source dans mediaInfo
		mediaInfo.SourceType = sourceType
		// Générer un nouveau nom et renommer
		ren := renamer.NewRenamer(group)
		newName = ren.GenerateName(movie, mediaInfo)
		newPath = filepath.Join(outDir, newName+filepath.Ext(absPath))

		fmt.Printf("📝 Renommage: %s\n", newName)
		if err := os.Rename(absPath, newPath); err != nil {
			return fmt.Errorf("erreur renommage: %w", err)
		}

		// Mettre à jour le chemin dans mediaInfo après le renommage
		mediaInfo.FilePath = newPath
	}

	// Générer le NFO
	fmt.Println("📄 Génération du NFO...")
	nfoGen := nfo.NewGenerator(group)
	nfoContent := nfoGen.Generate(movie, mediaInfo, newName+filepath.Ext(absPath))
	nfoPath := filepath.Join(outDir, newName+".nfo")
	if err := os.WriteFile(nfoPath, []byte(nfoContent), 0644); err != nil {
		return fmt.Errorf("erreur écriture NFO: %w", err)
	}
	fmt.Printf("✅ NFO créé: %s\n", nfoPath)

	fmt.Println("📋 Génération de la présentation...")
	// Générer la présentation BBCode
	presentationContent := presenter.GenerateBBcode(movie, mediaInfo)
	presentationPath := filepath.Join(outDir, newName+".bbcode")
	if err := os.WriteFile(presentationPath, []byte(presentationContent), 0644); err != nil {
		return fmt.Errorf("erreur écriture présentation: %w", err)
	}
	fmt.Printf("📋 Présentation créée: %s\n", presentationPath)

	// Générer le torrent
	if !skipTorrentFlag {
		fmt.Println("🧲 Génération du torrent...")
		torrentGen := torrent.NewGenerator()
		torrentPath := filepath.Join(outDir, newName+".torrent")
		if err := torrentGen.Create(newPath, torrentPath); err != nil {
			return fmt.Errorf("erreur génération torrent: %w", err)
		}
		fmt.Printf("✅ Torrent créé: %s\n", torrentPath)
	}

	fmt.Println("\n🎉 Traitement terminé avec succès!")
	return nil
}

func identifyMovie(ctx context.Context, client *tmdb.Client, prompter ui.Prompter, filename string) (*tmdb.Movie, error) {
	// Extraire les mots-clés du nom de fichier
	keywords := tmdb.ExtractKeywords(filename)

	for {
		// Rechercher sur TMDB
		results, err := client.SearchMovie(ctx, keywords)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			fmt.Println("Aucun résultat trouvé.")
		} else {
			// Afficher les résultats
			choice, err := prompter.SelectMovie(results)
			if err == nil {
				// Récupérer les détails complets du film
				return client.GetMovieDetails(ctx, choice.ID)
			}
		}

		// Demander une nouvelle recherche ou un ID direct
		input, err := prompter.AskForInput("Entrez un nouveau terme de recherche ou un ID TMDB (ex: id:12345):")
		if err != nil {
			return nil, err
		}

		// Vérifier si c'est un ID direct
		if id, ok := tmdb.ParseDirectID(input); ok {
			return client.GetMovieDetails(ctx, id)
		}

		// Nouvelle recherche avec les termes fournis
		keywords = input
	}
}
