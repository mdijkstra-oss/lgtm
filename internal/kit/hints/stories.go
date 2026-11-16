//go:build gallery

package hints

import (
	_ "embed"

	"strigil/internal/kit"
	"strigil/internal/kit/story"
	"strigil/internal/kit/theme"
)

//go:embed about.md
var about string

func Stories(th theme.Theme) story.Component {
	hs := []kit.Hint{{Key: "ctrl+k", Label: "search"}, {Key: "[ ]", Label: "tabs"}, {Key: "q", Label: "quit"}}
	draw := func(sep string) func(w, h int) string {
		return func(w, h int) string {
			pen := theme.On(th.Panel)
			return pen.Line(pen.Gap(1)+Render(th, pen, hs, sep), w)
		}
	}
	return story.Component{
		Name:  "hints",
		About: about,
		Stories: []story.Story{
			{Name: "dotted", Note: "The footer's separator.", W: 50, H: 1, Render: draw("  ·  ")},
			{Name: "spaced", Note: "The palette's separator.", W: 50, H: 1, Render: draw("   ")},
			{Name: "cut", Note: "Too narrow: the caller's line cuts it.", W: 24, H: 1, Render: draw("  ·  ")},
		},
	}
}
