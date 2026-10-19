package theaudiodb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSearchAlbumSendsArtistAndAlbumSeparately(t *testing.T) {
	var received url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"album":[{"idAlbum":2115888,"strArtist":"Daft Punk","strAlbum":"Homework","intYearReleased":1997}]}`))
	}))
	defer server.Close()

	t.Setenv("THEAUDIODB_BASE_URL", server.URL)
	client := NewClient()
	albums, err := client.SearchAlbum(context.Background(), "Daft Punk - Homework")
	if err != nil {
		t.Fatalf("SearchAlbum returned an error: %v", err)
	}
	if len(albums) != 1 || albums[0].ID != "2115888" {
		t.Fatalf("unexpected albums: %+v", albums)
	}
	if received.Get("s") != "Daft Punk" || received.Get("a") != "Homework" {
		t.Fatalf("unexpected query: %v", received)
	}
}

func TestDecodeTracksAcceptsStringAndNumberFields(t *testing.T) {
	tracks, err := decodeTracks([]byte(`{"track":[
		{"intTrackNumber":"01","strTrack":"Around the World","strTrackDuration":"07:09"},
		{"intTrackNumber":2,"strTrack":"Harder Better Faster Stronger","intDurationSeconds":224}
	]}`))
	if err != nil {
		t.Fatalf("decodeTracks returned an error: %v", err)
	}
	if len(tracks) != 2 || tracks[0].Number != 1 || tracks[0].Duration != 429 || tracks[1].Duration != 224 {
		t.Fatalf("unexpected tracks: %+v", tracks)
	}
}
