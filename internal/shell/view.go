package shell

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"strigil/internal/kit"
	"strigil/internal/kit/hints"
	"strigil/internal/kit/tabs"
	"strigil/internal/kit/text"
	"strigil/internal/kit/theme"
)

const (
	sidebarWidth = 30
	wideFrom     = 100
	minWidth     = 44
	minHeight    = 14
)

func (m Model) wide() bool { return m.w >= wideFrom }

func (m Model) pageSize() (w, h int) {
	if m.wide() {
		return m.w - sidebarWidth, m.h - 1
	}
	return m.w, m.h - 2
}

func (m Model) pageTop() int {
	if m.wide() {
		return 0
	}
	return 1
}

// paletteBox keeps the palette over the page column and in its upper half,
// so the sidebar and the lower part of the page stay in sight behind it.
func (m Model) paletteBox() (x, y, w, h int) {
	pw, _ := m.pageSize()
	w = min(pw, 108)
	// A few columns of page showing beside the palette read as a glitch.
	if pw-w < 12 {
		w = pw
	}
	h = max(10, min(m.h/2, 20))
	return max(0, (pw-w)/2), m.pageTop() + 2, w, h
}

// covered is how many rows of the page the palette hides, from the top.
func (m Model) covered() int {
	_, y, _, h := m.paletteBox()
	return y + h - m.pageTop()
}

func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.WindowTitle = "strigil"
	// The renderer leaves the bottom-right cell unpainted; this makes it
	// match the footer instead of the terminal's own colour.
	v.BackgroundColor = m.th.Panel
	switch {
	case m.w == 0:
	case m.w < minWidth || m.h < minHeight:
		pen := theme.On(m.th.Bg)
		rows := pen.Middle([]string{pen.Ink(m.th.Dim, "Make the window a little larger")}, m.w, m.h)
		v.SetContent(strings.Join(rows, "\n"))
	default:
		v.SetContent(m.frame())
		if c := m.pal.Cursor(); c != nil {
			x, y, _, _ := m.paletteBox()
			c.X += x
			c.Y += y
			v.Cursor = c
		}
	}
	return v
}

func (m Model) frame() string {
	pw, ph := m.pageSize()
	entries := make([]tabs.Entry, len(m.pages))
	for i, p := range m.pages {
		entries[i] = tabs.Entry{Title: p.Title(), Meta: p.Meta()}
	}

	var body string
	if m.wide() {
		body = lipgloss.JoinHorizontal(lipgloss.Top, m.page(pw, ph), tabs.Sidebar(m.th, entries, m.active, m.lit, sidebarWidth, ph))
	} else {
		body = tabs.Strip(m.th, entries, m.active, m.lit, m.w) + "\n" + m.page(pw, ph)
	}
	screen := body + "\n" + m.footer()

	if !m.pal.IsOpen() {
		return screen
	}
	x, y, _, _ := m.paletteBox()
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(screen),
		lipgloss.NewLayer(m.pal.View()).X(x).Y(y).Z(1),
	).Render()
}

func (m Model) page(w, h int) string {
	pen := theme.On(m.th.Bg)
	if len(m.pages) == 0 {
		rows := []string{
			pen.Style(m.th.Accent).Bold(true).Render("strigil"),
			"",
			pen.Ink(m.th.Dim, "Nothing open yet"),
			"",
			pen.Ink(m.th.Text, "ctrl+k") + pen.Ink(m.th.Dim, " to search"),
		}
		return strings.Join(pen.Middle(rows, w, h), "\n")
	}
	return strings.Join(pen.Block(strings.Split(m.pages[m.active].View(), "\n"), w, h), "\n")
}

func (m Model) footer() string {
	pen := theme.On(m.th.Panel)
	var keys []kit.Hint
	if !m.pal.IsOpen() {
		keys = append(keys, kit.Hint{Key: "ctrl+k", Label: "search"})
		if len(m.pages) > 1 {
			keys = append(keys, kit.Hint{Key: "[ ]", Label: "tabs"})
		}
		if len(m.pages) > 0 {
			keys = append(keys, m.pages[m.active].Hints()...)
			keys = append(keys, kit.Hint{Key: "x", Label: "close tab"})
		}
		keys = append(keys, kit.Hint{Key: "q", Label: "quit"})
	}
	left := pen.Gap(1) + hints.Render(m.th, pen, keys, "  ·  ")
	right := pen.Ink(m.th.Faint, text.Tail(m.dir, max(0, m.w/3))) + pen.Gap(1)
	return pen.Split(left, right, m.w)
}
