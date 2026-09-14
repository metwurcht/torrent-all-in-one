package tmdb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestYear(t *testing.T) {
	tests := []struct {
		name         string
		releaseDate  string
		expectedYear string
	}{
		{
			name:         "Format français avec date complète",
			releaseDate:  "19/08/2009 (FR)",
			expectedYear: "2009",
		},
		{
			name:         "Format français 2024",
			releaseDate:  "27/11/2024 (FR)",
			expectedYear: "2024",
		},
		{
			name:         "Format API standard",
			releaseDate:  "2009-08-19",
			expectedYear: "2009",
		},
		{
			name:         "Année seule",
			releaseDate:  "2009",
			expectedYear: "2009",
		},
		{
			name:         "Chaîne vide",
			releaseDate:  "",
			expectedYear: "",
		},
		{
			name:         "Format US",
			releaseDate:  "08/19/2009",
			expectedYear: "2009",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Movie{ReleaseDate: tt.releaseDate}
			got := m.Year()
			if got != tt.expectedYear {
				t.Errorf("Year() = %q, want %q (input: %q)", got, tt.expectedYear, tt.releaseDate)
			}
		})
	}
}

func TestOfficialAPIClient(t *testing.T) {
	t.Setenv("TMDB_API_KEY", "test-key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api_key") != "test-key" {
			t.Errorf("api_key = %q, want test-key", r.URL.Query().Get("api_key"))
		}
		if r.URL.Query().Get("language") != "fr-FR" {
			t.Errorf("language = %q, want fr-FR", r.URL.Query().Get("language"))
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/search/movie":
			_ = json.NewEncoder(w).Encode(map[string]any{"results": []any{map[string]any{"id": 10, "title": "Film", "original_title": "Movie", "release_date": "2024-01-02"}}})
		case r.URL.Path == "/movie/10":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 10, "title": "Film", "genres": []any{map[string]string{"name": "Drame"}}, "credits": map[string]any{"cast": []any{map[string]any{"name": "Acteur", "character": "Rôle", "order": 0}}, "crew": []any{map[string]string{"name": "Réalisateur", "job": "Director"}}}, "external_ids": map[string]string{"imdb_id": "tt123"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL
	movies, err := client.SearchMovie(context.Background(), "film")
	if err != nil || len(movies) != 1 || movies[0].ID != 10 {
		t.Fatalf("SearchMovie() = %#v, %v", movies, err)
	}
	details, err := client.GetMovieDetails(context.Background(), 10)
	if err != nil || details.IMDbID != "tt123" || len(details.Directors) != 1 || details.Genres[0] != "Drame" {
		t.Fatalf("GetMovieDetails() = %#v, %v", details, err)
	}
}

func TestClientRequiresAPIKey(t *testing.T) {
	previous := os.Getenv("TMDB_API_KEY")
	t.Cleanup(func() { _ = os.Setenv("TMDB_API_KEY", previous) })
	_ = os.Unsetenv("TMDB_API_KEY")
	_, err := NewClient().SearchMovie(context.Background(), "film")
	if err == nil || !strings.Contains(err.Error(), "TMDB_API_KEY") {
		t.Fatalf("SearchMovie() error = %v, want TMDB_API_KEY error", err)
	}
}

func TestRequestQueryEscaping(t *testing.T) {
	values := url.Values{"query": {"Dune & Part Two"}}
	if !strings.Contains(values.Encode(), "Dune+%26+Part+Two") {
		t.Fatalf("query was not escaped: %s", values.Encode())
	}
}
