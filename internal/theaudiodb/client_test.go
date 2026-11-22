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
	client := NewClient("")
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

func TestSearchAlbumWithoutSeparatorTriesArtistAlbumSplits(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		requests = append(requests, q.Get("s")+"|"+q.Get("a"))
		w.Header().Set("Content-Type", "application/json")
		if q.Get("s") == "Michael Jackson" && q.Get("a") == "Off the Wall" {
			_, _ = w.Write([]byte(`{"album":[{"idAlbum":"2118218","strArtist":"Michael Jackson","strAlbum":"Off the Wall","intYearReleased":"1979"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"album":null}`))
	}))
	defer server.Close()

	t.Setenv("THEAUDIODB_BASE_URL", server.URL)
	keywords := ExtractKeywords("Michael.Jackson.Off.the.Wall.1979.MP3[320kbps]-MW")
	albums, err := NewClient("").SearchAlbum(context.Background(), keywords)
	if err != nil {
		t.Fatalf("SearchAlbum returned an error: %v", err)
	}
	if len(albums) != 1 || albums[0].ID != "2118218" {
		t.Fatalf("unexpected albums: %+v (requests: %v)", albums, requests)
	}
	want := []string{"Michael Jackson Off the Wall|", "Michael|Jackson Off the Wall", "Michael Jackson|Off the Wall"}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v, want %v", requests, want)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("requests = %v, want %v", requests, want)
		}
	}
}
