package viewer

import (
	"image/color"
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"strigil/internal/kit/markdown"
)

// block is one cell of a document on screen, drawn from its own state
// alone. One with things to pick has one of them picked.
type block interface {
	view(w int, on bool) []string
	keys() []key.Binding
	count() int // how many things there are to pick; 0 for text
	picked() int
	pick(i int) block
	// about is the row of view where what is said of the pick starts:
	// everything from there down changes with the pick.
	about(w int) int
	// code is what the block is showing lines of now, if anything: enter
	// gives those the whole view.
	code() *markdown.Code
}

var (
	keyFull = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "whole view"))
)

// withCode adds the key for the whole view when there are lines to give it.
func withCode(c *markdown.Code, ks ...key.Binding) []key.Binding {
	if c != nil {
		ks = append(ks, keyFull)
	}
	return ks
}

// lines draws the code a block is showing: where it is, then the lines.
func lines(c *markdown.Code) []string {
	switch {
	case c == nil:
		return nil
	case c.Line == 0:
		return append([]string{codeTitle(c, 0, 0)}, codeRows(c, 0, 0)...)
	}
	return append([]string{codeTitle(c, c.Show[0], c.Show[1])}, codeRows(c, c.Show[0], c.Show[1])...)
}

func build(c *markdown.Cell) block {
	switch {
	case c.Subsections != nil:
		p := subsections{d: c.Subsections, in: -1}
		for i := range c.Subsections.List {
			var sub []block
			for j := range c.Subsections.List[i].Cells {
				sub = append(sub, build(&c.Subsections.List[i].Cells[j]))
			}
			p.blocks = append(p.blocks, sub)
		}
		return p
	case c.Code != nil:
		return excerpt{c.Code}
	case c.Swatch != nil:
		return swatch{d: c.Swatch}
	case c.Bars != nil:
		return bars{d: c.Bars}
	}
	return prose(c.Prose)
}

// prose is markdown.
type prose string

func (p prose) view(w int, _ bool) []string { return typeset(string(p), w) }
func (p prose) keys() []key.Binding         { return nil }
func (p prose) about(int) int               { return 0 }
func (p prose) count() int                  { return 0 }
func (p prose) picked() int                 { return 0 }
func (p prose) pick(int) block              { return p }
func (p prose) code() *markdown.Code        { return nil }

// excerpt is lines of a file, standing alone.
type excerpt struct{ c *markdown.Code }

func (e excerpt) view(int, bool) []string {
	var rows []string
	for _, r := range panel(lines(e.c), 0) {
		rows = append(rows, " "+r)
	}
	return rows
}
func (e excerpt) keys() []key.Binding  { return withCode(e.c) }
func (e excerpt) about(int) int        { return 0 }
func (e excerpt) count() int           { return 1 }
func (e excerpt) picked() int          { return 0 }
func (e excerpt) pick(int) block       { return e }
func (e excerpt) code() *markdown.Code { return e.c }

// subsections is a section's subsections, one of them shown, under what they are picked from: a
// sequence, a graph, or a row of buttons. A subsection is blocks, as a section
// is, so a block in a subsection is drawn and takes keys as it would outside.
type subsections struct {
	d      *markdown.Subsections
	blocks [][]block
	at     int
	in     int // the block of the subsection the cursor is on, or -1
}

// nested is a block with blocks in it: under tells it which of its rows
// the cursor is on.
type nested interface {
	under(row, w int) block
}

// inner is the block of the subsection the cursor is on when that block has
// things of its own to pick: left and right are its.
func (p subsections) inner() (block, bool) {
	if len(p.blocks) == 0 || p.in < 0 || p.in >= len(p.blocks[p.at]) {
		return nil, false
	}
	b := p.blocks[p.at][p.in]
	return b, b.count() > 1
}

// set gives a copy with one block of the open subsection replaced: the subsections are
// shared with the model this one was copied from.
func (p subsections) set(b block) subsections {
	p.blocks = slices.Clone(p.blocks)
	p.blocks[p.at] = slices.Clone(p.blocks[p.at])
	p.blocks[p.at][p.in] = b
	return p
}

func (p subsections) under(row, w int) block {
	_, owner := p.draw(w, false)
	p.in = -1
	if row >= 0 && row < len(owner) {
		p.in = owner[row]
	}
	return p
}

func (p subsections) code() *markdown.Code {
	if len(p.blocks) == 0 {
		return nil
	}
	if b, _ := p.inner(); b != nil && b.code() != nil {
		return b.code()
	}
	for _, b := range p.blocks[p.at] {
		if c := b.code(); c != nil {
			return c
		}
	}
	return nil
}

func (p subsections) keys() []key.Binding {
	if b, _ := p.inner(); b != nil && len(b.keys()) > 0 {
		return b.keys()
	}
	return withCode(p.code())
}

func (p subsections) count() int {
	if b, own := p.inner(); own {
		return b.count()
	}
	return len(p.d.List)
}

func (p subsections) picked() int {
	if b, own := p.inner(); own {
		return b.picked()
	}
	return p.at
}

