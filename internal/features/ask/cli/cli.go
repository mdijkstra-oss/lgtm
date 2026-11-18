// Package cli is the command that opens the documents, and asks for a new
// one.
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"lgtm/internal/features/ask"
	"lgtm/internal/features/ask/ui"
	"lgtm/internal/kit/llm"
)

func Command() *cobra.Command {
	var plain bool
	cmd := &cobra.Command{
		Use:   "ask [question]",
		Short: "Ask a question about the codebase, or read what was answered before",
		Long: "With a question, a model reads the repository and answers with a document,\n" +
			"kept in the system's temporary folder. The model is asked every time: the\n" +
			"code may have changed. With no question, the documents there are.",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			question := strings.TrimSpace(strings.Join(args, " "))
			if question == "" {
				return ui.Run(cmd.Context(), ask.Docs(), nil)
			}
			root := ask.Root()
			out, _ := os.Stdout.Stat()
			plain = plain || out == nil || out.Mode()&os.ModeCharDevice == 0
			model, err := llm.FromEnv(cmd.Context())
			if err != nil {
				return err
			}
			write := func(ctx context.Context, on func(ask.Event)) (ask.Written, error) {
				return ask.Write(ctx, model, llm.Options(), root, question, on)
			}
			if plain {
				return print(cmd.Context(), write)
			}
			return ui.Run(cmd.Context(), ask.Docs(), &ui.Asked{Question: question, Write: write})
		},
	}
	cmd.Flags().BoolVar(&plain, "plain", false, "print the document instead of opening it")
	return cmd
}

// print is the command with no terminal to draw on: what the model does
// goes to stderr a line a call, and the document to stdout.
func print(ctx context.Context, write func(context.Context, func(ask.Event)) (ask.Written, error)) error {
	w, err := write(ctx, func(e ask.Event) {
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
