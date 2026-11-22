package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const completionMarker = "# torrent-aio : autocomplétion"

// addCompletionInstallCmd ajoute « completion install » à la commande
// « completion » générée par Cobra.
func addCompletionInstallCmd() {
	rootCmd.InitDefaultCompletionCmd()
	for _, c := range rootCmd.Commands() {
		if c.Name() == "completion" {
			c.AddCommand(newCompletionInstallCmd())
			return
		}
	}
}

func newCompletionInstallCmd() *cobra.Command {
	var shell string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Installer automatiquement l'autocomplétion pour votre shell",
		Long: `Configure votre shell pour que la touche Tab complète les commandes,
sous-commandes, options et noms de paramètres de torrent-aio.

Le shell est détecté automatiquement (bash, zsh, fish ou PowerShell) ;
utilisez --shell pour le forcer. La commande peut être relancée sans risque :
elle ne modifie rien si l'autocomplétion est déjà installée.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			if shell == "" {
				shell = detectShell()
			}
			if shell == "" {
				return errors.New("shell non détecté, précisez-le avec --shell (bash, zsh, fish, powershell)")
			}
			return installCompletion(shell, commandRef(cmd.Root().Name()))
		},
	}
	cmd.Flags().StringVar(&shell, "shell", "", "Shell cible : bash, zsh, fish ou powershell (défaut: détecté)")
	_ = cmd.RegisterFlagCompletionFunc("shell", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"bash", "zsh", "fish", "powershell"}, cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}

// detectShell devine le shell de l'utilisateur.
func detectShell() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	switch filepath.Base(os.Getenv("SHELL")) {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	case "pwsh":
		return "powershell"
	}
	return ""
}

// commandRef retourne le nom de la commande si elle est trouvée dans le PATH,
// sinon le chemin absolu de l'exécutable courant.
func commandRef(name string) string {
	exe, err := os.Executable()
	if err != nil {
		return name
	}
	exe, _ = filepath.EvalSymlinks(exe)
	if found, err := exec.LookPath(name); err == nil {
		if found, err = filepath.EvalSymlinks(found); err == nil && found == exe {
			return name
		}
	}
	if strings.ContainsAny(exe, " '\"") {
		return fmt.Sprintf("%q", exe)
	}
	return exe
}

func installCompletion(shell, ref string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	switch shell {
	case "bash":
		rc := filepath.Join(home, ".bashrc")
		changed, err := appendBlock(rc,
			// Charger bash-completion s'il ne l'est pas déjà (requis par le script généré)
			"if ! declare -F _get_comp_words_by_ref >/dev/null; then",
			"  for f in "+strings.Join(bashCompletionPaths, " ")+"; do [ -r \"$f\" ] && . \"$f\" && break; done",
			"fi",
			fmt.Sprintf("source <(%s completion bash)", ref))
		if err != nil {
			return err
		}
		report(changed, rc, "source ~/.bashrc")
		if !hasBashCompletion() {
			fmt.Println("⚠️  Le paquet « bash-completion » semble absent, il est nécessaire à l'autocomplétion bash :")
			fmt.Println("   sudo apt install bash-completion   # Debian/Ubuntu")
			fmt.Println("   brew install bash-completion@2     # macOS")
		}

	case "zsh":
		rc := filepath.Join(home, ".zshrc")
		content, _ := os.ReadFile(rc)
		lines := []string{}
		if !strings.Contains(string(content), "compinit") {
			lines = append(lines, "autoload -Uz compinit && compinit")
		}
		lines = append(lines, fmt.Sprintf("source <(%s completion zsh)", ref))
		changed, err := appendBlock(rc, lines...)
		if err != nil {
			return err
		}
		report(changed, rc, "source ~/.zshrc")

	case "fish":
		// Les fichiers de ce dossier sont chargés automatiquement par fish
		path := filepath.Join(home, ".config", "fish", "completions", "torrent-aio.fish")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		content := fmt.Sprintf("%s\n%s completion fish | source\n", completionMarker, ref)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
		report(true, path, "")

	case "powershell":
		profile, err := powershellProfile()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(profile), 0o755); err != nil {
			return err
		}
		changed, err := appendBlock(profile, fmt.Sprintf("%s completion powershell | Out-String | Invoke-Expression", ref))
		if err != nil {
			return err
		}
		report(changed, profile, ". $PROFILE")

	default:
		return fmt.Errorf("shell %q non pris en charge (bash, zsh, fish, powershell)", shell)
	}
	return nil
}

// appendBlock ajoute les lignes à la fin du fichier, sauf si l'autocomplétion
// y est déjà installée. Retourne true si le fichier a été modifié.
func appendBlock(path string, lines ...string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	if strings.Contains(string(content), completionMarker) {
		return false, nil
	}

	block := "\n" + completionMarker + "\n" + strings.Join(lines, "\n") + "\n"
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		block = "\n" + block
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.WriteString(block)
	return err == nil, err
}

func report(changed bool, path, reload string) {
	if !changed {
		fmt.Printf("✅ L'autocomplétion est déjà installée (%s)\n", path)
		return
	}
	fmt.Printf("✅ Autocomplétion installée dans %s\n", path)
	if reload != "" {
		fmt.Printf("   Ouvrez un nouveau terminal ou exécutez : %s\n", reload)
	} else {
		fmt.Println("   Elle sera active dans les nouveaux terminaux.")
	}
}

var bashCompletionPaths = []string{
	"/usr/share/bash-completion/bash_completion",
	"/etc/bash_completion",
	"/opt/homebrew/etc/profile.d/bash_completion.sh",
	"/usr/local/etc/profile.d/bash_completion.sh",
}

func hasBashCompletion() bool {
	for _, p := range bashCompletionPaths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// powershellProfile demande à PowerShell le chemin de son profil utilisateur.
func powershellProfile() (string, error) {
	for _, bin := range []string{"pwsh", "powershell"} {
		out, err := exec.Command(bin, "-NoProfile", "-Command", "$PROFILE").Output()
		if err == nil && strings.TrimSpace(string(out)) != "" {
			return strings.TrimSpace(string(out)), nil
		}
	}
	return "", errors.New("PowerShell introuvable : impossible de déterminer le fichier de profil")
}
