package cli

import (
	"context"

	"github.com/metwurcht/torrent-all-in-one/internal/processor"
	"github.com/metwurcht/torrent-all-in-one/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func runProcess(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	inputFile := args[0]

	// Créer le prompter interactif
	prompter := ui.NewInteractivePrompter()

	// Créer le processor
	proc := processor.NewProcessor(prompter)

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
		SourceType:       nil, // Sera demandé par le processor si nécessaire
		ProgressReporter: &ConsoleReporter{},
	}

	// Exécuter le traitement
	_, err := proc.Process(ctx, inputFile, opts)
	return err
}
