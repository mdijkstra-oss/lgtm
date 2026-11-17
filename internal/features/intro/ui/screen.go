package ui

// cell is one character of the screen with its colours.
type cell struct {
	ch     string
	fg, bg rgb
	// "1" bold, "3" italic; "" plain.
	attr string
}

type screen struct {
	w, h  int
	cells []cell
}

func (s *screen) set(x, y int, c cell) {
	if x >= 0 && y >= 0 && x < s.w && y < s.h {
		s.cells[y*s.w+x] = c
	}
}

// say writes text from (x, y) and returns where it ended.
func (s *screen) say(x, y int, text string, fg rgb) int {
	for _, r := range text {
		s.set(x, y, cell{string(r), fg, ground, ""})
		x++
	}
	return x
}

// app draws what the blade uncovers: nothing, and at the bottom the keys.
func app(w, h int) *screen {
	s := &screen{w: w, h: h, cells: make([]cell, w*h)}
	for i := range s.cells {
		s.cells[i] = cell{" ", ground, ground, ""}
	}
	x := 4
	for i, k := range [][2]string{{"space", "again"}, {"q", "quit"}} {
		x = s.say(x, h-1, k[0], calm)
		x = s.say(x+1, h-1, k[1], hush)
		if i == 0 {
			x = s.say(x+1, h-1, "•", rule) + 1
		}
	}
	return s
}
