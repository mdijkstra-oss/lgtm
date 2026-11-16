// Package gallery shows every component's stories, one at a time, in the
// terminal. Only a command behind the gallery build tag imports it, so a
// normal build does not contain it.
package gallery

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"strigil/internal/kit"
	"strigil/internal/kit/hints"
	"strigil/internal/kit/story"
	"strigil/internal/kit/text"
	"strigil/internal/kit/theme"
)

const (
	listWidth = 30
	aboutRows = 9
)

func Run(ctx context.Context, th theme.Theme, components []story.Component) error {
	m := model{th: th, comps: components}
	for ci, c := range components {
		for si := range c.Stories {
			m.at = append(m.at, [2]int{ci, si})
		}
	}
	_, err := tea.NewProgram(m, tea.WithContext(ctx)).Run()
	return err
}

type model struct {
	th    theme.Theme
	comps []story.Component
	at    [][2]int // every story as component, story
	sel   int
	w, h  int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up":
			m.sel = max(0, m.sel-1)
		case "down":
			m.sel = min(len(m.at)-1, m.sel+1)
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.WindowTitle = "strigil gallery"
	v.BackgroundColor = m.th.Panel
	if m.w == 0 || len(m.at) == 0 {
		return v
	}
	h := m.h - 1
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.list(listWidth, h), m.stage(m.w-listWidth, h))
	v.SetContent(body + "\n" + m.footer())
	return v
}

func (m model) list(w, h int) string {
	pen := theme.On(m.th.Panel)
	rows := []string{"", pen.Gap(2) + pen.Style(m.th.Accent).Bold(true).Render("gallery"), ""}
	selRow := 0
	i := 0
	for _, c := range m.comps {
		rows = append(rows, pen.Gap(2)+pen.Style(m.th.Soft).Bold(true).Render(text.Head(c.Name, w-4)))
		for _, s := range c.Stories {
			if i == m.sel {
				on := theme.On(m.th.Band)
				selRow = len(rows)
				rows = append(rows, on.Line(on.Ink(m.th.Accent, "▌")+on.Gap(3)+on.Ink(m.th.Text, text.Head(s.Name, w-6)), w))
			} else {
				rows = append(rows, pen.Gap(4)+pen.Ink(m.th.Dim, text.Head(s.Name, w-6)))
			}
			i++
		}
		rows = append(rows, "")
	}
	if selRow >= h {
		rows = rows[selRow-h+2:]
	}
	return strings.Join(pen.Block(rows, w, h), "\n")
}

// stage draws the selected story at its own size, centred on a darker ground
// so its edges show.
func (m model) stage(w, h int) string {
	pen := theme.On(m.th.Bg)
	c := m.comps[m.at[m.sel][0]]
	s := c.Stories[m.at[m.sel][1]]

	size := pen.Ink(m.th.Faint, fmt.Sprintf("%d × %d", s.W, s.H)) + pen.Gap(2)
	rule := pen.Ink(m.th.Rule, strings.Repeat("─", w))
	rows := []string{
		pen.Split(pen.Gap(2)+pen.Ink(m.th.Dim, c.Name+" · ")+pen.Style(m.th.Text).Bold(true).Render(s.Name), size, w),
		pen.Gap(2) + pen.Ink(m.th.Soft, text.Head(s.Note, w-4)),
		rule,
	}

	ground := theme.On(m.th.Deep)
	room := max(1, h-len(rows)-aboutRows-1)
	drawn := strings.Split(s.Render(s.W, s.H), "\n")
	left := max(0, (w-s.W)/2)
	stage := make([]string, 0, room)
	for range max(0, (room-len(drawn))/2) {
		stage = append(stage, "")
	}
	for _, line := range drawn {
		stage = append(stage, ground.Gap(left)+line)
	}
	rows = append(rows, ground.Block(stage, w, room)...)

	rows = append(rows, rule)
	for _, line := range strings.Split(strings.TrimSpace(c.About), "\n") {
		rows = append(rows, pen.Gap(2)+pen.Ink(m.th.Dim, text.Head(line, w-4)))
	}
	return strings.Join(pen.Block(rows, w, h), "\n")
}

func (m model) footer() string {
	pen := theme.On(m.th.Panel)
	keys := []kit.Hint{{Key: "↑↓", Label: "story"}, {Key: "q", Label: "quit"}}
	count := pen.Ink(m.th.Faint, fmt.Sprintf("%d of %d", m.sel+1, len(m.at))) + pen.Gap(1)
	return pen.Split(pen.Gap(1)+hints.Render(m.th, pen, keys, "  ·  "), count, m.w)
}
