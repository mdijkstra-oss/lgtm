//go:build gallery

package tabs

import (
	_ "embed"

	"strigil/internal/kit/story"
	"strigil/internal/kit/theme"
)

//go:embed about.md
var about string

func Stories(th theme.Theme) story.Component {
	entries := []Entry{
		{Title: "search · lantern", Meta: "27"},
		{Title: "search · ledger", Meta: "21"},
		{Title: "search · a very long thing somebody typed", Meta: "3"},
	}
	side := func(es []Entry, active, lit int) func(w, h int) string {
		return func(w, h int) string { return Sidebar(th, es, active, lit, w, h) }
	}
	strip := func(es []Entry, active, lit int) func(w, h int) string {
		return func(w, h int) string { return Strip(th, es, active, lit, w) }
	}
	return story.Component{
		Name:  "tabs",
		About: about,
		Stories: []story.Story{
			{Name: "sidebar", Note: "Second tab active.", W: 30, H: 12, Render: side(entries, 1, -1)},
			{Name: "sidebar lit", Note: "First tab lit by a hovered result.", W: 30, H: 12, Render: side(entries, 1, 0)},
			{Name: "sidebar empty", Note: "Nothing open.", W: 30, H: 12, Render: side(nil, 0, -1)},
			{Name: "strip", Note: "The narrow layout.", W: 70, H: 1, Render: strip(entries, 0, 1)},
			{Name: "strip crowded", Note: "Leading tabs drop so the active one fits.", W: 36, H: 1, Render: strip(entries, 2, -1)},
			{Name: "strip empty", Note: "Nothing open.", W: 40, H: 1, Render: strip(nil, 0, -1)},
		},
	}
}
