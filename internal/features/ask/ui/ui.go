// Package ui is where a document is asked for, opened and read. It opens
// on the history of the repository, to pick a change from or to ask a
// question over; while the model works the page shows what it is doing;
// and each document it writes opens on a tab, drawn by kit/viewer.
package ui

import (
	"context"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"lgtm/internal/features/ask"
	"lgtm/internal/kit/markdown"
	"lgtm/internal/kit/scrape"
	"lgtm/internal/kit/theme"
	"lgtm/internal/kit/viewer"
)

var (
	pal theme.Theme

	sHead, sName, sBody, sMeta, sDim, sRule, sPick, sPrompt lipgloss.Style
	sLive, sBad                                             lipgloss.Style
)

func init() { use(theme.Dark) }

// use draws everything from here on in that theme, the documents too.
func use(t theme.Theme) {
	pal = t
	viewer.Use(t)
	scrape.Use(t)
	sLive = lipgloss.NewStyle().Foreground(t.Secondary)
	sBad = lipgloss.NewStyle().Foreground(t.Error)
	sHead = lipgloss.NewStyle().Foreground(t.Text1).Bold(true)
	sName = lipgloss.NewStyle().Foreground(t.Text1)
	sBody = lipgloss.NewStyle().Foreground(t.Text2)
	sMeta = lipgloss.NewStyle().Foreground(t.Text3)
	sDim = lipgloss.NewStyle().Foreground(t.Text4)
	sRule = lipgloss.NewStyle().Foreground(t.Line)
	sPick = lipgloss.NewStyle().Background(t.Primary).Foreground(t.OnPrimary).Bold(true)
	sPrompt = lipgloss.NewStyle().Foreground(t.Accent)
}

func wid(s string) int { return lipgloss.Width(s) }

// fit makes s exactly w cells wide.
func fit(s string, w int) string {
	s = ansi.Truncate(s, w, "…")
	return s + strings.Repeat(" ", max(0, w-wid(s)))
}

// Start is what the page opens with.
type Start struct {
	Repo, Branch string
	History      []ask.Commit
	Uncommitted  int          // files that differ from the last commit
	Answers      []ask.Answer // documents written before, the newest first
	// Write has a document written; the page shows what it is told.
	Write func(ctx context.Context, sub ask.Subject, on func(ask.Event)) (ask.Written, error)
	// First is asked for at once, when the command line named something.
	First *ask.Subject
}

type Model struct {
	w, h int
	ctx  context.Context
	from Start

	answers []ask.Answer
	tabs    []viewer.Model
	cur     int

	// The opening page: the history to pick from, and a line to narrow
	// it or to ask on.
	home   bool
	input  textinput.Model
	pick   int
	spread []string // the days whose commits are all listed
	help   help.Model

	work *work // the document being written, if one is

	// The slop the page opens under, until the blade has been over it.
	slop    scrape.Model
	covered bool
}

const placeholder = "Search the history, or ask a question"

// recents is how many earlier questions the opening page offers.
const recents = 5

// New opens on the history, with nothing open yet.
func New(from Start) Model {
	in := textinput.New()
	in.Prompt, in.Placeholder = "> ", placeholder
	look := textinput.StyleState{Text: sName, Placeholder: sDim, Prompt: sPrompt}
	in.SetStyles(textinput.Styles{
		Focused: look, Blurred: look,
		Cursor: textinput.CursorStyle{Color: pal.Secondary, Shape: tea.CursorBlock, Blink: true},
	})
	in.SetVirtualCursor(false)
	in.Focus()

	h := help.New()
	h.ShortSeparator = " • "
	h.Styles = help.Styles{ShortKey: sMeta, ShortDesc: sDim, ShortSeparator: sRule, Ellipsis: sRule}
	return Model{ctx: context.Background(), from: from, answers: from.Answers, home: true, input: in, help: h}
}

// covers says whether the page opens under slop for a blade to scrape off.
const covers = true