func (p subsections) pick(i int) block {
	if b, own := p.inner(); own {
		return p.set(b.pick(i))
	}
	p.at, p.in = i, -1
	return p
}

func (p subsections) about(w int) int {
	rows, owner := p.draw(w, false)
	if b, own := p.inner(); own {
		return slices.Index(owner, p.in) + b.about(w)
	}
	return len(rows) - len(owner) + slices.IndexFunc(owner, func(o int) bool { return o >= 0 })
}

func (p subsections) view(w int, on bool) []string {
	rows, _ := p.draw(w, on)
	return rows
}

// draw gives the rows, and for each row which block of the subsection it is
// of: -1 for what the subsections are picked from.
func (p subsections) draw(w int, on bool) (rows []string, owner []int) {
	_, own := p.inner()
	rows = p.from(w, on && !own)
	for range rows {
		owner = append(owner, -1)
	}
	if len(p.blocks) == 0 {
		return rows, owner
	}
	for j, b := range p.blocks[p.at] {
		rows, owner = append(rows, ""), append(owner, -1)
		// Text is set in a cell; a block has its own edge.
		wide, lead := w, ""
		if _, text := b.(prose); text {
			wide, lead = w-1, " "
		}
		for _, r := range b.view(wide, on && own && j == p.in) {
			rows, owner = append(rows, lead+r), append(owner, j)
		}
	}
	return rows, owner
}

// from draws what the subsections are picked from.
func (p subsections) from(w int, on bool) []string {
	switch {
	case len(p.d.Messages) > 0:
		return p.sequence(w, on)
	case len(p.d.Edges) > 0:
		return p.graph(on)
	}
	return p.buttons(w, on)
}

// name draws a message or a node: a button when it is a subsection, and words when
// it is not.
func (p subsections) name(s string, on bool) string {
	if i := p.d.Index(s); i >= 0 {
		return button("", s, at(i, p.at, on))
	}
	return sMeta.Render(" " + s + " ")
}

// sequence is a row a message, each in the lane it goes to.
func (p subsections) sequence(w int, on bool) []string {
	lane := w / len(p.d.Lanes)
	var head []seg
	for i, l := range p.d.Lanes {
		head = append(head, seg{i*lane + 1, ansi.Truncate(l, lane-3, "…"), sDim})
	}
	rows := []string{placed(head...), ""}
	prev := -1
	for _, h := range p.d.Messages {
		line := sDim
		if len(p.d.List) > 0 && p.d.List[p.at].Name == h.Name {
			line = sLive
		}
		x, msg := h.Lane*lane, p.name(h.Name, on)
		plain := lipgloss.NewStyle()
		switch {
		case prev >= 0 && h.Lane > prev:
			from := prev*lane + 1
			rows = append(rows, placed(seg{from, "╰" + strings.Repeat("─", x-from-3) + "▶", line}, seg{x, msg, plain}))
		case prev >= 0 && h.Lane < prev:
			to, back := prev*lane+1, x+wid(msg)+1
			rows = append(rows, placed(seg{x, msg, plain}, seg{back, "◀" + strings.Repeat("─", max(0, to-back-1)) + "╯", line}))
		default:
			rows = append(rows, placed(seg{x, msg, plain}))
		}
		prev = h.Lane
	}
	return rows
}

// graph is a row a node, each under the first node that points at it. One
// pointed at again is named again, and what it points at is not repeated.
func (p subsections) graph(on bool) []string {
	var order []string
	next, into := map[string][]markdown.Edge{}, map[string]int{}
	seen := map[string]bool{}
	for _, e := range p.d.Edges {
		for _, n := range []string{e.From, e.To} {
			if !seen[n] {
				seen[n], order = true, append(order, n)
			}
		}
		next[e.From] = append(next[e.From], e)
		into[e.To]++
	}
	var rows []string
	drawn := map[string]bool{}
	var draw func(n, lead, arrow string)
	draw = func(n, lead, arrow string) {
		row := sRule.Render(lead+arrow) + p.name(n, on)
		if drawn[n] && len(next[n]) > 0 {
			rows = append(rows, row+sDim.Render("↑"))
			return
		}
		drawn[n] = true
		rows = append(rows, row)
		if arrow != "" {
			// Under a last child nothing runs on down.
			if strings.HasPrefix(arrow, "└") {
				lead += "   "
			} else {
				lead += "│  "
			}
		}
		for i, e := range next[n] {
			a := "├─"
			if i == len(next[n])-1 {
				a = "└─"
			}
			if e.Label != "" {
				a += " " + e.Label + " ─"
			}
			draw(e.To, lead, a+"▶")
		}
	}
	for _, n := range order {
		if into[n] == 0 {
			draw(n, " ", "")
		}
	}
	// Nothing is pointed at by nothing when the arrows go round.
	if len(rows) == 0 {
		draw(order[0], " ", "")
	}
	return rows
}

