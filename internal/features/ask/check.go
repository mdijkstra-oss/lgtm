package ask

import (
	"fmt"
	"regexp"
	"strings"

	"lgtm/internal/kit/markdown"
)

// longest is how many lines a link may show.
const longest = 40

// How much a document may say, in words of text: code and drawings are
// not counted. A model left to itself writes three times this.
const (
	mostOpening  = 80
	mostPart     = 70  // a subsection of the answer
	mostPanel    = 130 // a panel of a practice question
	mostWords    = 800
	mostSections = 4 // before the practice
	mostPractice = 2
)

// words counts the words of text in cells, and in what is under them.
func words(cells []markdown.Cell) int {
	n := 0
	for _, c := range cells {
		n += len(strings.Fields(c.Prose))
		if c.Subsections != nil {
			for _, sub := range c.Subsections.List {
				n += words(sub.Cells)
			}
		}
	}
	return n
}

// asks says whether a section is a practice question.
func asks(s markdown.Section) bool {
	if strings.HasPrefix(s.Title, "Try:") {
		return true
	}
	for _, c := range s.Cells {
		if c.Subsections != nil && c.Subsections.Index("Question") >= 0 {
			return true
		}
	}
	return false
}

var (
	arrow = regexp.MustCompile(`^[\w./-]+(\[[^\]]*\])?\s*-->(\|[^|]*\|)?\s*[\w./-]+(\[[^\]]*\])?$`)
	skip  = regexp.MustCompile(`^(end$|subgraph\b|direction\b)`)
)

// check reads a document a model wrote and says everything wrong with it,
// in words the model can act on. seen says whether the model read lines a
// to b of a file; from names the file it read that has all of those lines.
//
// What is wrong comes in two kinds. hard is what would show the reader
// something broken or untrue, and is never let through. soft is length:
// worth one request to shorten, and not worth keeping a document from the
// reader.
func check(src string, root string, seen func(file string, a, b int) bool, from func(lines []string) string) (doc markdown.Doc, hard, soft []string) {
	say := func(format string, args ...any) { hard = append(hard, fmt.Sprintf(format, args...)) }
	nag := func(format string, args ...any) { soft = append(soft, fmt.Sprintf(format, args...)) }

	doc, err := markdown.Parse(src, reader(root))
	if err != nil {
		return doc, []string{err.Error()}, nil
	}
	if len(doc.Intro) == 0 {
		say("there is no text before the first ## heading: the answer comes first, in two to four sentences")
	}

	var walk func(where string, cells []markdown.Cell)
	walk = func(where string, cells []markdown.Cell) {
		for _, c := range cells {
			switch {
			case c.Code != nil:
				link := fmt.Sprintf("%s#L%d-L%d", c.Code.File, c.Code.Show[0], c.Code.Show[1])
				switch {
				case c.Code.Line == 0:
					say("%s: the link %s: %s", where, link, strings.Join(c.Code.Lines, " "))
				case c.Code.Show[1]-c.Code.Show[0] >= longest:
					say("%s: the link %s shows more than %d lines: link the few that matter", where, link, longest)
				case !seen(c.Code.File, c.Code.Show[0], c.Code.Show[1]):
					say("%s: the link %s is to lines you have not read: read them, then link the lines you saw", where, link)
				}
			case c.Subsections != nil:
				p := c.Subsections
				if len(p.Messages)+len(p.Edges) > 0 {
					drawn := map[string]bool{}
					for _, m := range p.Messages {
						drawn[m.Name] = true
					}
					for _, e := range p.Edges {
						drawn[e.From], drawn[e.To] = true, true
					}
					for _, sub := range p.List {
						if !drawn[sub.Name] {
							say("%s: the subsection %q is under a drawing that has nothing of that name: name it as a message or a node is named, letter for letter, or take it out", where, sub.Name)
						}
						shows := false
						for _, in := range sub.Cells {
							shows = shows || in.Code != nil
						}
						if !shows {
							say("%s: the subsection %q can be picked from the drawing and shows no code: give it a link to the lines that do what it says, or take the subsection out and leave it in the drawing", where, sub.Name)
						}
					}
				}
				for _, sub := range p.List {
					walk(where+", "+fmt.Sprintf("%q", sub.Name), sub.Cells)
				}
			}
		}
	}
	walk("the opening", doc.Intro)
	if n := words(doc.Intro); n > mostOpening {
		nag("the opening has %d words: answer in two to four plain sentences, under %d words", n, mostOpening)
	}
	total, parts, tries := words(doc.Intro), 0, 0
	for _, s := range doc.Sections {
		where := fmt.Sprintf("section %q", s.Title)
		walk(where, s.Cells)
		practice(where, s, say)
		total += words(s.Cells)
		most := mostPart
		if asks(s) {
			tries, most = tries+1, mostPanel
		} else {
			parts++
		}
		for _, c := range s.Cells {
			if c.Subsections == nil {
				continue
			}
			for _, sub := range c.Subsections.List {
				if n := words(sub.Cells); n > most {
					nag("%s: the subsection %q has %d words: say it in under %d and let the code show the rest", where, sub.Name, n, most)
				}
			}
		}
	}
	if parts > mostSections {
		nag("there are %d sections before the practice: keep the %d that answer the question and drop or merge the rest", parts, mostSections)
	}
	if tries > mostPractice {
		nag("there are %d practice questions: keep the best %d", tries, mostPractice)
	}
	if total > mostWords {
		nag("the document has %d words of text: cut it to under %d by dropping what the reader did not ask about, not by squeezing sentences", total, mostWords)
	}

	fences(src, from, say)
	return doc, hard, soft
}

