package ask

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"charm.land/fantasy"

	"lgtm/internal/kit/markdown"
)

// What one call may give back, so a careless one does not fill the context.
const (
	mostEntries = 400
	mostMatches = 200
	mostLines   = 250
	mostGit     = 300
	mostBytes   = 1 << 20
)

// session is one question being answered: the repository, what the model
// has read of it so far, and the document once it is accepted.
type session struct {
	root  string
	sub   Subject
	size  int // of the change, when the subject is one
	at    string
	files []string // what git does not ignore, sorted
	known map[string]bool
	on    func(Event)

	mu     sync.Mutex
	calls  int
	read   map[string][][2]int
	draft  string // the document as it was last handed in
	tab    string
	nagged bool // it has been asked once to shorten the document
	out    *Written
}

func open(root string, sub Subject, on func(Event)) *session {
	s := &session{root: root, sub: sub, on: on, known: map[string]bool{}, read: map[string][][2]int{}}
	if out, err := git(root, "ls-files", "-co", "--exclude-standard", "-z"); err == nil {
		s.files = strings.Split(strings.TrimRight(out, "\x00"), "\x00")
	} else {
		filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() && p != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			if !d.IsDir() {
				rel, _ := filepath.Rel(root, p)
				s.files = append(s.files, filepath.ToSlash(rel))
			}
			return nil
		})
	}
	kept := s.files[:0]
	for _, f := range s.files {
		if f != "" {
			kept = append(kept, f)
			s.known[f] = true
		}
	}
	s.files = kept
	sort.Strings(s.files)
	s.at = at(root)
	return s
}

// file gives a file of the repository as the model named it. Only what
// git does not ignore can be read: what is ignored is often a secret.
func (s *session) file(name string) (string, []byte, error) {
	name = path.Clean(strings.TrimPrefix(filepath.ToSlash(name), "./"))
	if !s.known[name] {
		return name, nil, fmt.Errorf("there is no file %s: list the folder to see what there is", name)
	}
	src, err := os.ReadFile(filepath.Join(s.root, name))
	if err != nil {
		return name, nil, fmt.Errorf("cannot read %s", name)
	}
	return name, src, nil
}

func (s *session) seen(file string, a, b int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
next:
	for n := a; n <= b; n++ {
		for _, r := range s.read[file] {
			if n >= r[0] && n <= r[1] {
				continue next
			}
		}
		return false
	}
	return true
}

// from names the file read that has every one of those lines in it.
func (s *session) from(lines []string) string {
	s.mu.Lock()
	files := make([]string, 0, len(s.read))
	for f := range s.read {
		files = append(files, f)
	}
	s.mu.Unlock()
	sort.Strings(files)
next:
	for _, f := range files {
		src, err := os.ReadFile(filepath.Join(s.root, f))
		if err != nil {
			continue
		}
		for _, l := range lines {
			if !bytes.Contains(src, []byte(l)) {
				continue next
			}
		}
		return f
	}
	return ""
}

// call runs one tool: it tells whoever is watching, counts it against the
// budget, and tells the model how much of the budget is left.
func (s *session) call(tool, what string, run func() (said, note string, err error)) (fantasy.ToolResponse, error) {
	s.mu.Lock()
	s.calls++
	id, left := s.calls, budget-s.calls
	s.mu.Unlock()
	s.on(Event{ID: id, Tool: tool, What: what})
	final := tool == "write_document" || tool == "fix_document"
	if left < 0 && !final {
		s.on(Event{ID: id, Tool: tool, What: what, Done: true, Failed: true, Note: "out of steps"})
		return fantasy.NewTextErrorResponse("No steps left. Call write_document now with the whole document, from what you have read."), nil
	}
	said, note, err := run()
	tail := fmt.Sprintf("\n\n[steps left: %d]", max(0, left))
	if left <= 5 {
		tail = fmt.Sprintf("\n\n[steps left: %d. Hand the document in with write_document.]", max(0, left))
	}
	if err != nil {
		s.on(Event{ID: id, Tool: tool, What: what, Done: true, Failed: true, Note: note, Said: err.Error()})
		return fantasy.NewTextErrorResponse(err.Error() + tail), nil
	}
	s.on(Event{ID: id, Tool: tool, What: what, Done: true, Note: note})
	if final {
		return fantasy.ToolResponse{Type: "text", Content: said, StopTurn: true}, nil
	}
	return fantasy.NewTextResponse(said + tail), nil
}

