package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"strigil/internal/kit/text"
	"strigil/internal/kit/theme"
)

// excerpt draws count lines of d from line index from, each w cells wide, on
// bg. sel is the hit to single out.
func excerpt(th theme.Theme, bg color.Color, d *doc, sel, from, count, w int) []string {
	digits := max(3, len(strconv.Itoa(len(d.lines))))
	selLine := -1
	if sel >= 0 && sel < len(d.hits) {
		selLine = d.hits[sel].line
	}
	rows := make([]string, count)
	for r := range rows {
		i := from + r
		if i < 0 || i >= len(d.lines) {
			rows[r] = theme.On(bg).Gap(w)
			continue
		}
		pen, num := theme.On(bg), th.Faint
		if i == selLine {
			pen, num = theme.On(th.Band), th.Accent
		}
		gutter := pen.Gap(1) + pen.Ink(num, fmt.Sprintf("%*d", digits, i+1)) + pen.Gap(2)
		rows[r] = pen.Line(gutter+content(th, pen, d, i, sel, w-digits-4), w)
	}
	return rows
}

func content(th theme.Theme, pen theme.Pen, d *doc, line, sel, width int) string {
	ids := d.byLine[line]
	ranges := make([][2]int, len(ids))
	var keep [2]int
	for k, id := range ids {
		ranges[k] = [2]int{d.hits[id].start, d.hits[id].end}
	}
	s, ranges := text.Expand(d.lines[line], ranges)
	for k, id := range ids {
		if id == sel {
			keep = ranges[k]
		}
	}
	a, b, left, right := text.Fit(s, keep, width)

	plain := pen.Style(th.Text)
	mark := theme.On(th.Match).Style(th.Deep)
	current := theme.On(th.Lit).Style(th.Deep).Bold(true)

	var sb strings.Builder
	if left {
		sb.WriteString(pen.Ink(th.Faint, "…"))
	}
	pos := a
	for k, id := range ids {
		from, to := max(ranges[k][0], a), min(ranges[k][1], b)
		if from >= to || from < pos {
			continue
		}
		sb.WriteString(plain.Render(s[pos:from]))
		if id == sel {
			sb.WriteString(current.Render(s[from:to]))
		} else {
			sb.WriteString(mark.Render(s[from:to]))
		}
		pos = to
	}
	if pos < b {
		sb.WriteString(plain.Render(s[pos:b]))
	}
	if right {
		sb.WriteString(pen.Ink(th.Faint, "…"))
	}
	return sb.String()
}

func preview(th theme.Theme, d *doc, sel, w, h int) string {
	pen := theme.On(th.Panel)
	rows := []string{pen.Line(pen.Ink(th.Soft, text.Tail(d.rel, w)), w), pen.Gap(w)}
	body := h - len(rows)
	if body > 0 {
		from := d.hits[sel].line - body/2
		from = max(0, min(from, len(d.lines)-body))
		rows = append(rows, excerpt(th, th.Panel, d, sel, from, body, w)...)
	}
	return strings.Join(rows, "\n")
}
