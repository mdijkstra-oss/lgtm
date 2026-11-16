//go:build gallery

package palette

import (
	"context"
	_ "embed"
	"strings"

	"strigil/internal/kit"
	"strigil/internal/kit/story"
	"strigil/internal/kit/theme"
)

//go:embed about.md
var about string

type fixedSource string

func (s fixedSource) Name() string                                         { return string(s) }
func (s fixedSource) Search(context.Context, string, func(kit.Item)) error { return nil }

func Stories(th theme.Theme) story.Component {
	pen := theme.On(th.Panel)
	preview := func(w, h int) string {
		rows := []string{pen.Ink(th.Soft, "notes.txt"), ""}
		for i := len(rows); i < h; i++ {
			rows = append(rows, pen.Ink(th.Dim, "a line of the thing being previewed"))
		}
		return strings.Join(rows, "\n")
	}
	found := []kit.Item{
		{Title: "Trimmed the lantern wick before dusk.", Marks: [][2]int{{12, 19}}, Location: "harbour-log.txt:27", Source: "files", Preview: preview},
		{Title: "The lantern was lit early because of the fog.", Marks: [][2]int{{4, 11}}, Location: "harbour-log.txt:44", Source: "files", Preview: preview},
		{Title: "A long line where the match sits a good way in, past what fits: lantern at last.", Marks: [][2]int{{64, 71}}, Location: "ledger/arrivals.txt:7", Source: "files", Preview: preview},
		{Title: "Light the lantern half an hour before sunset.", Marks: [][2]int{{10, 17}}, Location: "lantern-room.txt:9", Source: "files", Preview: preview},
	}
	state := func(query string, set func(*Model)) func(w, h int) string {
		return func(w, h int) string {
			m, _ := New(th, []kit.Source{fixedSource("files"), fixedSource("commands")}).Open()
			m.input.SetValue(query)
			if set != nil {
				set(&m)
			}
			return m.Resize(w, h).View()
		}
	}
	return story.Component{
		Name:  "palette",
		About: about,
		Stories: []story.Story{
			{Name: "empty", Note: "Just opened: says what can be searched.", W: 90, H: 16, Render: state("", nil)},
			{Name: "searching", Note: "A query with nothing back yet.", W: 90, H: 16,
				Render: state("lantern", func(m *Model) { m.searching = true })},
			{Name: "results", Note: "Second result under the cursor, preview beside the list.", W: 90, H: 16,
				Render: state("lantern", func(m *Model) { m.items, m.cursor = found, 1 })},
			{Name: "results narrow", Note: "No room for the preview.", W: 60, H: 14,
				Render: state("lantern", func(m *Model) { m.items, m.cursor = found, 2 })},
			{Name: "more arriving", Note: "Results on screen while the search still runs.", W: 90, H: 16,
				Render: state("lantern", func(m *Model) { m.items, m.searching = found[:2], true })},
			{Name: "capped", Note: "The search was stopped at the limit.", W: 90, H: 16,
				Render: state("e", func(m *Model) { m.items, m.capped = found, true })},
			{Name: "nothing", Note: "A finished search with no results.", W: 90, H: 16, Render: state("zzzq", nil)},
		},
	}
}
