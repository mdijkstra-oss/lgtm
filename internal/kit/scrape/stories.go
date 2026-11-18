//go:build gallery

package scrape

import (
	_ "embed"
	"strings"

	"lgtm/internal/kit/story"
	"lgtm/internal/kit/theme"
)

//go:embed about.md
var about string

func Stories() story.Component {
	// A page with words on it, so that what falls can be seen to miss them.
	page := func(w, h int) []string {
		rows := make([]string, h)
		for y := range rows {
			text := ""
			if y%4 == 1 {
				text = "    a line of the page under the slop"
			}
			rows[y] = text + strings.Repeat(" ", w-len(text))
		}
		return rows
	}
	after := func(frames int) func(w, h int) string {
		return func(w, h int) string {
			m := Seeded(7).Size(w, h)
			for range frames {
				m = m.Step()
			}
			return m.Over(page(w, h))
		}
	}
	return story.Component{
		Name:  "scrape",
		About: about,
		Stories: []story.Story{
			{Name: "covered", Note: "Before the blade comes.", W: 110, H: 34, Render: after(15)},
			{Name: "scraping", Note: "The blade part of the way across, the page behind it.", W: 110, H: 34, Render: after(52)},
			{Name: "scraped", Note: "The page, and nothing over it.", W: 110, H: 34, Render: after(200)},
			{Name: "light scraping", Note: "On a light terminal.", W: 110, H: 34, Render: func(w, h int) string {
				Use(theme.Light)
				defer Use(theme.Dark)
				return after(52)(w, h)
			}},
		},
	}
}
