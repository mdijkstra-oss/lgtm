// Package viewer is a document on screen: its text down one column, its
// blocks answering the keys while the cursor is on them, and an index of
// its sections beside it. It knows nothing of where a document came from.
package viewer

import (
	"slices"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"strigil/internal/kit/markdown"
)

type tab struct {
	d     markdown.Doc
	intro []block
	secs  []sect
	at    int // the rest the cursor's line is, counting down the page
	top   int // the first row of the document on screen

	// The page as last laid out, by settle. A swipe sends keys by the
	// dozen, and laying the page out is the slow part of answering one, so
	// it is laid out again only when stale: bare is the page with the
	// cursor in no block, which moving the cursor does not change.
	rows, bare []string
	rests      []rest
	stale      bool

	// Code given the whole view, and its first line on screen.
	whole    *markdown.Code
	wholeTop int
	// Where the cursor was when the page last settled.
	was, wasRow int
}

type sect struct {
	d      *markdown.Section
	blocks []block
}

// stop names a block: one of the intro (sec -1) or of a section.
type stop struct{ sec, blk int }

// rest is a row the cursor can be on: a line of text, or a line of a block.
type rest struct {
	row int
	sec int // the section it is in; -1 in the intro
	blk int // the block it is in; -1 for a line of text

	// For a line of a block: the block's first and last rows, and the row
	// where what it says of its pick starts.
	start, end, about int

	// For a line of the title or of a heading: what it says.
	head string
}

// in says the rest is a line of that block.
func (r rest) in(s stop) bool { return r.blk >= 0 && r.sec == s.sec && r.blk == s.blk }

// clone copies what a key may change, so an earlier model stays as it was.
func (t tab) clone() tab {
	t.intro = slices.Clone(t.intro)
	t.secs = slices.Clone(t.secs)
	for i := range t.secs {
		t.secs[i].blocks = slices.Clone(t.secs[i].blocks)
	}
	return t
}

func (t *tab) block(s stop) *block {
	if s.sec < 0 {
		return &t.intro[s.blk]
	}
	return &t.secs[s.sec].blocks[s.blk]
}

// Model is one document being read, in the room it was last given.
type Model struct {
	t    tab
	w, h int
}

func New(d markdown.Doc) Model {
	t := tab{d: d}
	for i := range d.Intro {
		t.intro = append(t.intro, build(&d.Intro[i]))
	}
	for i := range d.Sections {
		s := sect{d: &d.Sections[i]}
		for j := range s.d.Cells {
			s.blocks = append(s.blocks, build(&s.d.Cells[j]))
		}
		t.secs = append(t.secs, s)
	}
	return Model{t: t}
}

func (m Model) Doc() markdown.Doc { return m.t.d }

// Size gives the viewer the window's width and the rows it may fill.
func (m Model) Size(w, h int) Model {
	m.w, m.h, m.t.stale = w, h, true
	return m.settle()
}

var (
	keyMove = key.NewBinding(key.WithKeys("up", "down", "k", "j"), key.WithHelp("↑/↓", "move"))
	keyRoll = key.NewBinding(key.WithKeys("up", "down", "k", "j"), key.WithHelp("↑/↓", "scroll"))
	keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))
	keyStep = key.NewBinding(key.WithKeys("left", "right", "h", "l"), key.WithHelp("←/→", "step"))
)

// keyJump is the binding for the index: a section's number goes to it.
func keyJump(n int) key.Binding {
	return key.NewBinding(key.WithKeys("1"), key.WithHelp("1-"+strconv.Itoa(min(n+1, 9)), "section"))
}

// Key is a key pressed while the document is on screen.
func (m Model) Key(msg tea.KeyPressMsg) Model {
	m.t = m.t.clone()
	return m.read(msg).settle()
}

