// Package cli is the gallery subcommand. Only a file behind the gallery
// build tag imports it, so a normal build does not contain it.
package cli

import (
	"github.com/spf13/cobra"

	"strigil/internal/features/gallery"
	"strigil/internal/kit/story"
	"strigil/internal/kit/theme"
)

func Command(th theme.Theme, components []story.Component) *cobra.Command {
	return &cobra.Command{
		Use:   "gallery",
		Short: "Browse the interface's components one state at a time",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return gallery.Run(cmd.Context(), th, components)
		},
	}
}
