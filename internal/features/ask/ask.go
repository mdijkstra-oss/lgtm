// Package ask is having a model write a document about a repository: to
// answer a question, or to explain a commit or what is not committed yet.
// It is also the documents written before, and the history to pick from.
package ask

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"lgtm/internal/kit/markdown"
)

// The document built in shows every kind of block and says what each is
// for. It was written by hand about this repository at the commit it names.
//
//go:embed docs/*.md
var files embed.FS

var order = []string{"blocks"}

// Root gives the repository the working directory is in, or the working
// directory when it is in none.
func Root() string {
	root, _ := os.Getwd()
	for dir := root; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		if dir == filepath.Dir(dir) {
			return root
		}
	}
}

// reader reads the files a document points at from that repository.
func reader(root string) func(string) ([]byte, error) {
	return func(path string) ([]byte, error) { return os.ReadFile(filepath.Join(root, path)) }
}

// Builtin gives the documents built in. The code they point at is read
// from root.
func Builtin(root string) []markdown.Doc {
	var out []markdown.Doc
	for _, name := range order {
		src, err := files.ReadFile("docs/" + name + ".md")
		if err == nil {
			var a markdown.Doc
			if a, err = markdown.Parse(string(src), reader(root)); err == nil {
				out = append(out, a)
				continue
			}
		}
		panic(fmt.Sprintf("docs/%s.md: %v", name, err))
	}
	return out
}
