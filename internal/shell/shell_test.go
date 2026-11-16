package shell

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"strigil/internal/kit"
	"strigil/internal/kit/theme"
)

type page struct {
	id       string
	revealed string
	ready    bool
}

type loaded string

func (p page) ID() string               { return p.id }
func (p page) Title() string            { return p.id }
func (p page) Meta() string             { return "" }
func (p page) Init() tea.Cmd            { return func() tea.Msg { return loaded(p.id) } }
func (p page) Hints() []kit.Hint        { return nil }
func (p page) Resize(int, int) kit.Page { return p }
func (p page) Update(msg tea.Msg) (kit.Page, tea.Cmd) {
	if id, ok := msg.(loaded); ok && string(id) == p.id {
		p.ready = true
	}
	return p, nil
}
func (p page) View() string { return "" }
func (p page) Reveal(it kit.Item, _ int) kit.Page {
	p.revealed = it.Title
	return p
}

func result(doc, title string) kit.Item {
	return kit.Item{Title: title, DocID: doc, Open: func() kit.Page { return page{id: doc} }}
}

func withTabs(ids ...string) Model {
	m := New(Options{Theme: theme.Mocha()})
	m.w, m.h = 120, 40
	for _, id := range ids {
		m.pages = append(m.pages, page{id: id})
	}
	return m.resized()
}

func revealed(m Model, i int) string { return m.pages[i].(page).revealed }

func TestHoverOnTheCurrentPageMovesItAndClosingPutsItBack(t *testing.T) {
	m := withTabs("a", "b")

	m = m.hover(result("a", "first"), true)
	if revealed(m, 0) != "first" || m.lit != -1 {
		t.Fatalf("hover: revealed %q, lit %d", revealed(m, 0), m.lit)
	}

	m = m.hover(result("a", "second"), true)
	if revealed(m, 0) != "second" {
		t.Fatalf("second hover: revealed %q", revealed(m, 0))
	}

	m = m.settle()
	if revealed(m, 0) != "" {
		t.Errorf("closing left the page at %q", revealed(m, 0))
	}
}

func TestHoverOnAnotherTabLightsItWithoutSwitching(t *testing.T) {
	m := withTabs("a", "b")

	m = m.hover(result("b", "x"), true)
	if m.lit != 1 || m.active != 0 || revealed(m, 1) != "" {
		t.Fatalf("lit %d, active %d, revealed %q", m.lit, m.active, revealed(m, 1))
	}

	m = m.hover(result("c", "y"), true)
	if m.lit != -1 {
		t.Errorf("a result that is not open lit tab %d", m.lit)
	}
}

func TestHoverLeavingTheCurrentPagePutsItBack(t *testing.T) {
	m := withTabs("a", "b")
	m = m.hover(result("a", "first"), true)
	m = m.hover(result("b", "x"), true)
	if revealed(m, 0) != "" || m.lit != 1 {
		t.Errorf("revealed %q, lit %d", revealed(m, 0), m.lit)
	}
}

func TestChoosing(t *testing.T) {
	m := withTabs("a", "b")

	m = m.hover(result("b", "x"), true)
	m, _ = m.choose(result("b", "x"))
	if m.active != 1 || m.lit != -1 || revealed(m, 1) != "x" {
		t.Fatalf("existing tab: active %d, lit %d, revealed %q", m.active, m.lit, revealed(m, 1))
	}

	m, cmd := m.choose(result("c", "y"))
	if cmd == nil {
		t.Fatal("a new page was not started")
	}
	if len(m.pages) != 3 || m.active != 2 || revealed(m, 2) != "y" {
		t.Fatalf("new tab: %d pages, active %d", len(m.pages), m.active)
	}

	m = m.hover(result("c", "z"), true)
	m, _ = m.choose(result("c", "z"))
	if revealed(m, 2) != "z" {
		t.Errorf("choosing on the current page: revealed %q", revealed(m, 2))
	}
	m = m.settle()
	if revealed(m, 2) != "z" {
		t.Errorf("a chosen result was undone: %q", revealed(m, 2))
	}
}

func TestWhatAPageLoadsReachesItEvenWhileHovered(t *testing.T) {
	m := withTabs("a", "b")
	m = m.hover(result("a", "first"), true)

	tm, _ := m.Update(loaded("a"))
	m = tm.(Model)
	if !m.pages[0].(page).ready {
		t.Fatal("the page did not get its message")
	}
	m = m.settle()
	if !m.pages[0].(page).ready {
		t.Error("putting the page back lost what it had loaded")
	}
	if m.pages[1].(page).ready {
		t.Error("another page took the message")
	}
}

func TestXClosesTheCurrentTab(t *testing.T) {
	var tm tea.Model = withTabs("a", "b", "c")
	press := func(r rune) Model {
		tm, _ = tm.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		return tm.(Model)
	}
	press(']')
	if m := press('x'); len(m.pages) != 2 || m.pages[m.active].ID() != "c" {
		t.Fatalf("closing the middle tab: %d pages, on %q", len(m.pages), m.pages[m.active].ID())
	}
	if m := press('x'); len(m.pages) != 1 || m.pages[m.active].ID() != "a" {
		t.Fatalf("closing the last tab: %d pages, on %q", len(m.pages), m.pages[m.active].ID())
	}
	press('x')
	if m := press('x'); len(m.pages) != 0 || m.active != 0 {
		t.Errorf("closing with nothing open: %d pages, active %d", len(m.pages), m.active)
	}
}

func TestBracketsCycleTabs(t *testing.T) {
	var tm tea.Model = withTabs("a", "b", "c")
	press := func(r rune) int {
		tm, _ = tm.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		return tm.(Model).active
	}
	if got := press(']'); got != 1 {
		t.Errorf("] went to %d", got)
	}
	if got := press('['); got != 0 {
		t.Errorf("[ went to %d", got)
	}
	if got := press('['); got != 2 {
		t.Errorf("[ from the first tab went to %d", got)
	}
}
