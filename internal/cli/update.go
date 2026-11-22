package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/metwurcht/torrent-all-in-one/internal/config"
	"github.com/metwurcht/torrent-all-in-one/internal/update"
	"github.com/spf13/cobra"
)

// updateNotice reçoit le message à afficher si une nouvelle version existe.
var updateNotice chan string

// startUpdateCheck lance en arrière-plan la vérification de mise à jour.
func startUpdateCheck(cmd *cobra.Command) {
	version := cmd.Root().Version
	if cmd.Name() == "update" || !update.IsRelease(version) {
		return
	}
	if cfg, err := loadConfig(); err == nil && !cfg.Bool(config.KeyCheckUpdates) {
		return
	}

	updateNotice = make(chan string, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		latest, err := update.Latest(ctx)
		if err != nil || !update.IsNewer(latest.Tag, version) {
			updateNotice <- ""
			return
		}
		updateNotice <- fmt.Sprintf(`
🆕 Nouvelle version disponible : %s (version actuelle : %s)
   Mettez à jour avec : %s update
   Nouveautés : %s`, latest.Tag, version, cmd.Root().Name(), latest.URL)
	}()
}

// printUpdateNotice affiche le message de mise à jour s'il est disponible
// (attente courte pour ne pas retarder la fin du programme).
func printUpdateNotice() {
	if updateNotice == nil {
		return
	}
	select {
	case msg := <-updateNotice:
		if msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
	case <-time.After(1500 * time.Millisecond):
	}
}

func newUpdateCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Télécharger et installer la dernière version",
		Long: `Télécharge la dernière version publiée sur GitHub et remplace l'exécutable actuel.

Si l'exécutable est installé dans un dossier système (ex: /usr/local/bin),
lancez la commande avec les droits administrateur (sudo torrent-aio update).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			ctx := context.Background()
			current := cmd.Root().Version

			fmt.Println("🔎 Recherche de la dernière version...")
			latest, err := update.Latest(ctx)
			if err != nil {
				return err
			}

			if !force && update.IsRelease(current) && !update.IsNewer(latest.Tag, current) {
				fmt.Printf("✅ Vous utilisez déjà la dernière version (%s)\n", current)
				return nil
			}

			fmt.Printf("📦 Mise à jour %s → %s\n", current, latest.Tag)
			path, err := update.Install(ctx, latest.Tag, os.Stdout)
			if err != nil {
				return err
			}
			fmt.Printf("✅ torrent-aio %s installé dans %s\n", latest.Tag, path)
			fmt.Printf("   Nouveautés : %s\n", latest.URL)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Réinstaller même si la version est déjà à jour")
	return cmd
}
