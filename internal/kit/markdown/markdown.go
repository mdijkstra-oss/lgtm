// Package markdown is what a document is made of, and how one is read from
// markdown. It draws nothing and knows nothing of what a document is about.
package markdown

// Doc is a document: an intro, and sections under headings. Both are made
// of cells. How one is written is in Parse.
type Doc struct {
	Title    string
	Tab      string // its name on a tab
	At       string // the commit it was read at
	Intro    []Cell
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

// Verdict is what the marks on the section's subsections come to: one that
// breaks is enough to break it.
func (s Section) Verdict() Mark {
	mark := None
	for _, c := range s.Cells {
		if c.Subsections == nil {
			continue
		}
		for _, it := range c.Subsections.List {
			if it.Mark == Breaks || mark == None {
				mark = max(mark, it.Mark)
			}
		}
	}
	return mark
}

// Cell is one thing in a document: markdown, or a block. Exactly one field
// is set.
type Cell struct {
	Prose       string
	Subsections *Subsections
	Code        *Code
	Swatch      *Swatch
	Bars        *Bars
}

// Subsections is the third-level headings of a section and what stands
// under each. A reader may show one at a time. A drawing can stand over
// them to pick from: a sequence when there are messages, a graph when
// there are edges. With no subsections it is the drawing alone.
type Subsections struct {
	List     []Subsection
	Lanes    []string // who takes part in the sequence, left to right
	Messages []Message
	Edges    []Edge
}

// Subsection is a heading and what stands under it: cells, as a section
// is made of.
type Subsection struct {
	Name  string
	Mark  Mark
	Cells []Cell
}

// Index gives the subsection of that name, or -1.
func (p Subsections) Index(name string) int {
	for i, it := range p.List {
		if it.Name == name {
			return i
		}
	}
	return -1
}

// Message is one message of a sequence, drawn in the lane it goes to.
type Message struct {
	Lane int
	Name string
}

type Edge struct {
	From, To, Label string
}

// Code is lines of a file to show.
type Code struct {
	File  string
	Line  int    // the number of the first line; 0 when the file could not be read
	Show  [2]int // the first and last line shown before it is given the whole view
	Hot   [2]int // the first and last line that matter
	Lang  string
	Lines []string
}

// Swatch is colours of the theme by name, in groups.
type Swatch struct {
	Groups []Group
}

type Group struct {
	Name  string
	Names []string
}

// Bars is rows of two numbers: lines added and lines removed.
type Bars struct {
	Rows []Bar
}

type Bar struct {
	Name    string
	Added   int
	Removed int
	Note    string
}
