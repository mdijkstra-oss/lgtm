//go:build gallery

package ui

import (
	_ "embed"
	"errors"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"lgtm/internal/features/ask"
	"lgtm/internal/kit/scrape"
	"lgtm/internal/kit/story"
	"lgtm/internal/kit/theme"
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
			m := New(ask.Builtin(ask.Root()))
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
	// asked draws the page while a question is being answered.
	asked := func(w work) func(w, h int) string {
		return func(wide, high int) string {
			next, _ := New(ask.Builtin(ask.Root())).Update(tea.WindowSizeMsg{Width: wide, Height: high})
			m := next.(Model)
			w.question = "How does a markdown file become what is on screen in the viewer?"
			m.choosing, m.work = false, &w
			return m.render()
		}
	}
	calls := []step{
		{id: 1, tool: "read", what: "README.md", note: "5 lines", done: true},
		{id: 2, tool: "list", what: "internal/kit", note: "14 entries", done: true},
		{id: 3, tool: "grep", what: `func Parse|viewer\.New in internal`, note: "6 in 4 files", done: true},
		{id: 4, tool: "read", what: "internal/kit/markdown/parse.go", note: "150 lines", done: true},
		{id: 5, tool: "read", what: "internal/kit/markdown/nothing.go", note: "no such file", done: true, bad: true},
		{id: 6, tool: "git", what: "log --oneline -20 -- internal/kit/viewer", note: "20 lines", done: true},
		{id: 7, tool: "read", what: "internal/kit/viewer/viewer.go 84–240"},
	}
	thinking := "The parser hands back a Doc of cells and never draws. So the next thing to find is where a cell becomes something with a view: build in blocks.go looks like it, and New calls it for every cell up front."
	var many []step
	for range 6 {
		many = append(many, calls[:6]...)
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
			{Name: "asked", Note: "A question just asked: nothing has come back yet.", W: 110, H: 34, Render: asked(work{})},
			{Name: "answering", Note: "The model at work: what it is at, what it thinks, the calls it made, the one it waits on last.", W: 110, H: 34, Render: asked(work{steps: calls, think: thinking, frame: 3, took: 42 * time.Second})},
			{Name: "answering long", Note: "More calls than there are rows: the latest stay.", W: 110, H: 34, Render: asked(work{steps: append(many, calls[6]), think: thinking, frame: 5, took: 95 * time.Second})},
			{Name: "writing", Note: "The document handed in and being checked.", W: 110, H: 34, Render: asked(work{steps: append(calls[:6:6], step{id: 8, tool: "write_document"}), frame: 1, took: 71 * time.Second})},
			{Name: "failed", Note: "The model could not be asked: what went wrong, and the way out.", W: 110, H: 34, Render: asked(work{steps: calls[:2], took: 3 * time.Second, err: errors.New("LGTM_MODEL is not set, here or in ~/.lgtm/config.json: it names the model to ask")})},
			{Name: "answering narrow", Note: "80 columns.", W: 80, H: 24, Render: asked(work{steps: calls, think: thinking, frame: 3, took: 42 * time.Second})},
			{Name: "light answering", Note: "The model at work, in the light theme.", W: 110, H: 34, Render: light(asked(work{steps: calls, think: thinking, frame: 3, took: 42 * time.Second}))},
			{Name: "covered", Note: "How it opens: the blade part of the way across, the page behind it.", W: 110, H: 34, Render: func(w, h int) string {
				slop := scrape.Seeded(7).Size(w, h)
				for range 52 {
					slop = slop.Step()
				}
				return slop.Over(strings.Split(asked(work{steps: calls[:3], think: thinking, took: 2 * time.Second})(w, h), "\n"))
			}},
		},
	}
}
