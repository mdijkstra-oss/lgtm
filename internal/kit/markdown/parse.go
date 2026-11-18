package markdown

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// around is how many lines either side of what is shown the whole view has.
const around = 25

// Parse reads a document as it is written, in markdown.
//
//	---
//	tab: its name on a tab
//	at: the commit it was read at
//	---
//
//	# The title
//
//	Markdown.
//
//	## A heading
//
//	### [ok] A subsection
//
// What comes before the first second-level heading is the intro; each such
// heading starts a section. Third-level headings are its subsections,
// each running to the next; [ok] or [bad] before its name marks one. A
// subsection holds what a section does, but subsections of its own. Read as
// plain markdown the document is whole; showing subsections one at a time
// is the reader's choice.
//
// Four fences are blocks, by the name they open with:
//
//	code     path:first-last !hot-hot    lines of a file, read with read
//	swatch   group: Name Name            colours of the theme
//	bars     name +added -removed note   a row a line
//	mermaid  a sequenceDiagram or a flowchart
//
// A mermaid fence followed by subsections is how they are picked: a message or a
// node named as a subsection is that subsection.
func Parse(src string, read func(path string) ([]byte, error)) (Doc, error) {
	var a Doc
	lines := strings.Split(src, "\n")
	i := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i = 1; i < len(lines) && strings.TrimSpace(lines[i]) != "---"; i++ {
			key, value, _ := strings.Cut(lines[i], ":")
			switch strings.TrimSpace(key) {
			case "tab":
				a.Tab = strings.TrimSpace(value)
			case "at":
				a.At = strings.TrimSpace(value)
			}
		}
		i++
	}

	outer := &a.Intro // the cells of the section being read
	cells := outer    // where what is read goes: those, or a subsection's
	var (
		text   []string
		subs   *Subsections // the drawing or subsections of the section
		within bool         // inside a subsection
	)
	flush := func() {
		if t := strings.TrimSpace(strings.Join(text, "\n")); t != "" {
			*cells = append(*cells, Cell{Prose: t})
		}
		text = nil
	}
	fenced := false // inside a fence that is not a block: its lines are text
	for ; i < len(lines); i++ {
		l := lines[i]
		kind, isBlock := "", false
		if !fenced && strings.HasPrefix(l, "```") {
			kind = strings.TrimSpace(l[3:])
			switch kind {
			case "code", "swatch", "bars", "mermaid":
				isBlock = true
			}
		}
		switch {
		case isBlock:
			var body []string
			for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
				if strings.TrimSpace(lines[i]) != "" {
					body = append(body, strings.TrimSpace(lines[i]))
				}
			}
			var c Cell
			var err error
			switch kind {
			case "code":
				c.Code, err = code(body, read)
			case "swatch":
				c.Swatch = swatch(body)
			case "bars":
				c.Bars, err = bars(body)
			case "mermaid":
				c.Subsections, err = mermaid(body)
			}
			if err != nil {
				return a, fmt.Errorf("the %s fence ending at line %d: %w", kind, i+1, err)
			}
			flush()
			*cells = append(*cells, c)
			if !within {
				subs = c.Subsections
			}
		case strings.HasPrefix(l, "```"):
			fenced = !fenced
			text = append(text, l)
		case fenced:
			text = append(text, l)
		case strings.HasPrefix(l, "# ") && a.Title == "":
			a.Title = strings.TrimSpace(l[2:])
		case strings.HasPrefix(l, "## "):
			flush()
			a.Sections = append(a.Sections, Section{Title: strings.TrimSpace(l[3:])})
			outer = &a.Sections[len(a.Sections)-1].Cells
			cells, subs, within = outer, nil, false
		case strings.HasPrefix(l, "### "):
			flush()
			if subs == nil {
				subs = &Subsections{}
				*outer = append(*outer, Cell{Subsections: subs})
			}
			it := Subsection{Name: strings.TrimSpace(l[4:])}
			if rest, ok := strings.CutPrefix(it.Name, "[ok]"); ok {
				it.Name, it.Mark = strings.TrimSpace(rest), Holds
			} else if rest, ok := strings.CutPrefix(it.Name, "[bad]"); ok {
				it.Name, it.Mark = strings.TrimSpace(rest), Breaks
			}
			subs.List, within = append(subs.List, it), true
			cells = &subs.List[len(subs.List)-1].Cells
		default:
			text = append(text, l)
		}
	}
	flush()
	if a.Title == "" {
		return a, fmt.Errorf("no title: a document starts with a # heading")
	}
	return a, nil
}

