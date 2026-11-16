// Package search finds literal text in the files of a folder.
package search

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const maxFileSize = 2 << 20

type Query struct {
	Text       string
	Dir        string
	IgnoreCase bool
}

type Match struct {
	Path   string `json:"path"`
	Line   int    `json:"line"`
	Col    int    `json:"col"`
	Length int    `json:"length"`
	Text   string `json:"text"`
}

type Matcher struct{ re *regexp.Regexp }

func (q Query) Matcher() Matcher {
	expr := regexp.QuoteMeta(q.Text)
	if q.IgnoreCase {
		expr = "(?i)" + expr
	}
	return Matcher{re: regexp.MustCompile(expr)}
}

// Find returns the byte ranges of every occurrence in line.
func (m Matcher) Find(line string) [][2]int {
	found := m.re.FindAllStringIndex(line, -1)
	out := make([][2]int, len(found))
	for i, f := range found {
		out[i] = [2]int{f[0], f[1]}
	}
	return out
}

// Run walks q.Dir and calls emit for every occurrence, in path then line order.
// Match.Path is relative to q.Dir; Line and Col are 1-based, Col and Length in bytes.
func Run(ctx context.Context, q Query, emit func(Match)) error {
	root := q.Dir
	if root == "" {
		root = "."
	}
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a folder", root)
	}
	if q.Text == "" {
		return nil
	}
	m := q.Matcher()
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if cerr := ctx.Err(); cerr != nil {
			return cerr
		}
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		lines, err := readLines(path)
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		for i, line := range lines {
			for _, r := range m.Find(line) {
				emit(Match{Path: rel, Line: i + 1, Col: r[0] + 1, Length: r[1] - r[0], Text: line})
			}
		}
		return nil
	})
}

// Load returns the lines of the file at rel inside dir.
func Load(dir, rel string) ([]string, error) {
	return readLines(filepath.Join(dir, rel))
}

func readLines(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("%s is too large", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if bytes.IndexByte(data[:min(len(data), 8000)], 0) >= 0 {
		return nil, fmt.Errorf("%s is not text", path)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}
