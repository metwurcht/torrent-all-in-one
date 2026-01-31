package processor

import "github.com/metwurcht/torrent-all-in-one/internal/mediainfo"

// Options contient les options de configuration pour le traitement
type Options struct {
	// OutputDir est le dossier de sortie pour les fichiers générés
	// Si vide, utilise le même dossier que le fichier d'entrée
	OutputDir string

	// GroupName est le nom du groupe de release
	GroupName string

	// SkipTorrent indique s'il faut ignorer la génération du fichier torrent
	SkipTorrent bool

	// NoRename indique s'il faut conserver le nom du fichier original
	NoRename bool

	// SourceType est le type de source (BluRay, WEB, etc.)
	// Si nil, sera demandé à l'utilisateur via le Prompter
	SourceType *mediainfo.SourceType

	// ProgressReporter permet de suivre la progression du traitement
	// Si nil, utilise un SilentReporter
	ProgressReporter ProgressReporter
}

// DefaultOptions retourne les options par défaut
func DefaultOptions() *Options {
	return &Options{
		GroupName:   "TORRENT-AIO",
		SkipTorrent: false,
		NoRename:    false,
	}
}
