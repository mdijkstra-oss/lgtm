// Package palette is the ctrl+k overlay: type, watch results arrive, pick one.
package palette

import (
	"context"
	"sync"
	"time"

	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"strigil/internal/kit"
	"strigil/internal/kit/theme"
)

const (
	debounce    = 120 * time.Millisecond
	batchWindow = 25 * time.Millisecond
	maxItems    = 500
)

// Outcome is what one Update changed that the owner has to react to.
type Outcome struct {
	HighlightChanged bool
	Chosen           *kit.Item
	Closed           bool
}

type Model struct {
	th      theme.Theme
	sources []kit.Source
	input   textinput.Model
	spin    spinner.Model

	open      bool
	gen       int // bumped on every edit; results carrying an older one are dropped
	cancel    context.CancelFunc
	items     []kit.Item
	stale     bool // items belong to an earlier query still on screen
	searching bool
	capped    bool
	cursor    int
	top       int
	w, h      int
}

type debounceMsg struct{ gen int }

type itemsMsg struct {
	gen   int
	items []kit.Item
	done  bool
	ch    <-chan kit.Item
}

func New(th theme.Theme, sources []kit.Source) Model {
	pen := theme.On(th.Panel)

	in := textinput.New()
	in.Prompt = "❯ "
	in.Placeholder = "Search"
	st := textinput.DefaultDarkStyles()
	st.Focused.Prompt = pen.Style(th.Accent).Bold(true)
	st.Focused.Text = pen.Style(th.Text)
	st.Focused.Placeholder = pen.Style(th.Faint)
	st.Blurred = st.Focused
	st.Cursor.Color = th.Accent
	st.Cursor.Shape = tea.CursorBar
	in.SetStyles(st)
	in.SetVirtualCursor(false)

	return Model{
		th:      th,
		sources: sources,
		input:   in,
		spin:    spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(pen.Style(th.Accent))),
	}
}

func (m Model) IsOpen() bool { return m.open }

func (m Model) Open() (Model, tea.Cmd) {
	m = m.stop()
	m.open = true
	m.input.Reset()
	m.items, m.cursor, m.top = nil, 0, 0
	m.stale, m.searching, m.capped = false, false, false
	return m, m.input.Focus()
}

func (m Model) Resize(w, h int) Model {
	m.w, m.h = w, h
	m.input.SetWidth(max(1, w-8))
	return m.follow()
}

// Cursor is where the text cursor goes, relative to the palette's top-left
// corner.
func (m Model) Cursor() *tea.Cursor {
	if !m.open {
		return nil
	}
	c := m.input.Cursor()
	if c == nil {
		return nil
	}
	c.X += 2
	c.Y++
	return c
}

// Highlighted is the item under the cursor.
func (m Model) Highlighted() (kit.Item, bool) {
	if !m.open || len(m.items) == 0 {
		return kit.Item{}, false
	}
	return m.items[m.cursor], true
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd, Outcome) {
	if !m.open {
		return m, nil, Outcome{}
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m.close(), nil, Outcome{Closed: true}
		case "enter":
			it, ok := m.Highlighted()
			if !ok {
				return m, nil, Outcome{}
			}
			return m.close(), nil, Outcome{Chosen: &it}
		case "up":
			return m.move(-1)
		case "down":
			return m.move(1)
		}
	case debounceMsg:
		if msg.gen != m.gen {
			return m, nil, Outcome{}
		}
		return m.start()
	case itemsMsg:
		return m.receive(msg)
	case spinner.TickMsg:
		if !m.searching {
			return m, nil, Outcome{}
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd, Outcome{}
	}

	before := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if q := m.input.Value(); q != before {
		return m.edited(q, cmd)
	}
	return m, cmd, Outcome{}
}

func (m Model) edited(q string, cmd tea.Cmd) (Model, tea.Cmd, Outcome) {
	m = m.stop()
	m.capped = false
	if strings.TrimSpace(q) == "" {
		had := len(m.items) > 0
		m.items, m.cursor, m.top = nil, 0, 0
		m.stale, m.searching = false, false
		return m, cmd, Outcome{HighlightChanged: had}
	}
	cmds := []tea.Cmd{cmd}
	if !m.searching {
		cmds = append(cmds, m.spin.Tick)
	}
	m.searching = true
	m.stale = len(m.items) > 0
	gen := m.gen
	cmds = append(cmds, tea.Tick(debounce, func(time.Time) tea.Msg { return debounceMsg{gen: gen} }))
	return m, tea.Batch(cmds...), Outcome{}
}

func (m Model) start() (Model, tea.Cmd, Outcome) {
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	ch := make(chan kit.Item, 256)
	query := m.input.Value()
	sources := m.sources
	go func() {
		var wg sync.WaitGroup
		for _, s := range sources {
			wg.Go(func() {
				name := s.Name()
				_ = s.Search(ctx, query, func(it kit.Item) {
					it.Source = name
					select {
					case ch <- it:
					case <-ctx.Done():
					}
				})
			})
		}
		wg.Wait()
		close(ch)
	}()
	return m, wait(m.gen, ch), Outcome{}
}

// wait gathers what arrives within a short window so a fast source does not
// cause a redraw per item.
func wait(gen int, ch <-chan kit.Item) tea.Cmd {
	return func() tea.Msg {
		first, ok := <-ch
		if !ok {
			return itemsMsg{gen: gen, done: true}
		}
		batch := []kit.Item{first}
		timer := time.NewTimer(batchWindow)
		defer timer.Stop()
		for len(batch) < maxItems {
			select {
			case it, ok := <-ch:
				if !ok {
					return itemsMsg{gen: gen, items: batch, done: true}
				}
				batch = append(batch, it)
			case <-timer.C:
				return itemsMsg{gen: gen, items: batch, ch: ch}
			}
		}
		return itemsMsg{gen: gen, items: batch, ch: ch}
	}
}

func (m Model) receive(msg itemsMsg) (Model, tea.Cmd, Outcome) {
	if msg.gen != m.gen {
		return m, nil, Outcome{}
	}
	var out Outcome
	// The earlier query's results stay up until the new one has something to
	// show, so the list does not blank on every keystroke.
	if m.stale {
		m.items, m.cursor, m.top = nil, 0, 0
		m.stale = false
		out.HighlightChanged = true
	}
	if len(msg.items) > 0 {
		if len(m.items) == 0 {
			out.HighlightChanged = true
		}
		m.items = append(m.items[:len(m.items):len(m.items)], msg.items...)
	}
	if len(m.items) >= maxItems {
		m.items = m.items[:maxItems]
		m.capped = true
		m.searching = false
		if m.cancel != nil {
			m.cancel()
			m.cancel = nil
		}
		return m, nil, out
	}
	if msg.done {
		m.searching = false
		return m, nil, out
	}
	return m, wait(msg.gen, msg.ch), out
}

func (m Model) move(by int) (Model, tea.Cmd, Outcome) {
	if len(m.items) == 0 {
		return m, nil, Outcome{}
	}
	to := min(max(m.cursor+by, 0), len(m.items)-1)
	if to == m.cursor {
		return m, nil, Outcome{}
	}
	m.cursor = to
	return m.follow(), nil, Outcome{HighlightChanged: true}
}

func (m Model) follow() Model {
	vis := m.visible()
	if m.cursor < m.top {
		m.top = m.cursor
	}
	if m.cursor >= m.top+vis {
		m.top = m.cursor - vis + 1
	}
	return m
}

func (m Model) stop() Model {
	m.gen++
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	return m
}

func (m Model) close() Model {
	m = m.stop()
	m.open = false
	m.searching = false
	m.input.Blur()
	return m
}
