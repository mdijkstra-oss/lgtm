//go:build gallery

package frame

import (
	_ "embed"

	"strigil/internal/kit/story"
	"strigil/internal/kit/theme"
)

//go:embed about.md
var about string

func Stories(th theme.Theme) story.Component {
	pen := theme.On(th.Panel)
	body := []string{
		pen.Ink(th.Text, " first row"),
		pen.Ink(th.Soft, " second row"),
		pen.Ink(th.Dim, " a row that is far too long for a frame this narrow and gets cut"),
	}
	draw := func(b Box) func(w, h int) string {
		return func(w, h int) string { return b.Render(th, w, h) }
	}
	return story.Component{
		Name:  "frame",
		About: about,
		Stories: []story.Story{
			{Name: "head and foot", Note: "All three parts, rules between them.", W: 40, H: 10,
				Render: draw(Box{Head: pen.Ink(th.Text, " head"), Body: body, Foot: pen.Ink(th.Dim, " foot")})},
			{Name: "head only", Note: "No foot: the body takes its rows.", W: 40, H: 10,
				Render: draw(Box{Head: pen.Ink(th.Text, " head"), Body: body})},
			{Name: "body only", Note: "Just a border around the rows.", W: 40, H: 6,
				Render: draw(Box{Body: body})},
			{Name: "too many rows", Note: "Rows past the body's height are dropped.", W: 30, H: 4,
				Render: draw(Box{Body: body})},
		},
	}
}
