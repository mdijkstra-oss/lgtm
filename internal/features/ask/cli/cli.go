// Package cli is the command that opens the page to ask on.
package cli

import (
	"github.com/spf13/cobra"

	"strigil/internal/features/ask"
	"strigil/internal/features/ask/ui"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "ask",
		Short: "Ask about the codebase and read the answer",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return ui.Run(cmd.Context(), ask.Answers()) },
	}
}
