package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"

	askcli "strigil/internal/features/ask/cli"
	introcli "strigil/internal/features/intro/cli"
)

// A subcommand lives in the cli package of whatever owns it and is listed
// here.
var commands = []func() *cobra.Command{askcli.Command, introcli.Command}

func main() {
	root := &cobra.Command{
		Use:   "strigil",
		Short: "Scrape away the slop",
		// With nothing asked for, the intro page.
		RunE: func(cmd *cobra.Command, _ []string) error { return introcli.Command().RunE(cmd, nil) },
	}
	for _, c := range commands {
		root.AddCommand(c())
	}
	if err := fang.Execute(context.Background(), root); err != nil {
		os.Exit(1)
	}
}
