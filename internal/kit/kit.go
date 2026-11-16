// Package kit holds the contracts a feature implements to appear in the TUI.
package kit

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

// Source feeds the palette. Search blocks until done or ctx is cancelled.
type Source interface {
	Name() string
	Search(ctx context.Context, query string, emit func(Item)) error
}

type Item struct {
	Title    string
	Marks    [][2]int // byte ranges of Title to emphasise
	Location string
	Source   string // set by the palette from Source.Name
	DocID    string // ID of the Page this result lives in
	Data     any    // private to the feature that made the item
	// Preview is drawn on theme.Panel; it must not do I/O.
	Preview func(w, h int) string
	Open    func() Page
}

type Hint struct{ Key, Label string }

// Page is the content of a tab. Implementations are values: every method
// returns the changed page, so the shell can keep an earlier one to go back to.
type Page interface {
	ID() string
	Title() string
	Meta() string
	// Init starts whatever the page needs to load. Its messages reach every
	// open page, so they have to say which page they are for.
	Init() tea.Cmd
	Hints() []Hint
	Resize(w, h int) Page
	// Reveal brings the item into view and marks it. covered is how many of
	// the page's top rows an overlay hides; 0 when nothing does.
	Reveal(it Item, covered int) Page
	Update(tea.Msg) (Page, tea.Cmd)
	View() string
}
