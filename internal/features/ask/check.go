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
		n += len(strings.Fields(unfenced(c.Prose)))
		if c.Subsections != nil {
			for _, sub := range c.Subsections.List {
				n += words(sub.Cells)
			}
		}
	}
	return n
}

// unfenced is text without what stands in its fences: code is shown, not
// read, and does not count against how much a document may say.
func unfenced(text string) string {
	var out []string
	fenced := false
	for _, l := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			fenced = !fenced
			continue
		}
		if !fenced {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

var stop = regexp.MustCompile(`[.!?](\s|$)`)

// sentences counts the sentences of the text in cells, near enough: a
// full stop inside a code span is not the end of one.
func sentences(cells []markdown.Cell) int {
	n := 0
	for _, c := range cells {
		text := regexp.MustCompile("`[^`]*`").ReplaceAllString(unfenced(c.Prose), "x")
		n += len(stop.FindAllString(text, -1))
	}
	return n
}

// steps says whether a section shows something before and after, as two
// subsections to step between.
func steps(s markdown.Section) bool {
	for _, c := range s.Cells {
		p := c.Subsections
		if p == nil {
			continue
		}
		// "Before: GET listed twice" is a Before all the same.
		before, after := false, false
		for _, sub := range p.List {
			before = before || strings.HasPrefix(sub.Name, "Before") || sub.Name == "Without it"
			after = after || strings.HasPrefix(sub.Name, "After") || sub.Name == "With it"
		}
		if before && after {
			return true
		}
	}
	return false
}

// marked says whether any subsection of a block carries a ✓ or a ✗.
func marked(p *markdown.Subsections) bool {
	for _, sub := range p.List {
		if sub.Mark != markdown.None {
			return true
		}
	}
	return false
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
//
// size is how big a change the document is about, and noChange when it
// answers a question. A change has its own shape: it has to show something
// before and after, it explains and marks nothing as right or wrong, and
// it is as short as the change is small.
func check(src string, root string, size int, seen func(file string, a, b int) bool, from func(lines []string) string) (doc markdown.Doc, hard, soft []string) {
	say := func(format string, args ...any) { hard = append(hard, fmt.Sprintf(format, args...)) }
	nag := func(format string, args ...any) { soft = append(soft, fmt.Sprintf(format, args...)) }

	doc, err := markdown.Parse(src, reader(root))
	if err != nil {
		return doc, []string{err.Error()}, nil
	}
	if len(doc.Intro) == 0 {
		say("there is no text before the first ## heading: the answer comes first, in two to four sentences")
	}

	shown := map[string][][2]int{}
	var walk func(where string, cells []markdown.Cell)
	walk = func(where string, cells []markdown.Cell) {
		for _, c := range cells {
			switch {
			case c.Code != nil:
				link := fmt.Sprintf("%s#L%d-L%d", c.Code.File, c.Code.Show[0], c.Code.Show[1])
				for _, r := range shown[c.Code.File] {
					// The same lines, or mostly the same, a second time.
					both := min(r[1], c.Code.Show[1]) - max(r[0], c.Code.Show[0]) + 1
					if both*2 > c.Code.Show[1]-c.Code.Show[0]+1 {
						nag("%s: %s shows lines the document already showed (%d to %d): show a range once. Link other lines here, such as where the missing test would go, or take the link out", where, link, r[0], r[1])
						break
					}
				}
				shown[c.Code.File] = append(shown[c.Code.File], c.Code.Show)
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
		most := mostPart
		if asks(s) {
			tries, most = tries+1, mostPanel
		} else {
			parts++
		}
		// The questions at the end of a change have their own measure.
		if !asks(s) || size == noChange {
			total += words(s.Cells)
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
	sections, most := mostSections, mostWords
	if size != noChange {
		sections, most = limits[size][0], limits[size][1]
		// What stands between a heading and its first subsection stays on
		// screen through every step, so there is next to none of it.
		for _, s := range doc.Sections {
			for i, c := range s.Cells {
				if c.Subsections == nil || len(c.Subsections.List) == 0 || asks(s) {
					continue
				}
				if n := words(s.Cells[:i]); n > 30 {
					say("section %q has %d words above its first ###: they stay on screen through every step. Move them into a subsection, or drop them; only one line beginning \"Not changed:\" belongs there", s.Title, n)
				}
				break
			}
		}
		stepped := false
		for _, s := range doc.Sections {
			stepped = stepped || steps(s)
			for _, c := range s.Cells {
				// A change is explained, not marked: saying what holds and
				// what breaks is a review, which is someone else's job.
				if c.Subsections != nil && marked(c.Subsections) {
					say("section %q: take the ✓ and ✗ off its subsections. A document about a change explains what the code did and does; it does not report what is covered, missing or wrong. Name each step for what happens in it", s.Title)
				}
			}
		}
		if !stepped {
			say("a document about a change shows what it did before and what it does now: its first section has two subsections to step between, named ### Before: <the finding> and ### After: <the finding>")
		}
		// A step shows one block of code and ends on it.
		for _, s := range doc.Sections {
			if asks(s) {
				continue
			}
			for _, c := range s.Cells {
				if c.Subsections == nil {
					continue
				}
				for _, sub := range c.Subsections.List {
					blocks, last := 0, false
					for _, in := range sub.Cells {
						last = false
						switch {
						case in.Code != nil:
							blocks, last = blocks+1, true
						case strings.Contains(in.Prose, "```"):
							blocks++
							last = strings.HasSuffix(strings.TrimSpace(in.Prose), "```")
						}
					}
					switch {
					case blocks > 1:
						nag("section %q, %q shows %d blocks of code: a step shows one, a link or a diff fence. Keep the one that shows what the step says", s.Title, sub.Name, blocks)
					case blocks == 1 && !last:
						nag("section %q, %q has text after its code: a step ends on its code. Move the sentence above the code, or drop it", s.Title, sub.Name)
					}
				}
			}
		}
		if n := sentences(doc.Intro); n > 3 {
			nag("the opening has %d sentences: for a change it is exactly three: what behaves differently; what kind of change it is and who it can break; where to look", n)
		}
		if strings.HasPrefix(doc.Title, "Commit ") {
			nag("the title is %q: say the change in plain words, with no hash", doc.Title)
		}
		if n := words(doc.Intro); n > 70 {
			nag("the opening has %d words: for a change it is three sentences, under 60 words: what behaves differently, what kind of change it is, where to look", n)
		}
	}
	if parts > sections {
		nag("there are %d sections before the practice: this document gets %d. Keep what changed and what to check, and drop or merge the rest", parts, sections)
	}
	switch want := map[int]int{tiny: 1, small: 1, medium: 2, large: 2}[size]; {
	case size != noChange && tries == 0:
		say("a document about a change ends with a question the reader should be able to answer before they accept it: a ## Try: section with ### Question, ### Hint 1, ### Hint 2 and ### Answer. It asks what the change does, not what is wrong with it")
	case size != noChange && tries > want:
		nag("there are %d questions at the end: a change of this size gets %d", tries, want)
	case tries > mostPractice:
		nag("there are %d practice questions: keep the best %d", tries, mostPractice)
	}
	if total > most {
		nag("the document has %d words of text: cut it to under %d by dropping what the reader does not need to understand the change, not by squeezing sentences", total, most)
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
