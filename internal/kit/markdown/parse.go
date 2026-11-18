package markdown

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// around is how many lines either side of what is shown the whole view has.
const around = 25

// Parse reads a document as it is written, in markdown. Nothing in one is
// syntax of ours: any markdown reader shows it whole, and what is listed
// here is only what a reader may show better.
//
//	---
//	tab: its name on a tab
//	at: the commit it was read at
//	---
//
//	# The title
//
// What comes before the first second-level heading is the intro; each such
// heading starts a section. Third-level headings are its subsections, each
// running to the next, and a reader may show one at a time. A subsection
// whose name starts with ✓ or ✗ carries that mark.
//
//   - A line that is only a link to lines of a file, [name](path#L12-L18),
//     is those lines, read with read.
//   - A table with a column of +n and a column of -n is bars: lines added
//     and removed, a row a part.
//   - A mermaid fence, a sequenceDiagram or a flowchart, is a drawing.
//     Followed by subsections it is how they are picked: a message or a node
//     named as a subsection is that subsection.
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
	fenced := false // inside a fence that is not a drawing: its lines are text
	for ; i < len(lines); i++ {
		l := lines[i]
		switch {
		case !fenced && strings.HasPrefix(l, "```") && strings.TrimSpace(l[3:]) == "mermaid":
			var body []string
			for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
				if strings.TrimSpace(lines[i]) != "" {
					body = append(body, strings.TrimSpace(lines[i]))
				}
			}
			drawing, err := mermaid(body)
			if err != nil {
				return a, fmt.Errorf("the mermaid fence ending at line %d: %w", i+1, err)
			}
			flush()
			*cells = append(*cells, Cell{Subsections: drawing})
			if !within {
				subs = drawing
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
			for mark, sign := range map[string]Mark{"✓": Holds, "✔": Holds, "✗": Breaks, "✘": Breaks, "×": Breaks} {
				if rest, ok := strings.CutPrefix(it.Name, mark+" "); ok {
					it.Name, it.Mark = strings.TrimSpace(rest), sign
				}
			}
			subs.List, within = append(subs.List, it), true
			cells = &subs.List[len(subs.List)-1].Cells
		case strings.HasPrefix(l, "|"):
			table := []string{l}
			for i+1 < len(lines) && strings.HasPrefix(lines[i+1], "|") {
				i++
				table = append(table, lines[i])
			}
			if b := bars(table); b != nil {
				flush()
				*cells = append(*cells, Cell{Bars: b})
			} else {
				text = append(text, table...)
			}
		default:
			// Set in by four it is an example of a link, and stays text.
			if m := lineLink.FindStringSubmatch(strings.TrimSpace(l)); m != nil && !strings.HasPrefix(l, "    ") && !strings.HasPrefix(l, "\t") {
				flush()
				*cells = append(*cells, Cell{Code: code(m[1], m[2], m[3], read)})
				continue
			}
			text = append(text, l)
		}
	}
	flush()
	if a.Title == "" {
		return a, fmt.Errorf("no title: a document starts with a # heading")
	}
	return a, nil
}

// lineLink is a link to lines of a file, as GitHub writes one.
var lineLink = regexp.MustCompile(`^\[[^\]]*\]\(([^)\s#]+)#L(\d+)(?:-L?(\d+))?\)$`)

// code is lines first to last of a file, read with read.
func code(file, first, last string, read func(string) ([]byte, error)) *Code {
	c := &Code{File: file}
	c.Show[0], _ = strconv.Atoi(first)
	c.Show[1] = c.Show[0]
	if n, err := strconv.Atoi(last); err == nil && n > c.Show[0] {
		c.Show[1] = n
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
		return c
	}
	src, err := read(c.File)
	if err != nil {
		c.Lines = []string{"cannot read " + c.File + " from here"}
		return c
	}
	all := strings.Split(strings.ReplaceAll(strings.TrimRight(string(src), "\n"), "\t", "    "), "\n")
	if c.Show[0] < 1 || c.Show[1] > len(all) {
		c.Lines = []string{c.File + " has " + strconv.Itoa(len(all)) + " lines now"}
		return c
	}
	c.Line = max(1, c.Show[0]-around)
	c.Lines = all[c.Line-1 : min(len(all), c.Show[1]+around)]
	return c
}

