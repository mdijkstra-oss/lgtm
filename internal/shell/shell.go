// Package shell is the TUI frame: tabs, the palette, and the routing between
// them. It knows features only through the kit contracts.
package shell

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"strigil/internal/kit"
	"strigil/internal/kit/palette"
	"strigil/internal/kit/theme"
)

type Options struct {
	Theme   theme.Theme
	Sources []kit.Source
	Dir     string
}

func Run(ctx context.Context, o Options) error {
	_, err := tea.NewProgram(New(o), tea.WithContext(ctx)).Run()
	return err
}

type Model struct {
	th   theme.Theme
	dir  string
	w, h int

	pages  []kit.Page
	active int
	lit    int // tab holding the hovered result, or -1

	// held is the active page as it was before a hovered result moved it.
	held    kit.Page
	holding bool

	pal palette.Model
}

func New(o Options) Model {
	return Model{th: o.Theme, dir: o.Dir, lit: -1, pal: palette.New(o.Theme, o.Sources)}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m.resized(), nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if !m.pal.IsOpen() {
			return m.key(msg)
		}
	}
	var cmd tea.Cmd
	var out palette.Outcome
	m.pal, cmd, out = m.pal.Update(msg)
	cmds := []tea.Cmd{cmd}

	// What a page started in Init reports back here; each page knows its own.
	if _, key := msg.(tea.KeyMsg); !key {
		for i, p := range m.pages {
			m.pages[i], cmd = p.Update(msg)
			cmds = append(cmds, cmd)
		}
		if m.holding {
			m.held, _ = m.held.Update(msg)
		}
	}

	m, cmd = m.apply(out)
	return m, tea.Batch(append(cmds, cmd)...)
}

func (m Model) key(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+k":
		var cmd tea.Cmd
		m.pal, cmd = m.pal.Open()
		return m, cmd
	case "q":
		return m, tea.Quit
	case "[":
		return m.step(-1), nil
	case "]":
		return m.step(1), nil
	case "x":
		return m.closeTab(), nil
	}
	if len(m.pages) == 0 {
		return m, nil
	}
	var cmd tea.Cmd
	m.pages[m.active], cmd = m.pages[m.active].Update(msg)
	return m, cmd
}

func (m Model) step(by int) Model {
	if n := len(m.pages); n > 0 {
		m.active = (m.active + by + n) % n
	}
	return m
}

func (m Model) closeTab() Model {
	if len(m.pages) == 0 {
		return m
	}
	i := m.active
	m.pages = append(m.pages[:i:i], m.pages[i+1:]...)
	m.active = max(0, min(i, len(m.pages)-1))
	return m
}

func (m Model) apply(out palette.Outcome) (Model, tea.Cmd) {
	switch {
	case out.Chosen != nil:
		return m.choose(*out.Chosen)
	case out.Closed:
		return m.settle(), nil
	case out.HighlightChanged:
		it, ok := m.pal.Highlighted()
		return m.hover(it, ok), nil
	}
	return m, nil
}

// hover shows where a result is without committing to it.
func (m Model) hover(it kit.Item, ok bool) Model {
	m = m.settle()
	if !ok {
		return m
	}
	switch i := m.find(it.DocID); {
	case i < 0:
	case i == m.active:
		m.held, m.holding = m.pages[i], true
		m.pages[i] = m.pages[i].Reveal(it, m.covered())
	default:
		m.lit = i
	}
	return m
}

func (m Model) choose(it kit.Item) (Model, tea.Cmd) {
	i := m.find(it.DocID)
	if m.holding && i == m.active {
		// Keep the page where the hover left it.
		m.held, m.holding = nil, false
	}
	m = m.settle()
	var cmd tea.Cmd
	if i < 0 {
		w, h := m.pageSize()
		p := it.Open().Resize(w, h)
		cmd = p.Init()
		m.pages = append(m.pages, p)
		i = len(m.pages) - 1
	}
	m.pages[i] = m.pages[i].Reveal(it, 0)
	m.active = i
	return m, cmd
}

// settle undoes whatever a hover did.
func (m Model) settle() Model {
	if m.holding {
		m.pages[m.active] = m.held
		m.held, m.holding = nil, false
	}
	m.lit = -1
	return m
}

func (m Model) find(id string) int {
	for i, p := range m.pages {
		if p.ID() == id {
			return i
		}
	}
	return -1
}

func (m Model) resized() Model {
	w, h := m.pageSize()
	for i, p := range m.pages {
		m.pages[i] = p.Resize(w, h)
	}
	if m.holding {
		m.held = m.held.Resize(w, h)
	}
	_, _, pw, ph := m.paletteBox()
	m.pal = m.pal.Resize(pw, ph)
	return m
}
