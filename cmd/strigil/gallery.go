//go:build gallery

package main

import (
	"github.com/spf13/cobra"

	gallerycli "strigil/internal/features/gallery/cli"
	searchui "strigil/internal/features/search/ui"
	"strigil/internal/kit/frame"
	"strigil/internal/kit/hints"
	"strigil/internal/kit/palette"
	"strigil/internal/kit/story"
	"strigil/internal/kit/tabs"
	"strigil/internal/kit/theme"
)

func init() {
	commands = append(commands, func() *cobra.Command {
		th := theme.Mocha()
		return gallerycli.Command(th, components(th))
	})
}

// components lists everything the gallery shows. A test checks that no
// stories.go is missing from it.
func components(th theme.Theme) []story.Component {
	return []story.Component{
		frame.Stories(th),
		hints.Stories(th),
		palette.Stories(th),
		tabs.Stories(th),
		searchui.Stories(th),
	}
}