// Run opens the page, and starts on what the command line named if it
// named something.
func Run(ctx context.Context, from Start) error {
	use(theme.For(lipgloss.HasDarkBackground(os.Stdin, os.Stdout)))
	m := New(from)
	m.ctx = ctx
	var start []tea.Cmd
	if covers {
		m.slop, m.covered = scrape.New(), true
		start = append(start, scrapeTick())
	}
	if from.First != nil {
		var begin tea.Cmd
		m, begin = m.begin(*from.First)
		start = append(start, begin)
	}
	_, err := tea.NewProgram(started{m, tea.Batch(start...)}, tea.WithContext(ctx)).Run()
	return err
}

// started is a page with what it has to set going first.
type started struct {
	Model
	first tea.Cmd
}

func (s started) Init() tea.Cmd { return s.first }

func (m Model) Init() tea.Cmd { return nil }

var (
	keyTab    = key.NewBinding(key.WithKeys("[", "]"), key.WithHelp("[/]", "tab"))
	keyHome   = key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "history"))
	keyQuit   = key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))
	keyPick   = key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "choose"))
	keyGo     = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "explain"))
	keyAsk    = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "ask"))
	keyOpen   = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open"))
	keyMore   = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show them"))
	keyBack   = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))
	keyLeave  = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit"))
	keyStop   = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "stop"))
	keyRecent = func(n int) key.Binding {
		if n == 1 {
			return key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "earlier question"))
		}
		return key.NewBinding(key.WithKeys("1"), key.WithHelp("1-"+strconv.Itoa(n), "earlier question"))
	}
)

// The tabs take three rows and a blank one above the document; the help
// line and the blank rows round it take three below.
const bodyY = 4

func (m Model) bodyH() int  { return max(1, m.h-7) }
func (m Model) small() bool { return m.w < 60 || m.h < 16 }
func (m Model) colW() int   { return min(viewer.Widest, m.w-2*viewer.Left(m.w)) }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.tabs = slices.Clone(m.tabs)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		for i := range m.tabs {
			m.tabs[i] = m.tabs[i].Size(m.w, m.bodyH())
		}
		m.slop = m.slop.Size(m.w, m.h)
		return m, nil
	case scrapeMsg:
		// Once the page is clean nothing of the slop moves, so its clock stops.
		if m.slop = m.slop.Step(); m.slop.Busy() {
			return m, scrapeTick()
		}
		m.covered = false
		return m, nil
	case workMsg, eventsMsg, doneMsg:
		return m.working(msg)
	case tea.PasteMsg:
		if m.home && m.work == nil {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.pick = 0
			return m, cmd
		}
	case tea.KeyPressMsg:
		switch {
		case msg.String() == "ctrl+c":
			return m.stop(), tea.Quit
		case m.work != nil:
			switch msg.String() {
			case "esc":
				return m.stop(), nil
			case "q":
				return m.stop(), tea.Quit
			}
			return m, nil
		case m.home:
			return m.choose(msg)
		}
		return m.read(msg)
	}
	return m, nil
}

// item is one line of the opening page's list: the name of a day, what
// is not committed, a commit, the rest of a long day, the offer to ask
// what was typed, or a blank between days.
type item struct {
	day     string // a day's name, over its commits: not to be picked
	pending bool
	commit  *ask.Commit
	more    string // the day there is more of
	hidden  int    // how many more
	asks    string
}

func (it item) picks() bool { return it.pending || it.commit != nil || it.more != "" || it.asks != "" }

// perDay is how many commits of one day are listed before the rest are
// folded away: a day of forty commits is not forty rows to read past.
const perDay = 4