type listIn struct {
	Path  string `json:"path,omitempty" description:"The folder to list, from the repository root. Leave out for the root."`
	Depth int    `json:"depth,omitempty" description:"How many levels down to show. 2 when left out."`
}

type grepIn struct {
	Pattern string `json:"pattern" description:"A regular expression in RE2 syntax, matched against each line. Case matters."`
	Glob    string `json:"glob,omitempty" description:"Which files to search: a folder (internal/kit), or a pattern (*.go, internal/**/*.md). Every file when left out."`
	Max     int    `json:"max,omitempty" description:"The most matches to give. 50 when left out."`
}

type readIn struct {
	Path string `json:"path" description:"The file, from the repository root."`
	From int    `json:"from,omitempty" description:"The first line to read. 1 when left out."`
	To   int    `json:"to,omitempty" description:"The last line to read. 150 lines from the first when left out."`
}

type gitIn struct {
	Command string `json:"command" enum:"log,show,diff,blame" description:"The git command."`
	Args    string `json:"args,omitempty" description:"Its arguments, as on a command line. For example: --oneline -20 -- internal/kit"`
}

type fixIn struct {
	Old string `json:"old" description:"Text from the document as you last handed it in, copied exactly, that stands in it only once."`
	New string `json:"new" description:"What to put in its place. Empty takes the text out."`
}

type writeIn struct {
	Tab      string `json:"tab" description:"One or two lowercase words that name the document on its tab, such as: parsing"`
	Markdown string `json:"markdown" description:"The whole document, starting with its # title."`
}

func (s *session) tools() []fantasy.AgentTool {
	return []fantasy.AgentTool{
		fantasy.NewAgentTool("list",
			"The files and folders under a folder, leaving out what .gitignore leaves out. A file is shown with how many lines it has, a folder with how many files are in it; no contents. Use it to find your way and to check that a path is real before reading it. If it is cut short, list a folder further in.",
			func(_ context.Context, in listIn, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
				dir := path.Clean(strings.TrimPrefix(filepath.ToSlash(in.Path), "./"))
				if dir == "" || dir == "/" {
					dir = "."
				}
				return s.call("list", dir, func() (string, string, error) { return s.list(dir, in.Depth) })
			}),
		fantasy.NewAgentTool("grep",
			"Searches the contents of files for a regular expression and gives each match as path:line:text. Use it to find where a name is defined (func Name, type Name) and where it is used. It finds where; it does not show enough to understand the code, and its line numbers may not be put in a link: read the lines first.",
			func(_ context.Context, in grepIn, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
				what := in.Pattern
				if in.Glob != "" {
					what += " in " + in.Glob
				}
				return s.call("grep", what, func() (string, string, error) { return s.grep(in) })
			}),
		fantasy.NewAgentTool("read",
			"Lines of a file, each with its real line number before it; at most 250 at a time. The first line of the result says how many lines the file has. These are the only line numbers that may go in a link, and only lines you have read may be linked.",
			func(_ context.Context, in readIn, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
				what := in.Path
				if in.From > 1 || in.To > 0 {
					what = fmt.Sprintf("%s %d–%d", in.Path, max(1, in.From), last(in))
				}
				return s.call("read", what, func() (string, string, error) { return s.reading(in) })
			}),
		fantasy.NewAgentTool("git",
			"Read-only git: log, show, diff or blame. Examples of args: for log, --oneline -20 -- path; for diff, --numstat main...HEAD; for show, --stat abc1234; for blame, -L 40,60 path. At most 300 lines come back. Line numbers in a diff are not line numbers of the file: read the file before linking it.",
			func(_ context.Context, in gitIn, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
				return s.call("git", strings.TrimSpace(in.Command+" "+in.Args), func() (string, string, error) { return s.history(in) })
			}),
		fantasy.NewAgentTool("write_document",
			"Hands in the whole document. It is the only way to answer: nothing you say outside this call reaches the reader. It is read as the viewer will read it, and either accepted or given back with what is wrong. When it is given back, fix what is named and call this again with the whole document.",
			func(_ context.Context, in writeIn, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
				return s.call("write_document", "", func() (string, string, error) { return s.write(in) })
			}),
		fantasy.NewAgentTool("fix_document",
			"Changes one piece of the document you last handed in, and hands it in again. Use it when write_document gave the document back with a few things to fix: it is much quicker than sending the whole document. One change a call; make several calls in one turn when there are several things to fix.",
			func(_ context.Context, in fixIn, _ fantasy.ToolCall) (fantasy.ToolResponse, error) {
				return s.call("fix_document", "", func() (string, string, error) { return s.fix(in) })
			}),
	}
}

