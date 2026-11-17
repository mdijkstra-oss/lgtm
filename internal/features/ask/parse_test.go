package ask

import "testing"

func TestEveryAnswerParses(t *testing.T) {
	for _, name := range order {
		src, err := files.ReadFile("answers/" + name + ".md")
		if err != nil {
			t.Fatal(err)
		}
		a, err := Parse(string(src))
		if err != nil {
			t.Errorf("answers/%s.md: %v", name, err)
		}
		if a.Tab == "" || a.At == "" || len(a.Lead) == 0 {
			t.Errorf("answers/%s.md: needs a tab, a commit and a lead", name)
		}
	}
}

func TestParse(t *testing.T) {
	a, err := Parse("---\ntab: t\nat: abc\n---\n# Q?\n\nlead\n\n```strigil-block\n{\"kind\":\"check\",\"claim\":\"c\",\"cases\":[{\"name\":\"n\",\"holds\":false,\"why\":\"w\"}]}\n```\n\n## One\n\n```go\n## not a heading\n```\n")
	if err != nil {
		t.Fatal(err)
	}
	if a.Question != "Q?" || a.Tab != "t" || a.At != "abc" {
		t.Errorf("head read as %q, %q, %q", a.Question, a.Tab, a.At)
	}
	if len(a.Lead) != 2 || a.Lead[0].Prose != "lead" || a.Lead[1].Check == nil {
		t.Fatalf("lead read as %+v", a.Lead)
	}
	if len(a.Sections) != 1 || a.Sections[0].Title != "One" || len(a.Sections[0].Cells) != 1 {
		t.Fatalf("sections read as %+v", a.Sections)
	}
	if _, err := Parse("# Q\n\n```strigil-block\n{\"kind\":\"nope\"}\n```\n"); err == nil {
		t.Error("a block of an unknown kind was taken")
	}
}
