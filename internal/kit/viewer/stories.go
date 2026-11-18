//go:build gallery

package viewer

import (
	_ "embed"
	"strings"

	tea "charm.land/bubbletea/v2"

	"lgtm/internal/kit/markdown"
	"lgtm/internal/kit/story"
	"lgtm/internal/kit/theme"
)

//go:embed about.md
var about string

// sample has one of everything, and points at no file.
const sample = "# A title\n\nAn intro of a line or two, to go down a line at a time.\n\nAnd a second paragraph of it.\n\n" +
	"## Cases\n\nText above the subsections.\n\n### ✓ wide\nIt holds when wide.\n\n### ✗ narrow\nIt breaks when narrow.\n\n| part | in | out |\n|---|---|---|\n| draw | +12 | -3 |\n| size | +2 | -40 |\n\n" +
	"## Order\n\n```mermaid\nsequenceDiagram\n  participant sh as cmd/shell\n  sh->>sh: start\n  sh->>parser: read\n  parser->>sh: done\n```\n\n### start\nIt starts.\n\n### read\nIt reads.\n\n### done\nIt is done.\n\n" +
	"## Shape\n\n```mermaid\nflowchart TD\n  app --> cli\n  app --> ui[\"app/ui\"]\n  cli -->|runs| ui\n  ui --> theme\n  cli --> flags\n```\n\n### app\nThe whole.\n\n### app/ui\nWhat is drawn.\n\n### theme\nThe colours.\n\n" +
	"## Code\n\n[sample.go](sample.go#L3-L5)\n\n" +
	"## Colours\n\nThe brand is `#6b50ff`, and the rest are in the table.\n\n| name | colour |\n|---|---|\n| live | `#ff60ff` |\n| good | `#00ffb2` |\n| wrong | `rgb(235, 66, 104)` |\n| calm | `hsl(190 77% 64%)` |\n\n" +
	"## Bars\n\n| part | added | removed | note |\n|---|---|---|---|\n| new | +120 | -0 | What came. |\n| old | +0 | -300 | What went. |\n| both | +40 | -25 | |\n"

const sampleGo = "package sample\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n\nfunc Sub(a, b int) int {\n\treturn a - b\n}\n"

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
	d, err := markdown.Parse(sample, func(string) ([]byte, error) { return []byte(sampleGo), nil })
	if err != nil {
		panic(err)
	}
	// A key is one Bubble Tea names, or "block": down to the next block.
	after := func(keys ...string) func(w, h int) string {
		return func(w, h int) string {
			m := New(d).Size(w, h)
			for _, k := range keys {
				if k != "block" {
					m = m.Key(press(k))
					continue
				}
				for range 200 {
					m = m.Key(press("down"))
					if m.t.rests[m.t.at].blk >= 0 || m.t.at == len(m.t.rests)-1 {
						break
					}
				}
			}
			rows := m.Rows()
			for i := range rows {
				rows[i] = fit(rows[i], w)
			}
			return strings.Join(rows, "\n")
		}
	}
	// light draws a story in the light theme, and puts the dark one back.
	light := func(render func(w, h int) string) func(w, h int) string {
		return func(w, h int) string {
			Use(theme.Light)
			defer Use(theme.Dark)
			return render(w, h)
		}
	}
	return story.Component{
		Name:  "viewer",
		About: about,
		Stories: []story.Story{
			{Name: "opened", Note: "A document as it opens, the cursor on its title.", W: 110, H: 27, Render: after()},
			{Name: "on a line", Note: "The cursor goes down text a line at a time.", W: 110, H: 27, Render: after("down", "down")},
			{Name: "on a heading", Note: "A section's number went to its heading; the index lights the section.", W: 110, H: 27, Render: after("3")},
			{Name: "subsections", Note: "Subsections with marks: a row of buttons, the first picked. The block the cursor is in is lit.", W: 110, H: 27, Render: after("2", "block")},
			{Name: "one that breaks", Note: "Right: the subsection that does not hold. It holds a block of its own.", W: 110, H: 27, Render: after("2", "block", "right")},
			{Name: "a block in a subsection", Note: "On a line of the bars in the subsection, right picks in the bars.", W: 110, H: 27, Render: after("2", "block", "right", "down", "down", "right")},
			{Name: "a sequence", Note: "A sequence the subsections are picked from.", W: 110, H: 27, Render: after("3", "block", "right")},
			{Name: "a graph", Note: "A graph the subsections are picked from. A node drawn twice is lit twice; one with no subsection is words.", W: 110, H: 27, Render: after("4", "block", "right")},
			{Name: "code", Note: "Lines of a file, standing alone.", W: 110, H: 27, Render: after("5", "block")},
			{Name: "whole view", Note: "Enter gives the code the whole view.", W: 110, H: 27, Render: after("5", "block", "enter")},
			{Name: "colours", Note: "A colour in a code span is shown as itself, as hex, rgb or hsl, in a sentence and in a table.", W: 110, H: 27, Render: after("6")},
			{Name: "bars", Note: "Two numbers a row, to one scale.", W: 110, H: 27, Render: after("7", "block", "right")},
			{Name: "wide", Note: "A wide window: the document and its index in the middle.", W: 160, H: 33, Render: after()},
			{Name: "narrow", Note: "80 columns: no room for the index.", W: 80, H: 17, Render: after()},
			{Name: "light", Note: "On a light terminal: the light theme.", W: 110, H: 27, Render: light(after("3"))},
			{Name: "light subsections", Note: "Buttons, marks and bars in the light theme.", W: 110, H: 27, Render: light(after("2", "block", "right", "down", "down", "right"))},
			{Name: "light code", Note: "Code in the light theme.", W: 110, H: 27, Render: light(after("5", "block"))},
			{Name: "light colours", Note: "Colours a document names are the same on a light terminal.", W: 110, H: 27, Render: light(after("6"))},
		},
	}
}
