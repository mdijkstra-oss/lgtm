package ui

import (
	"context"
	"image/color"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"strigil/internal/features/ask"
	"strigil/internal/kit/theme"
)

// block is one cell of a notebook on screen, drawn from its own state
// alone. One with things to pick has one of them picked.
type block interface {
	view(w int, on bool) []string
	update(msg tea.Msg) (block, tea.Cmd)
	keys() []key.Binding
	count() int // how many things there are to pick; 0 for text
	picked() int
	pick(i int) block
	// about is the row of view where what is said of the pick starts:
	// everything from there down changes with the pick.
	about(w int) int
	// code is what the block is showing lines of now, if anything: enter
	// gives those the whole view.
	code() *ask.Code
}

var (
	keyFull = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "whole view"))
	keyRun  = key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "run it now"))
)

// withCode adds the key for the whole view when there are lines to give it.
func withCode(c *ask.Code, ks ...key.Binding) []key.Binding {
	if c != nil {
		ks = append(ks, keyFull)
	}
	return ks
}

// lines draws the code a block is showing: where it is, then the lines.
func lines(c *ask.Code) []string {
	switch {
	case c == nil:
		return nil
	case c.Line == 0:
		return append([]string{codeTitle(c, 0, 0)}, codeRows(c, 0, 0)...)
	}
	return append([]string{codeTitle(c, c.Show[0], c.Show[1])}, codeRows(c, c.Show[0], c.Show[1])...)
}

func build(id *int, c *ask.Cell) block {
	*id++
	switch {
	case c.Flow != nil:
		return flow{d: c.Flow}
	case c.Check != nil:
		return check{d: c.Check}
	case c.Colours != nil:
		return colours{d: c.Colours}
	case c.Run != nil:
		return run{d: c.Run, id: *id, busy: -1, now: make([]*ask.Cmd, len(c.Run.Cmds)),
			spin: spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(sLive))}
	case c.Change != nil:
		return change{d: c.Change}
	}
	return prose(c.Prose)
}

// prose is markdown.
type prose string

func (p prose) view(w int, _ bool) []string     { return markdown(string(p), w) }
func (p prose) update(tea.Msg) (block, tea.Cmd) { return p, nil }
func (p prose) keys() []key.Binding             { return nil }
func (p prose) about(int) int                   { return 0 }
func (p prose) count() int                      { return 0 }
func (p prose) picked() int                     { return 0 }
func (p prose) pick(int) block                  { return p }
func (p prose) code() *ask.Code                 { return nil }

// flow draws the way through the packages, a row a hop, and says what the
// hop it is on does.
type flow struct {
	d  *ask.Flow
	at int
}

func (f flow) code() *ask.Code                 { return f.d.Hops[f.at].Code }
func (f flow) keys() []key.Binding             { return withCode(f.code()) }
func (f flow) update(tea.Msg) (block, tea.Cmd) { return f, nil }

// The lanes, a blank row, a row a hop, and a blank row.
func (f flow) about(int) int    { return 2 + len(f.d.Hops) + 1 }
func (f flow) count() int       { return len(f.d.Hops) }
func (f flow) picked() int      { return f.at }
func (f flow) pick(i int) block { f.at = i; return f }

func (f flow) view(w int, on bool) []string {
	lane := w / len(f.d.Lanes)
	var head []seg
	for i, l := range f.d.Lanes {
		head = append(head, seg{i*lane + 1, ansi.Truncate(l, lane-3, "…"), sDim})
	}
	rows := []string{placed(head...), ""}

	prev := -1
	for i, h := range f.d.Hops {
		line := sDim
		if i == f.at {
			line = sLive
		}
		x, hop := h.Lane*lane, button("", h.Name, at(i, f.at, on))
		plain := lipgloss.NewStyle()
		switch {
		case prev >= 0 && h.Lane > prev:
			from := prev*lane + 1
			rows = append(rows, placed(seg{from, "╰" + strings.Repeat("─", x-from-3) + "▶", line}, seg{x, hop, plain}))
		case prev >= 0 && h.Lane < prev:
			to, back := prev*lane+1, x+wid(hop)+1
			rows = append(rows, placed(seg{x, hop, plain}, seg{back, "◀" + strings.Repeat("─", max(0, to-back-1)) + "╯", line}))
		default:
			rows = append(rows, placed(seg{x, hop, plain}))
		}
		prev = h.Lane
	}

	// The note keeps the height of the longest. The code takes the room it
	// needs.
	tall := 0
	for _, h := range f.d.Hops {
		tall = max(tall, len(markdown(h.Note, w-1)))
	}
	note := markdown(f.d.Hops[f.at].Note, w-1)
	rows = append(rows, "")
	for i := range tall {
		if i < len(note) {
			rows = append(rows, " "+note[i])
		} else {
			rows = append(rows, "")
		}
	}
	if code := lines(f.code()); len(code) > 0 {
		rows = append(rows, "")
		for _, r := range panel(code, 0) {
			rows = append(rows, " "+r)
		}
	}
	return rows
}

