package markdown

import "testing"

const sample = "---\ntab: t\nat: abc\n---\n# Q?\n\nintro\n\n" +
	"```mermaid\ngraph TD\n  a -->|uses| b --> c[Sea]\n  d[\"deep/down\"]\n  c --> d\n```\n\n" +
	"### ✗ b\nwhy\n\n[f.go](f.go#L2-L3)\n\n### Sea\n\n" +
	"## One\n\n```go\n## not a heading\n```\n\n" +
	"| part | in | out | note |\n|---|---|---|---|\n| new | +12 | -0 | came |\n| old | +0 | -30 | |\n\n" +
	"| a | b |\n|---|---|\n| 1 | 2 |\n"

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
	want := []Edge{{"a", "b", "uses"}, {"b", "Sea", ""}, {"Sea", "deep/down", ""}}
	if len(p.Edges) != len(want) {
		t.Fatalf("edges read as %+v", p.Edges)
	}
	for i, e := range want {
		if p.Edges[i] != e {
			t.Errorf("edge %d read as %+v, not %+v", i, p.Edges[i], e)
		}
	}
	if len(p.List) != 2 || p.List[0].Name != "b" || p.List[0].Mark != Breaks || p.List[0].Cells[0].Prose != "why" || p.Index("Sea") != 1 {
		t.Fatalf("subsections read as %+v", p.List)
	}
	if c := p.List[0].Cells[1].Code; c == nil || c.File != "f.go" || c.Show != [2]int{2, 3} || c.Line != 1 || len(c.Lines) != 4 {
		t.Errorf("code read as %+v", c)
	}
	if (Section{Cells: a.Intro}).Verdict() != Breaks {
		t.Error("a subsection that breaks did not break its section")
	}
	if len(a.Sections) != 1 || len(a.Sections[0].Cells) != 3 {
		t.Fatalf("sections read as %+v", a.Sections)
	}
	b := a.Sections[0].Cells[1].Bars
	if b == nil || len(b.Rows) != 2 || b.Rows[0] != (Bar{"new", 12, 0, "came"}) || b.Rows[1] != (Bar{"old", 0, 30, ""}) {
		t.Errorf("bars read as %+v", b)
	}
	if a.Sections[0].Cells[2].Prose == "" {
		t.Error("a table with no +n and -n columns was not left as text")
	}
	if eg, _ := Parse("# Q\n\n    [f.go](f.go#L2-L3)\n", read); len(eg.Intro) != 1 || eg.Intro[0].Code != nil {
		t.Error("a link set in as an example was read as lines of a file")
	}
	if _, err := Parse("# Q\n\n```mermaid\npie\n```\n", nil); err == nil {
		t.Error("a drawing there is none for was taken")
	}
}

func TestSequenceNames(t *testing.T) {
	a, err := Parse("# Q\n\n```mermaid\nsequenceDiagram\n  participant m as cmd/app\n  participant lib\n  m->>lib: go\n  lib->>m: back\n```\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	p := a.Intro[0].Subsections
	if len(p.Lanes) != 2 || p.Lanes[0] != "cmd/app" || p.Lanes[1] != "lib" {
		t.Errorf("lanes read as %q", p.Lanes)
	}
	if len(p.Messages) != 2 || p.Messages[0] != (Message{1, "go"}) || p.Messages[1] != (Message{0, "back"}) {
		t.Errorf("messages read as %+v", p.Messages)
	}
}
