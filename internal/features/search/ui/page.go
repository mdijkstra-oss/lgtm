package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"strigil/internal/kit"
	"strigil/internal/kit/frame"
	"strigil/internal/kit/text"
	"strigil/internal/kit/theme"
)

const headerRows = 2

// Page is one search: every hit of a query, grouped by file. Enter opens the
// file of the selected hit over the list.
type Page struct {
	th    theme.Theme
	dir   string
	query string

	loaded bool
	failed string
	docs   []*doc
	rows   []row
	hits   []place
	sel    int
	top    int
	want   *ref // a hit asked for before the results were in

	viewing bool
	fileTop int

	w, h int
}

type rowKind int

const (
	rowGap rowKind = iota
	rowFile
	rowLine
)

type row struct {
	kind rowKind
	doc  int
	line int
}

// place is one hit: which file, which of its hits, which row shows it.
type place struct{ doc, idx, row int }

type loadedMsg struct {
	id   string
	docs []*doc
	err  error
}

func (p Page) ID() string    { return pageID(p.query) }
func (p Page) Title() string { return "search · " + p.query }

func (p Page) Meta() string {
	if !p.loaded {
		return "…"
	}
	return strconv.Itoa(len(p.hits))
}

func (p Page) Init() tea.Cmd {
	id, dir, query := p.ID(), p.dir, p.query
	return func() tea.Msg {
		docs, err := collect(context.Background(), dir, query)
		return loadedMsg{id: id, docs: docs, err: err}
	}
}

func (p Page) Hints() []kit.Hint {
	switch {
	case p.viewing:
		return []kit.Hint{{Key: "↑↓", Label: "scroll"}, {Key: "esc", Label: "back"}}
	case len(p.hits) > 0:
		return []kit.Hint{{Key: "↑↓", Label: "move"}, {Key: "enter", Label: "view file"}}
	}
	return nil
}

func (p Page) body() int { return max(1, p.h-headerRows) }

func (p Page) Resize(w, h int) kit.Page {
	p.w, p.h = w, h
	p.fileTop = p.clampFile(p.fileTop)
	return p.scrolled(p.top)
}

func (p Page) Reveal(it kit.Item, covered int) kit.Page {
	r, ok := it.Data.(ref)
	if !ok {
		return p
	}
	p.viewing = false
	if !p.loaded {
		p.want = &r
		return p
	}
	i := p.locate(r)
	if i < 0 {
		return p
	}
	p.sel = i
	at := p.hits[i].row

	from := max(0, covered-headerRows)
	if from >= p.body() {
		from = 0
	}
	if from == 0 {
		p = p.scrolled(p.top)
	}
	if on := at - p.top; on >= from && on < p.body() {
		return p
	}
	top := at - from - (p.body()-from)/2
	if from == 0 {
		return p.scrolled(top)
	}
	// A hit near the start of the list can only clear the overlay by
	// scrolling past the first row.
	p.top = min(top, max(0, len(p.rows)-p.body()))
	return p
}

func (p Page) locate(r ref) int {
	for i, pl := range p.hits {
		d := p.docs[pl.doc]
		if d.rel == r.path && d.hits[pl.idx].line == r.line && d.hits[pl.idx].start == r.start {
			return i
		}
	}
	return -1
}

func (p Page) Update(msg tea.Msg) (kit.Page, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		if msg.id != p.ID() || p.loaded {
			return p, nil
		}
		return p.filled(msg), nil
	case tea.KeyPressMsg:
		if len(p.hits) == 0 {
			return p, nil
		}
		if p.viewing {
			return p.fileKey(msg.String()), nil
		}
		return p.listKey(msg.String()), nil
	}
	return p, nil
}

func (p Page) filled(msg loadedMsg) Page {
	p.loaded = true
	if msg.err != nil {
		p.failed = msg.err.Error()
		return p
	}
	p.docs = msg.docs
	for di, d := range p.docs {
		if di > 0 {
			p.rows = append(p.rows, row{kind: rowGap})
		}
		p.rows = append(p.rows, row{kind: rowFile, doc: di})
		last := -1
		for hi, h := range d.hits {
			if h.line != last {
				p.rows = append(p.rows, row{kind: rowLine, doc: di, line: h.line})
				last = h.line
			}
			p.hits = append(p.hits, place{doc: di, idx: hi, row: len(p.rows) - 1})
		}
	}
	if p.want != nil {
		if i := p.locate(*p.want); i >= 0 {
			p.sel = i
		}
		p.want = nil
	}
	if len(p.hits) == 0 {
		return p
	}
	if at := p.hits[p.sel].row; at >= p.body() {
		return p.scrolled(at - p.body()/2)
	}
	return p
}

func (p Page) listKey(key string) Page {
	switch key {
	case "up":
		return p.selected(p.sel - 1)
	case "down":
		return p.selected(p.sel + 1)
	case "pgup":
		return p.selected(p.sel - p.body())
	case "pgdown":
		return p.selected(p.sel + p.body())
	case "home":
		return p.selected(0)
	case "end":
		return p.selected(len(p.hits) - 1)
	case "enter":
		p.viewing = true
		p.fileTop = p.clampFile(p.current().line - p.fileBody()/2)
	}
	return p
}

func (p Page) fileKey(key string) Page {
	switch key {
	case "esc", "enter":
		p.viewing = false
	case "up":
		p.fileTop = p.clampFile(p.fileTop - 1)
	case "down":
		p.fileTop = p.clampFile(p.fileTop + 1)
	case "pgup":
		p.fileTop = p.clampFile(p.fileTop - p.fileBody())
	case "pgdown":
		p.fileTop = p.clampFile(p.fileTop + p.fileBody())
	case "home":
		p.fileTop = 0
	case "end":
		p.fileTop = p.clampFile(len(p.file().lines))
	}
	return p
}

