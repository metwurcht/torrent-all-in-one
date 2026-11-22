package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolutionOrder(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("TMDB_API_KEY", "from-env")
	t.Setenv("TORRENT_AIO_GROUP_NAME", "")

	s, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if s.Found {
		t.Fatal("aucun fichier ne devrait être trouvé")
	}
	if v, src := s.Get(KeyTMDBAPIKey); v != "from-env" || src != SourceEnv {
		t.Fatalf("Get() = %q, %s; want env value", v, src)
	}
	if v, src := s.Get(KeyGroupName); v != "TORRENT-AIO" || src != SourceDefault {
		t.Fatalf("Get() = %q, %s; want default", v, src)
	}

	if err := s.Set(KeyTMDBAPIKey, "from-file"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(KeySkipTorrent, "true"); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	s, err = Load("")
	if err != nil {
		t.Fatal(err)
	}
	if !s.Found || s.Path != SearchPaths()[0] {
		t.Fatalf("fichier attendu dans %s, obtenu %s (found=%v)", SearchPaths()[0], s.Path, s.Found)
	}
	if v, src := s.Get(KeyTMDBAPIKey); v != "from-file" || src != SourceFile {
		t.Fatalf("le fichier doit primer sur l'environnement, obtenu %q (%s)", v, src)
	}
	if !s.Bool(KeySkipTorrent) {
		t.Fatal("skip_torrent devrait valoir true")
	}
	info, err := os.Stat(s.Path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions = %v, want 0600", info.Mode().Perm())
	}
}

func TestSecondSearchPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, fileName), []byte("group_name: HOMEGRP\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if v := s.String(KeyGroupName); v != "HOMEGRP" {
		t.Fatalf("group_name = %q, want HOMEGRP", v)
	}
}

func TestSetValidation(t *testing.T) {
	s := &Store{values: map[string]any{}}
	if err := s.Set("inconnu", "x"); err == nil {
		t.Fatal("un paramètre inconnu devrait être refusé")
	}
	if err := s.Set(KeyNoRename, "peut-être"); err == nil {
		t.Fatal("une valeur booléenne invalide devrait être refusée")
	}
}
