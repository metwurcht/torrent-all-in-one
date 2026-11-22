package cli

import (
	"github.com/metwurcht/torrent-all-in-one/internal/config"
	"github.com/spf13/cobra"
)

var cfgFile string

// Variables pour les flags CLI
var (
	outputDir   string
	groupName   string
	skipTorrent bool
	noRename    bool
)

var rootCmd = &cobra.Command{
	Use:   "torrent-aio <fichier_video|dossier>",
	Short: "Torrent All-In-One - Outil de préparation de releases",
	Long: `Torrent All-In-One est un outil CLI qui permet de:
- Identifier un film, une série TV (TMDB) ou un album (TheAudioDB)
- Analyser les métadonnées d'un ou plusieurs fichiers
- Générer un fichier NFO
- Renommer le(s) fichier(s) selon les conventions warez
- Générer un fichier torrent
- Générer une présentation BBCode

Utilisation:
  torrent-aio movie.mkv          # Traiter un film (fichier unique)
  torrent-aio ./ma-serie/        # Traiter une série TV (dossier d'épisodes)
  torrent-aio ./mon-album/       # Traiter un album (dossier de fichiers audio)

Configuration:
  torrent-aio config set tmdb_api_key <votre_clé>
  torrent-aio config list`,
	Args: cobra.ExactArgs(1),
	RunE: runProcess,
}

// Execute lance le CLI.
func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate("torrent-aio {{.Version}}\n")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "fichier de configuration (défaut: premier trouvé parmi ~/.config/torrent-aio.yml, ~/torrent-aio.yml)")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Dossier de sortie (défaut: même dossier que le fichier)")
	rootCmd.Flags().StringVarP(&groupName, "group", "g", "", "Nom du groupe de release")
	rootCmd.Flags().BoolVar(&skipTorrent, "skip-torrent", false, "Ne pas générer le fichier torrent")
	rootCmd.Flags().BoolVar(&noRename, "no-rename", false, "Ne pas renommer le fichier vidéo")

	rootCmd.AddCommand(newConfigCmd())
}

func loadConfig() (*config.Store, error) {
	return config.Load(cfgFile)
}
