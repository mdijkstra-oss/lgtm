// Package ask is what an answer to a question about a codebase is made of,
// and the answers there are.
package ask

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Answer is a notebook: a lead that is always shown, and sections that
// open. Both are made of cells. How one is written is in Parse.
type Answer struct {
	Question string
	Tab      string // its name on a tab
	At       string // the commit it was read at
	Lead     []Cell
	Sections []Section
}

// Section is what stands under one heading.
type Section struct {
	Title string
	Cells []Cell
}

type Mark int

const (
	None Mark = iota
	Holds
	Breaks
)

// Verdict is what the section's checks and runs come to, and the claim to
// show for it: the first check that breaks, or else the first check.
func (s Section) Verdict() (Mark, string) {
	mark, claim := None, ""
	note := func(broke bool, said string) {
		switch {
		case broke && mark != Breaks:
			mark, claim = Breaks, said
		case mark == None:
			mark, claim = Holds, said
		case claim == "" && mark == Holds && !broke:
			claim = said
		}
	}
	for _, c := range s.Cells {
		switch {
		case c.Check != nil:
			note(c.Check.Held() < len(c.Check.Cases), c.Check.Claim)
		case c.Run != nil:
			failed := 0
			for _, cmd := range c.Run.Cmds {
				if !cmd.OK {
					failed++
				}
			}
			note(failed > 0, "")
		}
	}
	return mark, claim
}

// Cell is one thing in a notebook: markdown, or a block. Exactly one field
// is set.
type Cell struct {
	Prose   string
	Flow    *Flow
	Check   *Check
	Colours *Colours
	Run     *Run
	Change  *Change
}

// A block is written as JSON with its kind beside its own fields.
func (c *Cell) UnmarshalJSON(b []byte) error {
	var k struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(b, &k); err != nil {
		return err
	}
	switch k.Kind {
	case "flow":
		c.Flow = new(Flow)
		return json.Unmarshal(b, c.Flow)
	case "check":
		c.Check = new(Check)
		return json.Unmarshal(b, c.Check)
	case "colours":
		c.Colours = new(Colours)
		return json.Unmarshal(b, c.Colours)
	case "run":
		c.Run = new(Run)
		return json.Unmarshal(b, c.Run)
	case "change":
		c.Change = new(Change)
		return json.Unmarshal(b, c.Change)
	}
	return fmt.Errorf("no block of kind %q", k.Kind)
}

func (c Cell) MarshalJSON() ([]byte, error) {
	type kind struct {
		Kind string `json:"kind"`
	}
	switch {
	case c.Flow != nil:
		return json.Marshal(struct {
			kind
			*Flow
		}{kind{"flow"}, c.Flow})
	case c.Check != nil:
		return json.Marshal(struct {
			kind
			*Check
		}{kind{"check"}, c.Check})
	case c.Colours != nil:
		return json.Marshal(struct {
			kind
			*Colours
		}{kind{"colours"}, c.Colours})
	case c.Run != nil:
		return json.Marshal(struct {
			kind
			*Run
		}{kind{"run"}, c.Run})
	case c.Change != nil:
		return json.Marshal(struct {
			kind
			*Change
		}{kind{"change"}, c.Change})
	}
	return nil, fmt.Errorf("a cell of text is not a block")
}

// Flow is the way through the code from one place to another.
type Flow struct {
	Lanes []string `json:"lanes"` // the packages it passes through, left to right
	Hops  []Hop    `json:"hops"`
}

type Hop struct {
	Lane int    `json:"lane"`
	Name string `json:"name"`
	Note string `json:"note"` // markdown
	Code *Code  `json:"code,omitempty"`
}

// Code is lines to show: from a file, or what a command printed.
type Code struct {
	File  string   `json:"file"`           // the file, or the command
	Line  int      `json:"line,omitempty"` // the number of the first line; 0 when the lines have none
	Show  [2]int   `json:"show"`           // the first and last line shown before it is given the whole view
	Hot   [2]int   `json:"hot"`            // the first and last line that matter
	Lang  string   `json:"lang,omitempty"`
	Lines []string `json:"lines"`
}

// Check is a claim and each case it was tried on.
type Check struct {
	Claim string `json:"claim"`
	Cases []Case `json:"cases"`
}

type Case struct {
	Name  string `json:"name"`
	Holds bool   `json:"holds"`
	Why   string `json:"why"` // markdown
	Code  *Code  `json:"code,omitempty"`
}

// Held counts the cases that hold.
func (c Check) Held() (n int) {
	for _, k := range c.Cases {
		if k.Holds {
			n++
		}
	}
	return n
}

// Colours is a palette, in groups.
type Colours struct {
	Groups []Group `json:"groups"`
}

type Group struct {
	Name  string   `json:"name"`
	Items []Colour `json:"items"`
}

type Colour struct {
	Name string `json:"name"`
	Tone string `json:"tone"`
	Job  string `json:"job"`
	Uses int    `json:"uses"`
}

// Run is commands and how each ended when the answer was written.
type Run struct {
	At   string `json:"at"`
	Cmds []Cmd  `json:"cmds"`
}

type Cmd struct {
	Line string   `json:"line"`
	OK   bool     `json:"ok"`
	Took string   `json:"took"`
	Out  []string `json:"out"`
}

// Change is what a commit did to each part it touched.
type Change struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Name    string `json:"name"`
	Files   int    `json:"files"`
	Added   int    `json:"added"`
	Removed int    `json:"removed"`
	Note    string `json:"note,omitempty"`
}

// Find gives the answers whose question has every word of the query.
func Find(in []Answer, query string) []Answer {
	words := strings.Fields(strings.ToLower(query))
	var out []Answer
next:
	for _, a := range in {
		q := strings.ToLower(a.Question)
		for _, w := range words {
			if !strings.Contains(q, w) {
				continue next
			}
		}
		out = append(out, a)
	}
	return out
}

// Exec runs a command line in the working directory and says how it ended.
func Exec(ctx context.Context, line string) Cmd {
	words := strings.Fields(line)
	began := time.Now()
	out, err := exec.CommandContext(ctx, words[0], words[1:]...).CombinedOutput()
	took := strconv.FormatFloat(time.Since(began).Seconds(), 'f', 1, 64) + "s"
	text := strings.TrimRight(strings.ReplaceAll(string(out), "\t", "  "), "\n")
	lines := []string{"(no output)"}
	if text != "" {
		lines = strings.Split(text, "\n")
	} else if err != nil {
		lines = []string{err.Error()}
	}
	return Cmd{Line: line, OK: err == nil, Took: took, Out: lines}
}
