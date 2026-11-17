package ui

import (
	_ "embed"
	"math/rand/v2"
	"strings"
)

// The sentences nobody needed, one a line; lines starting with # name the
// kind that follows and are skipped.
//
//go:embed slop.txt
var slopFile string

var sentences = func() [][]rune {
	var out [][]rune
	for _, l := range strings.Split(slopFile, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, []rune(l))
		}
	}
	return out
}()

// What lies under the sentences: code of the kind that piles up.
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

// said is one sentence laid over the junk.
type said struct {
	row, x int
	text   []rune
	hue    int // which marker went over it
}

// scatter picks sentences for a screen of w by h and gives each a row of
// its own, somewhere along it.
func scatter(rng *rand.Rand, w, h int) []said {
	var out []said
	if w < 20 || h < 4 {
		return out
	}
	rows := rng.Perm(h - 1)
	for _, i := range rng.Perm(len(sentences)) {
		if len(out) >= (h-1)*2/5 {
			break
		}
		s := sentences[i]
		if len(s) > w-4 {
			continue
		}
		out = append(out, said{row: rows[len(out)], x: 1 + rng.IntN(w-len(s)-2), text: s, hue: len(out)})
	}
	return out
}

// slopAt is what covers a cell: a letter of a sentence, the marker's
// overrun either side of one, or junk. mark is the sentence the marker went
// over, if one did; faint says it is junk from further back.
func slopAt(lines []said, x, row int) (ch rune, mark *said, faint bool) {
	for i := range lines {
		s := &lines[i]
		if s.row != row || x < s.x-1 || x > s.x+len(s.text) {
			continue
		}
		if x == s.x-1 || x == s.x+len(s.text) {
			return ' ', s, false
		}
		return s.text[x-s.x], s, false
	}
	r := junkRows[row%len(junkRows)]
	n := (x + row*7) % len(r)
	return r[n], nil, junkFaint[row%len(junkRows)][n]
}