// practice holds a section that looks like a practice question to the
// shape one has.
func practice(where string, s markdown.Section, say func(string, ...any)) {
	for i, c := range s.Cells {
		p := c.Subsections
		if p == nil || len(p.List) == 0 {
			continue
		}
		is := false
		for _, sub := range p.List {
			n := strings.ToLower(sub.Name)
			is = is || n == "question" || n == "answer" || strings.HasPrefix(n, "hint")
		}
		if !is {
			continue
		}
		if p.List[0].Name != "Question" || p.List[len(p.List)-1].Name != "Answer" {
			say("%s: a practice question has the subsections Question, Hint 1, Hint 2, Answer, in that order", where)
		}
		if i > 0 {
			say("%s: nothing stands above ### Question in a practice section, or it stays on screen", where)
		}
		for _, sub := range p.List {
			if sub.Mark != markdown.None {
				say("%s: no ✓ or ✗ on the subsections of a practice question", where)
				break
			}
		}
	}
}

// fences reads what the parser does not hold a document to: arrows a
// flowchart may be written with, and code in a fence that is in the
// repository and should be a link.
func fences(src string, from func(lines []string) string, say func(string, ...any)) {
	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		if strings.HasPrefix(l, "Read:") || strings.HasPrefix(l, "Not read:") {
			say("line %d: take out the line listing what you read and did not read: the document does not say it", i+1)
		}
		if strings.HasPrefix(l, "#### ") {
			say("line %d: there is no #### heading: use ### or plain text", i+1)
		}
		if !strings.HasPrefix(l, "```") {
			continue
		}
		lang, start := strings.TrimSpace(l[3:]), i
		var body []string
		for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
			if t := strings.TrimSpace(lines[i]); t != "" {
				body = append(body, t)
			}
		}
		switch {
		case lang == "mermaid":
			if len(body) == 0 {
				continue
			}
			if kind := strings.Fields(body[0])[0]; kind != "flowchart" && kind != "graph" {
				continue
			}
			for _, b := range body[1:] {
				if !skip.MatchString(b) && !arrow.MatchString(b) {
					say("the flowchart at line %d: %q is not an arrow as they may be written: a --> b, or a -->|label| b", start+1, b)
				}
			}
		case lang != "diff":
			if file := copied(body, from); file != "" {
				say("the fence at line %d is code from %s: take it out and link the lines instead, as [name](%s#L10-L20) alone on a line", start+1, file, file)
			}
		}
	}
}

// copied names the file a fence was copied out of, if it was: three lines
// of some length or more, all of them in one file the model read.
func copied(body []string, from func(lines []string) string) string {
	var long []string
	for _, b := range body {
		if len(b) >= 12 {
			long = append(long, b)
		}
	}
	if len(long) < 3 || from == nil {
		return ""
	}
	return from(long)
}
