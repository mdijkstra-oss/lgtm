package palette

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"strigil/internal/kit"
	"strigil/internal/kit/frame"
	"strigil/internal/kit/hints"
	"strigil/internal/kit/text"
	"strigil/internal/kit/theme"
)

const (
	rowsPerItem  = 2
	previewAbove = 76
)

func (m Model) bodyHeight() int {
	_, bh := frame.BodySize(m.w, m.h, true, true)
	return max(rowsPerItem, bh)
}

func (m Model) visible() int { return max(1, m.bodyHeight()/rowsPerItem) }

func (m Model) View() string {
	pen := theme.On(m.th.Panel)
	bw, _ := frame.BodySize(m.w, m.h, true, true)
	return frame.Box{
		Head: pen.Gap(1) + restyle(m.input.View(), pen.Style(m.th.Text)),
		Body: m.body(bw, m.bodyHeight()),
		Foot: m.status(bw),
	}.Render(m.th, m.w, m.h)
}

// restyle gives st to the text that textinput leaves bare: with a real
// cursor it draws the cell under it without any style.
func restyle(view string, st lipgloss.Style) string {
	const reset = "\x1b[m"
	var b strings.Builder
	for {
		i := strings.Index(view, reset)
		if i < 0 {
			b.WriteString(view)
			return b.String()
		}
		b.WriteString(view[:i+len(reset)])
		view = view[i+len(reset):]
		j := strings.Index(view, "\x1b[")
		if j < 0 {
			j = len(view)
		}
		if j > 0 {
			b.WriteString(st.Render(view[:j]))
		}
		view = view[j:]
	}
}

func (m Model) body(w, h int) []string {
	pen := theme.On(m.th.Panel)
	if len(m.items) == 0 {
		return pen.Middle(m.message(w), w, h)
	}
	lw := w
	if w >= previewAbove {
		lw = max(34, w*2/5)
	}
	left := m.list(lw, h)
	if lw == w {
		return left
	}

	pw := w - lw - 1
	var right []string
	if it := m.items[m.cursor]; it.Preview != nil {
		right = strings.Split(it.Preview(pw-2, h), "\n")
	}
	right = pen.Block(right, pw-2, h)
	sep := pen.Ink(m.th.Rule, "│")
	rows := make([]string, h)
	for i := range rows {
		rows[i] = left[i] + sep + pen.Gap(1) + right[i] + pen.Gap(1)
	}
	return rows
}

func (m Model) message(w int) []string {
	pen := theme.On(m.th.Panel)
	q := strings.TrimSpace(m.input.Value())
	switch {
	case q == "":
		names := make([]string, len(m.sources))
		for i, s := range m.sources {
			names[i] = s.Name()
		}
		return []string{
			pen.Ink(m.th.Soft, "Type to search"),
			"",
			pen.Ink(m.th.Faint, text.Head(strings.Join(names, " · "), w-4)),
		}
	case m.searching:
		return []string{pen.Ink(m.th.Dim, "Searching…")}
	default:
		return []string{
			pen.Ink(m.th.Soft, "Nothing matches"),
			"",
			pen.Ink(m.th.Match, text.Head("“"+q+"”", w-4)),
		}
	}
}

func (m Model) list(w, h int) []string {
	rows := make([]string, 0, h)
	for i := m.top; i < len(m.items) && i < m.top+m.visible(); i++ {
		it := m.items[i]
		pen := theme.On(m.th.Panel)
		mark := pen.Gap(2)
		if i == m.cursor {
			pen = theme.On(m.th.Band)
			mark = pen.Ink(m.th.Accent, "▌") + pen.Gap(1)
		}
		rows = append(rows, pen.Line(mark+m.title(pen, i == m.cursor, it, w-3), w))

		source := pen.Ink(m.th.Faint, it.Source) + pen.Gap(1)
		room := w - 2 - text.Width(it.Source) - 3
		rows = append(rows, pen.Split(pen.Gap(2)+pen.Ink(m.th.Dim, text.Tail(it.Location, room)), source, w))
	}
	return theme.On(m.th.Panel).Block(rows, w, h)
}

func (m Model) title(pen theme.Pen, selected bool, it kit.Item, width int) string {
	plain := pen.Style(m.th.Soft)
	if selected {
		plain = pen.Style(m.th.Text)
	}
	hit := pen.Style(m.th.Match).Bold(true)

	var keep [2]int
	if len(it.Marks) > 0 {
		keep = it.Marks[0]
	}
	a, b, left, right := text.Fit(it.Title, keep, width)

	var sb strings.Builder
	if left {
		sb.WriteString(pen.Ink(m.th.Faint, "…"))
	}
	pos := a
	for _, mk := range it.Marks {
		s, e := max(mk[0], a), min(mk[1], b)
		if s >= e || s < pos {
			continue
		}
		sb.WriteString(plain.Render(it.Title[pos:s]))
		sb.WriteString(hit.Render(it.Title[s:e]))
		pos = e
	}
	if pos < b {
		sb.WriteString(plain.Render(it.Title[pos:b]))
	}
	if right {
		sb.WriteString(pen.Ink(m.th.Faint, "…"))
	}
	return sb.String()
}

func (m Model) status(w int) string {
	pen := theme.On(m.th.Panel)
	left := pen.Gap(1)
	switch {
	case m.searching:
		left += m.spin.View() + pen.Ink(m.th.Dim, " searching")
	case len(m.items) > 0:
		total := fmt.Sprint(len(m.items))
		if m.capped {
			total += "+"
		}
		left += pen.Ink(m.th.Soft, fmt.Sprint(m.cursor+1)) + pen.Ink(m.th.Dim, " of "+total)
	}
	keys := []kit.Hint{{Key: "↑↓", Label: "select"}, {Key: "enter", Label: "open"}, {Key: "esc", Label: "close"}}
	return pen.Split(left, hints.Render(m.th, pen, keys, "   ")+pen.Gap(1), w)
}
