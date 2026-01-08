package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/metwurcht/torrent-all-in-one/internal/tmdb"
)

// InteractivePrompter implémente Prompter pour une utilisation CLI interactive
type InteractivePrompter struct {
	reader *bufio.Reader
}

// NewInteractivePrompter crée un nouveau prompter interactif
func NewInteractivePrompter() *InteractivePrompter {
	return &InteractivePrompter{
		reader: bufio.NewReader(os.Stdin),
	}
}

// SelectMovie affiche une liste de films et retourne le choix de l'utilisateur
func (p *InteractivePrompter) SelectMovie(movies []tmdb.Movie) (*tmdb.Movie, error) {
	if len(movies) == 0 {
		return nil, fmt.Errorf("aucun film à sélectionner")
	}

	fmt.Println("\n📽️  Résultats de recherche:")
	fmt.Println(strings.Repeat("─", 60))

	for i, movie := range movies {
		year := ""
		if len(movie.ReleaseDate) >= 4 {
			year = movie.ReleaseDate[len(movie.ReleaseDate)-4:]
		}

		rating := ""
		if movie.VoteAverage > 0 {
			rating = fmt.Sprintf(" ⭐ %.1f", movie.VoteAverage)
		}

		fmt.Printf("  [%d] %s (%s)%s\n", i+1, movie.Title, year, rating)

		if movie.OriginalTitle != "" && movie.OriginalTitle != movie.Title {
			fmt.Printf("      └─ %s\n", movie.OriginalTitle)
		}
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("  [0] Nouvelle recherche / Entrer un ID TMDB")
	fmt.Println()

	// Demander le choix
	for {
		fmt.Print("Votre choix: ")
		input, err := p.reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				return nil, fmt.Errorf("impossible de lire l'entrée (pas de TTY). Pour Docker, utilisez: docker run -it")
			}
			return nil, err
		}

		input = strings.TrimSpace(input)

		// Si c'est 0, on retourne une erreur pour déclencher une nouvelle recherche
		if input == "0" {
			return nil, fmt.Errorf("nouvelle recherche demandée")
		}

		// Essayer de parser comme un numéro
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(movies) {
			fmt.Printf("❌ Choix invalide. Entrez un nombre entre 1 et %d\n", len(movies))
			continue
		}

		return &movies[choice-1], nil
	}
}

// SelectSourceType demande à l'utilisateur de choisir le type de source
func (p *InteractivePrompter) SelectSourceType() (string, error) {
	type sourceOption struct {
		display string
		value   string
	}

	sources := []sourceOption{
		{"BluRay", "BluRay"},
		{"BluRay Rip", "BluRay.HDLight"},
		{"REMUX", "REMUX"},
		{"Téléchargement WEB", "WEB"},
		{"WEBRip", "WEBRip"},
	}

	fmt.Println("\n📀 Type de source:")
	fmt.Println(strings.Repeat("─", 60))
	for i, source := range sources {
		fmt.Printf("  [%d] %s\n", i+1, source.display)
	}
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	for {
		fmt.Print("Votre choix: ")
		input, err := p.reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				return "", fmt.Errorf("impossible de lire l'entrée (pas de TTY). Pour Docker, utilisez: docker run -it")
			}
			return "", err
		}

		input = strings.TrimSpace(input)
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(sources) {
			fmt.Printf("❌ Choix invalide. Entrez un nombre entre 1 et %d\n", len(sources))
			continue
		}

		return sources[choice-1].value, nil
	}
}

// AskForInput demande une entrée texte à l'utilisateur
func (p *InteractivePrompter) AskForInput(prompt string) (string, error) {
	promptUI := promptui.Prompt{
		Label: prompt,
	}

	result, err := promptUI.Run()
	if err != nil {
		// Fallback sur stdin simple si promptui échoue
		fmt.Print(prompt + " ")
		input, err := p.reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				return "", fmt.Errorf("impossible de lire l'entrée (pas de TTY). Pour Docker, utilisez: docker run -it")
			}
			return "", err
		}
		return strings.TrimSpace(input), nil
	}

	return result, nil
}

// Confirm demande une confirmation oui/non
func (p *InteractivePrompter) Confirm(prompt string) (bool, error) {
	promptUI := promptui.Prompt{
		Label:     prompt,
		IsConfirm: true,
	}

	_, err := promptUI.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false, nil
		}
		// Fallback
		fmt.Printf("%s [y/N]: ", prompt)
		input, err := p.reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		input = strings.ToLower(strings.TrimSpace(input))
		return input == "y" || input == "yes" || input == "o" || input == "oui", nil
	}

	return true, nil
}

// ShowProgress affiche une barre de progression
func (p *InteractivePrompter) ShowProgress(current, total int, message string) {
	percentage := float64(current) / float64(total) * 100
	barWidth := 40
	filled := int(float64(barWidth) * float64(current) / float64(total))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf("\r[%s] %.1f%% %s", bar, percentage, message)

	if current >= total {
		fmt.Println()
	}
}

// ShowMessage affiche un message
func (p *InteractivePrompter) ShowMessage(message string) {
	fmt.Println(message)
}

// ShowError affiche une erreur
func (p *InteractivePrompter) ShowError(message string) {
	fmt.Fprintf(os.Stderr, "❌ Erreur: %s\n", message)
}
