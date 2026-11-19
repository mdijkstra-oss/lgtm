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
	doc := ask.Builtin(ask.Root())[0]
	start := Start{
		Repo: "shop", Branch: "main", Uncommitted: 3,
		History: []ask.Commit{
			{Hash: "9c1e7a2", Subject: "feat(orders): reservations expire after fifteen minutes", Author: "Ada Okafor", Day: "Today"},
			{Hash: "41b0d93", Subject: "fix(payments): retry a charge once when the gateway times out", Author: "Jun Tanaka", Day: "Today"},
			{Hash: "e77f120", Subject: "refactor(store): one ledger for stock, in place of two tables", Author: "Ada Okafor", Day: "Yesterday"},
			{Hash: "0aa94c6", Subject: "feat(api): a view of an order with its payments", Author: "Mira Sandoval", Day: "Monday"},
			{Hash: "b3d5518", Subject: "chore: bump the router", Author: "Jun Tanaka", Day: "Monday"},
			{Hash: "6f2c8e0", Subject: "test(store): cover a reservation that is paid as it expires", Author: "Ada Okafor", Day: "Monday"},
			{Hash: "d019a7b", Subject: "feat(orders): cancel an order that was never paid", Author: "Mira Sandoval", Day: "Monday"},
			{Hash: "27e4f36", Subject: "docs: how stock is counted", Author: "Jun Tanaka", Day: "Monday"},
			{Hash: "c5510de", Subject: "feat(api): list the orders of a customer", Author: "Mira Sandoval", Day: "Monday"},
			{Hash: "88a13f4", Subject: "fix(store): count a returned item once", Author: "Ada Okafor", Day: "Monday"},
			{Hash: "1d0b6a9", Subject: "Start the shop", Author: "Ada Okafor", Day: "3 November"},
		},
		Answers: []ask.Answer{
			{Doc: doc, Asked: "What happens, step by step, when an order is placed?", When: "12 min ago"},
			{Doc: doc, Asked: "Commit e77f120: refactor(store): one ledger for stock", Change: "e77f120", When: "1 h ago"},
			{Doc: doc, Asked: "Why are there two kinds of reservation?", When: "3 h ago"},
		},
	}
	// after draws the page a start gives, once those keys were pressed.
	from := func(start Start, keys ...string) func(w, h int) string {
		return func(w, h int) string {
			m := New(start)
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
	after := func(keys ...string) func(w, h int) string { return from(start, keys...) }
	// reading is the page with the document that shows every block open,
	// which is the first of the earlier questions.
	reading := func(keys ...string) func(w, h int) string { return after(append([]string{"1"}, keys...)...) }
	// light draws a story in the light theme, and puts the dark one back.
	light := func(render func(w, h int) string) func(w, h int) string {
		return func(w, h int) string {
			use(theme.Light)
			defer use(theme.Dark)
			return render(w, h)
		}
	}
	// asked draws the page while a document is being written.
	asked := func(w work) func(w, h int) string {
		return func(wide, high int) string {
			next, _ := New(start).Update(tea.WindowSizeMsg{Width: wide, Height: high})
			m := next.(Model)
			if w.question == "" {
				w.question = "How does a markdown file become what is on screen in the viewer?"
			}
			m.work = &w
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
	bare := Start{Repo: "shop", Branch: "main", History: start.History[:3]}
	return story.Component{
		Name:  "ask",
		About: about,
		Stories: []story.Story{
			{Name: "history", Note: "How it opens: what is not committed, then the commits, and under them the questions answered before. A commit explained before has a tick.", W: 110, H: 34, Render: after()},
			{Name: "history picked", Note: "Down the list: the picked row is one fill. A day's name is passed over.", W: 110, H: 34, Render: after("down", "down", "down")},
			{Name: "history a long day", Note: "A day of many commits lists four and folds the rest; the fold can be picked.", W: 110, H: 34, Render: after("down", "down", "down", "down", "down", "down", "down", "down")},
			{Name: "history a day opened", Note: "Enter on the fold lists the day whole.", W: 110, H: 34, Render: after("down", "down", "down", "down", "down", "down", "down", "down", "enter")},
			{Name: "history typed", Note: "Typing narrows the commits to those with every word, in hash, subject or author.", W: 110, H: 34, Render: after("a", "d", "a", " ", "s", "t", "o", "r", "e")},
			{Name: "history a question", Note: "What no commit has is a question: enter asks it.", W: 110, H: 34, Render: after("w", "h", "y", " ", "t", "w", "o", " ", "t", "a", "b", "l", "e", "s", "?")},
			{Name: "history bare", Note: "Nothing uncommitted and nothing asked before: the commits alone.", W: 110, H: 34, Render: from(bare)},
			{Name: "history empty", Note: "No commits at all.", W: 110, H: 34, Render: from(Start{Repo: "scratch"})},
			{Name: "history narrow", Note: "80 columns, 24 rows: the list scrolls under the cursor.", W: 80, H: 24, Render: after("down", "down", "down", "down", "down", "down")},
			{Name: "history over a document", Note: "Back on the history with a document open: its tab stays, and esc goes back to it.", W: 110, H: 34, Render: reading("ctrl+k")},
			{Name: "document", Note: "A document open: its tab, the document, the keys.", W: 110, H: 34, Render: reading()},
			{Name: "in a block", Note: "The keys of the block the cursor is in come first in the help line.", W: 110, H: 34, Render: reading("3", "down", "down", "down", "down", "down", "down", "down", "down", "down")},
			{Name: "wide", Note: "A wide window: everything in the middle.", W: 160, H: 40, Render: reading()},
			{Name: "narrow", Note: "80 columns.", W: 80, H: 24, Render: reading()},
			{Name: "light history", Note: "On a light terminal: the history in the light theme.", W: 110, H: 34, Render: light(after("down"))},
			{Name: "light document", Note: "A document in the light theme.", W: 110, H: 34, Render: light(reading())},
			{Name: "too small", Note: "No room to draw in.", W: 50, H: 10, Render: after()},
			{Name: "asked", Note: "A question just asked: nothing has come back yet.", W: 110, H: 34, Render: asked(work{})},
			{Name: "explaining", Note: "A commit picked from the history: the page is named for it.", W: 110, H: 34, Render: asked(work{question: "Commit 9c1e7a2: feat(orders): reservations expire after fifteen minutes", steps: []step{{id: 1, tool: "git", what: "show 9c1e7a2 -- internal/orders", note: "84 lines", done: true}, {id: 2, tool: "read", what: "internal/orders/store.go 40–120"}}, frame: 2, took: 4 * time.Second})},
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
				return slop.Over(strings.Split(after()(w, h), "\n"))
			}},
		},
	}
}
