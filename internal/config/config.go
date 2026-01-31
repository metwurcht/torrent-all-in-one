package config

// Config contient les paramètres de configuration par défaut de l'application
type Config struct {
	GroupName   string
	SkipTorrent bool
	NoRename    bool
}

// Default retourne la configuration par défaut
func Default() Config {
	return Config{
		GroupName:   "TORRENT-AIO",
		SkipTorrent: false,
		NoRename:    false,
	}
}
