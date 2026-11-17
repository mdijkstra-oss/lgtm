//go:build gallery

package ui

import (
	_ "embed"

	tea "charm.land/bubbletea/v2"

	"strigil/internal/features/ask"
	"strigil/internal/kit/story"
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

// onBlock moves the cursor down to the first line of the next block.
func onBlock(m Model) Model {
	for range 200 {
		next, _ := m.Update(press("down"))
		m = next.(Model)
		t := m.tabs[m.cur]
		if _, rests := t.lay(m.colW()); rests[t.at].blk >= 0 || t.at == len(rests)-1 {
			break
		}
	}
	return m
}

func Stories() story.Component {
	// A key is one Bubble Tea names, or "block": down to the next block.
	after := func(keys ...string) func(w, h int) string {
		return func(w, h int) string {
			m := New(ask.Answers())
			step := func(msg tea.Msg) {
				next, _ := m.Update(msg)
				m = next.(Model)
			}
			step(tea.WindowSizeMsg{Width: w, Height: h})
			for _, k := range keys {
				if k == "block" {
					m = onBlock(m)
					continue
				}
				step(press(k))
			}
			return m.render()
		}
	}
	// The second and third answers, each opened from the palette.
	second := func(keys ...string) []string { return append([]string{"down", "enter"}, keys...) }
	third := func(keys ...string) []string { return append([]string{"down", "down", "enter"}, keys...) }
	return story.Component{
		Name:  "ask",
		About: about,
		Stories: []story.Story{
			{Name: "palette", Note: "How it opens: nothing asked, every question there is an answer for.", W: 110, H: 34, Render: after()},
			{Name: "palette typed", Note: "Typing narrows the questions.", W: 110, H: 34, Render: after("c", "o", "l")},
			{Name: "palette nothing", Note: "A question there is no answer for.", W: 110, H: 34, Render: after("w", "h", "y")},
			{Name: "nothing asked", Note: "The palette closed before anything was asked.", W: 110, H: 34, Render: after("esc")},
			{Name: "answer", Note: "An answer as it opens, the cursor on its first line. Blocks have a grey outline, and show the code of what is picked in them.", W: 110, H: 34, Render: after("enter")},
			{Name: "on a line", Note: "The cursor goes down text a line at a time.", W: 110, H: 34, Render: after("enter", "down", "down")},
			{Name: "on a block", Note: "On a line of a block: the block is lit, and left and right pick in it.", W: 110, H: 34, Render: after("enter", "block")},
			{Name: "a hop", Note: "Right twice: the third hop is picked. The cursor stays on its line.", W: 110, H: 34, Render: after("enter", "block", "right", "right")},
			{Name: "whole view", Note: "Enter gives the code the whole view.", W: 110, H: 34, Render: after("enter", "block", "right", "right", "enter")},
			{Name: "on a heading", Note: "A section's number went to its heading; the index lights the section.", W: 110, H: 34, Render: after("enter", "3")},
			{Name: "a case that breaks", Note: "In a check, stepped to the case that does not hold.", W: 110, H: 34, Render: after("enter", "3", "block", "right")},
			{Name: "colours", Note: "The second answer: the palette as itself, the cursor on a colour nothing uses.", W: 110, H: 34, Render: after(second("block", "right", "right", "right", "right", "right")...)},
			{Name: "four tries", Note: "A check whose cases are things that were tried.", W: 110, H: 34, Render: after(second("4", "block", "right", "right")...)},
			{Name: "a commit", Note: "The third answer: what a commit took out and put in.", W: 110, H: 34, Render: after(third("block", "right", "right", "right", "right", "right")...)},
			{Name: "a run", Note: "Commands as they ended when the answer was written.", W: 110, H: 34, Render: after(third("2", "block", "right", "right", "right")...)},
			{Name: "two tabs", Note: "A second question asked; its tab beside the first.", W: 110, H: 34, Render: after("enter", "ctrl+k", "down", "enter")},
			{Name: "asking again", Note: "The palette over an answer.", W: 110, H: 34, Render: after("enter", "ctrl+k", "i", "n")},
			{Name: "wide", Note: "A wide window: the notebook and its index in the middle.", W: 160, H: 40, Render: after("enter")},
			{Name: "narrow", Note: "80 columns: no room for the index.", W: 80, H: 24, Render: after("enter")},
			{Name: "too small", Note: "No room to draw in.", W: 50, H: 10, Render: after("enter")},
		},
	}
}
