// Package cli is the command that opens the intro page.
package cli

import (
	"github.com/spf13/cobra"

	"strigil/internal/features/intro/ui"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "intro",
		Short: "Show the intro page",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return ui.Run(cmd.Context()) },
	}
}