func last(in readIn) int {
	if in.To > 0 {
		return in.To
	}
	return max(1, in.From) + 149
}

func (s *session) list(dir string, depth int) (string, string, error) {
	if depth <= 0 {
		depth = 2
	}
	prefix := ""
	if dir != "." {
		prefix = dir + "/"
	}
	var names []string
	count := map[string]int{}
	for _, f := range s.files {
		rest, ok := strings.CutPrefix(f, prefix)
		if !ok {
			continue
		}
		parts := strings.Split(rest, "/")
		if len(parts) <= depth {
			names = append(names, f)
			continue
		}
		d := prefix + strings.Join(parts[:depth], "/") + "/"
		if count[d] == 0 {
			names = append(names, d)
		}
		count[d]++
	}
	if len(names) == 0 {
		return "", "nothing", fmt.Errorf("there is no folder %s, or nothing in it", dir)
	}
	sort.Strings(names)
	var b strings.Builder
	for i, n := range names {
		if i == mostEntries {
			fmt.Fprintf(&b, "… %d more: list a folder further in\n", len(names)-i)
			break
		}
		if strings.HasSuffix(n, "/") {
			fmt.Fprintf(&b, "%s  (%d files)\n", n, count[n])
		} else if src, err := os.ReadFile(filepath.Join(s.root, n)); err == nil && len(src) <= mostBytes && !bytes.Contains(src[:min(len(src), 8000)], []byte{0}) {
			fmt.Fprintf(&b, "%s  (%d lines)\n", n, bytes.Count(src, []byte("\n"))+1)
		} else {
			fmt.Fprintf(&b, "%s\n", n)
		}
	}
	return b.String(), fmt.Sprintf("%d entries", len(names)), nil
}

// globbed makes of a glob what tells whether a path matches it. A name
// with nothing wild in it is a folder, or a file.
func globbed(glob string) (func(string) bool, error) {
	glob = strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(glob)), "./")
	if glob == "" || glob == "." {
		return func(string) bool { return true }, nil
	}
	if !strings.ContainsAny(glob, "*?") {
		glob = strings.TrimSuffix(glob, "/")
		return func(p string) bool { return p == glob || strings.HasPrefix(p, glob+"/") }, nil
	}
	var re strings.Builder
	re.WriteString("^")
	if !strings.Contains(glob, "/") {
		re.WriteString("(.*/)?")
	}
	for i := 0; i < len(glob); i++ {
		switch {
		case strings.HasPrefix(glob[i:], "**/"):
			re.WriteString("(.*/)?")
			i += 2
		case strings.HasPrefix(glob[i:], "**"):
			re.WriteString(".*")
			i++
		case glob[i] == '*':
			re.WriteString("[^/]*")
		case glob[i] == '?':
			re.WriteString("[^/]")
		default:
			re.WriteString(regexp.QuoteMeta(glob[i : i+1]))
		}
	}
	re.WriteString("$")
	r, err := regexp.Compile(re.String())
	if err != nil {
		return nil, err
	}
	return r.MatchString, nil
}

