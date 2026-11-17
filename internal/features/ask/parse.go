package ask

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

// The answers there are, in the order the palette lists them. Each was
// written by hand about this repository at the commit it names; the code,
// counts and command output in them were read from it then and are not kept
// in step.
//
//go:embed answers/*.md
var files embed.FS

var order = []string{"run", "colours", "intro-commit"}

func Answers() []Answer {
	var out []Answer
	for _, name := range order {
		src, err := files.ReadFile("answers/" + name + ".md")
		if err == nil {
			var a Answer
			if a, err = Parse(string(src)); err == nil {
				out = append(out, a)
				continue
			}
		}
		// The files are built in, and a test parses every one.
		panic(fmt.Sprintf("answers/%s.md: %v", name, err))
	}
	return out
}

// fence opens a block inside an answer.
const fence = "```strigil-block"

// Parse reads an answer as it is written: a markdown document.
//
//	---
//	tab: its name on a tab
//	at: the commit it was read at
//	---
//
//	# The question
//
//	Markdown, and wherever a block goes, its JSON in a fence:
//
//	```strigil-block
//	{"kind": "check", "claim": "...", "cases": [...]}
//	```
//
//	## A heading
//
// What comes before the first second-level heading is the lead; each such
// heading starts a section.
func Parse(src string) (Answer, error) {
	var a Answer
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

	cells := &a.Lead
	var text []string
	flush := func() {
		if t := strings.TrimSpace(strings.Join(text, "\n")); t != "" {
			*cells = append(*cells, Cell{Prose: t})
		}
		text = nil
	}
	fenced := false // inside a fence that is not a block: its lines are text
	for ; i < len(lines); i++ {
		l := lines[i]
		switch {
		case strings.HasPrefix(l, fence):
			flush()
			var block []string
			for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
				block = append(block, lines[i])
			}
			var c Cell
			if err := json.Unmarshal([]byte(strings.Join(block, "\n")), &c); err != nil {
				return a, fmt.Errorf("the block ending at line %d: %w", i+1, err)
			}
			*cells = append(*cells, c)
		case strings.HasPrefix(l, "```"):
			fenced = !fenced
			text = append(text, l)
		case fenced:
			text = append(text, l)
		case strings.HasPrefix(l, "# ") && a.Question == "":
			a.Question = strings.TrimSpace(l[2:])
		case strings.HasPrefix(l, "## "):
			flush()
			a.Sections = append(a.Sections, Section{Title: strings.TrimSpace(l[3:])})
			cells = &a.Sections[len(a.Sections)-1].Cells
		default:
			text = append(text, l)
		}
	}
	flush()
	if a.Question == "" {
		return a, fmt.Errorf("no question: an answer starts with a # heading")
	}
	return a, nil
}
