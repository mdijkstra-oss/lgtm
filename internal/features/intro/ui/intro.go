// Package ui is the intro page: the screen starts covered in slop, a blade
// scrapes it off, and what is under it is an empty page.
package ui

import (
	"context"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"lgtm/internal/kit/scrape"
	"lgtm/internal/kit/theme"
)

var (
	pal theme.Theme

	sKey, sDoes, sRule lipgloss.Style
)

func init() { use(theme.Dark) }

// use draws everything from here on in that theme.
func use(t theme.Theme) {
	pal = t
	scrape.Use(t)
	sKey = lipgloss.NewStyle().Foreground(t.Text3)
	sDoes = lipgloss.NewStyle().Foreground(t.Text4)
	sRule = lipgloss.NewStyle().Foreground(t.Line)
}

type Model struct {
	w, h int
	slop scrape.Model
}

type tickMsg struct{}

// New makes the page with sentences picked afresh.
func New() Model { return Model{slop: scrape.New()} }

// seeded makes the page with the same pick every time, for showing it.
func seeded(seed uint64) Model { return Model{slop: scrape.Seeded(seed)} }

func Run(ctx context.Context) error {
	use(theme.For(lipgloss.HasDarkBackground(os.Stdin, os.Stdout)))
	_, err := tea.NewProgram(New(), tea.WithContext(ctx)).Run()
	return err
}

func (m Model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	return tea.Tick(time.Second/scrape.FPS, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.slop = m.slop.Size(m.w, m.h)
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "space":
			was := m.slop.Busy()
			m.slop = m.slop.Soil()
			if !was {
				return m, tick()
			}
		}
	case tickMsg:
		// Once the page is clean nothing moves, so the clock stops.
		if m = m.step(); m.slop.Busy() {
			return m, tick()
		}
	}
	return m, nil
}

func (m Model) step() Model {
	m.slop = m.slop.Step()
	return m
}

func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.WindowTitle = "lgtm"
	v.BackgroundColor = pal.Surface0
	if m.w >= 60 && m.h >= 12 {
		v.SetContent(m.render())
	}
	return v
}

func (m Model) render() string { return m.slop.Over(app(m.w, m.h)) }

// app is what the blade uncovers: nothing, and at the bottom the keys.
func app(w, h int) []string {
	rows := make([]string, h)
	for y := range rows {
		rows[y] = strings.Repeat(" ", w)
	}
	keys := "    " + sKey.Render("space") + " " + sDoes.Render("again") + " " + sRule.Render("•") + " " + sKey.Render("q") + " " + sDoes.Render("quit")
	rows[h-1] = keys + strings.Repeat(" ", max(0, w-lipgloss.Width(keys)))
	return rows
}