// code reads "path:first-last !hot-hot", and the lines from the file.
func code(body []string, read func(string) ([]byte, error)) (*Code, error) {
	if len(body) != 1 {
		return nil, fmt.Errorf("wants one line, path:first-last")
	}
	words := strings.Fields(body[0])
	at := strings.LastIndex(words[0], ":")
	if at < 0 {
		return nil, fmt.Errorf("%q has no line numbers", words[0])
	}
	c := &Code{File: words[0][:at]}
	var err error
	if c.Show, err = span(words[0][at+1:]); err != nil {
		return nil, err
	}
	c.Hot = c.Show
	if len(words) > 1 && strings.HasPrefix(words[1], "!") {
		if c.Hot, err = span(words[1][1:]); err != nil {
			return nil, err
		}
	}
	switch filepath.Ext(c.File) {
	case ".go":
		c.Lang = "go"
	case ".yml", ".yaml":
		c.Lang = "yaml"
	case ".md":
		c.Lang = "markdown"
	}
	if read == nil {
		return c, nil
	}
	src, err := read(c.File)
	if err != nil {
		c.Lines = []string{"cannot read " + c.File + " from here"}
		return c, nil
	}
	all := strings.Split(strings.ReplaceAll(strings.TrimRight(string(src), "\n"), "\t", "    "), "\n")
	if c.Show[1] > len(all) {
		c.Lines = []string{c.File + " has " + strconv.Itoa(len(all)) + " lines now"}
		return c, nil
	}
	c.Line = max(1, c.Show[0]-around)
	c.Lines = all[c.Line-1 : min(len(all), c.Show[1]+around)]
	return c, nil
}

// span reads "12-18", or "12" for the one line.
func span(s string) (out [2]int, err error) {
	first, last, both := strings.Cut(s, "-")
	if !both {
		last = first
	}
	if out[0], err = strconv.Atoi(first); err == nil {
		out[1], err = strconv.Atoi(last)
	}
	if err != nil || out[0] < 1 || out[1] < out[0] {
		return out, fmt.Errorf("%q is not a line or a range of lines", s)
	}
	return out, nil
}

func swatch(body []string) *Swatch {
	c := &Swatch{}
	for _, l := range body {
		name, names, _ := strings.Cut(l, ":")
		c.Groups = append(c.Groups, Group{Name: strings.TrimSpace(name), Names: strings.Fields(names)})
	}
	return c
}

func bars(body []string) (*Bars, error) {
	b := &Bars{}
	for _, l := range body {
		words := strings.Fields(l)
		if len(words) < 3 || !strings.HasPrefix(words[1], "+") || !strings.HasPrefix(words[2], "-") {
			return nil, fmt.Errorf("%q is not: name +added -removed note", l)
		}
		in, err := strconv.Atoi(words[1][1:])
		out, err2 := strconv.Atoi(words[2][1:])
		if err != nil || err2 != nil {
			return nil, fmt.Errorf("%q has a count that is not a number", l)
		}
		b.Rows = append(b.Rows, Bar{Name: words[0], Added: in, Removed: out, Note: strings.Join(words[3:], " ")})
	}
	return b, nil
}

// mermaid reads the part of Mermaid there is a drawing for: the messages
// of a sequenceDiagram, and the arrows of a flowchart. A node is named by
// its label when it has one.
func mermaid(body []string) (*Subsections, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("is empty")
	}
	p := &Subsections{}
	kind := strings.Fields(body[0])[0]
	switch kind {
	case "sequenceDiagram":
		lane := func(name string) int {
			name = strings.TrimSpace(name)
			for i, l := range p.Lanes {
				if l == name {
					return i
				}
			}
			p.Lanes = append(p.Lanes, name)
			return len(p.Lanes) - 1
		}
		for _, l := range body[1:] {
			if name, ok := strings.CutPrefix(l, "participant "); ok {
				lane(name)
				continue
			}
			ends, said, _ := strings.Cut(l, ":")
			from, to, ok := strings.Cut(ends, "->>")
			if !ok {
				return nil, fmt.Errorf("%q is not a message, from->>to: name", l)
			}
			lane(strings.TrimRight(from, "-"))
			p.Messages = append(p.Messages, Message{Lane: lane(to), Name: strings.TrimSpace(said)})
		}
	case "flowchart", "graph":
		name := func(node string) string {
			node = strings.TrimSpace(node)
			if at := strings.IndexAny(node, "[({"); at >= 0 {
				return strings.Trim(node[at:], "[](){}\"")
			}
			return node
		}
		for _, l := range body[1:] {
			if l == "end" || strings.HasPrefix(l, "subgraph") || strings.HasPrefix(l, "direction") {
				continue
			}
			nodes := strings.Split(l, "-->")
			for i := 1; i < len(nodes); i++ {
				e := Edge{From: nodes[i-1], To: nodes[i]}
				// a -->|label| b
				if rest, ok := strings.CutPrefix(strings.TrimSpace(e.To), "|"); ok {
					e.Label, e.To, _ = strings.Cut(rest, "|")
					nodes[i] = e.To
				}
				e.From, e.To = name(e.From), name(e.To)
				p.Edges = append(p.Edges, e)
			}
		}
	default:
		return nil, fmt.Errorf("no drawing for a %s yet", kind)
	}
	if len(p.Messages)+len(p.Edges) == 0 {
		return nil, fmt.Errorf("has nothing to draw")
	}
	return p, nil
}