func (s *session) grep(in grepIn) (string, string, error) {
	re, err := regexp.Compile(in.Pattern)
	if err != nil {
		return "", "bad pattern", fmt.Errorf("the pattern is not a regular expression: %v", err)
	}
	match, err := globbed(in.Glob)
	if err != nil {
		return "", "bad glob", fmt.Errorf("the glob cannot be read: %v", err)
	}
	most := in.Max
	if most <= 0 {
		most = 50
	}
	most = min(most, mostMatches)
	var b strings.Builder
	found, files := 0, 0
	for _, f := range s.files {
		if !match(f) {
			continue
		}
		src, err := os.ReadFile(filepath.Join(s.root, f))
		if err != nil || len(src) > mostBytes || bytes.Contains(src[:min(len(src), 8000)], []byte{0}) {
			continue
		}
		hit := false
		for n, l := range strings.Split(string(src), "\n") {
			if !re.MatchString(l) {
				continue
			}
			found, hit = found+1, true
			if found <= most {
				if l = strings.TrimSpace(l); len(l) > 200 {
					l = l[:200] + "…"
				}
				fmt.Fprintf(&b, "%s:%d:%s\n", f, n+1, l)
			}
		}
		if hit {
			files++
		}
	}
	if found == 0 {
		return "No match.", "no match", nil
	}
	if found > most {
		fmt.Fprintf(&b, "… %d more matches not shown: narrow the pattern or the glob\n", found-most)
	}
	return b.String(), fmt.Sprintf("%d in %d files", found, files), nil
}

func (s *session) reading(in readIn) (string, string, error) {
	name, src, err := s.file(in.Path)
	if err != nil {
		return "", "no such file", err
	}
	if bytes.Contains(src[:min(len(src), 8000)], []byte{0}) {
		return "", "not text", fmt.Errorf("%s is not text", name)
	}
	lines := strings.Split(strings.TrimRight(string(src), "\n"), "\n")
	from := max(1, in.From)
	if from > len(lines) {
		return "", "past the end", fmt.Errorf("%s has %d lines", name, len(lines))
	}
	to := min(last(in), from+mostLines-1, len(lines))
	var b strings.Builder
	fmt.Fprintf(&b, "%s (%d lines), lines %d to %d:\n", name, len(lines), from, to)
	for n := from; n <= to; n++ {
		fmt.Fprintf(&b, "%5d  %s\n", n, lines[n-1])
	}
	s.mu.Lock()
	s.read[name] = append(s.read[name], [2]int{from, to})
	s.mu.Unlock()
	return b.String(), fmt.Sprintf("%d lines", to-from+1), nil
}

func (s *session) history(in gitIn) (string, string, error) {
	switch in.Command {
	case "log", "show", "diff", "blame":
	default:
		return "", "refused", fmt.Errorf("git %s is not one of log, show, diff, blame", in.Command)
	}
	args := []string{in.Command}
	if in.Command != "blame" {
		args = append(args, "--no-ext-diff", "--no-textconv", "--no-color")
	}
	for _, a := range strings.Fields(in.Args) {
		a = strings.Trim(a, "'\"")
		// These write a file, or read one outside the repository.
		if strings.HasPrefix(a, "--output") || a == "--no-index" || strings.HasPrefix(a, "--ext-diff") || strings.HasPrefix(a, "--textconv") {
			return "", "refused", fmt.Errorf("%s is not allowed", a)
		}
		args = append(args, a)
	}
	if in.Command == "log" && !strings.Contains(in.Args, "-n") && !regexp.MustCompile(`(^|\s)-\d+`).MatchString(in.Args) {
		args = append(args[:1], append([]string{"-50"}, args[1:]...)...)
	}
	out, err := git(s.root, args...)
	if err != nil {
		return "", "failed", fmt.Errorf("git %s: %v", in.Command, err)
	}
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	note := fmt.Sprintf("%d lines", len(lines))
	if len(lines) > mostGit {
		lines = append(lines[:mostGit], fmt.Sprintf("… %d more lines not shown: narrow it to a path or a commit", len(lines)-mostGit))
	}
	if out == "" {
		return "Nothing.", "nothing", nil
	}
	return strings.Join(lines, "\n"), note, nil
}