// check is a claim and the cases it was tried on, one of them shown.
type check struct {
	d  *ask.Check
	at int
}

func (c check) code() *ask.Code                 { return c.d.Cases[c.at].Code }
func (c check) keys() []key.Binding             { return withCode(c.code()) }
func (c check) update(tea.Msg) (block, tea.Cmd) { return c, nil }
func (c check) count() int                      { return len(c.d.Cases) }
func (c check) picked() int                     { return c.at }
func (c check) pick(i int) block                { c.at = i; return c }

// chips says which cases go on which row: as many to a row as fit.
func (c check) chips(w int) [][]int {
	rows, used := [][]int{nil}, 0
	for i, k := range c.d.Cases {
		n := wid(k.Name) + 4
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

// The claim, a blank row, the rows of cases with a blank row between, and
// a blank row.
func (c check) about(w int) int { return 2 + 2*len(c.chips(w)) }

func (c check) view(w int, on bool) []string {
	held := c.d.Held()
	rows := []string{
		" " + sign(held == len(c.d.Cases)) + " " + sName.Render(c.d.Claim) + "  " +
			sDim.Render(strconv.Itoa(held)+" of "+strconv.Itoa(len(c.d.Cases))+" hold"),
		"",
	}

	for n, row := range c.chips(w) {
		if n > 0 {
			rows = append(rows, "")
		}
		var chips []string
		for _, i := range row {
			chips = append(chips, button(tick(c.d.Cases[i].Holds), c.d.Cases[i].Name, at(i, c.at, on)))
		}
		rows = append(rows, strings.Join(chips, " "))
	}
	rows = append(rows, "")

	k := c.d.Cases[c.at]
	// What is said of a case keeps the height of the longest.
	tall := 0
	for _, o := range c.d.Cases {
		tall = max(tall, len(markdown(o.Why, w-1)))
	}
	why := markdown(k.Why, w-1)
	for i := range tall {
		if i < len(why) {
			rows = append(rows, " "+why[i])
		} else {
			rows = append(rows, "")
		}
	}
	if code := lines(k.Code); len(code) > 0 {
		rows = append(rows, "")
		for _, r := range panel(code, 0) {
			rows = append(rows, " "+r)
		}
	}
	return rows
}

// colours shows a palette as the colours themselves.
type colours struct {
	d  *ask.Colours
	at int
}

var swatch = map[string]color.Color{
	"Primary": theme.Primary, "Secondary": theme.Secondary, "Accent": theme.Accent, "OnPrimary": theme.OnPrimary,
	"Text1": theme.Text1, "Text2": theme.Text2, "Text3": theme.Text3, "Text4": theme.Text4, "Line": theme.Line,
	"Surface0": theme.Surface0, "Surface1": theme.Surface1, "Surface2": theme.Surface2, "Surface3": theme.Surface3,
	"Error": theme.Error, "Warning": theme.Warning, "Info": theme.Info, "Success": theme.Success,
}

func (c colours) all() []ask.Colour {
	var out []ask.Colour
	for _, g := range c.d.Groups {
		out = append(out, g.Items...)
	}
	return out
}

func (c colours) code() *ask.Code                 { return nil }
func (c colours) keys() []key.Binding             { return nil }
func (c colours) update(tea.Msg) (block, tea.Cmd) { return c, nil }
func (c colours) count() int                      { return len(c.all()) }
func (c colours) picked() int                     { return c.at }
func (c colours) pick(i int) block                { c.at = i; return c }

func (c colours) about(int) int { return 2 * len(c.d.Groups) }

func (c colours) view(w int, on bool) []string {
	var rows []string
	n := 0
	for _, g := range c.d.Groups {
		var line strings.Builder
		line.WriteString(" " + sDim.Render(fit(g.Name, 10)))
		for _, it := range g.Items {
			line.WriteString(lipgloss.NewStyle().Foreground(swatch[it.Name]).Render("██"))
			line.WriteString(" " + button("", it.Name, at(n, c.at, on)) + "  ")
			n++
		}
		rows = append(rows, line.String(), "")
	}
	it := c.all()[c.at]
	uses := strconv.Itoa(it.Uses) + " uses"
	switch it.Uses {
	case 0:
		uses = "no users yet"
	case 1:
		uses = "1 use"
	}
	sep := sDim.Render(" • ")
	return append(rows, " "+sName.Render(it.Name)+sep+sMeta.Render("Charmtone "+it.Tone)+sep+sMeta.Render(it.Job)+sep+sMeta.Render(uses))
}

// run is commands with how each ended, and can run one again.
type run struct {
	d    *ask.Run
	id   int
	at   int
	busy int        // the command running now, or -1
	now  []*ask.Cmd // how each ended when run from here, if it was
	spin spinner.Model
}

type ranMsg struct {
	id, at int
	cmd    ask.Cmd
}

func (r run) about(int) int    { return len(r.d.Cmds) + 1 }
func (r run) count() int       { return len(r.d.Cmds) }
func (r run) picked() int      { return r.at }
func (r run) pick(i int) block { r.at = i; return r }

func (r run) code() *ask.Code {
	c := r.shown(r.at)
	return &ask.Code{File: c.Line, Lines: c.Out}
}

func (r run) keys() []key.Binding {
	if r.busy < 0 {
		return []key.Binding{keyRun, keyFull}
	}
	return []key.Binding{keyFull}
}

func (r run) update(msg tea.Msg) (block, tea.Cmd) {
	switch msg := msg.(type) {
	case ranMsg:
		if msg.id == r.id {
			r.now = append([]*ask.Cmd(nil), r.now...)
			r.now[msg.at], r.busy = &msg.cmd, -1
		}
		return r, nil
	case spinner.TickMsg:
		if r.busy < 0 {
			return r, nil
		}
		var cmd tea.Cmd
		r.spin, cmd = r.spin.Update(msg)
		return r, cmd
	case tea.KeyPressMsg:
		if msg.String() == "r" && r.busy < 0 {
			r.busy = r.at
			id, at, line := r.id, r.at, r.d.Cmds[r.at].Line
			return r, tea.Batch(r.spin.Tick, func() tea.Msg {
				return ranMsg{id, at, ask.Exec(context.Background(), line)}
			})
		}
	}
	return r, nil
}

// shown is the command at i as it last ended, here or when recorded.
func (r run) shown(i int) ask.Cmd {
	if r.now[i] != nil {
		return *r.now[i]
	}
	return r.d.Cmds[i]
}

func (r run) view(w int, on bool) []string {
	var rows []string
	wide := 0
	for _, c := range r.d.Cmds {
		wide = max(wide, wid(c.Line))
	}
	wide = min(wide, w-16)
	for i := range r.d.Cmds {
		c := r.shown(i)
		took := sDim.Render(c.Took)
		if i == r.busy {
			took = r.spin.View() + " " + sDim.Render("running…")
		}
		rows = append(rows, button(tick(c.OK), fit(c.Line, wide), at(i, r.at, on))+"  "+took)
	}

	c := r.shown(r.at)
	when := "recorded at " + r.d.At
	if r.now[r.at] != nil {
		when = "run here just now"
	}
	out := []string{sDim.Render(when)}
	const tall = 7
	for i, l := range c.Out {
		if i == tall-2 && len(c.Out) > tall-1 {
			out = append(out, sDim.Render("… "+strconv.Itoa(len(c.Out)-i)+" more lines"))
			break
		}
		out = append(out, sBody.Render(l))
	}
	rows = append(rows, "")
	for _, p := range panel(out, tall) {
		rows = append(rows, " "+p)
	}
	return rows
}

// change draws what a commit did to each part as a bar: lines out to the
// left of a shared edge, lines in to the right.
type change struct {
	d  *ask.Change
	at int
}

func (c change) code() *ask.Code                 { return nil }
func (c change) keys() []key.Binding             { return nil }
func (c change) update(tea.Msg) (block, tea.Cmd) { return c, nil }

// A bar a part, and a blank row.
func (c change) about(int) int    { return len(c.d.Parts) + 1 }
func (c change) count() int       { return len(c.d.Parts) }
func (c change) picked() int      { return c.at }
func (c change) pick(i int) block { c.at = i; return c }

func (c change) view(w int, on bool) []string {
	nameW, most := 0, 1
	for _, p := range c.d.Parts {
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
	for i, p := range c.d.Parts {
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
			seg{0, button("", fit(p.Name, nameW), at(i, c.at, on)), lipgloss.NewStyle()},
			seg{mid - wid(out), out, sBad},
			seg{mid, in, sOK},
			seg{w - wid(count), count, sDim},
		))
	}
	p := c.d.Parts[c.at]
	files := strconv.Itoa(p.Files) + " files"
	if p.Files == 1 {
		files = "1 file"
	}
	last := " " + sName.Render(p.Name) + sDim.Render(" • ") + sMeta.Render(files)
	if p.Note != "" {
		last += sDim.Render(" • ") + sMeta.Render(p.Note)
	}
	return append(rows, "", last)
}
