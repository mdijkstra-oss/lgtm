// Package cli is the search subcommand: flags in, matches out.
package cli

import (
	"encoding/json"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"strigil/internal/features/search"
)

var (
	pathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Magenta)
	placeStyle = lipgloss.NewStyle().Foreground(lipgloss.Green)
	sepStyle   = lipgloss.NewStyle().Foreground(lipgloss.BrightBlack)
	matchStyle = lipgloss.NewStyle().Foreground(lipgloss.Red).Bold(true)
)

func Command() *cobra.Command {
	var (
		dir        string
		ignoreCase bool
		asJSON     bool
	)
	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Find text in the files of a folder",
		Example: "strigil search harbour --dir samples\nstrigil search harbour --ignore-case --json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			q := search.Query{Text: args[0], Dir: dir, IgnoreCase: ignoreCase}
			out := cmd.OutOrStdout()
			if asJSON {
				matches := []search.Match{}
				if err := search.Run(cmd.Context(), q, func(m search.Match) { matches = append(matches, m) }); err != nil {
					return err
				}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(matches)
			}
			return search.Run(cmd.Context(), q, func(m search.Match) {
				_, _ = lipgloss.Fprintln(out, formatMatch(m))
			})
		},
	}
	cmd.Flags().StringVar(&dir, "dir", ".", "folder to search")
	cmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "match regardless of letter case")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print matches as JSON")
	return cmd
}

func formatMatch(m search.Match) string {
	start := m.Col - 1
	end := start + m.Length
	colon := sepStyle.Render(":")
	return pathStyle.Render(m.Path) + colon +
		placeStyle.Render(strconv.Itoa(m.Line)) + colon +
		placeStyle.Render(strconv.Itoa(m.Col)) + colon + " " +
		m.Text[:start] + matchStyle.Render(m.Text[start:end]) + m.Text[end:]
}
