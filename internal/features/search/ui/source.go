// Package ui is how search appears in the TUI: a palette source and a results page.
package ui

import (
	"context"
	"fmt"
	"strings"

	"strigil/internal/features/search"
	"strigil/internal/kit"
	"strigil/internal/kit/theme"
)

type Source struct {
	th  theme.Theme
	dir string
}

func NewSource(th theme.Theme, dir string) Source { return Source{th: th, dir: dir} }

func (s Source) Name() string { return "files" }

func (s Source) Search(ctx context.Context, query string, emit func(kit.Item)) error {
	q := newQuery(s.dir, query)
	matcher := q.Matcher()
	docs := map[string]*doc{}
	return search.Run(ctx, q, func(m search.Match) {
		d, seen := docs[m.Path]
		if !seen {
			d = load(s.dir, matcher, m.Path)
			docs[m.Path] = d
		}
		if d == nil {
			return
		}
		idx := d.index(m.Line-1, m.Col-1)
		if idx < 0 {
			return
		}
		emit(s.item(query, d, idx, m))
	})
}

func (s Source) item(query string, d *doc, idx int, m search.Match) kit.Item {
	th, dir := s.th, s.dir
	lead := len(m.Text) - len(strings.TrimLeft(m.Text, " \t"))
	start := m.Col - 1 - lead
	return kit.Item{
		Title:    strings.ReplaceAll(m.Text[lead:], "\t", " "),
		Marks:    [][2]int{{start, start + m.Length}},
		Location: fmt.Sprintf("%s:%d", m.Path, m.Line),
		DocID:    pageID(query),
		Data:     ref{path: m.Path, line: m.Line - 1, start: m.Col - 1},
		Preview:  func(w, h int) string { return preview(th, d, idx, w, h) },
		Open:     func() kit.Page { return Page{th: th, dir: dir, query: query} },
	}
}

func newQuery(dir, text string) search.Query {
	return search.Query{Text: text, Dir: dir, IgnoreCase: true}
}

// pageID is the same for queries that find the same things.
func pageID(query string) string { return "search:" + strings.ToLower(query) }

// ref points an item at one hit of its query.
type ref struct {
	path        string
	line, start int
}

// doc is a file as one query sees it. It is not changed after load returns.
type doc struct {
	rel    string
	lines  []string
	hits   []hit
	byLine map[int][]int
}

type hit struct {
	line       int
	start, end int
}

func load(dir string, matcher search.Matcher, rel string) *doc {
	lines, err := search.Load(dir, rel)
	if err != nil {
		return nil
	}
	return index(rel, lines, matcher)
}

func index(rel string, lines []string, matcher search.Matcher) *doc {
	d := &doc{rel: rel, lines: lines, byLine: map[int][]int{}}
	for i, line := range lines {
		for _, r := range matcher.Find(line) {
			d.byLine[i] = append(d.byLine[i], len(d.hits))
			d.hits = append(d.hits, hit{line: i, start: r[0], end: r[1]})
		}
	}
	return d
}

// collect loads every file the query hits, in the order search finds them.
func collect(ctx context.Context, dir, query string) ([]*doc, error) {
	q := newQuery(dir, query)
	matcher := q.Matcher()
	seen := map[string]bool{}
	var docs []*doc
	err := search.Run(ctx, q, func(m search.Match) {
		if seen[m.Path] {
			return
		}
		seen[m.Path] = true
		if d := load(dir, matcher, m.Path); d != nil {
			docs = append(docs, d)
		}
	})
	return docs, err
}

func (d *doc) index(line, start int) int {
	for _, i := range d.byLine[line] {
		if d.hits[i].start == start {
			return i
		}
	}
	return -1
}
