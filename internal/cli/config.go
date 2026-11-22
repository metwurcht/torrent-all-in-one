package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/metwurcht/torrent-all-in-one/internal/config"
	"github.com/spf13/cobra"
)

func keysHelp() string {
	var sb strings.Builder
	sb.WriteString("Paramètres disponibles:\n")
	for _, k := range config.Keys {
		fmt.Fprintf(&sb, "  %-20s %s (env: %s)\n", k.Name, k.Description, k.Env)
	}
	return sb.String()
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Gérer la configuration (clés API, nom de groupe...)",
		Long: `Gère le fichier de configuration de torrent-aio.

Chaque paramètre est lu dans cet ordre : flag de la ligne de commande,
fichier de configuration, variable d'environnement, valeur par défaut.

` + keysHelp(),
		Example: `  torrent-aio config set tmdb_api_key 0123456789abcdef
  torrent-aio config set group_name MONGROUPE
  torrent-aio config get group_name
  torrent-aio config list
  torrent-aio config unset output`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "set <paramètre> <valeur>",
			Short: "Définir un paramètre dans le fichier de configuration",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				store, err := loadConfig()
				if err != nil {
					return err
				}
				if err := store.Set(args[0], args[1]); err != nil {
					return err
				}
				if err := store.Save(); err != nil {
					return err
				}
				fmt.Printf("✅ %s enregistré dans %s\n", args[0], store.Path)
				return nil
			},
		},
		&cobra.Command{
			Use:   "get <paramètre>",
			Short: "Afficher la valeur effective d'un paramètre",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				if _, err := config.LookupKey(args[0]); err != nil {
					return err
				}
				store, err := loadConfig()
				if err != nil {
					return err
				}
				fmt.Println(store.String(args[0]))
				return nil
			},
		},
		&cobra.Command{
			Use:   "unset <paramètre>",
			Short: "Retirer un paramètre du fichier de configuration",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				store, err := loadConfig()
				if err != nil {
					return err
				}
				if err := store.Unset(args[0]); err != nil {
					return err
				}
				if !store.Found {
					fmt.Println("Aucun fichier de configuration, rien à faire.")
					return nil
				}
				if err := store.Save(); err != nil {
					return err
				}
				fmt.Printf("✅ %s retiré de %s\n", args[0], store.Path)
				return nil
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "Lister tous les paramètres avec leur valeur et leur provenance",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				store, err := loadConfig()
				if err != nil {
					return err
				}
				printPath(store)
				fmt.Println()
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "PARAMÈTRE\tVALEUR\tSOURCE")
				for _, k := range config.Keys {
					v, src := store.Get(k.Name)
					if k.Secret {
						v = config.Mask(v)
					}
					if v == "" {
						v = "-"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\n", k.Name, v, src)
				}
				return w.Flush()
			},
		},
		&cobra.Command{
			Use:   "path",
			Short: "Afficher le chemin du fichier de configuration",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				store, err := loadConfig()
				if err != nil {
					return err
				}
				printPath(store)
				return nil
			},
		},
	)
	return cmd
}

func printPath(store *config.Store) {
	if store.Found {
		fmt.Printf("Fichier de configuration : %s\n", store.Path)
	} else {
		fmt.Printf("Aucun fichier de configuration (sera créé dans %s)\n", store.Path)
	}
}
