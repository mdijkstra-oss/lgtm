// Package storytest checks a component's stories: each one has to fill its
// declared size exactly and match the snapshot stored under testdata.
package storytest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"strigil/internal/kit/story"
)

// UpdateEnv, when set, makes Check rewrite the snapshots instead of comparing.
const UpdateEnv = "STORY_UPDATE"

func Check(t *testing.T, c story.Component) {
	t.Helper()
	if c.Name == "" || strings.TrimSpace(c.About) == "" {
		t.Errorf("component %q needs a name and an about.md", c.Name)
	}
	if len(c.Stories) == 0 {
		t.Errorf("component %q has no stories", c.Name)
	}
	seen := map[string]bool{}
	for _, s := range c.Stories {
		if s.Name == "" || s.Note == "" || s.W <= 0 || s.H <= 0 || s.Render == nil {
			t.Errorf("story %q needs a name, a note, a size and a Render", s.Name)
			continue
		}
		if seen[s.Name] {
			t.Errorf("story %q is listed twice", s.Name)
		}
		seen[s.Name] = true
		t.Run(strings.ReplaceAll(s.Name, " ", "_"), func(t *testing.T) {
			out := s.Render(s.W, s.H)
			lines := strings.Split(out, "\n")
			if len(lines) != s.H {
				t.Errorf("drew %d rows, want %d", len(lines), s.H)
			}
			for i, line := range lines {
				if w := lipgloss.Width(line); w != s.W {
					t.Errorf("row %d is %d cells wide, want %d", i, w, s.W)
				}
			}
			compare(t, filepath.Join("testdata", strings.ReplaceAll(s.Name, " ", "_")+".golden"), out)
		})
	}
}

func compare(t *testing.T, file, out string) {
	t.Helper()
	if os.Getenv(UpdateEnv) != "" {
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(out), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("no snapshot yet; run the tests with %s=1 to write it", UpdateEnv)
	}
	if string(want) == out {
		return
	}
	got, was := strings.Split(out, "\n"), strings.Split(string(want), "\n")
	for i := range max(len(got), len(was)) {
		var g, w string
		if i < len(got) {
			g = got[i]
		}
		if i < len(was) {
			w = was[i]
		}
		if g != w {
			t.Fatalf("differs from %s at row %d\n got: %q\nwant: %q\nrun the tests with %s=1 to accept the change", file, i, g, w, UpdateEnv)
		}
	}
}
