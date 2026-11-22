package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"v1.5.0", "v1.4.1", true},
		{"v1.4.10", "v1.4.9", true},
		{"v2.0.0", "v1.99.99", true},
		{"v1.4.1", "v1.4.1", false},
		{"v1.4.0", "v1.4.1", false},
		{"v1.4.1", "v1.4.1-3-gabcdef-dirty", false},
		{"v1.5.0", "dev", false},
		{"garbage", "v1.0.0", false},
	}
	for _, tt := range tests {
		if got := IsNewer(tt.latest, tt.current); got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestLatest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9","html_url":"https://example.com/v9.9.9"}`))
	}))
	defer server.Close()
	previous := latestReleaseURL
	latestReleaseURL = server.URL
	t.Cleanup(func() { latestReleaseURL = previous })

	r, err := Latest(context.Background())
	if err != nil || r.Tag != "v9.9.9" || r.URL != "https://example.com/v9.9.9" {
		t.Fatalf("Latest() = %+v, %v", r, err)
	}
}
