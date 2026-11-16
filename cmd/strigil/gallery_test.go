//go:build gallery

package main

import (
	"io/fs"
	"path/filepath"
	"testing"

	"strigil/internal/kit/theme"
)

func TestEveryStoriesFileIsInTheGallery(t *testing.T) {
	files := 0
	err := filepath.WalkDir(filepath.Join("..", "..", "internal"), func(_ string, d fs.DirEntry, err error) error {
		if err == nil && d.Name() == "stories.go" {
			files++
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	names := map[string]bool{}
	for _, c := range components(theme.Mocha()) {
		if names[c.Name] {
			t.Errorf("component %q is listed twice", c.Name)
		}
		names[c.Name] = true
	}
	if len(names) != files {
		t.Errorf("the gallery lists %d components but there are %d stories.go files", len(names), files)
	}
}
