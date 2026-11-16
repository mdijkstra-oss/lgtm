package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"strigil/internal/kit"
	"strigil/internal/kit/theme"
)

func opened(t *testing.T, query string, reveal *ref) Page {
	t.Helper()
	dir := t.TempDir()
	for name, body := range map[string]string{
		"a.txt":     "one pear\ntwo\npear and pear\n",
		"sub/b.txt": "nothing\na Pear\n",
	} {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var p kit.Page = Page{th: theme.Mocha(), dir: dir, query: query}
	p = p.Resize(80, 24)
	if reveal != nil {
		p = p.Reveal(kit.Item{Data: *reveal}, 0)
	}
	p, _ = p.Update(p.Init()())
	return p.(Page)
}

func TestPageGroupsHitsByFile(t *testing.T) {
	p := opened(t, "pear", nil)

	if len(p.docs) != 2 || len(p.hits) != 4 || p.Meta() != "4" {
		t.Fatalf("%d files, %d hits, meta %q", len(p.docs), len(p.hits), p.Meta())
	}
	// file, line 1, line 3 (two hits on one row), gap, file, line 2
	if len(p.rows) != 6 || p.hits[1].row != p.hits[2].row {
		t.Errorf("rows %+v, hits %+v", p.rows, p.hits)
	}
	if p.Title() != "search · pear" {
		t.Errorf("title %q", p.Title())
	}
}

func TestRevealBeforeTheResultsAreIn(t *testing.T) {
	p := opened(t, "pear", &ref{path: filepath.Join("sub", "b.txt"), line: 1, start: 2})
	if p.sel != 3 {
		t.Errorf("selected hit %d, want 3", p.sel)
	}
}

func TestEnterOpensTheFileAndEscReturns(t *testing.T) {
	var p kit.Page = opened(t, "pear", nil)
	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !p.(Page).viewing || !strings.Contains(p.View(), "two") {
		t.Fatal("the file is not showing")
	}
	p, _ = p.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if p.(Page).viewing {
		t.Error("esc did not close the file")
	}
}

func TestSameWordInAnotherCaseIsTheSamePage(t *testing.T) {
	if pageID("Pear") != pageID("pear") {
		t.Error("ids differ")
	}
}
