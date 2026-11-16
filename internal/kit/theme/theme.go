// Package theme is the single place colours are decided.
package theme

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/x/ansi"
)

// Theme names colours by the job they do.
type Theme struct {
	Bg    color.Color // page
	Panel color.Color // sidebar, palette
	Deep  color.Color // text drawn on a bright fill
	Band  color.Color // selected row, current line
	Rule  color.Color
	Faint color.Color
	Dim   color.Color
	Soft  color.Color
	Text  color.Color

	Accent color.Color // focus
	Lit    color.Color // hover target, current match
	Match  color.Color
}

func Mocha() Theme { return from(catppuccin.Mocha) }

func from(f catppuccin.Flavor) Theme {
	return Theme{
		Bg:     f.Base(),
		Panel:  f.Mantle(),
		Deep:   f.Crust(),
		Band:   f.Surface0(),
		Rule:   f.Surface1(),
		Faint:  f.Surface2(),
		Dim:    f.Overlay1(),
		Soft:   f.Subtext0(),
		Text:   f.Text(),
		Accent: f.Mauve(),
		Lit:    f.Peach(),
		Match:  f.Yellow(),
	}
}

// Pen writes on one background. Every cell it produces carries that
// background, so joined pieces never show the terminal's own colour.
type Pen struct{ bg color.Color }

func On(bg color.Color) Pen { return Pen{bg: bg} }

func (p Pen) Style(fg color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(fg).Background(p.bg)
}

func (p Pen) Ink(fg color.Color, s string) string { return p.Style(fg).Render(s) }

func (p Pen) Gap(n int) string {
	if n <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Background(p.bg).Render(strings.Repeat(" ", n))
}

// Line makes s exactly w cells wide.
func (p Pen) Line(s string, w int) string {
	n := lipgloss.Width(s)
	if n > w {
		return ansi.Truncate(s, w, "")
	}
	return s + p.Gap(w-n)
}

// Split puts left and right at the two ends of a w-cell line; right is
// dropped when both do not fit.
func (p Pen) Split(left, right string, w int) string {
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return p.Line(left, w)
	}
	return left + p.Gap(gap) + right
}

func (p Pen) Center(s string, w int) string {
	gap := w - lipgloss.Width(s)
	if gap <= 0 {
		return p.Line(s, w)
	}
	return p.Gap(gap/2) + s + p.Gap(gap-gap/2)
}

// Block makes rows exactly h lines of w cells.
func (p Pen) Block(rows []string, w, h int) []string {
	out := make([]string, h)
	for i := range out {
		if i < len(rows) {
			out[i] = p.Line(rows[i], w)
		} else {
			out[i] = p.Gap(w)
		}
	}
	return out
}

// Middle centres rows vertically in h lines of w cells.
func (p Pen) Middle(rows []string, w, h int) []string {
	top := max(0, (h-len(rows))/2)
	out := make([]string, 0, h)
	for range top {
		out = append(out, "")
	}
	for _, r := range rows {
		out = append(out, p.Center(r, w))
	}
	return p.Block(out, w, h)
}
