package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "torrent-aio",
	Short: "Torrent All-In-One - Outil de préparation de releases",
	Long: `Torrent All-In-One est un outil CLI qui permet de:
- Identifier un film via TMDB (scraping)
- Analyser les métadonnées d'un fichier vidéo
- Générer un fichier NFO
- Renommer le fichier selon les conventions warez
- Générer un fichier torrent

Exemple d'utilisation:
  torrent-aio process movie.mkv`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "fichier de configuration (défaut: $HOME/.config/torrent-aio.yml)")
}

func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Ajouter les chemins de recherche corrects
		viper.AddConfigPath(home)
		viper.AddConfigPath(filepath.Join(home, ".config"))
		viper.SetConfigType("yml")
		viper.SetConfigName("torrent-aio")
	}

	viper.SetEnvPrefix("TORRENT_AIO")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			fmt.Fprintf(os.Stderr, "⚠️  Erreur lecture config: %v\n", err)
		}
	}
}
