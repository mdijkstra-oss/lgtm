package ask

import (
	"testing"

	"strigil/internal/kit/markdown"
)

func TestEveryDocParses(t *testing.T) {
	for _, d := range Docs() {
		if d.Tab == "" || d.At == "" || len(d.Intro) == 0 {
			t.Errorf("%q: needs a tab, a commit and an intro", d.Title)
		}
	}
	if len(Docs()) != len(order) {
		t.Errorf("%d documents read, of %d", len(Docs()), len(order))
	}
}

// A subsection under a drawing that the drawing does not name can be
// stepped to, and lights nothing.
func TestEverySubsectionUnderADrawingIsInIt(t *testing.T) {
	for _, d := range Docs() {
		all := append([]markdown.Section{{Title: d.Title, Cells: d.Intro}}, d.Sections...)
		for _, s := range all {
			for _, c := range s.Cells {
				p := c.Subsections
				if p == nil || len(p.Messages)+len(p.Edges) == 0 {
					continue
				}
				drawn := map[string]bool{}
				for _, m := range p.Messages {
					drawn[m.Name] = true
				}
				for _, e := range p.Edges {
					drawn[e.From], drawn[e.To] = true, true
				}
				for _, sub := range p.List {
					if !drawn[sub.Name] {
						t.Errorf("%q, %q: the drawing has no %q", d.Title, s.Title, sub.Name)
					}
				}
			}
		}
	}
}
