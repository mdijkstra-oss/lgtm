package viewer

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"charm.land/glamour/v2"
	gansi "charm.land/glamour/v2/ansi"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"strigil/internal/kit/markdown"
	"strigil/internal/kit/theme"
)

// What each tier is for is in docs/ux/README.md.
var (
	sHead  = lipgloss.NewStyle().Foreground(theme.Text1).Bold(true)
	sName  = lipgloss.NewStyle().Foreground(theme.Text1)
	sBody  = lipgloss.NewStyle().Foreground(theme.Text2)
	sMeta  = lipgloss.NewStyle().Foreground(theme.Text3)
	sDim   = lipgloss.NewStyle().Foreground(theme.Text4)
	sRule  = lipgloss.NewStyle().Foreground(theme.Line)
	sOK    = lipgloss.NewStyle().Foreground(theme.Success)
	sBad   = lipgloss.NewStyle().Foreground(theme.Error)
	sBlock = lipgloss.NewStyle().Foreground(theme.Surface3)  // the outline of a block the cursor is not on
	sHere  = lipgloss.NewStyle().Foreground(theme.Primary)   // the cursor's bar, the edge of what floats
	sLive  = lipgloss.NewStyle().Foreground(theme.Secondary) // the live point inside a block
	sPick  = lipgloss.NewStyle().Background(theme.Primary).Foreground(theme.OnPrimary).Bold(true)
	sAsk   = lipgloss.NewStyle().Foreground(theme.Accent)
)

func wid(s string) int { return lipgloss.Width(s) }

// fit makes a row exactly w cells: cut with … or padded.
func fit(s string, w int) string {
	s = ansi.Truncate(s, w, "…")
	return s + strings.Repeat(" ", max(0, w-wid(s)))
}

// seg is a piece of a row and the column it starts in.
type seg struct {
	x  int
	s  string
	st lipgloss.Style
}

// placed lays pieces out on one row, left to right.
func placed(segs ...seg) string {
	var b strings.Builder
	x := 0
	for _, s := range segs {
		if s.x > x {
			b.WriteString(strings.Repeat(" ", s.x-x))
			x = s.x
		}
		b.WriteString(s.st.Render(s.s))
		x += wid(s.s)
	}
	return b.String()
}

func sign(holds bool) string {
	if holds {
		return sOK.Render("✓")
	}
	return sBad.Render("×")
}

func hex(c color.Color) *string {
	r, g, b, _ := c.RGBA()
	s := fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
	return &s
}

// sheet is Glamour's stylesheet, built from the theme.
func sheet() gansi.StyleConfig {
	no, yes := uint(0), true
	ink := func(c color.Color) gansi.StylePrimitive { return gansi.StylePrimitive{Color: hex(c)} }
	strong := gansi.StylePrimitive{Color: hex(theme.Text1), Bold: &yes}
	return gansi.StyleConfig{
		Document: gansi.StyleBlock{StylePrimitive: ink(theme.Text2), Margin: &no},
		// The title is the one chip on the page; headings are told apart by
		// their marks, as Glamour's own styles do.
		Heading: gansi.StyleBlock{StylePrimitive: gansi.StylePrimitive{Color: hex(theme.Info), Bold: &yes}},
		H1: gansi.StyleBlock{StylePrimitive: gansi.StylePrimitive{
			Color: hex(theme.OnPrimary), BackgroundColor: hex(theme.Primary), Bold: &yes, Prefix: " ", Suffix: " ",
		}},
		H2:     gansi.StyleBlock{StylePrimitive: gansi.StylePrimitive{Prefix: "## "}},
		H3:     gansi.StyleBlock{StylePrimitive: gansi.StylePrimitive{Prefix: "### "}},
		H4:     gansi.StyleBlock{StylePrimitive: gansi.StylePrimitive{Prefix: "#### "}},
		Strong: strong,
		Emph:   gansi.StylePrimitive{Color: hex(theme.Text1)},
		// No surface under inline code: a surface means the cursor can go there.
		Code: gansi.StyleBlock{StylePrimitive: ink(theme.Text1)},
		List: gansi.StyleList{LevelIndent: 2},
		Item: gansi.StylePrimitive{BlockPrefix: "• "},
		CodeBlock: gansi.StyleCodeBlock{
			StyleBlock: gansi.StyleBlock{Margin: &no},
			// Chroma says what each word of a language is; these say what
			// colour each kind of word gets, whatever the language.
			Chroma: &gansi.Chroma{
				Text:                ink(theme.Text2),
				Name:                ink(theme.Text2),
				Comment:             ink(theme.Text4),
				CommentPreproc:      ink(theme.Text4),
				Punctuation:         ink(theme.Text3),
				Operator:            ink(theme.Text3),
				Keyword:             ink(theme.CodeKeyword),
				KeywordReserved:     ink(theme.CodeKeyword),
				KeywordNamespace:    ink(theme.CodeKeyword),
				KeywordType:         ink(theme.CodeType),
				NameClass:           ink(theme.CodeType),
				NameTag:             ink(theme.CodeKeyword),
				NameAttribute:       ink(theme.CodeFunction),
				NameFunction:        ink(theme.CodeFunction),
				NameBuiltin:         ink(theme.CodeFunction),
				NameConstant:        ink(theme.CodeNumber),
				Literal:             ink(theme.CodeNumber),
				LiteralNumber:       ink(theme.CodeNumber),
				LiteralString:       ink(theme.CodeString),
				LiteralStringEscape: ink(theme.CodeNumber),
			},
		},
	}
}

