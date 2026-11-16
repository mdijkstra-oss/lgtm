// Package cli is the tui subcommand. It is where the interface is put
// together: the shell plus what each feature contributes to it.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	searchui "strigil/internal/features/search/ui"
	"strigil/internal/kit"
	"strigil/internal/kit/theme"
	"strigil/internal/shell"
)

func Command() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Open the interactive interface",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if info, err := os.Stat(dir); err != nil {
				return err
			} else if !info.IsDir() {
				return fmt.Errorf("%s is not a folder", dir)
			}
			th := theme.Mocha()
			return shell.Run(cmd.Context(), shell.Options{
				Theme:   th,
				Dir:     dir,
				Sources: []kit.Source{searchui.NewSource(th, dir)},
			})
		},
	}
	cmd.Flags().StringVar(&dir, "dir", ".", "folder to search")
	return cmd
}
