package cli

import (
	"fmt"
	"os"

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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "fichier de configuration (défaut: $HOME/.torrent-aio.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".torrent-aio")
	}

	viper.SetEnvPrefix("TORRENT_AIO")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Configuration:", viper.ConfigFileUsed())
	}
}
