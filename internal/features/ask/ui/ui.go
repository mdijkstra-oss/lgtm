// Package ui is where a question is asked and its answer read: a palette to
// ask with, a tab per question, and the answer as a notebook whose blocks
// respond to the keys while the cursor is on them.
package ui

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"strigil/internal/features/ask"
	"strigil/internal/kit/theme"
)

type tab struct {
	a    ask.Answer
	lead []block
	secs []sect
	at   int // the rest the cursor's line is, counting down the page
	top  int // the first row of the notebook on screen

	// The page as last laid out, by settle. A swipe sends keys by the
	// dozen, and laying the page out is the slow part of answering one, so
	// it is laid out again only when stale: bare is the page with the
	// cursor in no block, which moving the cursor does not change.
	rows, bare []string
	rests      []rest
	stale      bool

	// Code given the whole view, and its first line on screen.
	whole    *ask.Code
	wholeTop int
	// Where the cursor was when the page last settled.
	was, wasRow int
}

type sect struct {
	d      *ask.Section
	blocks []block
}

// stop names a block: one of the lead (sec -1) or of a section.
type stop struct{ sec, blk int }

// rest is a row the cursor can be on: a line of text, or a line of a block.
type rest struct {
	row int
	sec int // the section it is in; -1 in the lead
	blk int // the block it is in; -1 for a line of text

	// For a line of a block: the block's first and last rows, and the row
	// where what it says of its pick starts.
	start, end, about int

	// For a line of the question or of a heading: what it says.
	head string
}

// in says the rest is a line of that block.
func (r rest) in(s stop) bool { return r.blk >= 0 && r.sec == s.sec && r.blk == s.blk }

// clone copies what a key may change, so an earlier model stays as it was.
func (t tab) clone() tab {
	t.lead = slices.Clone(t.lead)
	t.secs = slices.Clone(t.secs)
	for i := range t.secs {
		t.secs[i].blocks = slices.Clone(t.secs[i].blocks)
	}
	return t
}

func (t *tab) block(s stop) *block {
	if s.sec < 0 {
		return &t.lead[s.blk]
	}
	return &t.secs[s.sec].blocks[s.blk]
}

type Model struct {
	w, h  int
	known []ask.Answer
	tabs  []tab
	cur   int
	ids   int

	asking bool
	input  textinput.Model
	pick   int
	help   help.Model
}

const placeholder = "Ask about this codebase"

// New opens on the palette, with nothing asked yet.
func New(known []ask.Answer) Model {
	in := textinput.New()
	in.Prompt, in.Placeholder = "> ", placeholder
	look := textinput.StyleState{Text: sName, Placeholder: sDim, Prompt: sAsk}
	in.SetStyles(textinput.Styles{
		Focused: look, Blurred: look,
		Cursor: textinput.CursorStyle{Color: theme.Secondary, Shape: tea.CursorBlock, Blink: true},
	})
	in.SetVirtualCursor(false)
	in.Focus()

	h := help.New()
	h.ShortSeparator = " • "
	h.Styles = help.Styles{ShortKey: sMeta, ShortDesc: sDim, ShortSeparator: sRule, Ellipsis: sRule}
	return Model{known: known, asking: true, input: in, help: h}
}

func Run(ctx context.Context, known []ask.Answer) error {
	_, err := tea.NewProgram(New(known), tea.WithContext(ctx)).Run()
	return err
}

func (m Model) Init() tea.Cmd { return nil }

var (
	keyMove  = key.NewBinding(key.WithKeys("up", "down", "k", "j"), key.WithHelp("↑/↓", "move"))
	keyTab   = key.NewBinding(key.WithKeys("[", "]"), key.WithHelp("[/]", "tab"))
	keyAsk   = key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "ask"))
	keyQuit  = key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))
	keyPick  = key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "choose"))
	keyOpen  = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open"))
	keyClose = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close"))
	keyRoll  = key.NewBinding(key.WithKeys("up", "down", "k", "j"), key.WithHelp("↑/↓", "scroll"))
	keyBack  = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))
	keyStep  = key.NewBinding(key.WithKeys("left", "right", "h", "l"), key.WithHelp("←/→", "step"))
)

