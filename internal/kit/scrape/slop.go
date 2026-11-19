package scrape

// What covers the page: code of the kind that piles up.
var junk = []string{
	"utils2_final_v3", "helper", "manager", "wrapper", "AbstractFactoryFactory", "// increment i", "TODO", "fixme",
	"handleStuff()", "doTheThing()", "data2", "temp_new_copy", "processData", "misc", "common", "base", "impl",
	"if err != nil { return nil }", "catch (e) {}", "console.log('here')", "// TODO: refactor", "any", "obj",
	"var result2 = result", "legacy_old_backup", "// this should never happen", "sleep(100)", "isValidValid",
	"getManagerFactoryHelper()", "newNewHandler", "flag2", "// temporary fix", "return true // ?", "x1",
}

// junkFaint says, cell for cell, which of the junk sits further back.
var junkRows, junkFaint = func() ([][]rune, [][]bool) {
	rows, faint := make([][]rune, 23), make([][]bool, 23)
	for i := range rows {
		for n := 0; len(rows[i]) < 400; n++ {
			word := []rune(junk[(i*5+n*3)%len(junk)] + "  ")
			rows[i] = append(rows[i], word...)
			for range word {
				faint[i] = append(faint[i], (i+n)%3 != 0)
			}
		}
	}
	return rows, faint
}()

// The word stamped over the junk, a letter five dots wide and seven high.
// It is what was said of all that code.
var word = [][7]string{
	{"X....", "X....", "X....", "X....", "X....", "X....", "XXXXX"},
	{".XXXX", "X....", "X....", "X.XXX", "X...X", "X...X", ".XXXX"},
	{"XXXXX", "..X..", "..X..", "..X..", "..X..", "..X..", "..X.."},
	{"X...X", "XX.XX", "X.X.X", "X...X", "X...X", "X...X", "X...X"},
}

// Dots between two letters.
const between = 2

// run is a stretch of one row that the marker went over: part of a letter.
type run struct {
	row, x, n  int
	from, span int // where the word starts and how wide it is, for its colour
}

// stamp lays the word over a page of w by h, in the middle, as large as
// fits: a dot is a few cells wide and a row or two high, since a cell is
// twice as high as it is wide. On a page too small for it there is none.
func stamp(w, h int) []run {
	dots := len(word)*5 + (len(word)-1)*between
	for _, size := range [][2]int{{4, 2}, {3, 2}, {2, 1}} {
		wide, high := dots*size[0], 7*size[1]
		if wide > w-6 || high > h-5 {
			continue
		}
		x0, y0 := (w-wide)/2, (h-1-high)/2
		var out []run
		for l, letter := range word {
			for y, dotted := range letter {
				for x := 0; x < 5; x++ {
					if dotted[x] != 'X' {
						continue
					}
					n := 1
					for x+n < 5 && dotted[x+n] == 'X' {
						n++
					}
					for r := range size[1] {
						out = append(out, run{
							row: y0 + y*size[1] + r, x: x0 + (l*(5+between)+x)*size[0], n: n * size[0],
							from: x0, span: wide,
						})
					}
					x += n - 1
				}
			}
		}
		return out
	}
	return nil
}

// slopAt is what covers a cell: junk, and whether the marker went over it
// there. faint says it is junk from further back.
func slopAt(runs []run, x, row int) (ch rune, mark *run, faint bool) {
	r := junkRows[row%len(junkRows)]
	n := (x + row*7) % len(r)
	for i := range runs {
		if s := &runs[i]; s.row == row && x >= s.x && x < s.x+s.n {
			return r[n], s, false
		}
	}
	return r[n], nil, junkFaint[row%len(junkRows)][n]
}
