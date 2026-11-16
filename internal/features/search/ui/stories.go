//go:build gallery

package ui

import (
	_ "embed"
	"errors"

	"strigil/internal/kit/story"
	"strigil/internal/kit/theme"
)

//go:embed about.md
var about string

func Stories(th theme.Theme) story.Component {
	matcher := newQuery("", "lantern").Matcher()
	docs := []*doc{
		index("harbour-log.txt", []string{
			"HARBOUR LOG",
			"",
			"Day 1",
			"  Trimmed the lantern wick and cleaned the glass before dusk.",
			"  The ferry was late by an hour.",
			"",
			"Day 2",
			"  The lantern was lit early; a second lantern was hung at the east wall.",
			"\tA line indented with a tab, about the lantern.",
		}, matcher),
		index("ledger/arrivals.txt", []string{
			"LEDGER OF ARRIVALS",
			"Northern Lantern, from the north, with timber. Waited outside for the evening tide and came in by the light of the lantern.",
		}, matcher),
	}
	page := func(fill *loadedMsg, set func(*Page)) func(w, h int) string {
		return func(w, h int) string {
			p := Page{th: th, query: "lantern"}.Resize(w, h).(Page)
			if fill != nil {
				p = p.filled(*fill)
			}
			if set != nil {
				set(&p)
			}
			return p.View()
		}
	}
	found := &loadedMsg{docs: docs}
	return story.Component{
		Name:  "search",
		About: about,
		Stories: []story.Story{
			{Name: "results", Note: "Hits grouped by file, first one selected.", W: 80, H: 18, Render: page(found, nil)},
			{Name: "second hit on a line", Note: "Two hits share a row; the selected one is marked.", W: 80, H: 18,
				Render: page(found, func(p *Page) { p.sel = 2 })},
			{Name: "long line", Note: "The row is cut around the selected hit.", W: 80, H: 18,
				Render: page(found, func(p *Page) { p.sel = len(p.hits) - 1 })},
			{Name: "file open", Note: "Enter: the file over the list, centred on the hit.", W: 80, H: 18,
				Render: page(found, func(p *Page) { p.sel, p.viewing = 1, true })},
			{Name: "narrow", Note: "Less room, same layout.", W: 46, H: 14, Render: page(found, nil)},
			{Name: "loading", Note: "Opened, results not in yet.", W: 80, H: 10, Render: page(nil, nil)},
			{Name: "nothing", Note: "The search found nothing.", W: 80, H: 10, Render: page(&loadedMsg{}, nil)},
			{Name: "failed", Note: "The search could not run.", W: 80, H: 10,
				Render: page(&loadedMsg{err: errors.New("samples: no such file or directory")}, nil)},
			{Name: "preview", Note: "What the palette shows beside its list.", W: 50, H: 10,
				Render: func(w, h int) string { return preview(th, docs[0], 1, w, h) }},
		},
	}
}