// keyJump is the binding for the index: a section's number goes to it.
func keyJump(n int) key.Binding {
	return key.NewBinding(key.WithKeys("1"), key.WithHelp("1-"+strconv.Itoa(min(n+1, 9)), "section"))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.tabs = slices.Clone(m.tabs)
		for i := range m.tabs {
			m.tabs[i].stale = true
		}
		return m.settle(), nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.tabs = slices.Clone(m.tabs)
		for i := range m.tabs {
			m.tabs[i] = m.tabs[i].clone()
		}
		if m.asking {
			next, cmd := m.ask(msg)
			return next.settle(), cmd
		}
		next, cmd := m.read(msg)
		return next.settle(), cmd
	}

	// Anything else is a block's own news: each gets to see it.
	var cmds []tea.Cmd
	m.tabs = slices.Clone(m.tabs)
	for i := range m.tabs {
		t := m.tabs[i].clone()
		tell := func(bs []block) {
			for j, b := range bs {
				var cmd tea.Cmd
				bs[j], cmd = b.update(msg)
				cmds = append(cmds, cmd)
			}
		}
		tell(t.lead)
		for _, s := range t.secs {
			tell(s.blocks)
		}
		t.stale = true
		m.tabs[i] = t
	}
	return m.settle(), tea.Batch(cmds...)
}

// ask is a key while the palette is open.
func (m Model) ask(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	found := ask.Find(m.known, m.input.Value())
	switch msg.String() {
	case "esc", "ctrl+k":
		m.asking = false
	case "up":
		m.pick = max(0, m.pick-1)
	case "down":
		m.pick = max(0, min(len(found)-1, m.pick+1))
	case "enter":
		if m.pick < len(found) {
			m = m.open(found[m.pick])
		}
	default:
		was := m.input.Value()
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if m.input.Value() != was {
			m.pick = 0
		}
		return m, cmd
	}
	return m, nil
}

// open shows an answer on its tab, making the tab if the question is new.
func (m Model) open(a ask.Answer) Model {
	m.asking = false
	for i, t := range m.tabs {
		if t.a.Question == a.Question {
			m.cur = i
			return m
		}
	}
	t := tab{a: a}
	for i := range a.Lead {
		t.lead = append(t.lead, build(&m.ids, &a.Lead[i]))
	}
	for i := range a.Sections {
		s := sect{d: &a.Sections[i]}
		for j := range s.d.Cells {
			s.blocks = append(s.blocks, build(&m.ids, &s.d.Cells[j]))
		}
		t.secs = append(t.secs, s)
	}
	m.tabs = append(m.tabs, t)
	m.cur = len(m.tabs) - 1
	return m
}

// read is a key while an answer, or nothing, is on screen.
func (m Model) read(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+k":
		m.asking, m.pick = true, 0
		m.input.Reset()
		return m, nil
	case "q":
		return m, tea.Quit
	}
	if len(m.tabs) == 0 {
		return m, nil
	}
	t := &m.tabs[m.cur]
	if t.whole != nil {
		switch msg.String() {
		case "esc", "enter":
			t.whole = nil
		case "up", "k":
			t.wholeTop--
		case "down", "j":
			t.wholeTop++
		}
		return m, nil
	}
	rests := t.rests
	if n := msg.String(); len(n) == 1 && n[0] >= '1' && n[0] <= '9' && int(n[0]-'1') <= len(t.secs) {
		// 1 is the question, and the sections follow. A section's first
		// rest is its heading.
		for i, r := range rests {
			// As ints: a byte cannot go below nought, and the question is -1.
			if r.sec == int(n[0])-'2' {
				t.at = i
				break
			}
		}
		return m, nil
	}
	here := rests[t.at]
	switch msg.String() {
	case "up", "k":
		t.at = max(0, t.at-1)
	case "down", "j":
		t.at = min(len(rests)-1, t.at+1)
	case "[":
		m.cur = (m.cur + len(m.tabs) - 1) % len(m.tabs)
	case "]":
		m.cur = (m.cur + 1) % len(m.tabs)
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
		if here.row < here.about {
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
			t.whole, t.wholeTop = c, c.Hot[0]-c.Line-m.bodyH()/3
			return m, nil
		}
		var cmd tea.Cmd
		*b, cmd = (*b).update(msg)
		t.stale = true
		return m, cmd
	}
	return m, nil
}

const (
	edge   = 4 // the least room left of the text; the cell two before the text is the cursor's
	bodyY  = 3
	widest = 88
	indexW = 30
	gap    = 4
)

