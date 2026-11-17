//go:build gallery

package ui

import (
	_ "embed"

	tea "charm.land/bubbletea/v2"

	"strigil/internal/kit/story"
)

//go:embed about.md
var about string

func Stories() story.Component {
	after := func(frames int) func(w, h int) string {
		return func(w, h int) string {
			next, _ := seeded(7).Update(tea.WindowSizeMsg{Width: w, Height: h})
			m := next.(Model)
			for range frames {
				m = m.step()
			}
			return m.render()
		}
	}
	return story.Component{
		Name:  "intro",
		About: about,
		Stories: []story.Story{
			{Name: "slop", Note: "Before the blade comes.", W: 110, H: 34, Render: after(15)},
			{Name: "scraping", Note: "The blade part of the way across, drawing the rule as it goes.", W: 110, H: 34, Render: after(52)},
			{Name: "clean", Note: "Scraped: the app.", W: 110, H: 34, Render: after(200)},
			{Name: "small", Note: "The same, in a small window.", W: 80, H: 24, Render: after(200)},
		},
	}
}
