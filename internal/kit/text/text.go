// Package text measures and cuts plain strings by terminal cells.
package text

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

const tabStop = 4

func Width(s string) int { return ansi.StringWidth(s) }

// Expand turns tabs into spaces and moves the byte ranges along.
func Expand(s string, ranges [][2]int) (string, [][2]int) {
	if !strings.Contains(s, "\t") {
		return s, ranges
	}
	var b strings.Builder
	pos := make([]int, len(s)+1)
	col := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		for k := 0; k < size; k++ {
			pos[i+k] = b.Len()
		}
		if r == '\t' {
			n := tabStop - col%tabStop
			b.WriteString(strings.Repeat(" ", n))
			col += n
		} else {
			b.WriteString(s[i : i+size])
			col += Width(s[i : i+size])
		}
		i += size
	}
	pos[len(s)] = b.Len()
	out := make([][2]int, len(ranges))
	for i, r := range ranges {
		out[i] = [2]int{pos[r[0]], pos[r[1]]}
	}
	return b.String(), out
}

// Fit picks the byte window [a, b) of s to show in width cells so that keep
// stays visible. A cut side takes one cell for an ellipsis the caller draws.
func Fit(s string, keep [2]int, width int) (a, b int, left, right bool) {
	if width <= 0 {
		return 0, 0, false, false
	}
	if Width(s) <= width {
		return 0, len(s), false, false
	}
	end := advance(s, 0, width-1)
	if keep[1] <= end || width < 4 {
		return 0, end, false, true
	}
	budget := width - 2
	a = retreat(s, keep[0], budget/3)
	b = advance(s, a, budget)
	if b >= len(s) {
		a = retreat(s, len(s), budget+1)
		return a, len(s), true, false
	}
	return a, b, true, true
}

// Head keeps the start of s; Tail keeps the end.
func Head(s string, width int) string {
	if Width(s) <= width {
		return s
	}
	if width <= 0 {
		return ""
	}
	return s[:advance(s, 0, width-1)] + "…"
}

func Tail(s string, width int) string {
	if Width(s) <= width {
		return s
	}
	if width <= 0 {
		return ""
	}
	return "…" + s[retreat(s, len(s), width-1):]
}

func advance(s string, start, cells int) int {
	i, used := start, 0
	for i < len(s) {
		_, size := utf8.DecodeRuneInString(s[i:])
		w := Width(s[i : i+size])
		if used+w > cells {
			break
		}
		used += w
		i += size
	}
	return i
}

func retreat(s string, start, cells int) int {
	i, used := start, 0
	for i > 0 {
		_, size := utf8.DecodeLastRuneInString(s[:i])
		w := Width(s[i-size : i])
		if used+w > cells {
			break
		}
		used += w
		i -= size
	}
	return i
}