// items gives the list as what is typed leaves it. The commits are those
// that have every word typed, in their hash, their subject or their
// author, a day at a time. When none has, what was typed is a question.
func (m Model) items() []item {
	typed := strings.TrimSpace(m.input.Value())
	words := strings.Fields(strings.ToLower(typed))
	has := func(text string) bool {
		text = strings.ToLower(text)
		for _, w := range words {
			if !strings.Contains(text, w) {
				return false
			}
		}
		return true
	}
	var out []item
	if m.from.Uncommitted > 0 && has("uncommitted changes") {
		out = append(out, item{pending: true})
	}
	day, shown := "", 0
	fold := func() {
		if n := len(out); n > 0 && out[n-1].more != "" && out[n-1].hidden == 1 {
			// One more is shown, not folded: the fold would take its row.
			out[n-1] = item{commit: out[n-1].commit}
		}
	}
	for i := range m.from.History {
		c := &m.from.History[i]
		if !has(c.Hash + " " + c.Subject + " " + c.Author) {
			continue
		}
		if c.Day != day || len(out) == 0 || (len(out) == 1 && out[0].pending) {
			fold()
			if len(out) > 0 {
				out = append(out, item{})
			}
			out = append(out, item{day: c.Day})
			day, shown = c.Day, 0
		}
		// Searching, every match is wanted; so is a day that was opened.
		if shown < perDay || typed != "" || slices.Contains(m.spread, c.Day) {
			out = append(out, item{commit: c})
			shown++
		} else if n := len(out); out[n-1].more != "" {
			out[n-1].hidden++
		} else {
			out = append(out, item{more: c.Day, hidden: 1, commit: c})
		}
	}
	fold()
	if len(out) == 0 && typed != "" {
		out = append(out, item{asks: typed})
	}
	return out
}

// picked gives the item the cursor is on, among those that can be picked.
func (m Model) picked(items []item) (item, int, bool) {
	n := 0
	for i, it := range items {
		if !it.picks() {
			continue
		}
		if n == m.pick {
			return it, i, true
		}
		n++
	}
	return item{}, 0, false
}

// earlier gives the questions answered before, the newest first. A change
// explained before is not among them: it is in the history, marked.
func (m Model) earlier() []ask.Answer {
	var out []ask.Answer
	seen := map[string]bool{}
	for _, a := range m.answers {
		// Asked twice, a question is listed once, with its latest answer.
		if q := strings.ToLower(a.Asked); a.Change == "" && !seen[q] && len(out) < recents {
			seen[q] = true
			out = append(out, a)
		}
	}
	return out
}

// explained gives the document there already is for a commit. A commit
// does not change, so what was written about it is opened again.
func (m Model) explained(hash string) (markdown.Doc, bool) {
	for _, a := range m.answers {
		if a.Change == hash {
			return a.Doc, true
		}
	}
	return markdown.Doc{}, false
}