// chips says which subsections go on which row of buttons: as many as fit.
func (p subsections) chips(w int) [][]int {
	rows, used := [][]int{nil}, 0
	for i, it := range p.d.List {
		n := wid(it.Name) + 4
		if it.Mark != markdown.None {
			n += 2
		}
		if used > 0 && used+1+n > w {
			rows, used = append(rows, nil), 0
		}
		if used > 0 {
			used++
		}
		rows[len(rows)-1], used = append(rows[len(rows)-1], i), used+n
	}
	return rows
}

func (p subsections) buttons(w int, on bool) []string {
	var rows []string
	for n, row := range p.chips(w) {
		if n > 0 {
			rows = append(rows, "")
		}
		var chips []string
		for _, i := range row {
			mark := ""
			if m := p.d.List[i].Mark; m != markdown.None {
				mark = tick(m == markdown.Holds)
			}
			chips = append(chips, button(mark, p.d.List[i].Name, at(i, p.at, on)))
		}
		rows = append(rows, strings.Join(chips, " "))
	}
	return rows
}

// swatch shows colours of the theme as themselves.
type swatch struct {
	d  *markdown.Swatch
	at int
}

func hues() map[string]color.Color {
	return map[string]color.Color{
		"Primary": pal.Primary, "Secondary": pal.Secondary, "Accent": pal.Accent, "OnPrimary": pal.OnPrimary,
		"Text1": pal.Text1, "Text2": pal.Text2, "Text3": pal.Text3, "Text4": pal.Text4, "Line": pal.Line,
		"Surface0": pal.Surface0, "Surface1": pal.Surface1, "Surface2": pal.Surface2, "Surface3": pal.Surface3,
		"CodeKeyword": pal.CodeKeyword, "CodeType": pal.CodeType, "CodeFunction": pal.CodeFunction,
		"CodeString": pal.CodeString, "CodeNumber": pal.CodeNumber,
		"Error": pal.Error, "Warning": pal.Warning, "Info": pal.Info, "Success": pal.Success,
	}
}

func (c swatch) all() []string {
	var out []string
	for _, g := range c.d.Groups {
		out = append(out, g.Names...)
	}
	return out
}

func (c swatch) code() *markdown.Code { return nil }
func (c swatch) keys() []key.Binding  { return nil }
func (c swatch) count() int           { return len(c.all()) }
func (c swatch) picked() int          { return c.at }
func (c swatch) pick(i int) block     { c.at = i; return c }

func (c swatch) about(int) int { return 2 * len(c.d.Groups) }

func (c swatch) view(w int, on bool) []string {
	var rows []string
	n, hues := 0, hues()
	for _, g := range c.d.Groups {
		var line strings.Builder
		line.WriteString(" " + sDim.Render(fit(g.Name, 10)))
		for _, name := range g.Names {
			line.WriteString(lipgloss.NewStyle().Foreground(hues[name]).Render("██"))
			line.WriteString(" " + button("", name, at(n, c.at, on)) + "  ")
			n++
		}
		rows = append(rows, line.String(), "")
	}
	name := c.all()[c.at]
	said := "not a colour of the theme"
	if col, ok := hues[name]; ok {
		said = *hex(col)
	}
	return append(rows, " "+sName.Render("theme."+name)+sDim.Render(" • ")+sMeta.Render(said))
}

// bars draws two numbers a row as a bar: lines out to the left of a shared
// edge, lines in to the right.
type bars struct {
	d  *markdown.Bars
	at int
}

func (b bars) code() *markdown.Code { return nil }
func (b bars) keys() []key.Binding  { return nil }

// A bar a row, and a blank row.
func (b bars) about(int) int    { return len(b.d.Rows) + 1 }
func (b bars) count() int       { return len(b.d.Rows) }
func (b bars) picked() int      { return b.at }
func (b bars) pick(i int) block { b.at = i; return b }

func (b bars) view(w int, on bool) []string {
	nameW, most := 0, 1
	for _, p := range b.d.Rows {
		nameW = max(nameW, wid(p.Name))
		most = max(most, p.Added, p.Removed)
	}
	const nums = 12
	room := w - nameW - 2 - 2 - nums
	bar := func(n int) string {
		if n == 0 {
			return ""
		}
		return strings.Repeat("▄", max(1, n*room/(most*2)))
	}
	var rows []string
	for i, p := range b.d.Rows {
		out, in := bar(p.Removed), bar(p.Added)
		mid := nameW + 2 + 2 + room/2
		count := ""
		if p.Added > 0 {
			count += "+" + strconv.Itoa(p.Added)
		}
		if p.Removed > 0 {
			count += " -" + strconv.Itoa(p.Removed)
		}
		count = strings.TrimSpace(count)
		rows = append(rows, placed(
			seg{0, button("", fit(p.Name, nameW), at(i, b.at, on)), lipgloss.NewStyle()},
			seg{mid - wid(out), out, sBad},
			seg{mid, in, sOK},
			seg{w - wid(count), count, sDim},
		))
	}
	p := b.d.Rows[b.at]
	last := " " + sName.Render(p.Name)
	if p.Note != "" {
		last += sDim.Render(" • ") + sMeta.Render(p.Note)
	}
	return append(rows, "", last)
}
