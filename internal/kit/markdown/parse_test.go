package markdown

import "testing"

const sample = "---\ntab: t\nat: abc\n---\n# Q?\n\nintro\n\n" +
	"```mermaid\ngraph TD\n  a -->|uses| b --> c[Sea]\n```\n\n" +
	"### [bad] b\nwhy\n\n```code\nf.go:2-3 !3\n```\n\n### Sea\n\n" +
	"## One\n\n```go\n## not a heading\n```\n\n```swatch\nbrand: Primary\n```\n"

func TestParse(t *testing.T) {
	read := func(string) ([]byte, error) { return []byte("one\ntwo\nthree\nfour\n"), nil }
	a, err := Parse(sample, read)
	if err != nil {
		t.Fatal(err)
	}
	if a.Title != "Q?" || a.Tab != "t" || a.At != "abc" {
		t.Errorf("head read as %q, %q, %q", a.Title, a.Tab, a.At)
	}
	if len(a.Intro) != 2 || a.Intro[0].Prose != "intro" || a.Intro[1].Subsections == nil {
		t.Fatalf("intro read as %+v", a.Intro)
	}
	p := a.Intro[1].Subsections
	if len(p.Edges) != 2 || p.Edges[0] != (Edge{"a", "b", "uses"}) || p.Edges[1] != (Edge{"b", "Sea", ""}) {
		t.Errorf("edges read as %+v", p.Edges)
	}
	if len(p.List) != 2 || p.List[0].Mark != Breaks || p.List[0].Cells[0].Prose != "why" || p.Index("Sea") != 1 {
		t.Fatalf("subsections read as %+v", p.List)
	}
	if c := p.List[0].Cells[1].Code; c == nil || c.Line != 1 || len(c.Lines) != 4 || c.Hot != [2]int{3, 3} {
		t.Errorf("code read as %+v", c)
	}
	if (Section{Cells: a.Intro}).Verdict() != Breaks {
		t.Error("a subsection that breaks did not break its section")
	}
	if len(a.Sections) != 1 || len(a.Sections[0].Cells) != 2 || a.Sections[0].Cells[1].Swatch == nil {
		t.Fatalf("sections read as %+v", a.Sections)
	}
	if _, err := Parse("# Q\n\n```mermaid\npie\n```\n", nil); err == nil {
		t.Error("a drawing there is none for was taken")
	}
}