var (
	rule  = regexp.MustCompile(`^:?-+:?$`)
	plus  = regexp.MustCompile(`^\+\d+$`)
	minus = regexp.MustCompile(`^-\d+$`)
)

// bars reads a table as bars when it has a column of +n and a column of
// -n. Its first column is the names, and what other columns say is the
// note. Any other table is nil, and stays text.
func bars(table []string) *Bars {
	var rows [][]string
	for _, l := range table {
		cells := strings.Split(strings.Trim(strings.TrimSpace(l), "|"), "|")
		for i := range cells {
			cells[i] = strings.Trim(strings.TrimSpace(cells[i]), "`")
		}
		rows = append(rows, cells)
	}
	if len(rows) < 3 || !every(rows[1:2], 0, rule) {
		return nil
	}
	body, in, out := rows[2:], -1, -1
	for col := 1; col < len(rows[0]); col++ {
		switch {
		case in < 0 && every(body, col, plus):
			in = col
		case out < 0 && every(body, col, minus):
			out = col
		}
	}
	if in < 0 || out < 0 {
		return nil
	}
	b := &Bars{}
	for _, r := range body {
		bar := Bar{Name: r[0]}
		bar.Added, _ = strconv.Atoi(r[in][1:])
		bar.Removed, _ = strconv.Atoi(r[out][1:])
		var note []string
		for col := 1; col < len(r); col++ {
			if col != in && col != out && r[col] != "" {
				note = append(note, r[col])
			}
		}
		bar.Note = strings.Join(note, " ")
		b.Rows = append(b.Rows, bar)
	}
	return b
}

// every says that column of every row matches.
func every(rows [][]string, col int, re *regexp.Regexp) bool {
	for _, r := range rows {
		if col >= len(r) || !re.MatchString(r[col]) {
			return false
		}
	}
	return true
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
		// participant id as Name
		named := map[string]string{}
		who := func(id string) string {
			id = strings.TrimSpace(id)
			if name, ok := named[id]; ok {
				return name
			}
			return id
		}
		for _, l := range body[1:] {
			if rest, ok := strings.CutPrefix(l, "participant "); ok {
				id, name, aliased := strings.Cut(rest, " as ")
				if aliased {
					named[strings.TrimSpace(id)] = strings.TrimSpace(name)
				}
				lane(who(id))
				continue
			}
			ends, said, _ := strings.Cut(l, ":")
			from, to, ok := strings.Cut(ends, "->>")
			if !ok {
				return nil, fmt.Errorf("%q is not a message, from->>to: name", l)
			}
			lane(who(strings.TrimRight(from, "-")))
			p.Messages = append(p.Messages, Message{Lane: lane(who(to)), Name: strings.TrimSpace(said)})
		}
	case "flowchart", "graph":
		// id["Name"] names a node once; after that its id is enough.
		named := map[string]string{}
		name := func(node string) string {
			node = strings.TrimSpace(node)
			if at := strings.IndexAny(node, "[({"); at >= 0 {
				named[node[:at]] = strings.Trim(node[at:], "[](){}\"")
				return named[node[:at]]
			}
			if n, ok := named[node]; ok {
				return n
			}
			return node
		}
		for _, l := range body[1:] {
			if l == "end" || strings.HasPrefix(l, "subgraph") || strings.HasPrefix(l, "direction") {
				continue
			}
			nodes := strings.Split(l, "-->")
			if len(nodes) == 1 {
				name(nodes[0])
			}
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
