// Package cli is the command that opens the documents.
package cli

import (
	"github.com/spf13/cobra"

	"strigil/internal/features/ask"
	"strigil/internal/features/ask/ui"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "ask",
		Short: "Read a document about the codebase",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return ui.Run(cmd.Context(), ask.Docs()) },
	}
}
