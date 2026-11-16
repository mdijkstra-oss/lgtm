package text

import "testing"

func TestExpandMovesRanges(t *testing.T) {
	s, r := Expand("\tab\tcd", [][2]int{{1, 3}, {4, 6}})
	if s != "    ab  cd" {
		t.Fatalf("expanded = %q", s)
	}
	if s[r[0][0]:r[0][1]] != "ab" || s[r[1][0]:r[1][1]] != "cd" {
		t.Errorf("ranges = %v", r)
	}
}

func TestFit(t *testing.T) {
	s := "the quick brown fox jumps over the lazy dog"

	if a, b, l, r := Fit(s, [2]int{0, 0}, 80); a != 0 || b != len(s) || l || r {
		t.Errorf("short line was cut: %d %d %v %v", a, b, l, r)
	}

	a, b, l, r := Fit(s, [2]int{4, 9}, 12)
	if a != 0 || !r || l || Width(s[a:b]) != 11 {
		t.Errorf("early match: %q %v %v", s[a:b], l, r)
	}

	keep := [2]int{35, 39}
	a, b, l, r = Fit(s, keep, 12)
	if !l || a > keep[0] || b < keep[1] {
		t.Errorf("late match not visible: %q %v %v", s[a:b], l, r)
	}
	cells := Width(s[a:b])
	if l {
		cells++
	}
	if r {
		cells++
	}
	if cells > 12 {
		t.Errorf("window is %d cells, want at most 12", cells)
	}
}

func TestHeadAndTail(t *testing.T) {
	if got := Head("abcdefgh", 5); got != "abcd…" {
		t.Errorf("Head = %q", got)
	}
	if got := Tail("abcdefgh", 5); got != "…efgh" {
		t.Errorf("Tail = %q", got)
	}
}
