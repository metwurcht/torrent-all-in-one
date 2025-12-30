package torrent

import (
	"fmt"
	"os"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

// Generator génère des fichiers torrent
type Generator struct {
	trackerURL string
	pieceSize  int64
	comment    string
	createdBy  string
}

// NewGenerator crée un nouveau générateur de torrent
func NewGenerator(trackerURL string) *Generator {
	return &Generator{
		trackerURL: trackerURL,
		pieceSize:  256 * 1024, // 256 KB par défaut
		comment:    "Created by Torrent All-In-One",
		createdBy:  "Torrent-AIO",
	}
}

// SetPieceSize définit la taille des pièces
func (g *Generator) SetPieceSize(size int64) {
	g.pieceSize = size
}

// SetComment définit le commentaire du torrent
func (g *Generator) SetComment(comment string) {
	g.comment = comment
}

// Create crée un fichier torrent à partir d'un fichier source
func (g *Generator) Create(sourcePath, outputPath string) error {
	// Vérifier que le fichier source existe
	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("fichier source introuvable: %w", err)
	}

	// Calculer la taille optimale des pièces
	pieceLength := g.calculatePieceLength(info.Size())

	// Créer le metainfo
	mi := metainfo.MetaInfo{
		Comment:   g.comment,
		CreatedBy: g.createdBy,
	}

	// Ajouter le tracker
	if g.trackerURL != "" {
		mi.Announce = g.trackerURL
	}

	// Construire les informations du fichier
	builder := metainfo.Info{
		PieceLength: pieceLength,
	}

	if err := builder.BuildFromFilePath(sourcePath); err != nil {
		return fmt.Errorf("erreur construction torrent: %w", err)
	}

	mi.InfoBytes, err = bencode.Marshal(builder)
	if err != nil {
		return fmt.Errorf("erreur encodage info: %w", err)
	}

	// Écrire le fichier torrent
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer f.Close()

	if err := mi.Write(f); err != nil {
		return fmt.Errorf("erreur écriture torrent: %w", err)
	}

	return nil
}

// CreateFromDirectory crée un torrent à partir d'un dossier
func (g *Generator) CreateFromDirectory(dirPath, outputPath string) error {
	// Vérifier que le dossier existe
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("dossier introuvable: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s n'est pas un dossier", dirPath)
	}

	// Calculer la taille totale pour la taille des pièces
	totalSize, err := g.calculateDirSize(dirPath)
	if err != nil {
		return err
	}

	pieceLength := g.calculatePieceLength(totalSize)

	// Créer le metainfo
	mi := metainfo.MetaInfo{
		Comment:   g.comment,
		CreatedBy: g.createdBy,
	}

	if g.trackerURL != "" {
		mi.Announce = g.trackerURL
	}

	// Construire les informations du dossier
	builder := metainfo.Info{
		PieceLength: pieceLength,
	}

	if err := builder.BuildFromFilePath(dirPath); err != nil {
		return fmt.Errorf("erreur construction torrent: %w", err)
	}

	mi.InfoBytes, err = bencode.Marshal(builder)
	if err != nil {
		return fmt.Errorf("erreur encodage info: %w", err)
	}

	// Écrire le fichier torrent
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %w", err)
	}
	defer f.Close()

	if err := mi.Write(f); err != nil {
		return fmt.Errorf("erreur écriture torrent: %w", err)
	}

	return nil
}

// calculatePieceLength calcule la taille optimale des pièces
func (g *Generator) calculatePieceLength(fileSize int64) int64 {
	// Taille en Mo
	const (
		GB1  = int64(1024 * 1024 * 1024)
		GB2  = int64(2 * 1024 * 1024 * 1024)
		GB4  = int64(4 * 1024 * 1024 * 1024)
		GB8  = int64(8 * 1024 * 1024 * 1024)
		MB1  = int64(1024 * 1024)
		MB2  = int64(2 * 1024 * 1024)
		MB4  = int64(4 * 1024 * 1024)
		MB8  = int64(8 * 1024 * 1024)
		MB16 = int64(16 * 1024 * 1024)
	)

	switch {
	case fileSize <= GB1:
		return MB1
	case fileSize <= GB2:
		return MB2
	case fileSize <= GB4:
		return MB4
	case fileSize <= GB8:
		return MB8
	default:
		return MB16
	}
}

// calculateDirSize calcule la taille totale d'un dossier
func (g *Generator) calculateDirSize(path string) (int64, error) {
	var size int64

	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			subSize, err := g.calculateDirSize(path + "/" + entry.Name())
			if err != nil {
				continue
			}
			size += subSize
		} else {
			size += info.Size()
		}
	}

	return size, nil
}

// GetInfoHash retourne le hash info d'un fichier torrent existant
func GetInfoHash(torrentPath string) (string, error) {
	mi, err := metainfo.LoadFromFile(torrentPath)
	if err != nil {
		return "", fmt.Errorf("erreur lecture torrent: %w", err)
	}

	hash := mi.HashInfoBytes()
	return hash.HexString(), nil
}
