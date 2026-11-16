// Package hints draws key hints: which key, and what it does.
package hints

import (
	"strings"

	"strigil/internal/kit"
	"strigil/internal/kit/theme"
)

// Render joins the hints on pen's background with sep between them.
func Render(th theme.Theme, pen theme.Pen, hs []kit.Hint, sep string) string {
	parts := make([]string, len(hs))
	for i, h := range hs {
		parts[i] = pen.Ink(th.Soft, h.Key) + pen.Ink(th.Dim, " "+h.Label)
	}
	return strings.Join(parts, pen.Ink(th.Faint, sep))
}
