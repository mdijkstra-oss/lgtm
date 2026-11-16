// Package tabs draws the list of open pages.
package tabs

import (
	"strings"

	"strigil/internal/kit/text"
	"strigil/internal/kit/theme"
)

type Entry struct {
	Title string
	Meta  string
}

// Sidebar draws the entries as a column. lit is the entry a hovered result
// lives in, or -1.
func Sidebar(th theme.Theme, entries []Entry, active, lit, w, h int) string {
	pen := theme.On(th.Panel)
	rows := []string{
		"",
		pen.Gap(2) + pen.Style(th.Accent).Bold(true).Render("strigil"),
		"",
		pen.Gap(2) + pen.Ink(th.Dim, "open ") + pen.Ink(th.Rule, strings.Repeat("─", max(0, w-9))),
		"",
	}
	if len(entries) == 0 {
		rows = append(rows, pen.Gap(2)+pen.Ink(th.Faint, "nothing yet"))
	}
	for i, e := range entries {
		rows = append(rows, entry(th, e, i == active, i == lit, w))
	}
	return strings.Join(pen.Block(rows, w, h), "\n")
}

func entry(th theme.Theme, e Entry, active, lit bool, w int) string {
	pen := theme.On(th.Panel)
	mark := pen.Gap(2)
	title := pen.Style(th.Dim)
	meta := th.Faint
	switch {
	case lit:
		pen = theme.On(th.Band)
		mark = pen.Ink(th.Lit, "◆") + pen.Gap(1)
		title = pen.Style(th.Lit).Bold(true)
		meta = th.Lit
	case active:
		pen = theme.On(th.Band)
		mark = pen.Ink(th.Accent, "▌") + pen.Gap(1)
		title = pen.Style(th.Text).Bold(true)
		meta = th.Soft
	}
	right := pen.Ink(meta, e.Meta) + pen.Gap(2)
	room := w - 2 - text.Width(e.Meta) - 3
	return pen.Split(mark+title.Render(text.Head(e.Title, room)), right, w)
}

// Strip draws the entries on one line, for terminals too narrow for Sidebar.
func Strip(th theme.Theme, entries []Entry, active, lit, w int) string {
	pen := theme.On(th.Panel)
	if len(entries) == 0 {
		return pen.Line(pen.Gap(1)+pen.Style(th.Accent).Bold(true).Render("strigil"), w)
	}
	cells := make([]string, len(entries))
	widths := make([]int, len(entries))
	for i, e := range entries {
		label := " " + text.Head(e.Title, 24) + " "
		switch i {
		case lit:
			cells[i] = theme.On(th.Band).Style(th.Lit).Bold(true).Render(label)
		case active:
			cells[i] = theme.On(th.Band).Style(th.Text).Bold(true).Render(label)
		default:
			cells[i] = pen.Ink(th.Dim, label)
		}
		widths[i] = text.Width(label)
	}
	// Drop leading entries until the one that matters fits.
	must := active
	if lit >= 0 {
		must = max(active, lit)
	}
	first, used := 0, 0
	for i := 0; i <= must && i < len(entries); i++ {
		used += widths[i]
	}
	for first < must && used > w {
		used -= widths[first]
		first++
	}
	return pen.Line(strings.Join(cells[first:], ""), w)
}