// choose is a key on the opening page.
func (m Model) choose(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	items := m.items()
	count := 0
	for _, it := range items {
		if it.picks() {
			count++
		}
	}
	switch k := msg.String(); k {
	case "esc":
		if len(m.tabs) == 0 {
			return m, tea.Quit
		}
		m.home = false
	case "ctrl+k":
		m.home = len(m.tabs) == 0
	case "up":
		m.pick = max(0, m.pick-1)
	case "down":
		m.pick = max(0, min(count-1, m.pick+1))
	case "enter":
		it, _, ok := m.picked(items)
		switch {
		case !ok:
		case it.pending:
			return m.begin(ask.Subject{Pending: true})
		case it.more != "":
			m.spread = append(slices.Clone(m.spread), it.more)
		case it.commit != nil:
			if doc, ok := m.explained(it.commit.Hash); ok {
				return m.open(doc), nil
			}
			return m.begin(ask.Subject{Commit: it.commit.Hash, About: it.commit.Subject})
		default:
			return m.begin(ask.Subject{Question: it.asks})
		}
	default:
		// A number opens an earlier question while nothing is typed; once
		// something is, it is part of what is typed.
		if was := m.earlier(); m.input.Value() == "" && len(k) == 1 && k[0] >= '1' && int(k[0]-'1') < len(was) {
			return m.open(was[k[0]-'1'].Doc), nil
		}
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

// open shows a document on its tab, making the tab if it has none.
func (m Model) open(d markdown.Doc) Model {
	m.home = false
	for i, t := range m.tabs {
		if t.Doc().Title == d.Title && t.Doc().At == d.At {
			m.cur = i
			return m
		}
	}
	m.tabs = append(m.tabs, viewer.New(d).Size(m.w, m.bodyH()))
	m.cur = len(m.tabs) - 1
	return m
}

// read is a key while a document is on screen.
func (m Model) read(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+k":
		m.home, m.pick = true, 0
		m.input.Reset()
		return m, nil
	case "q":
		return m, tea.Quit
	}
	if len(m.tabs) == 0 {
		m.home = true
		return m, nil
	}
	switch msg.String() {
	case "[":
		m.cur = (m.cur + len(m.tabs) - 1) % len(m.tabs)
	case "]":
		m.cur = (m.cur + 1) % len(m.tabs)
	default:
		m.tabs[m.cur] = m.tabs[m.cur].Key(msg)
	}
	return m, nil
}

// keys are the bindings that work now, for the help line.
func (m Model) keys() []key.Binding {
	switch {
	case m.work != nil && m.work.err == nil:
		return []key.Binding{keyStop, keyQuit}
	case m.work != nil:
		return []key.Binding{keyBack, keyQuit}
	case m.home:
		ks := []key.Binding{keyPick}
		if it, _, ok := m.picked(m.items()); ok {
			switch {
			case it.asks != "":
				ks = append(ks, keyAsk)
			case it.more != "":
				ks = append(ks, keyMore)
			case it.commit != nil && m.known(it.commit.Hash):
				ks = append(ks, keyOpen)
			default:
				ks = append(ks, keyGo)
			}
		}
		if n := len(m.earlier()); n > 0 && m.input.Value() == "" {
			ks = append(ks, keyRecent(n))
		}
		if len(m.tabs) == 0 {
			return append(ks, keyLeave)
		}
		return append(ks, keyBack)
	}
	ks := m.tabs[m.cur].Keys()
	if len(m.tabs) > 1 {
		ks = append(ks, keyTab)
	}
	return append(ks, keyHome, keyQuit)
}

func (m Model) known(hash string) bool {
	_, ok := m.explained(hash)
	return ok
}

func (m Model) render() string {
	lines := make([]string, m.h)
	if m.small() {
		lines[min(1, m.h-1)] = "  " + sDim.Render("Make the window larger.")
		return m.page(lines)
	}
	left := strings.Repeat(" ", viewer.Left(m.w))
	if len(m.tabs) > 0 && m.work == nil {
		for i, row := range m.tabline() {
			lines[i] = left + row
		}
	}
	switch {
	case m.work != nil:
		for i, row := range m.work.rows(m.colW(), m.bodyH()) {
			lines[bodyY+i] = left + row
		}
	case m.home:
		for i, row := range m.opening(m.colW(), m.bodyH()) {
			lines[bodyY+i] = left + row
		}
	default:
		copy(lines[bodyY:], m.tabs[m.cur].Rows())
	}
	h := m.help
	h.SetWidth(m.w - 2*viewer.Left(m.w))
	lines[m.h-2] = left + h.ShortHelpView(m.keys())
	return m.page(lines)
}

// Where the line to type on stands on the opening page, under the name
// of the repository.
const inputY = 2

// opening draws the opening page: the repository, the line to type on,
// the history as what is typed leaves it, and under it the questions
// answered before. The line to type on and the questions keep their rows
// whatever the list does.
func (m Model) opening(width, height int) []string {
	out := make([]string, height)
	name := sHead.Render(m.from.Repo)
	if m.from.Branch != "" {
		name += sDim.Render("  on " + m.from.Branch)
	}
	out[0] = name
	in := m.input
	in.SetWidth(width - 3)
	out[inputY] = in.View()

	was := m.earlier()
	under := 0
	if len(was) > 0 {
		under = len(was) + 2
	}
	room := max(1, height-(inputY+2)-under)
	items := m.items()
	_, at, _ := m.picked(items)
	// The list stays where it is until the picked row would leave it,
	// and then shows a row past it, so there is always somewhere to go.
	first := min(max(0, at-room+2), max(0, len(items)-room))
	for i := 0; i < room && first+i < len(items); i++ {
		out[inputY+2+i] = m.line(items[first+i], first+i == at, width)
	}
	if len(items) == 0 {
		out[inputY+2] = sDim.Render("No commits here yet. Type a question and press enter.")
	}

	if len(was) > 0 {
		y := height - len(was) - 1
		label := "earlier questions "
		out[y] = sDim.Render(label) + sRule.Render(strings.Repeat("─", max(0, width-wid(label))))
		for i, a := range was {
			num := sMeta.Render(strconv.Itoa(i+1)) + "  "
			when := sDim.Render(a.When)
			out[y+1+i] = num + sBody.Render(fit(a.Asked, max(0, width-wid(num)-wid(when)-2))) + "  " + when
		}
	}
	return out
}

// kind is the "feat(viewer): " a commit's subject may start with.
var kind = regexp.MustCompile(`^[a-z]+(\([^)]*\))?!?: `)

// line draws one row of the list. A commit is what it says, with its hash
// to the right, quiet: the day it was made is the heading over it. The
// picked row is one fill from edge to edge.
func (m Model) line(it item, picked bool, width int) string {
	mark, markSt, name, right := " ", sDim, "", ""
	switch {
	case it.day != "":
		return sMeta.Render(it.day)
	case it.pending:
		mark, markSt, name = "●", sLive, "Uncommitted changes"
		right = strconv.Itoa(m.from.Uncommitted) + " files"
		if m.from.Uncommitted == 1 {
			right = "1 file"
		}
	case it.more != "":
		name = "+ " + strconv.Itoa(it.hidden) + " more"
	case it.commit != nil:
		if m.known(it.commit.Hash) {
			mark = "✓"
		}
		name, right = it.commit.Subject, it.commit.Hash
	case it.asks != "":
		mark, markSt, name = "?", sLive, "Ask: "+it.asks
	default:
		return ""
	}
	room := max(0, width-2-wid(right)-2)
	if picked {
		return sPick.Render(fit(mark+" "+fit(name, room)+"  "+right, width))
	}
	text := sName.Render(fit(name, room))
	switch {
	case it.more != "":
		text = sDim.Render(fit(name, room))
	case it.commit != nil:
		// What kind of change it is, is quieter than what it did.
		name = fit(name, room)
		pre := kind.FindString(name)
		text = sMeta.Render(pre) + sBody.Render(name[len(pre):])
	}
	return markSt.Render(mark) + " " + text + "  " + sDim.Render(right)
}

// page makes every row exactly as wide as the window.
func (m Model) page(lines []string) string {
	for i := range lines {
		lines[i] = fit(lines[i], m.w)
	}
	return strings.Join(lines, "\n")
}

// tabline is three rows: a tab a document, standing on a rule that the
// open one breaks. On the opening page no tab is the open one.
func (m Model) tabline() []string {
	w := m.w - 2*viewer.Left(m.w)
	top, mid, low := " ", " ", "─"
	for i, t := range m.tabs {
		name, n := sDim.Render(t.Doc().Tab), wid(t.Doc().Tab)+2
		foot := "┴" + strings.Repeat("─", n) + "┴"
		if i == m.cur && !m.home {
			name, foot = sHead.Render(t.Doc().Tab), "┘"+strings.Repeat(" ", n)+"└"
		}
		top += "╭" + strings.Repeat("─", n) + "╮"
		mid += sRule.Render("│") + " " + name + " " + sRule.Render("│")
		low += foot
	}
	if !m.home {
		at := "read at " + m.tabs[m.cur].Doc().At
		if room := w - wid(mid) - wid(at); room >= 4 {
			mid += strings.Repeat(" ", room) + sDim.Render(at)
		}
	}
	low += strings.Repeat("─", max(0, w-wid(low)))
	return []string{sRule.Render(fit(top, w)), fit(mid, w), sRule.Render(fit(low, w))}
}

func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.WindowTitle = "lgtm"
	v.BackgroundColor = pal.Surface0
	if m.w == 0 {
		return v
	}
	if m.covered {
		v.SetContent(m.slop.Over(strings.Split(m.render(), "\n")))
		return v
	}
	v.SetContent(m.render())
	if m.home && m.work == nil && !m.small() {
		if c := m.input.Cursor(); c != nil {
			c.X, c.Y = c.X+viewer.Left(m.w), c.Y+bodyY+inputY
			v.Cursor = c
		}
	}
	return v
}
