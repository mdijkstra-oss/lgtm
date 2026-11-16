package story_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Packages under kit that draw nothing on their own.
var exempt = map[string]bool{"story": true, "text": true, "theme": true}

// Every package that draws has to describe itself in the story format.
func TestEveryComponentHasStories(t *testing.T) {
	var dirs []string
	kit, err := os.ReadDir("..")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range kit {
		if e.IsDir() && !exempt[e.Name()] {
			dirs = append(dirs, filepath.Join("..", e.Name()))
		}
	}
	ui, err := filepath.Glob(filepath.Join("..", "..", "features", "*", "ui"))
	if err != nil {
		t.Fatal(err)
	}
	dirs = append(dirs, ui...)
	if len(dirs) == 0 {
		t.Fatal("found no components to check")
	}

	for _, dir := range dirs {
		name := strings.TrimLeft(filepath.ToSlash(dir), "./")
		src, err := os.ReadFile(filepath.Join(dir, "stories.go"))
		switch {
		case err != nil:
			t.Errorf("%s: no stories.go", name)
		case !strings.HasPrefix(string(src), "//go:build gallery\n"):
			t.Errorf("%s: stories.go must start with //go:build gallery, so it stays out of the shipped binary", name)
		case !strings.Contains(string(src), ") story.Component"):
			t.Errorf("%s: stories.go must have func Stories(...) story.Component", name)
		}
		if about, err := os.ReadFile(filepath.Join(dir, "about.md")); err != nil || strings.TrimSpace(string(about)) == "" {
			t.Errorf("%s: no about.md", name)
		}
		test, err := os.ReadFile(filepath.Join(dir, "stories_test.go"))
		if err != nil || !strings.Contains(string(test), "storytest.Check(") {
			t.Errorf("%s: no stories_test.go calling storytest.Check", name)
		}
	}
}
