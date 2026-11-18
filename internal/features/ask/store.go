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

// Kept is the folder the documents written for the repository at root
// are kept in: one per repository under the machine's temporary folder,
// wherever this system has that, named for the repository and its path.
// Nothing lgtm writes shows up in anyone's working tree, and the system
// clears it away in its own time.
func Kept(root string) string {
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
func save(root, tab, question, commit, src string) (string, error) {
	dir := Kept(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	name := slug(question, 48)
	if name == "" {
		name = "document"
	}
	path := filepath.Join(dir, name+".md")
	for n := 2; exists(path); n++ {
		path = filepath.Join(dir, name+"-"+strconv.Itoa(n)+".md")
	}
	head := "---\ntab: " + tab + "\nat: " + commit + "\nasked: " + strings.Join(strings.Fields(question), " ") + "\n---\n\n"
	return path, os.WriteFile(path, []byte(head+src), 0o644)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// kept gives the files of the documents written for root, the newest first.
func kept(root string) []string {
	paths, _ := filepath.Glob(filepath.Join(Kept(root), "*.md"))
	when := map[string]int64{}
	for _, p := range paths {
		if info, err := os.Stat(p); err == nil {
			when[p] = info.ModTime().UnixNano()
		}
	}
	sort.Slice(paths, func(i, j int) bool { return when[paths[i]] > when[paths[j]] })
	return paths
}

// Saved gives the documents written for the repository at root, the newest
// first. One that no longer reads as a document is left out.
func Saved(root string) []markdown.Doc {
	var out []markdown.Doc
	for _, p := range kept(root) {
		src, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if d, err := markdown.Parse(string(src), reader(root)); err == nil {
			out = append(out, d)
		}
	}
	return out
}