var (
	renderers = map[string]*glamour.TermRenderer{}
	rendered  = map[string][]string{}
	shaded    = map[*markdown.Code][]string{}
)

// markdown renders text w cells wide and gives its rows.
func typeset(text string, w int) []string {
	which := strconv.Itoa(w)
	key := which + "\x00" + text
	if rows, ok := rendered[key]; ok {
		return rows
	}
	r := renderers[which]
	if r == nil {
		r, _ = glamour.NewTermRenderer(glamour.WithStyles(sheet()), glamour.WithWordWrap(w), glamour.WithChromaFormatter("terminal16m"))
		renderers[which] = r
	}
	out, err := r.Render(text)
	if err != nil {
		out = text
	}
	rows := strings.Split(strings.Trim(out, "\n"), "\n")
	rendered[key] = rows
	return rows
}

// shade gives a file's lines with their syntax told apart, one row a line.
func shade(c *markdown.Code) []string {
	if rows, ok := shaded[c]; ok {
		return rows
	}
	rows := typeset("```"+c.Lang+"\n"+strings.Join(c.Lines, "\n")+"\n```", 400)
	// Glamour sets a code block off with a blank row either side. They come
	// off evenly: the code's own first or last line may be blank too.
	if extra := len(rows) - len(c.Lines); extra > 0 && extra%2 == 0 {
		rows = rows[extra/2 : len(rows)-extra/2]
	}
	if len(rows) != len(c.Lines) {
		rows = make([]string, len(c.Lines))
		for i, l := range c.Lines {
			rows[i] = sBody.Render(l)
		}
	}
	shaded[c] = rows
	return rows
}

func dent(s string) int { return len(s) - len(strings.TrimLeft(s, " ")) }

// codeTitle names the lines shown: the file and which lines, or the command.
func codeTitle(c *markdown.Code, from, to int) string {
	if c.Line == 0 {
		return sMeta.Render(c.File)
	}
	return sMeta.Render(c.File + ":" + strconv.Itoa(from) + "-" + strconv.Itoa(to))
}

// codeRows gives lines from to to of a file, numbered, the ones that
// matter marked, shown from the left edge. For output it gives every line.
func codeRows(c *markdown.Code, from, to int) []string {
	var rows []string
	if c.Line == 0 {
		for _, l := range c.Lines {
			rows = append(rows, sBody.Render(l))
		}
		return rows
	}
	lit, left := shade(c), 1<<30
	for no := from; no <= to; no++ {
		if l := c.Lines[no-c.Line]; strings.TrimSpace(l) != "" {
			left = min(left, dent(l))
		}
	}
	numW := wid(strconv.Itoa(to))
	for no := from; no <= to; no++ {
		raw := c.Lines[no-c.Line]
		// Glamour pads each row out to its wrap width; the cut drops that too.
		text := ansi.Cut(lit[no-c.Line], min(left, wid(raw)), wid(raw))
		n, bar := sDim, " "
		if no >= c.Hot[0] && no <= c.Hot[1] {
			n, bar = sName, sLive.Render("▌")
		}
		num := strconv.Itoa(no)
		rows = append(rows, strings.Repeat(" ", numW-wid(num))+n.Render(num)+" "+bar+" "+text)
	}
	return rows
}

// panel sets rows behind a quiet rule, as something a block is showing,
// and keeps it at least h rows tall.
func panel(rows []string, h int) []string {
	out := make([]string, 0, h)
	for i := range max(h, len(rows)) {
		row := ""
		if i < len(rows) {
			row = rows[i]
		}
		out = append(out, sDim.Render("│")+" "+row)
	}
	return out
}

// How far along a thing the cursor can go to is.
type reach int

const (
	idle   reach = iota // it can be gone to
	chosen              // it is the one picked in its block, and the cursor is elsewhere
	live                // the cursor is on it
)

// button is how anything the cursor can go to is drawn: its own surface,
// lifted when it is the one picked, filled when the cursor is on it. mark
// is "", "✓" or "×".
func button(mark, label string, r reach) string {
	if mark != "" {
		mark += " "
	}
	switch r {
	case live:
		return sPick.Render(" " + mark + label + " ")
	case chosen:
		back := lipgloss.NewStyle().Background(theme.Surface3)
		return back.Render(" ") + hue(mark).Background(theme.Surface3).Render(mark) +
			back.Foreground(theme.Text1).Bold(true).Render(label+" ")
	}
	back := lipgloss.NewStyle().Background(theme.Surface1)
	return back.Render(" ") + hue(mark).Background(theme.Surface1).Render(mark) +
		back.Foreground(theme.Text2).Render(label+" ")
}

func hue(mark string) lipgloss.Style {
	switch strings.TrimSpace(mark) {
	case "✓":
		return sOK
	case "×":
		return sBad
	}
	return sMeta
}

// at says how far along item i is, in a block whose pick is sel.
func at(i, sel int, on bool) reach {
	switch {
	case i == sel && on:
		return live
	case i == sel:
		return chosen
	}
	return idle
}

func tick(holds bool) string {
	if holds {
		return "✓"
	}
	return "×"
}

// band is a heading with the cursor on it: n cells wide, on the cursor's
// colour fading out to the page's.
func band(text string, n int) string {
	ramp := lipgloss.Blend1D(n, theme.Primary, theme.Surface0)
	var b strings.Builder
	for i, r := range []rune(fit(text, n)) {
		b.WriteString(lipgloss.NewStyle().Background(ramp[min(i, len(ramp)-1)]).Foreground(theme.OnPrimary).Bold(true).Render(string(r)))
	}
	return b.String()
}
