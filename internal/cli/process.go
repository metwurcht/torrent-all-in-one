package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/metwurcht/torrent-all-in-one/internal/media/movie"
	"github.com/metwurcht/torrent-all-in-one/internal/media/music"
	"github.com/metwurcht/torrent-all-in-one/internal/media/tvshow"
	"github.com/metwurcht/torrent-all-in-one/internal/processor"
	"github.com/metwurcht/torrent-all-in-one/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runProcess(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	inputPath := args[0]

	// Créer le prompter interactif
	prompter := ui.NewInteractivePrompter()

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

	outDir := outputDir
	if outDir == "" {
		outDir = viper.GetString("output")
	}

	// Préparer les options
	opts := &processor.Options{
		OutputDir:        outDir,
		GroupName:        group,
		SkipTorrent:      skipTorrentFlag,
		NoRename:         noRenameFlag,
		SourceType:       nil,
		ProgressReporter: &ConsoleReporter{},
	}

	// Détecter si l'entrée est un dossier (série TV) ou un fichier (film)
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("chemin introuvable: %s", inputPath)
	}

	if info.IsDir() {
		// Détecter le type de contenu du dossier
		dirType := processor.DetectDirectoryType(inputPath)

		switch dirType {
		case "music":
			// Dossier contenant des fichiers audio → Album de musique
			pipeline := music.NewPipeline(group)
			proc := processor.NewProcessor(pipeline, prompter)
			_, err := proc.ProcessMusicDirectory(ctx, inputPath, opts)
			return err
		default:
			// Dossier contenant des fichiers vidéo → Série TV
			pipeline := tvshow.NewPipeline(group)
			proc := processor.NewProcessor(pipeline, prompter)
			_, err := proc.ProcessDirectory(ctx, inputPath, opts)
			return err
		}
	}

	// Fichier → Film
	pipeline := movie.NewPipeline(group)
	proc := processor.NewProcessor(pipeline, prompter)
	_, err = proc.Process(ctx, inputPath, opts)
	return err
}
