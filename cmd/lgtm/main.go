package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"

	askcli "lgtm/internal/features/ask/cli"
	introcli "lgtm/internal/features/intro/cli"
)

// A subcommand lives in the cli package of whatever owns it and is listed
// here.
var commands = []func() *cobra.Command{askcli.Command, introcli.Command}

func main() {
	root := &cobra.Command{
		Use:   "lgtm [question]",
		Short: "Looks good to me, said by someone who read it",
		// What is not a command is a question; with nothing asked, the
		// documents there are.
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error { return askcli.Command().RunE(cmd, args) },
	}
	for _, c := range commands {
		root.AddCommand(c())
	}
	if err := fang.Execute(context.Background(), root); err != nil {
		os.Exit(1)
	}
}
