package ask

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repo makes a repository with one file of thirty lines and one git ignores.
func repo(t *testing.T) *session {
	t.Helper()
	// What is written is kept under the temporary folder.
	t.Setenv("TMPDIR", t.TempDir())
	root := t.TempDir()
	var src strings.Builder
	for i := 1; i <= 30; i++ {
		src.WriteString("line number " + strings.Repeat("x", i) + "\n")
	}
	for name, text := range map[string]string{"pkg/a.go": src.String(), ".gitignore": ".env\n", ".env": "KEY=secret\n"} {
		os.MkdirAll(filepath.Join(root, filepath.Dir(name)), 0o755)
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "first"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	return open(root, Subject{Question: "how does a work?"}, func(Event) {})
}

func TestWhatGitIgnoresCannotBeRead(t *testing.T) {
	s := repo(t)
	if _, _, err := s.reading(readIn{Path: ".env"}); err == nil {
		t.Error("read .env")
	}
	if out, _, _ := s.list(".", 2); strings.Contains(out, ".env") || !strings.Contains(out, "pkg/a.go  (31 lines)") {
		t.Errorf("listed:\n%s", out)
	}
	if _, _, err := s.reading(readIn{Path: "../outside"}); err == nil {
		t.Error("read outside the repository")
	}
}

func TestGitOnlyReads(t *testing.T) {
	s := repo(t)
	if out, _, err := s.history(gitIn{Command: "log", Args: "--oneline"}); err != nil || !strings.Contains(out, "first") {
		t.Errorf("log: %q %v", out, err)
	}
	for _, in := range []gitIn{{Command: "commit", Args: "-m x"}, {Command: "diff", Args: "--output=x"}, {Command: "diff", Args: "--no-index /etc/passwd /etc/hosts"}} {
		if _, _, err := s.history(in); err == nil {
			t.Errorf("git %s %s was allowed", in.Command, in.Args)
		}
	}
}

func TestGrepFindsInWhatTheGlobNames(t *testing.T) {
	s := repo(t)
	for glob, want := range map[string]bool{"": true, "pkg": true, "*.go": true, "pkg/**/*.go": true, "**/*.md": false, "other": false} {
		out, _, err := s.grep(grepIn{Pattern: "number x{30}$", Glob: glob})
		if err != nil || strings.Contains(out, "pkg/a.go:30:") != want {
			t.Errorf("glob %q: %q %v", glob, out, err)
		}
	}
}

func TestADocumentIsHeldToWhatWasRead(t *testing.T) {
	s := repo(t)
	doc := func(body string) string { return "# A title\n\nThe answer.\n\n## A section\n\n" + body + "\n" }
	link := "[a.go](pkg/a.go#L3-L8)"
	wrong := func(src string) string {
		_, hard, soft := check(src, s.root, noChange, s.seen, s.from)
		return strings.Join(append(hard, soft...), "\n")
	}

	if w := wrong(doc(link)); !strings.Contains(w, "not read") {
		t.Errorf("a link to lines not read: %q", w)
	}
	s.reading(readIn{Path: "pkg/a.go", From: 1, To: 10})
	for name, c := range map[string]struct{ src, want string }{
		"lines read":            {doc(link), ""},
		"lines past what was":   {doc("[a.go](pkg/a.go#L8-L12)"), "not read"},
		"lines the file lacks":  {doc("[a.go](pkg/a.go#L80-L90)"), "has 30 lines"},
		"no opening":            {"# A title\n\n## A section\n\nText.\n", "before the first"},
		"code pasted":           {doc("```go\nline number x\nline number xx\nline number xxx\n```"), "link the lines"},
		"code of its own":       {doc("```go\nsomething := other()\nanother := thing()\nand a third one\n```"), ""},
		"an arrow chained":      {doc("```mermaid\nflowchart TD\n  a --> b --> c\n```"), "not an arrow"},
		"arrows as they may be": {doc("```mermaid\nflowchart TD\n  a[\"The a\"] --> b\n  a -->|uses| c\n```\n\n### The a\nText.\n\n" + link), ""},
		"a part with no code":   {doc("```mermaid\nflowchart TD\n  a --> b\n```\n\n### a\nText."), "shows no code"},
		"a name not drawn":      {doc("```mermaid\nflowchart TD\n  a --> b\n```\n\n### c\nText."), "nothing of that name"},
		"practice":              {doc("### Question\nWhat?\n\n### Hint 1\nThink.\n\n### Hint 2\nLook.\n\n### Answer\nThat."), ""},
		"too many sections":     {doc("Text.") + strings.Repeat("\n## Another\n\nText.\n", 4), "sections before the practice"},
		"too many words":        {doc(strings.Repeat("word ", 900)), "words of text"},
		"practice out of order": {doc("### Hint 1\nThink.\n\n### Question\nWhat?\n\n### Answer\nThat."), "in that order"},
	} {
		if w := wrong(c.src); (c.want == "") != (w == "") || !strings.Contains(w, c.want) {
			t.Errorf("%s: %q, want %q", name, w, c.want)
		}
	}
}

func TestTooLongIsAskedAboutOnceAndThenLetThrough(t *testing.T) {
	s := repo(t)
	long := writeIn{Tab: "long", Markdown: "# A title\n\nThe answer.\n\n## A section\n\n" + strings.Repeat("word ", 900)}
	if _, _, err := s.write(long); err == nil || !strings.Contains(err.Error(), "words of text") {
		t.Fatalf("the first time: %v", err)
	}
	if _, note, err := s.write(long); err != nil || s.out == nil {
		t.Fatalf("the second time: %s %v", note, err)
	}
}

func TestADocumentNeverAcceptedIsStillGiven(t *testing.T) {
	s := repo(t)
	if s.settle() {
		t.Error("gave a document when none was handed in")
	}
	s.write(writeIn{Tab: "x", Markdown: "# A title\n\nThe answer.\n\n[a.go](pkg/a.go#L3-L8)\n"})
	if s.out != nil {
		t.Fatal("accepted a link to lines not read")
	}
	if !s.settle() || s.out == nil || s.out.Doc.Title != "A title" {
		t.Errorf("settled: %+v", s.out)
	}
}

func TestRemovedCodeLosesTheIndentationItShares(t *testing.T) {
	got := dedent("Text.\n\n```diff\n-\t\t\tfor x := range xs {\n-\t\t\t\tuse(x)\n-\t\t\t}\n```\n\n\tkept as it is\n")
	want := "Text.\n\n```diff\n- for x := range xs {\n-     use(x)\n- }\n```\n\n\tkept as it is\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestOneWordGitKnowsIsACommit(t *testing.T) {
	s := repo(t)
	head := Named(s.root, "HEAD")
	if head.Commit == "" || head.About != "first" || head.Title() != "Commit "+head.Commit+": first" {
		t.Errorf("HEAD: %+v", head)
	}
	for _, text := range []string{"how does HEAD work", "nothing-of-that-name", "--help"} {
		if sub := Named(s.root, text); sub.Question != text || sub.Commit != "" {
			t.Errorf("%q: %+v", text, sub)
		}
	}
	if h := History(s.root, 10); len(h) != 1 || h[0].Hash != head.Commit || h[0].Subject != "first" {
		t.Errorf("history: %+v", h)
	}
	if n := Uncommitted(s.root); n != 0 {
		t.Errorf("%d files uncommitted in a clean repository", n)
	}
	os.WriteFile(filepath.Join(s.root, "new.txt"), []byte("x\n"), 0o644)
	if n := Uncommitted(s.root); n != 1 {
		t.Errorf("%d files uncommitted, with one new", n)
	}
}

func TestAChangeIsKeptAsOne(t *testing.T) {
	s := repo(t)
	s.sub = Named(s.root, "HEAD")
	say := s.asking()
	if !strings.Contains(say, "Explain commit "+s.sub.Commit) || !strings.Contains(say, "=== pkg/a.go (numbered") || !strings.Contains(say, "    3  + line number xxx") {
		t.Errorf("asked as: %s", say)
	}
	// The lines of the change came with the question, so they can be
	// linked without being read again.
	if !s.seen("pkg/a.go", 3, 8) {
		t.Error("the lines of the change do not count as read")
	}
	// A change has to be shown before and after, to step between.
	if _, _, err := s.write(writeIn{Tab: "first", Markdown: "# The first commit\n\nIt adds a.\n\n[a.go](pkg/a.go#L3-L8)\n"}); err == nil || !strings.Contains(err.Error(), "### Before") {
		t.Fatalf("a change with nothing to step through: %v", err)
	}
	stepped := "# The first commit\n\nIt adds a.\n\n## What a counts\n\nIt counts lines.\n\n### Before\nThere was nothing.\n\n```diff\n- nothing at all was here before\n- nothing at all was here before\n```\n\n### After\nNow there is a.\n\n[a.go](pkg/a.go#L3-L8)\n\n## Try: a line is added\n\n### Question\nWhat is counted?\n\n### Hint 1\nThink.\n\n### Hint 2\nLook.\n\n### Answer\nLines.\n"
	if _, note, err := s.write(writeIn{Tab: "first", Markdown: stepped}); err != nil {
		t.Fatal(note, err)
	}
	if got := Answers(s.root); len(got) != 1 || got[0].Change != s.sub.Commit {
		t.Errorf("kept: %+v", got)
	}
}

func TestAnAcceptedDocumentIsKeptOutsideTheRepository(t *testing.T) {
	s := repo(t)
	s.reading(readIn{Path: "pkg/a.go"})
	if _, note, err := s.write(writeIn{Tab: "A Work", Markdown: "---\ntab: theirs\n---\n# How a works\n\nIt counts.\n\n[a.go](pkg/a.go#L3-L8)\n"}); err != nil {
		t.Fatal(note, err)
	}
	if s.out == nil || s.out.Doc.Tab != "a-work" || s.out.Doc.At == "" {
		t.Fatalf("written: %+v", s.out)
	}
	saved := Answers(s.root)
	if len(saved) != 1 || saved[0].Doc.Title != "How a works" || saved[0].Doc.Tab != "a-work" || saved[0].Asked != "how does a work?" || saved[0].Change != "" {
		t.Errorf("saved: %+v", saved)
	}
	if out, _ := git(s.root, "status", "--porcelain", "--ignored"); strings.TrimSpace(out) != "!! .env" {
		t.Errorf("the repository changed: %q", out)
	}
}