// index says whether there is room for the index beside the notebook.
func (m Model) index() bool { return m.w-2*edge >= 64+gap+indexW }

func (m Model) colW() int {
	if m.index() {
		return min(widest, m.w-2*edge-gap-indexW)
	}
	return min(widest, m.w-2*edge)
}

// left is where text starts: the notebook, and the index if it is there,
// sit in the middle of the window.
func (m Model) left() int {
	all := m.colW()
	if m.index() {
		all += gap + indexW
	}
	return max(edge, (m.w-all)/2)
}

func (m Model) bodyH() int  { return max(1, m.h-6) }
func (m Model) small() bool { return m.w < 60 || m.h < 16 }

// lay gives the notebook's rows w cells wide, and the rows the cursor can
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
			for n, l := range lines[1 : len(lines)-1] {
				if worded(l) {
					rests = append(rests, rest{
						row: len(rows) + 1 + n, sec: sec, blk: i,
						start: len(rows), end: len(rows) + len(lines), about: len(rows) + 1 + b.about(inner),
					})
				}
			}
			rows = append(append(rows, lines...), "")
		}
	}
	text(-1, "", markdown("# "+t.a.Question, w-1), true)
	cells(-1, "", t.lead)
	for i, s := range t.secs {
		text(i, "", markdown("## "+s.d.Title, w-1), true)
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
	if len(m.tabs) == 0 || m.w == 0 || m.small() {
		return m
	}
	t := &m.tabs[m.cur]
	if t.whole != nil {
		t.wholeTop = max(0, min(t.wholeTop, len(t.whole.Lines)-(m.bodyH()-2)))
	}
	fresh := t.stale || t.bare == nil
	if fresh {
		t.bare, t.rests = t.lay(m.colW())
		t.stale = false
	}
	t.at = min(t.at, len(t.rests)-1)
	t.rows = t.lit(m.colW())
	rows, rests := t.rows, t.rests
	h := m.bodyH()
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

// keys are the bindings that work now, for the help line.
func (m Model) keys() []key.Binding {
	if m.asking {
		return []key.Binding{keyPick, keyOpen, keyClose}
	}
	if len(m.tabs) == 0 {
		return []key.Binding{keyAsk, keyQuit}
	}
	t := m.tabs[m.cur]
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
	if len(m.tabs) > 1 {
		ks = append(ks, keyTab)
	}
	return append(ks, keyAsk, keyQuit)
}

func (m Model) render() string {
	lines := make([]string, m.h)
	if m.small() {
		lines[min(1, m.h-1)] = "  " + sDim.Render("Make the window larger.")
		return m.page(lines)
	}
	left := strings.Repeat(" ", m.left())
	switch {
	case len(m.tabs) == 0:
		lines[bodyY] = left + sDim.Render("Nothing asked yet.")
	case m.tabs[m.cur].whole != nil:
		lines[1] = left + m.tabline()
		m.drawWhole(lines)
	default:
		lines[1] = left + m.tabline()
		m.drawNotebook(lines)
	}
	h := m.help
	h.SetWidth(m.w - m.left() - edge)
	lines[m.h-2] = left + h.ShortHelpView(m.keys())
	page := m.page(lines)
	if !m.asking {
		return page
	}
	box, x, y := m.palette()
	band := strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", m.w)+"\n", lipgloss.Height(box)+2), "\n")
	over := lipgloss.NewCompositor(
		lipgloss.NewLayer(page),
		lipgloss.NewLayer(band).Y(y-1).Z(1),
		lipgloss.NewLayer(box).X(x).Y(y).Z(2),
	).Render()
	return m.page(strings.Split(over, "\n")[:m.h])
}

// drawNotebook draws the answer, its gutter, and the index beside it.
func (m Model) drawNotebook(lines []string) {
	t := m.tabs[m.cur]
	w := m.colW()
	rows, at := t.rows, t.rests[t.at]
	var index []string
	if m.index() {
		index = t.index(at.sec)
	}
	for y := range m.bodyH() {
		row, bar := "", " "
		if n := t.top + y; n < len(rows) {
			row = rows[n]
			if n == at.row {
				bar = sHere.Render("▌")
			}
		}
		line := strings.Repeat(" ", m.left()-2) + bar + fit(row, w+2)
		if y < len(index) {
			line += strings.Repeat(" ", gap-1) + index[y]
		}
		lines[bodyY+y] = line
	}
}

// index is the sections down a rail: each a stop on it, marked with what
// its checks came to, with its number to go to it by and the claim it
// stands or falls on. The one the cursor is in is lit; index is handed
// that section, read off the cursor, and keeps nothing itself.
func (t tab) index(in int) []string {
	rail := sRule.Render("│")
	rows := []string{" " + sDim.Render("contents") + " " + sRule.Render(strings.Repeat("─", indexW-10)), ""}
	// The lead comes first, under the question.
	all := append([]ask.Section{{Title: t.a.Question, Cells: t.a.Lead}}, t.a.Sections...)
	for n, d := range all {
		i := n - 1
		mark, claim := d.Verdict()
		bar, title, said, stop := " ", sMeta, sDim, sDim.Render("○")
		if mark != ask.None {
			stop = sign(mark == ask.Holds)
		}
		if i == in {
			bar, title, said = sHere.Render("▌"), sHead, sMeta
		}
		if n > 0 {
			rows = append(rows, " "+rail)
		}
		rows = append(rows, bar+stop+" "+sDim.Render(strconv.Itoa(n+1))+"  "+title.Render(ansi.Truncate(d.Title, indexW-6, "…")))
		// The claim takes two rows at most.
		lines := strings.Split(lipgloss.Wrap(claim, indexW-6, ""), "\n")
		if len(lines) > 2 {
			lines = append(lines[:1], ansi.Truncate(strings.Join(lines[1:], " "), indexW-6, "…"))
		}
		for _, l := range lines {
			if claim != "" {
				rows = append(rows, " "+rail+"    "+said.Render(l))
			}
		}
	}
	return rows
}

// drawWhole gives the view to one piece of code.
func (m Model) drawWhole(lines []string) {
	t := m.tabs[m.cur]
	c := t.whole
	left := strings.Repeat(" ", m.left())
	last := c.Line + len(c.Lines) - 1
	lines[bodyY] = left + codeTitle(c, c.Line, last)
	rows := codeRows(c, c.Line, last)
	for y := 0; y < m.bodyH()-2 && t.wholeTop+y < len(rows); y++ {
		lines[bodyY+2+y] = left + rows[t.wholeTop+y]
	}
}

// page makes every row exactly as wide as the window.
func (m Model) page(lines []string) string {
	for i := range lines {
		lines[i] = fit(lines[i], m.w)
	}
	return strings.Join(lines, "\n")
}

func (m Model) tabline() string {
	at := "read at " + m.tabs[m.cur].a.At
	line := ""
	for i, t := range m.tabs {
		if i > 0 {
			line += sRule.Render(" │ ")
		}
		if i == m.cur {
			line += sHead.Render(t.a.Tab)
		} else {
			line += sDim.Render(t.a.Tab)
		}
	}
	if room := m.w - 2*m.left() - wid(at) - wid(line); room >= 4 {
		line += strings.Repeat(" ", room) + sDim.Render(at)
	}
	return line
}

// palette is the box to ask in, and where it goes.
func (m Model) palette() (box string, x, y int) {
	inner := min(72, m.w-8) - 6
	in := m.input
	in.SetWidth(inner - 3)
	rows := []string{fit(in.View(), inner), ""}
	found := ask.Find(m.known, in.Value())
	for i := range m.known {
		switch {
		case i >= len(found) && i == 0:
			rows = append(rows, sDim.Render("  No answer for that yet."))
		case i >= len(found):
			rows = append(rows, "")
		case i == m.pick:
			rows = append(rows, sPick.Render(fit("  "+found[i].Question, inner)))
		default:
			rows = append(rows, sBody.Render(fit("  "+found[i].Question, inner)))
		}
	}
	box = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(theme.Primary).
		Padding(1, 2).Render(strings.Join(rows, "\n"))
	return box, (m.w - wid(box)) / 2, max(2, (m.h-lipgloss.Height(box))/3)
}

func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.WindowTitle = "strigil"
	v.BackgroundColor = theme.Surface0
	if m.w == 0 {
		return v
	}
	v.SetContent(m.render())
	if m.asking && !m.small() {
		if c := m.input.Cursor(); c != nil {
			_, x, y := m.palette()
			c.X, c.Y = c.X+x+3, c.Y+y+2
			v.Cursor = c
		}
	}
	return v
}
