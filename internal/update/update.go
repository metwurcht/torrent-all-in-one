// Package update vérifie et installe les nouvelles versions publiées sur GitHub.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const repo = "metwurcht/torrent-all-in-one"

// URLs modifiables pour les tests.
var (
	latestReleaseURL = "https://api.github.com/repos/" + repo + "/releases/latest"
	downloadBaseURL  = "https://github.com/" + repo + "/releases/download"
)

// Release décrit une version publiée.
type Release struct {
	Tag string `json:"tag_name"`
	URL string `json:"html_url"`
}

// Latest retourne la dernière release publiée.
func Latest(ctx context.Context) (*Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("impossible de contacter GitHub : %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub a répondu %s", resp.Status)
	}
	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("réponse GitHub invalide : %w", err)
	}
	if r.Tag == "" {
		return nil, errors.New("aucune version trouvée sur GitHub")
	}
	return &r, nil
}

var semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)`)

func parse(v string) ([3]int, bool) {
	m := semverRe.FindStringSubmatch(strings.TrimSpace(v))
	if m == nil {
		return [3]int{}, false
	}
	var out [3]int
	for i := range out {
		out[i], _ = strconv.Atoi(m[i+1])
	}
	return out, true
}

// IsRelease indique si la version est une vraie version (et pas « dev »).
func IsRelease(version string) bool {
	_, ok := parse(version)
	return ok
}

// IsNewer indique si latest est strictement plus récente que current.
// Une version courante non reconnue (ex. « dev ») n'est jamais considérée comme obsolète.
func IsNewer(latest, current string) bool {
	l, okL := parse(latest)
	c, okC := parse(current)
	if !okL || !okC {
		return false
	}
	for i := range l {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

// AssetName retourne le nom du binaire de release pour la plateforme courante.
func AssetName() string {
	name := fmt.Sprintf("torrent-aio-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

// oldPath est l'emplacement de l'ancien binaire sous Windows (un exécutable
// en cours d'utilisation ne peut pas être remplacé, seulement renommé).
func oldPath(exe string) string { return exe + ".old" }

// CleanupOld supprime l'ancien binaire laissé par une mise à jour précédente.
func CleanupOld() {
	if exe, err := currentExecutable(); err == nil {
		_ = os.Remove(oldPath(exe))
	}
}

func currentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}

// Install télécharge la version tag et remplace l'exécutable courant.
// Retourne le chemin de l'exécutable mis à jour.
func Install(ctx context.Context, tag string, progress io.Writer) (string, error) {
	exe, err := currentExecutable()
	if err != nil {
		return "", fmt.Errorf("impossible de localiser l'exécutable courant : %w", err)
	}
	dir := filepath.Dir(exe)

	tmp, err := os.CreateTemp(dir, ".torrent-aio-update-*")
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return "", fmt.Errorf("pas de droit d'écriture dans %s : relancez la commande avec les droits administrateur (ex: sudo torrent-aio update)", dir)
		}
		return "", fmt.Errorf("création du fichier temporaire : %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // sans effet si le fichier a été renommé

	url := fmt.Sprintf("%s/%s/%s", downloadBaseURL, tag, AssetName())
	fmt.Fprintf(progress, "⬇️  Téléchargement de %s\n", url)
	if err := download(ctx, url, tmp); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return "", err
	}

	// Vérifier que le binaire téléchargé s'exécute bien sur cette machine
	if err := verify(ctx, tmpPath); err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		old := oldPath(exe)
		_ = os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			return "", fmt.Errorf("impossible de déplacer l'ancien exécutable : %w", err)
		}
		if err := os.Rename(tmpPath, exe); err != nil {
			_ = os.Rename(old, exe) // restauration
			return "", fmt.Errorf("installation de la nouvelle version : %w", err)
		}
		return exe, nil
	}

	if err := os.Rename(tmpPath, exe); err != nil {
		return "", fmt.Errorf("installation de la nouvelle version : %w", err)
	}
	return exe, nil
}

func download(ctx context.Context, url string, w io.Writer) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("téléchargement : %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("aucun binaire %s dans cette version (plateforme non supportée ?)", AssetName())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("téléchargement : GitHub a répondu %s", resp.Status)
	}
	n, err := io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("téléchargement : %w", err)
	}
	if n == 0 {
		return errors.New("téléchargement : fichier vide")
	}
	return nil
}

// verify lance « <binaire> --help » pour s'assurer que le fichier téléchargé est
// bien un exécutable valide pour cette machine avant de remplacer l'actuel.
func verify(ctx context.Context, path string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, path, "--help").CombinedOutput(); err != nil {
		return fmt.Errorf("le binaire téléchargé ne s'exécute pas correctement : %w\n%s", err, out)
	}
	return nil
}