func (s *session) write(in writeIn) (string, string, error) {
	src := strings.TrimSpace(in.Markdown)
	// The head is ours to write.
	if strings.HasPrefix(src, "---") {
		if _, rest, ok := strings.Cut(src[3:], "\n---"); ok {
			src = strings.TrimSpace(rest)
		}
	}
	s.mu.Lock()
	s.tab = in.Tab
	s.mu.Unlock()
	return s.hand(dedent(src))
}

func (s *session) fix(in fixIn) (string, string, error) {
	s.mu.Lock()
	src := s.draft
	s.mu.Unlock()
	switch n := strings.Count(src, in.Old); {
	case src == "":
		return "", "nothing to fix", fmt.Errorf("there is no document yet: hand it in with write_document first")
	case in.Old == "" || n == 0:
		return "", "not found", fmt.Errorf("the document does not have that text: copy it exactly as you wrote it")
	case n > 1:
		return "", "found twice", fmt.Errorf("that text is in the document %d times: give more of it, so that it is there once", n)
	}
	return s.hand(strings.TrimSpace(strings.Replace(src, in.Old, in.New, 1)))
}

// hand reads a document the model handed in, and keeps it if nothing is
// wrong with it.
func (s *session) hand(src string) (string, string, error) {
	s.mu.Lock()
	s.draft = src
	tab := s.tab
	s.mu.Unlock()
	size := s.size
	if size == noChange && s.sub.Change() != "" {
		size = medium
	}
	doc, hard, soft := check(src, s.root, size, s.seen, s.from)
	s.mu.Lock()
	// Too long is asked about once. A second time it is let through: the
	// reader is better off with a long document than with none.
	if s.nagged && len(hard) == 0 {
		soft = nil
	}
	s.nagged = s.nagged || len(soft) > 0
	s.mu.Unlock()
	if wrong := append(hard, soft...); len(wrong) > 0 {
		var b strings.Builder
		b.WriteString("Not accepted. Fix these with fix_document, one change a call and several calls in one turn. Send the whole document again with write_document only if most of it has to change.\n")
		for i, w := range wrong {
			fmt.Fprintf(&b, "%d. %s\n", i+1, w)
		}
		note := "1 thing to fix"
		if len(wrong) > 1 {
			note = fmt.Sprintf("%d things to fix", len(wrong))
		}
		return "", note, fmt.Errorf("%s", b.String())
	}
	if err := s.keep(doc, src, tab); err != nil {
		return "", "not saved", fmt.Errorf("the document could not be saved: %v", err)
	}
	return "Accepted.", "accepted", nil
}

// keep saves a document as the answer.
func (s *session) keep(doc markdown.Doc, src, tab string) error {
	if tab = slug(tab, 18); tab == "" {
		tab = slug(doc.Title, 18)
	}
	file, err := save(s.root, tab, s.sub, s.at, src+"\n")
	if err != nil {
		return err
	}
	doc.Tab, doc.At = tab, s.at
	s.mu.Lock()
	s.out = &Written{Doc: doc, File: file, Source: src + "\n"}
	s.mu.Unlock()
	return nil
}

// settle is for a model that ran out of turns with a document handed in
// and never accepted. If it reads as a document the reader gets it as it
// is: something was still wrong with it, and it is still worth more to
// them than an apology.
func (s *session) settle() bool {
	s.mu.Lock()
	src, tab := s.draft, s.tab
	s.mu.Unlock()
	if src == "" {
		return false
	}
	doc, err := markdown.Parse(src, reader(s.root))
	return err == nil && s.keep(doc, src, tab) == nil
}