// read is a key.
func (m Model) read(msg tea.KeyPressMsg) Model {
	if m.w == 0 || m.h == 0 {
		return m
	}
	t := &m.t
	if t.whole != nil {
		switch msg.String() {
		case "esc", "enter":
			t.whole = nil
		case "up", "k":
			t.wholeTop--
		case "down", "j":
			t.wholeTop++
		}
		return m
	}
	rests := t.rests
	if n := msg.String(); len(n) == 1 && n[0] >= '1' && n[0] <= '9' && int(n[0]-'1') <= len(t.secs) {
		// 1 is the title, and the sections follow. A section's first
		// rest is its heading.
		for i, r := range rests {
			// As ints: a byte cannot go below nought, and the title is -1.
			if r.sec == int(n[0])-'2' {
				t.at = i
				break
			}
		}
		return m
	}
	here := rests[t.at]
	switch msg.String() {
	case "up", "k":
		t.at = max(0, t.at-1)
	case "down", "j":
		t.at = min(len(rests)-1, t.at+1)
	case "left", "h", "right", "l":
		if here.blk < 0 {
			break
		}
		in := stop{here.sec, here.blk}
		b := t.block(in)
		// Past either end it comes round to the other.
		n, to := (*b).count(), (*b).picked()+1
		if k := msg.String(); k == "left" || k == "h" {
			to -= 2
		}
		*b = (*b).pick((to + n) % n)
		t.stale = true
		// Everything said of the pick is new. A cursor that was in it goes
		// to its first line, and the page shows the block from its top, so
		// what was picked and what to pick next are both in view.
		if here.row < here.start+1+(*b).about(m.inside(here.sec)) {
			break
		}
		_, rests = t.lay(m.colW())
		for i, r := range rests {
			if r.in(in) && r.row >= r.about {
				t.at, t.top = i, r.start-2
				break
			}
		}
	default:
		if here.blk < 0 {
			break
		}
		b := t.block(stop{here.sec, here.blk})
		if c := (*b).code(); c != nil && msg.String() == "enter" {
			// The line that matters starts a third of the way down.
			t.whole, t.wholeTop = c, c.Hot[0]-c.Line-m.h/3
		}
	}
	return m
}

const (
	edge   = 4 // the least room left of the text; the cell two before the text is the cursor's
	widest = 88
	indexW = 30
	gap    = 4
)

// Left is where text starts in a window that wide: the document, and the
// index if there is room for it, sit in the middle.
func Left(w int) int {
	m := Model{w: w}
	all := m.colW()
	if m.index() {
		all += gap + indexW
	}
	return max(edge, (w-all)/2)
}

// index says whether there is room for the index beside the document.
func (m Model) index() bool { return m.w-2*edge >= 64+gap+indexW }

func (m Model) colW() int {
	if m.index() {
		return min(widest, m.w-2*edge-gap-indexW)
	}
	return min(widest, m.w-2*edge)
}

// inside is how wide a block is within its outline: a section's blocks
// are set in from the intro's.
func (m Model) inside(sec int) int {
	if sec < 0 {
		return m.colW() - 2
	}
	return m.colW() - 4
}

// lay gives the document's rows w cells wide, and the rows the cursor can
// be on: each row, of text or of a block, that has words on it.
func (t tab) lay(w int) (bare []string, rests []rest) {
	var rows []string
	text := func(sec int, indent string, lines []string, head bool) {
		for _, l := range lines {
			if worded(l) {
				r := rest{row: len(rows), sec: sec, blk: -1}
				if head {
					// Kept from its first cell, so lighting it moves no letter.
					r.head = strings.TrimRight(ansi.Strip(l), " ")
				}
				rests = append(rests, r)
			}
			rows = append(rows, indent+" "+ansi.Truncate(l, w-wid(indent)-1, "…"))
		}
		rows = append(rows, "")
	}
	cells := func(sec int, indent string, blocks []block) {
		for i, b := range blocks {
			inner := w - wid(indent) - 2
			if b.count() == 0 {
				text(sec, indent, b.view(w-wid(indent)-1, false), false)
				continue
			}
			lines := box(w, indent, b, false)
			about := len(rows) + 1 + b.about(inner)
			for n, l := range lines[1 : len(lines)-1] {
				if worded(l) {
					rests = append(rests, rest{
						row: len(rows) + 1 + n, sec: sec, blk: i,
						start: len(rows), end: len(rows) + len(lines), about: about,
					})
				}
			}
			rows = append(append(rows, lines...), "")
		}
	}
	text(-1, "", typeset("# "+t.d.Title, w-1), true)
	cells(-1, "", t.intro)
	for i, s := range t.secs {
		text(i, "", typeset("## "+s.d.Title, w-1), true)
		cells(i, "  ", s.blocks)
	}
	return rows[:len(rows)-1], rests
}

