package tmdb

import "testing"

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
