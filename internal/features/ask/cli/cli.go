// Package cli is the command that has a document written and opens the
// page to read it on.
package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"lgtm/internal/features/ask"
	"lgtm/internal/features/ask/ui"
	"lgtm/internal/kit/llm"
)

// history is how many commits the page can search.
const history = 500

func Command() *cobra.Command {
	var plain bool
	cmd := &cobra.Command{
		Use:   "ask [question | commit]",
		Short: "Have a question answered, or a change explained",
		Long: "With a question, a model reads the repository and answers with a document.\n" +
			"With a commit, as git names one, the document explains that commit. With\n" +
			"nothing, the history opens to pick a change from: what is not committed yet\n" +
			"comes first. Documents are kept in the system's temporary folder.",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := ask.Root()
			// The model is looked for only when something is asked of it:
			// reading what was written before needs none.
			write := func(ctx context.Context, sub ask.Subject, on func(ask.Event)) (ask.Written, error) {
				model, err := llm.FromEnv(ctx)
				if err != nil {
					return ask.Written{}, err
				}
				return ask.Write(ctx, model, llm.Options(), root, sub, on)
			}
			var first *ask.Subject
			if text := strings.TrimSpace(strings.Join(args, " ")); text != "" {
				sub := ask.Named(root, text)
				first = &sub
			}
			if out, _ := os.Stdout.Stat(); plain || out == nil || out.Mode()&os.ModeCharDevice == 0 {
				// With no page to pick on, nothing named is what is not committed.
				if first == nil {
					if ask.Uncommitted(root) == 0 {
						return errors.New("nothing is uncommitted: name a commit, or ask a question")
					}
					first = &ask.Subject{Pending: true}
				}
				return print(cmd.Context(), *first, write)
			}
			return ui.Run(cmd.Context(), ui.Start{
				Repo:        filepath.Base(root),
				Branch:      ask.Branch(root),
				History:     ask.History(root, history),
				Uncommitted: ask.Uncommitted(root),
				Answers:     ask.Answers(root),
				Write:       write,
				First:       first,
			})
		},
	}
	cmd.Flags().BoolVar(&plain, "plain", false, "print the document instead of opening it")
	return cmd
}

// print is the command with no terminal to draw on: what the model does
// goes to stderr a line a call, and the document to stdout.
func print(ctx context.Context, sub ask.Subject, write func(context.Context, ask.Subject, func(ask.Event)) (ask.Written, error)) error {
	w, err := write(ctx, sub, func(e ask.Event) {
		if e.Done {
			mark := "✓"
			if e.Failed {
				mark = "✗"
			}
			fmt.Fprintf(os.Stderr, "%s %s %s  %s\n", mark, e.Tool, e.What, e.Note)
			if e.Said != "" && (e.Tool == "write_document" || e.Tool == "fix_document") {
				fmt.Fprintln(os.Stderr, e.Said)
			}
		}
	})
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "saved to", w.File)
	_, err = fmt.Print(w.Source)
	return err
}
