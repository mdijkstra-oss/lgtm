package main

import (
	"context"
	"os"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"

	searchcli "strigil/internal/features/search/cli"
	shellcli "strigil/internal/shell/cli"
)

// A subcommand lives in the cli package of whatever owns it and is listed
// here. Files behind a build tag add theirs from init.
var commands = []func() *cobra.Command{searchcli.Command, shellcli.Command}

func main() {
	root := &cobra.Command{
		Use:   "strigil",
		Short: "Search a folder from the command line or an interactive interface",
	}
	for _, c := range commands {
		root.AddCommand(c())
	}
	if err := fang.Execute(context.Background(), root); err != nil {
		os.Exit(1)
	}
}
