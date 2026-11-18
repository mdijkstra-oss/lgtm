package viewer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"lgtm/internal/kit/markdown"
)

func TestHeadingHasItsSectionsMark(t *testing.T) {
	d, err := markdown.Parse("# T\n\n## Plain\n\nText.\n\n## Cases\n\n### ✓ one\n\nHolds.\n\n### ✗ two\n\nBreaks.\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	page := ansi.Strip(strings.Join(New(d).Size(100, 30).Rows(), "\n"))
	if !strings.Contains(page, "## Cases ×") || strings.Contains(page, "## Plain ×") {
		t.Errorf("headings drawn as:\n%s", page)
	}
}

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

func TestCodeAmongWordsHasARail(t *testing.T) {
	rows := typeset("Words.\n\n    ### not a heading\n    is := 1\n\nMore words.", 60)
	// Rows are padded; only the shape is compared.
	var got []string
	for _, r := range rows {
		got = append(got, strings.TrimRight(ansi.Strip(r), " "))
	}
	if want := "Words.\n\n│ ### not a heading\n│ is := 1\n\nMore words."; strings.Join(got, "\n") != want {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), want)
	}
	// Under a list item it is more of the item.
	if list := ansi.Strip(strings.Join(typeset("- one\n\n    more of one", 60), "\n")); strings.Contains(list, "│") {
		t.Errorf("a list's own lines were taken for code:\n%s", list)
	}
	// A colour is not parted from its cell by the end of a line.
	for _, r := range typeset("It can be one or another, even `rgb(104, 255, 214)`.", 34) {
		if strings.HasSuffix(strings.TrimRight(ansi.Strip(r), " "), "█") {
			t.Errorf("a cell was left at the end of a line: %q", ansi.Strip(r))
		}
	}
}