func (p Page) file() *doc   { return p.docs[p.hits[p.sel].doc] }
func (p Page) current() hit { return p.file().hits[p.hits[p.sel].idx] }

// selected moves to hit i and keeps its row, and the file name above the
// first row of a group, on screen.
func (p Page) selected(i int) Page {
	p.sel = max(0, min(i, len(p.hits)-1))
	at := p.hits[p.sel].row
	if at-1 < p.top {
		p.top = at - 1
	}
	if at >= p.top+p.body() {
		p.top = at - p.body() + 1
	}
	return p.scrolled(p.top)
}

func (p Page) scrolled(top int) Page {
	p.top = max(0, min(top, len(p.rows)-p.body()))
	return p
}

func (p Page) View() string {
	pen := theme.On(p.th.Bg)
	rows := []string{p.header(), pen.Ink(p.th.Rule, strings.Repeat("─", p.w))}
	switch {
	case !p.loaded:
		rows = append(rows, pen.Middle([]string{pen.Ink(p.th.Dim, "Searching…")}, p.w, p.body())...)
	case p.failed != "":
		rows = append(rows, pen.Middle([]string{pen.Ink(p.th.Lit, text.Head(p.failed, p.w-4))}, p.w, p.body())...)
	case len(p.hits) == 0:
		rows = append(rows, pen.Middle([]string{pen.Ink(p.th.Soft, "Nothing matches")}, p.w, p.body())...)
	default:
		for i := range p.body() {
			rows = append(rows, p.row(p.top+i))
		}
	}
	list := strings.Join(rows, "\n")
	if !p.viewing {
		return list
	}
	x, y, w, h := p.fileBox()
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(list),
		lipgloss.NewLayer(p.fileView(w, h)).X(x).Y(y).Z(1),
	).Render()
}

func (p Page) header() string {
	pen := theme.On(p.th.Bg)
	right := ""
	if len(p.hits) > 0 {
		files := fmt.Sprintf("%d files", len(p.docs))
		if len(p.docs) == 1 {
			files = "1 file"
		}
		right = pen.Ink(p.th.Dim, fmt.Sprintf("%d of %d", p.sel+1, len(p.hits))) +
			pen.Ink(p.th.Faint, "  ·  "+files) + pen.Gap(2)
	}
	room := p.w - lipgloss.Width(right) - 14
	left := pen.Gap(2) + pen.Ink(p.th.Dim, "search · ") +
		pen.Style(p.th.Match).Bold(true).Render(text.Head(p.query, room))
	return pen.Split(left, right, p.w)
}

func (p Page) row(i int) string {
	pen := theme.On(p.th.Bg)
	if i < 0 || i >= len(p.rows) {
		return pen.Gap(p.w)
	}
	r := p.rows[i]
	d := p.docs[r.doc]
	switch r.kind {
	case rowFile:
		count := strconv.Itoa(len(d.hits))
		right := pen.Ink(p.th.Faint, count) + pen.Gap(2)
		return pen.Split(pen.Gap(2)+path(p.th, pen, d.rel, p.w-6-len(count)), right, p.w)
	case rowLine:
		sel, num := -1, p.th.Faint
		if cur := p.hits[p.sel]; cur.row == i {
			pen, sel, num = theme.On(p.th.Band), cur.idx, p.th.Accent
		}
		digits := p.digits()
		gutter := pen.Gap(4) + pen.Ink(num, fmt.Sprintf("%*d", digits, r.line+1)) + pen.Gap(2)
		return pen.Line(gutter+content(p.th, pen, d, r.line, sel, p.w-digits-8), p.w)
	}
	return pen.Gap(p.w)
}

func (p Page) digits() int {
	n := 3
	for _, d := range p.docs {
		n = max(n, len(strconv.Itoa(len(d.lines))))
	}
	return n
}

// path draws a relative path with its last part standing out.
func path(th theme.Theme, pen theme.Pen, rel string, width int) string {
	shown := text.Tail(rel, width)
	cut := max(0, len(shown)-len(filepath.Base(rel)))
	return pen.Ink(th.Dim, shown[:cut]) + pen.Style(th.Text).Bold(true).Render(shown[cut:])
}

// fileBox leaves the list showing on every side of the opened file.
func (p Page) fileBox() (x, y, w, h int) {
	x, y = 6, 2
	if p.w < 80 {
		x = 2
	}
	if p.h < 20 {
		y = 1
	}
	return x, y, p.w - 2*x, p.h - 2*y
}

func (p Page) fileBody() int {
	_, _, w, h := p.fileBox()
	_, bh := frame.BodySize(w, h, true, false)
	return max(1, bh)
}

func (p Page) clampFile(top int) int {
	if len(p.hits) == 0 {
		return 0
	}
	return max(0, min(top, len(p.file().lines)-p.fileBody()))
}

func (p Page) fileView(w, h int) string {
	pen := theme.On(p.th.Panel)
	d, cur := p.file(), p.current()
	bw, bh := frame.BodySize(w, h, true, false)
	where := pen.Ink(p.th.Dim, fmt.Sprintf("line %d", cur.line+1)) + pen.Gap(1)
	return frame.Box{
		Head: pen.Split(pen.Gap(1)+path(p.th, pen, d.rel, bw-lipgloss.Width(where)-3), where, bw),
		Body: excerpt(p.th, p.th.Panel, d, p.hits[p.sel].idx, p.fileTop, bh, bw),
	}.Render(p.th, w, h)
}
