// Package frame draws a bordered panel around rows it is given.
package frame

import (
	"strings"

	"charm.land/lipgloss/v2"

	"strigil/internal/kit/theme"
)

// Box is a panel. Where it goes and how big it is are the caller's business.
type Box struct {
	Head string // a row above a rule; left out when empty
	Body []string
	Foot string // a row below a rule; left out when empty
}

// BodySize is the room Body gets in a box of w by h cells.
func BodySize(w, h int, head, foot bool) (bw, bh int) {
	bh = h - 2
	if head {
		bh -= 2
	}
	if foot {
		bh -= 2
	}
	return max(0, w-2), max(0, bh)
}

func (b Box) Render(th theme.Theme, w, h int) string {
	pen := theme.On(th.Panel)
	bw, bh := BodySize(w, h, b.Head != "", b.Foot != "")
	rule := pen.Ink(th.Rule, strings.Repeat("─", bw))

	var rows []string
	if b.Head != "" {
		rows = append(rows, pen.Line(b.Head, bw), rule)
	}
	rows = append(rows, pen.Block(b.Body, bw, bh)...)
	if b.Foot != "" {
		rows = append(rows, rule, pen.Line(b.Foot, bw))
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(th.Accent).
		BorderBackground(th.Panel).
		Render(strings.Join(rows, "\n"))
}
