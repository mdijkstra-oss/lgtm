//go:build gallery

package ui

import (
	_ "embed"

	tea "charm.land/bubbletea/v2"

	"strigil/internal/features/ask"
	"strigil/internal/kit/story"
	"strigil/internal/kit/theme"
)

//go:embed about.md
var about string

// press is a key by the name Bubble Tea gives it.
func press(key string) tea.KeyPressMsg {
	switch key {
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "ctrl+k":
		return tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl}
	}
	return tea.KeyPressMsg{Code: []rune(key)[0], Text: key}
}

func Stories() story.Component {
	after := func(keys ...string) func(w, h int) string {
		return func(w, h int) string {
			m := New(ask.Docs())
			step := func(msg tea.Msg) {
				next, _ := m.Update(msg)
				m = next.(Model)
			}
			step(tea.WindowSizeMsg{Width: w, Height: h})
			for _, k := range keys {
				step(press(k))
			}
			return m.render()
		}
	}
	// light draws a story in the light theme, and puts the dark one back.
	light := func(render func(w, h int) string) func(w, h int) string {
		return func(w, h int) string {
			use(theme.Light)
			defer use(theme.Dark)
			return render(w, h)
		}
	}
	return story.Component{
		Name:  "ask",
		About: about,
		Stories: []story.Story{
			{Name: "palette", Note: "How it opens: nothing open, the title of every document there is.", W: 110, H: 34, Render: after()},
			{Name: "palette typed", Note: "Typing narrows the titles.", W: 110, H: 34, Render: after("m", "a", "d", "e")},
			{Name: "palette nothing", Note: "A title no document has.", W: 110, H: 34, Render: after("w", "h", "y")},
			{Name: "nothing open", Note: "The palette closed before anything was opened.", W: 110, H: 34, Render: after("esc")},
			{Name: "document", Note: "A document open: its tab, the document, the keys.", W: 110, H: 34, Render: after("enter")},
			{Name: "in a block", Note: "The keys of the block the cursor is in come first in the help line.", W: 110, H: 34, Render: after("enter", "3", "down", "down", "down", "down", "down", "down", "down", "down", "down")},
			{Name: "opening another", Note: "The palette over a document.", W: 110, H: 34, Render: after("enter", "ctrl+k", "c", "a")},
			{Name: "wide", Note: "A wide window: everything in the middle.", W: 160, H: 40, Render: after("enter")},
			{Name: "narrow", Note: "80 columns.", W: 80, H: 24, Render: after("enter")},
			{Name: "light palette", Note: "On a light terminal: the palette in the light theme.", W: 110, H: 34, Render: light(after())},
			{Name: "light document", Note: "A document in the light theme.", W: 110, H: 34, Render: light(after("enter"))},
			{Name: "too small", Note: "No room to draw in.", W: 50, H: 10, Render: after("enter")},
		},
	}
}
