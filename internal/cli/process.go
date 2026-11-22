package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/metwurcht/torrent-all-in-one/internal/config"
	"github.com/metwurcht/torrent-all-in-one/internal/media/movie"
	"github.com/metwurcht/torrent-all-in-one/internal/media/music"
	"github.com/metwurcht/torrent-all-in-one/internal/media/tvshow"
	"github.com/metwurcht/torrent-all-in-one/internal/processor"
	"github.com/metwurcht/torrent-all-in-one/internal/ui"
	"github.com/spf13/cobra"
)

func runProcess(cmd *cobra.Command, args []string) error {
	// Les erreurs de traitement ne doivent pas réafficher l'aide
	cmd.SilenceUsage = true

	ctx := context.Background()
	inputPath := args[0]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Priorité: flag CLI > fichier de config > variable d'environnement > défaut
	group := cfg.String(config.KeyGroupName)
	if cmd.Flags().Changed("group") {
		group = groupName
	}

	skipTorrentFlag := cfg.Bool(config.KeySkipTorrent)
	if cmd.Flags().Changed("skip-torrent") {
		skipTorrentFlag = skipTorrent
	}

	noRenameFlag := cfg.Bool(config.KeyNoRename)
	if cmd.Flags().Changed("no-rename") {
		noRenameFlag = noRename
	}

	outDir := cfg.String(config.KeyOutput)
	if cmd.Flags().Changed("output") {
		outDir = outputDir
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

	// Détecter si l'entrée est un dossier (série TV / musique) ou un fichier (film)
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("chemin introuvable: %s", inputPath)
	}

	prompter := ui.NewInteractivePrompter()

	if info.IsDir() && processor.DetectDirectoryType(inputPath) == "music" {
		// Dossier contenant des fichiers audio → Album de musique
		pipeline := music.NewPipeline(group, cfg.String(config.KeyTheAudioDBAPIKey))
		proc := processor.NewProcessor(pipeline, prompter)
		_, err := proc.ProcessMusicDirectory(ctx, inputPath, opts)
		return err
	}

	// Films et séries TV nécessitent une clé TMDB
	tmdbKey, err := requireKey(cmd, cfg, config.KeyTMDBAPIKey, "TMDB", "https://www.themoviedb.org/settings/api")
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Dossier contenant des fichiers vidéo → Série TV
		pipeline := tvshow.NewPipeline(group, tmdbKey)
		proc := processor.NewProcessor(pipeline, prompter)
		_, err := proc.ProcessDirectory(ctx, inputPath, opts)
		return err
	}

	// Fichier → Film
	pipeline := movie.NewPipeline(group, tmdbKey)
	proc := processor.NewProcessor(pipeline, prompter)
	_, err = proc.Process(ctx, inputPath, opts)
	return err
}

// requireKey retourne la clé API demandée ou une erreur expliquant comment la configurer.
func requireKey(cmd *cobra.Command, cfg *config.Store, name, service, signupURL string) (string, error) {
	if v := cfg.String(name); v != "" {
		return v, nil
	}
	key, _ := config.LookupKey(name)
	return "", fmt.Errorf(`clé API %s introuvable (ni dans le fichier de configuration, ni dans la variable d'environnement %s).

Obtenez une clé sur %s puis configurez-la avec :
  %s config set %s <votre_clé>`, service, key.Env, signupURL, cmd.Root().Name(), name)
}
