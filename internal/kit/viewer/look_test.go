package viewer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestColoursAreShown(t *testing.T) {
	// The same colour three ways, and three that are not colours.
	rows := typeset("`#00f` `rgb(0, 0, 255)` `hsl(240, 100%, 50%)` `#12` `rgb(1, 2)` #0000ff", 100)
	row := strings.Join(rows, "")
	if got := strings.Count(ansi.Strip(row), "█"); got != 3 {
		t.Fatalf("%d colours shown, not 3: %q", got, ansi.Strip(row))
	}
	// Each is drawn the way the first is.
	first := row[strings.LastIndex(row[:strings.Index(row, "█")], "\x1b"):strings.Index(row, "█")]
	if got := strings.Count(row, first+"█"); got != 3 {
		t.Errorf("%d of 3 came to the colour %q", got, first)
	}
	// A fence is code, whatever it has in it.
	if code := strings.Join(typeset("```css\na { color: `#6b50ff` }\n```", 80), ""); strings.Contains(code, "█") {
		t.Error("a colour in a fence was shown")
	}
}
