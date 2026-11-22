package renamer

import "testing"

func TestCleanTitle(t *testing.T) {
	tests := map[string]string{
		"La fête est finie":           "La.fete.est.finie",
		"Amélie Poulain":              "Amelie.Poulain",
		"L'Âge de glace":              "L.Age.de.glace",
		"Aujourd’hui":                 "Aujourd.hui",
		"Spider-Man: No Way Home":     "Spider.Man.No.Way.Home",
		"Cœur de pirate":              "Coeur.de.pirate",
		"Fast & Furious":              "Fast.and.Furious",
		"Mr. Robot":                   "Mr.Robot",
		"Qu'est-ce qu'on a fait ?":    "Qu.est.ce.qu.on.a.fait",
		"Beyoncé (Deluxe) [Remaster]": "Beyonce.Deluxe.Remaster",
		"  Hello,   World!  ":         "Hello.World",
		"WALL·E":                      "WALLE",
		"Ça":                          "Ca",
		"Björk - Homogenic":           "Bjork.Homogenic",
		"千と千尋の神隠し":                    "千と千尋の神隠し",
		"":                            "",
	}
	for in, want := range tests {
		if got := CleanTitle(in); got != want {
			t.Errorf("CleanTitle(%q) = %q, want %q", in, got, want)
		}
	}
}
