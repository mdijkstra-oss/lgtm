package ask

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"lgtm/internal/kit/markdown"
)

var unsafe = regexp.MustCompile(`[^a-z0-9]+`)

// slug makes a name fit for a file and a tab.
func slug(s string, most int) string {
	s = strings.Trim(unsafe.ReplaceAllString(strings.ToLower(s), "-"), "-")
	if len(s) > most {
		s = strings.Trim(s[:most], "-")
	}
	return s
}

// Folder is where the documents written for the repository at root are
// kept: one per repository under the machine's temporary folder, wherever
// this system has that, named for the repository and its path. Nothing
// lgtm writes shows up in anyone's working tree, and the system clears it
// away in its own time.
func Folder(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = root
	}
	sum := sha256.Sum256([]byte(abs))
	return filepath.Join(os.TempDir(), "lgtm", slug(filepath.Base(abs), 32)+"-"+hex.EncodeToString(sum[:4]))
}

// git runs a git command in the repository at root.
func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"--no-pager", "-C", root}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_PAGER=cat", "GIT_OPTIONAL_LOCKS=0")
	var out, bad bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &bad
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s", strings.TrimSpace(bad.String()))
	}
	return out.String(), nil
}

// at names the commit the repository at root is at.
func at(root string) string {
	if out, err := git(root, "rev-parse", "--short", "HEAD"); err == nil {
		return strings.TrimSpace(out)
	}
	return "the working tree"
}

// save keeps a document written for root and gives the file it went to.
func save(root, tab string, sub Subject, commit, src string) (string, error) {
	dir := Folder(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := slug(sub.Title(), 48)
	if name == "" {
		name = "document"
	}
	path := filepath.Join(dir, name+".md")
	for n := 2; exists(path); n++ {
		path = filepath.Join(dir, name+"-"+strconv.Itoa(n)+".md")
	}
	head := "---\ntab: " + tab + "\nat: " + commit + "\nasked: " + strings.Join(strings.Fields(sub.Title()), " ") + "\n"
	if c := sub.Change(); c != "" {
		head += "change: " + c + "\n"
	}
	return path, os.WriteFile(path, []byte(head+"---\n\n"+src), 0o644)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Answer is a document written earlier for this repository.
type Answer struct {
	Doc    markdown.Doc
	Asked  string // what it was asked for, as the page named it
	Change string // the commit it explains, or "uncommitted"; empty for a question
	When   string // how long ago
}

// ago says how long ago something was, as short as a list wants it.
func ago(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return strconv.Itoa(int(d.Minutes())) + " min ago"
	case d < 24*time.Hour:
		return strconv.Itoa(int(d.Hours())) + " h ago"
	}
	return strconv.Itoa(int(d.Hours()/24)) + " d ago"
}

// Answers gives the documents written for the repository at root, the
// newest first. One that no longer reads as a document is left out.
func Answers(root string) []Answer {
	paths, _ := filepath.Glob(filepath.Join(Folder(root), "*.md"))
	when := map[string]time.Time{}
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil {
			when[p] = info.ModTime()
		}
	}
	sort.Slice(paths, func(i, j int) bool { return when[paths[i]].After(when[paths[j]]) })
	var out []Answer
	for _, p := range paths {
		src, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		d, err := markdown.Parse(string(src), reader(root))
		if err != nil {
			continue
		}
		a := Answer{Doc: d, Asked: d.Title, When: ago(time.Since(when[p]))}
		if head, _, ok := strings.Cut(strings.TrimPrefix(string(src), "---\n"), "\n---\n"); ok {
			for _, l := range strings.Split(head, "\n") {
				switch k, v, _ := strings.Cut(l, ":"); strings.TrimSpace(k) {
				case "asked":
					a.Asked = strings.TrimSpace(v)
				case "change":
					a.Change = strings.TrimSpace(v)
				}
			}
		}
		out = append(out, a)
	}
	return out
}
