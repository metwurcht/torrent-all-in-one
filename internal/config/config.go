// Package config gère les paramètres de l'application.
//
// Ordre de résolution d'un paramètre :
//  1. fichier de configuration (premier fichier trouvé parmi SearchPaths)
//  2. variable d'environnement associée
//  3. valeur par défaut
//
// Les flags de la ligne de commande, quand ils sont fournis, priment sur tout le reste
// (géré par le package cli).
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Kind est le type d'une valeur de configuration.
type Kind int

const (
	KindString Kind = iota
	KindBool
)

// Noms des paramètres de configuration.
const (
	KeyTMDBAPIKey       = "tmdb_api_key"
	KeyTheAudioDBAPIKey = "theaudiodb_api_key"
	KeyGroupName        = "group_name"
	KeyOutput           = "output"
	KeySkipTorrent      = "skip_torrent"
	KeyNoRename         = "no_rename"
	KeyCheckUpdates     = "check_updates"
)

// Key décrit un paramètre de configuration.
type Key struct {
	Name        string
	Env         string
	Default     string
	Description string
	Kind        Kind
	Secret      bool
}

// Keys liste tous les paramètres configurables.
var Keys = []Key{
	{Name: KeyTMDBAPIKey, Env: "TMDB_API_KEY", Description: "Clé API TMDB (films et séries)", Secret: true},
	{Name: KeyTheAudioDBAPIKey, Env: "THEAUDIODB_API_KEY", Default: "123", Description: "Clé API TheAudioDB (musique, 123 = clé publique gratuite)", Secret: true},
	{Name: KeyGroupName, Env: "TORRENT_AIO_GROUP_NAME", Default: "TORRENT-AIO", Description: "Nom du groupe de release"},
	{Name: KeyOutput, Env: "TORRENT_AIO_OUTPUT", Description: "Dossier de sortie (vide = dossier du fichier source)"},
	{Name: KeySkipTorrent, Env: "TORRENT_AIO_SKIP_TORRENT", Default: "false", Description: "Ne pas générer le fichier torrent", Kind: KindBool},
	{Name: KeyNoRename, Env: "TORRENT_AIO_NO_RENAME", Default: "false", Description: "Ne pas renommer les fichiers", Kind: KindBool},
	{Name: KeyCheckUpdates, Env: "TORRENT_AIO_CHECK_UPDATES", Default: "true", Description: "Vérifier les mises à jour à chaque exécution", Kind: KindBool},
}

// Source indique d'où provient une valeur.
type Source string

const (
	SourceFile    Source = "fichier"
	SourceEnv     Source = "environnement"
	SourceDefault Source = "défaut"
)

const fileName = "torrent-aio.yml"

// LookupKey retourne la définition d'un paramètre.
func LookupKey(name string) (Key, error) {
	for _, k := range Keys {
		if k.Name == name {
			return k, nil
		}
	}
	names := make([]string, len(Keys))
	for i, k := range Keys {
		names[i] = k.Name
	}
	return Key{}, fmt.Errorf("paramètre inconnu %q (paramètres disponibles : %s)", name, strings.Join(names, ", "))
}

// SearchPaths retourne, par ordre de priorité, les emplacements où le fichier de
// configuration est recherché.
func SearchPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, ".config", fileName),
		filepath.Join(home, fileName),
	}
}

// Store contient la configuration chargée depuis le fichier.
type Store struct {
	// Path est le fichier lu, ou celui qui sera créé lors de Save.
	Path string
	// Found indique si le fichier existait au chargement.
	Found  bool
	values map[string]any
}

// Load charge la configuration. Si explicitPath est vide, le premier fichier
// existant parmi SearchPaths est utilisé.
func Load(explicitPath string) (*Store, error) {
	candidates := SearchPaths()
	if explicitPath != "" {
		candidates = []string{explicitPath}
	}
	if len(candidates) == 0 {
		return nil, errors.New("impossible de déterminer le dossier utilisateur")
	}

	s := &Store{Path: candidates[0], values: map[string]any{}}
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("lecture de %s : %w", p, err)
		}
		if err := yaml.Unmarshal(data, &s.values); err != nil {
			return nil, fmt.Errorf("fichier de configuration invalide %s : %w", p, err)
		}
		if s.values == nil {
			s.values = map[string]any{}
		}
		s.Path, s.Found = p, true
		break
	}
	return s, nil
}

// Get retourne la valeur d'un paramètre et sa provenance.
func (s *Store) Get(name string) (string, Source) {
	key, err := LookupKey(name)
	if err != nil {
		return "", SourceDefault
	}
	if v, ok := s.values[name]; ok && v != nil {
		return strings.TrimSpace(fmt.Sprint(v)), SourceFile
	}
	if v, ok := os.LookupEnv(key.Env); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), SourceEnv
	}
	return key.Default, SourceDefault
}

// String retourne la valeur d'un paramètre.
func (s *Store) String(name string) string {
	v, _ := s.Get(name)
	return v
}

// Bool retourne la valeur booléenne d'un paramètre (false si invalide).
func (s *Store) Bool(name string) bool {
	b, _ := strconv.ParseBool(s.String(name))
	return b
}

// Set modifie un paramètre (en mémoire, voir Save).
func (s *Store) Set(name, value string) error {
	key, err := LookupKey(name)
	if err != nil {
		return err
	}
	value = strings.TrimSpace(value)
	if key.Kind == KindBool {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%s attend une valeur booléenne (true/false), reçu %q", name, value)
		}
		s.values[name] = b
		return nil
	}
	s.values[name] = value
	return nil
}

// Unset supprime un paramètre du fichier (en mémoire, voir Save).
func (s *Store) Unset(name string) error {
	if _, err := LookupKey(name); err != nil {
		return err
	}
	delete(s.values, name)
	return nil
}

// Save écrit la configuration dans s.Path (permissions 0600 car elle peut contenir des clés API).
func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return fmt.Errorf("création du dossier de configuration : %w", err)
	}
	// Ordre stable des clés dans le fichier
	names := make([]string, 0, len(s.values))
	for k := range s.values {
		names = append(names, k)
	}
	sort.Strings(names)
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, k := range names {
		var v yaml.Node
		if err := v.Encode(s.values[k]); err != nil {
			return err
		}
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: k}, &v)
	}
	data, err := yaml.Marshal(node)
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.Path, data, 0o600); err != nil {
		return fmt.Errorf("écriture de %s : %w", s.Path, err)
	}
	s.Found = true
	return os.Chmod(s.Path, 0o600)
}

// Mask masque une valeur secrète pour l'affichage.
func Mask(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 6 {
		return strings.Repeat("*", len(v))
	}
	return v[:3] + strings.Repeat("*", len(v)-6) + v[len(v)-3:]
}
