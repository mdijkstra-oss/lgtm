package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// See AGENTS.md, "Folder READMEs".
func TestFolderReadmes(t *testing.T) {
	root := filepath.Join("..", "..")
	dirs := []string{"cmd", "internal"}
	for _, parent := range []string{"internal", "internal/features"} {
		for _, name := range subdirs(t, filepath.Join(root, parent)) {
			dirs = append(dirs, parent+"/"+name)
		}
	}

	for _, dir := range dirs {
		readme, err := os.ReadFile(filepath.Join(root, dir, "README.md"))
		if err != nil {
			t.Errorf("%s: no README.md", dir)
			continue
		}
		for _, name := range subdirs(t, filepath.Join(root, dir)) {
			if !strings.Contains(string(readme), "`"+name+"/") {
				t.Errorf("%s/README.md does not say what `%s/` holds", dir, name)
			}
		}
	}
}

func subdirs(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "testdata" {
			names = append(names, e.Name())
		}
	}
	return names
}