// box draws a block in its outline. The outline is always there, so the
// cursor coming and going changes its colour and moves nothing.
func box(w int, indent string, b block, on bool) []string {
	edge, inner := sBlock, w-wid(indent)-2
	if on {
		edge = sHere
	}
	out := []string{indent + edge.Render("╭"+strings.Repeat("─", inner)+"╮")}
	for _, l := range b.view(inner, on) {
		out = append(out, indent+edge.Render("│")+fit(l, inner)+edge.Render("│"))
	}
	return append(out, indent+edge.Render("╰"+strings.Repeat("─", inner)+"╯"))
}

// lit is the bare page with the block the cursor is in drawn as that: the
// same height, its outline and its pick in the cursor's colour.
func (t tab) lit(w int) []string {
	at := t.rests[t.at]
	if at.head != "" {
		// A heading under the cursor is lit across the page, so going to
		// one by its number shows at once where that was.
		rows := slices.Clone(t.bare)
		rows[at.row] = band(" "+at.head, w+2)
		return rows
	}
	if at.blk < 0 {
		return t.bare
	}
	indent := "  "
	if at.sec < 0 {
		indent = ""
	}
	rows := slices.Clone(t.bare)
	copy(rows[at.start:], box(w, indent, *t.block(stop{at.sec, at.blk}), true))
	return rows
}

// worded says a row has a letter or a digit on it: something to read, and
// not blank or only the lines of a drawing.
func worded(row string) bool {
	return strings.ContainsFunc(ansi.Strip(row), func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) })
}

// settle scrolls the page to where the cursor is.
func (m Model) settle() Model {
	if m.w == 0 || m.h == 0 {
		return m
	}
	t := &m.t
	if t.whole != nil {
		t.wholeTop = max(0, min(t.wholeTop, len(t.whole.Lines)-(m.h-2)))
	}
	fresh := t.stale || t.bare == nil
	if fresh {
		t.bare, t.rests = t.lay(m.colW())
		t.stale = false
	}
	t.at = min(t.at, len(t.rests)-1)
	if r := t.rests[t.at]; r.blk >= 0 {
		b := t.block(stop{r.sec, r.blk})
		if n, ok := (*b).(nested); ok {
			*b = n.under(r.row-r.start-1, m.inside(r.sec))
		}
	}
	t.rows = t.lit(m.colW())
	rows, rests := t.rows, t.rests
	h := m.h
	// The cursor keeps to a band: from the third row to a third of the way
	// down. The page moves only as far as it must to keep it there, so two
	// thirds of the window show what is ahead and what was just left stays
	// in view above.
	at := rests[t.at]
	from := t.top
	low, high := at.row-h/3, at.row-2
	if at.blk >= 0 {
		// In a block the page holds back: the block's top stays in view
		// for as long as the cursor keeps a third of the window below it.
		low = max(at.row-(h-h/3), min(low, at.start-2))
	}
	t.top = max(low, min(t.top, high))
	// A block cut off at the bottom is pulled up, though never so far
	// that its top goes.
	if at.blk >= 0 && at.end > t.top+h {
		t.top = max(t.top, min(high, at.end-h, at.start-2))
	}
	// One line down or up never throws the page: it goes as far as the
	// cursor went and one row more, and catches up over the next lines.
	if step := at.row - t.wasRow; !fresh && (t.at-t.was == 1 || t.was-t.at == 1) {
		t.top = max(from+min(step, 0)-1, min(t.top, from+max(step, 0)+1))
		t.top = max(at.row-h+1, min(t.top, at.row))
	}
	t.was, t.wasRow = t.at, at.row
	t.top = max(0, min(t.top, len(rows)-h))
	return m
}

