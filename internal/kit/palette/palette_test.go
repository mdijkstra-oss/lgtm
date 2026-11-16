package palette

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"strigil/internal/kit"
	"strigil/internal/kit/theme"
)

func opened(t *testing.T) Model {
	t.Helper()
	m, _ := New(theme.Mocha(), nil).Open()
	return m.Resize(80, 20)
}

func typed(m Model, s string) Model {
	for _, r := range s {
		m, _, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func items(titles ...string) []kit.Item {
	out := make([]kit.Item, len(titles))
	for i, title := range titles {
		out[i] = kit.Item{Title: title}
	}
	return out
}

func TestResultsOfAnOlderQueryAreDropped(t *testing.T) {
	m := typed(opened(t), "a")
	old := m.gen
	m = typed(m, "b")

	m, _, out := m.Update(itemsMsg{gen: old, items: items("late"), done: true})

	if len(m.items) != 0 || out.HighlightChanged {
		t.Errorf("late results were shown: %v", m.items)
	}
	if !m.searching {
		t.Error("the newer search should still be running")
	}
}

func TestEarlierResultsStayUntilTheNewOnesArrive(t *testing.T) {
	m := typed(opened(t), "a")
	m, _, _ = m.Update(itemsMsg{gen: m.gen, items: items("one", "two", "three"), done: true})
	m, _, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	m = typed(m, "b")
	if got, _ := m.Highlighted(); got.Title != "three" {
		t.Fatalf("highlight moved before new results arrived: %q", got.Title)
	}

	m, _, out := m.Update(itemsMsg{gen: m.gen, items: items("fresh"), done: true})
	got, ok := m.Highlighted()
	if !ok || got.Title != "fresh" || len(m.items) != 1 || !out.HighlightChanged {
		t.Errorf("after new results: %v, highlighted %q, changed %v", m.items, got.Title, out.HighlightChanged)
	}
}

func TestClearingTheQueryClearsTheList(t *testing.T) {
	m := typed(opened(t), "a")
	m, _, _ = m.Update(itemsMsg{gen: m.gen, items: items("one"), done: true})

	m, _, out := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	if _, ok := m.Highlighted(); ok || !out.HighlightChanged || m.searching {
		t.Errorf("list not cleared: %v", m.items)
	}
}

func TestEnterChoosesAndEscCloses(t *testing.T) {
	m := typed(opened(t), "a")
	m, _, _ = m.Update(itemsMsg{gen: m.gen, items: items("one", "two"), done: true})
	m, _, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	chosen, _, out := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if out.Chosen == nil || out.Chosen.Title != "two" || chosen.IsOpen() {
		t.Errorf("enter: %+v", out)
	}

	closed, _, out := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if !out.Closed || closed.IsOpen() {
		t.Errorf("esc: %+v", out)
	}
}
