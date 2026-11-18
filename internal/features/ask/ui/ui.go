// Package ui is where a document is opened and read: a palette to open one
// by its title, and a tab per document open. The document itself is drawn
// by kit/viewer.
package ui

import (
	"context"
	"os"
	"slices"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"strigil/internal/features/ask"
	"strigil/internal/kit/markdown"
	"strigil/internal/kit/theme"
	"strigil/internal/kit/viewer"
)

var (
	pal theme.Theme

	sHead, sName, sBody, sMeta, sDim, sRule, sPick, sPrompt lipgloss.Style
)

func init() { use(theme.Dark) }

// use draws everything from here on in that theme, the documents too.
func use(t theme.Theme) {
	pal = t
	viewer.Use(t)
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

type Model struct {
	w, h  int
	known []markdown.Doc
	tabs  []viewer.Model
	cur   int

	choosing bool
	input    textinput.Model
	pick     int
	help     help.Model
}

const placeholder = "Open a document by its title"

// New opens on the palette, with nothing open yet.
func New(known []markdown.Doc) Model {
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
	return Model{known: known, choosing: true, input: in, help: h}
}

func Run(ctx context.Context, known []markdown.Doc) error {
	use(theme.For(lipgloss.HasDarkBackground(os.Stdin, os.Stdout)))
	_, err := tea.NewProgram(New(known), tea.WithContext(ctx)).Run()
	return err
}

func (m Model) Init() tea.Cmd { return nil }

var (
	keyTab   = key.NewBinding(key.WithKeys("[", "]"), key.WithHelp("[/]", "tab"))
	keyDocs  = key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "documents"))
	keyQuit  = key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))
	keyPick  = key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "choose"))
	keyOpen  = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open"))
	keyClose = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close"))
)

// The tabs take three rows and a blank one above the document; the help
// line and the blank rows round it take three below.
const bodyY = 4

func (m Model) bodyH() int  { return max(1, m.h-7) }
func (m Model) small() bool { return m.w < 60 || m.h < 16 }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.tabs = slices.Clone(m.tabs)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		for i := range m.tabs {
			m.tabs[i] = m.tabs[i].Size(m.w, m.bodyH())
		}
		return m, nil
	case tea.KeyPressMsg:
		switch {
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		case m.choosing:
			return m.choose(msg)
		}
		return m.read(msg)
	}
	return m, nil
}

// choose is a key while the palette is open.
func (m Model) choose(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	found := ask.Find(m.known, m.input.Value())
	switch msg.String() {
	case "esc", "ctrl+k":
		m.choosing = false
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

// open shows a document on its tab, making the tab if it has none.
func (m Model) open(d markdown.Doc) Model {
	m.choosing = false
	for i, t := range m.tabs {
		if t.Doc().Title == d.Title {
			m.cur = i
			return m
		}
	}
	m.tabs = append(m.tabs, viewer.New(d).Size(m.w, m.bodyH()))
	m.cur = len(m.tabs) - 1
	return m
}

// read is a key while a document, or nothing, is on screen.
func (m Model) read(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+k":
		m.choosing, m.pick = true, 0
		m.input.Reset()
		return m, nil
	case "q":
		return m, tea.Quit
	}
	if len(m.tabs) == 0 {
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
	if m.choosing {
		return []key.Binding{keyPick, keyOpen, keyClose}
	}
	if len(m.tabs) == 0 {
		return []key.Binding{keyDocs, keyQuit}
	}
	ks := m.tabs[m.cur].Keys()
	if len(m.tabs) > 1 {
		ks = append(ks, keyTab)
	}
	return append(ks, keyDocs, keyQuit)
}

func (m Model) render() string {
	lines := make([]string, m.h)
	if m.small() {
		lines[min(1, m.h-1)] = "  " + sDim.Render("Make the window larger.")
		return m.page(lines)
	}
	left := strings.Repeat(" ", viewer.Left(m.w))
	if len(m.tabs) == 0 {
		lines[bodyY] = left + sDim.Render("Nothing open yet.")
	} else {
		for i, row := range m.tabline() {
			lines[i] = left + row
		}
		copy(lines[bodyY:], m.tabs[m.cur].Rows())
	}
	h := m.help
	h.SetWidth(m.w - 2*viewer.Left(m.w))
	lines[m.h-2] = left + h.ShortHelpView(m.keys())
	page := m.page(lines)
	if !m.choosing {
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

// page makes every row exactly as wide as the window.
func (m Model) page(lines []string) string {
	for i := range lines {
		lines[i] = fit(lines[i], m.w)
	}
	return strings.Join(lines, "\n")
}

// tabline is three rows: a tab a document, standing on a rule that the
// open one breaks.
func (m Model) tabline() []string {
	w := m.w - 2*viewer.Left(m.w)
	top, mid, low := " ", " ", "─"
	for i, t := range m.tabs {
		name, n := sDim.Render(t.Doc().Tab), wid(t.Doc().Tab)+2
		foot := "┴" + strings.Repeat("─", n) + "┴"
		if i == m.cur {
			name, foot = sHead.Render(t.Doc().Tab), "┘"+strings.Repeat(" ", n)+"└"
		}
		top += "╭" + strings.Repeat("─", n) + "╮"
		mid += sRule.Render("│") + " " + name + " " + sRule.Render("│")
		low += foot
	}
	at := "read at " + m.tabs[m.cur].Doc().At
	if room := w - wid(mid) - wid(at); room >= 4 {
		mid += strings.Repeat(" ", room) + sDim.Render(at)
	}
	low += strings.Repeat("─", max(0, w-wid(low)))
	return []string{sRule.Render(fit(top, w)), fit(mid, w), sRule.Render(fit(low, w))}
}

// palette is the box to open a document from, and where it goes.
func (m Model) palette() (box string, x, y int) {
	inner := min(72, m.w-8) - 6
	in := m.input
	in.SetWidth(inner - 3)
	rows := []string{fit(in.View(), inner), ""}
	found := ask.Find(m.known, in.Value())
	for i := range m.known {
		switch {
		case i >= len(found) && i == 0:
			rows = append(rows, sDim.Render("  No document by that title."))
		case i >= len(found):
			rows = append(rows, "")
		case i == m.pick:
			rows = append(rows, sPick.Render(fit("  "+found[i].Title, inner)))
		default:
			rows = append(rows, sBody.Render(fit("  "+found[i].Title, inner)))
		}
	}
	box = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(pal.Primary).
		Padding(1, 2).Render(strings.Join(rows, "\n"))
	return box, (m.w - wid(box)) / 2, max(2, (m.h-lipgloss.Height(box))/3)
}

func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.WindowTitle = "strigil"
	v.BackgroundColor = pal.Surface0
	if m.w == 0 {
		return v
	}
	v.SetContent(m.render())
	if m.choosing && !m.small() {
		if c := m.input.Cursor(); c != nil {
			_, x, y := m.palette()
			c.X, c.Y = c.X+x+3, c.Y+y+2
			v.Cursor = c
		}
	}
	return v
}
