package processor

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindAudioFilesRecursivelyPreservesDiscOrdering(t *testing.T) {
	root := t.TempDir()
	for _, relative := range []string{
		"CD2/04 - Demo.mp3",
		"CD1/02 - Billie Jean.mp3",
		"CD1/01 - Beat It.mp3",
		"booklet.pdf",
	} {
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := FindAudioFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		filepath.Join(root, "CD1/01 - Beat It.mp3"),
		filepath.Join(root, "CD1/02 - Billie Jean.mp3"),
		filepath.Join(root, "CD2/04 - Demo.mp3"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected audio files:\n got: %v\nwant: %v", got, want)
	}
	if gotType := DetectDirectoryType(root); gotType != "music" {
		t.Fatalf("DetectDirectoryType() = %q, want music", gotType)
	}
}