// Keys are the bindings that work now, for a help line.
func (m Model) Keys() []key.Binding {
	if m.w == 0 || m.h == 0 {
		return nil
	}
	t := m.t
	if t.whole != nil {
		return []key.Binding{keyRoll, keyBack}
	}
	ks := []key.Binding{keyMove}
	if r := t.rests[t.at]; r.blk >= 0 {
		b := *t.block(stop{r.sec, r.blk})
		if b.count() > 1 {
			ks = append(ks, keyStep)
		}
		ks = append(ks, b.keys()...)
	}
	if m.index() {
		ks = append(ks, keyJump(len(t.secs)))
	}
	return ks
}

// Rows is the document as it stands: as many rows as it was given.
func (m Model) Rows() []string {
	lines := make([]string, m.h)
	if m.w == 0 || m.h == 0 {
		return lines
	}
	if m.t.whole != nil {
		m.drawWhole(lines)
	} else {
		m.drawDoc(lines)
	}
	return lines
}

// drawDoc draws the document, its gutter, and the index beside it.
func (m Model) drawDoc(lines []string) {
	t := m.t
	w := m.colW()
	rows, at := t.rows, t.rests[t.at]
	var index []string
	if m.index() {
		index = t.index(at.sec)
	}
	for y := range m.h {
		row, bar := "", " "
		if n := t.top + y; n < len(rows) {
			row = rows[n]
			if n == at.row {
				bar = sHere.Render("▌")
			}
		}
		line := strings.Repeat(" ", Left(m.w)-2) + bar + fit(row, w+2)
		if y < len(index) {
			line += strings.Repeat(" ", gap-1) + index[y]
		}
		lines[y] = line
	}
}

// index is the sections down a rail: each a stop on it, marked with what
// the marks on its pages came to, with its number to go to it by. The one
// the cursor is in is lit; index is handed that section, read off the
// cursor, and keeps nothing itself.
func (t tab) index(in int) []string {
	rail := sRule.Render("│")
	rows := []string{" " + sDim.Render("contents") + " " + sRule.Render(strings.Repeat("─", indexW-10)), ""}
	// The intro comes first, under the title.
	all := append([]markdown.Section{{Title: t.d.Title, Cells: t.d.Intro}}, t.d.Sections...)
	for n, d := range all {
		i := n - 1
		mark := d.Verdict()
		bar, title, stop := " ", sMeta, sDim.Render("○")
		if mark != markdown.None {
			stop = sign(mark == markdown.Holds)
		}
		if i == in {
			bar, title = sHere.Render("▌"), sHead
		}
		if n > 0 {
			rows = append(rows, " "+rail)
		}
		rows = append(rows, bar+stop+" "+sDim.Render(strconv.Itoa(n+1))+"  "+title.Render(ansi.Truncate(d.Title, indexW-6, "…")))
	}
	return rows
}

// drawWhole gives the view to one piece of code.
func (m Model) drawWhole(lines []string) {
	t := m.t
	c := t.whole
	left := strings.Repeat(" ", Left(m.w))
	last := c.Line + len(c.Lines) - 1
	lines[0] = left + codeTitle(c, c.Line, last)
	rows := codeRows(c, c.Line, last)
	for y := 0; y < m.h-2 && t.wholeTop+y < len(rows); y++ {
		lines[2+y] = left + rows[t.wholeTop+y]
	}
}
