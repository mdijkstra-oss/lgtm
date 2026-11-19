package ask

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// mostPatch is how many lines of a change are handed over with the
// question. What is past it the model asks for.
const mostPatch = 700

var hunk = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)

// patch gives a change as the model is handed it: the diff, with every
// line that is still in the file numbered as the file numbers it now.
// Those lines count as read, so the model can link them without asking
// for them again. A file that has changed since the change was made is
// shown without numbers, and has to be read before it is linked.
//
// diff is what git printed for the change; same says whether a file is
// now as the change left it.
func (s *session) patch(diff string, same func(file string) bool) string {
	var b strings.Builder
	file, numbered, n, shown := "", false, 0, 0
	for _, l := range strings.Split(strings.TrimRight(diff, "\n"), "\n") {
		if shown >= mostPatch {
			b.WriteString("… the rest of the change is not shown: see it with the git tool.\n")
			break
		}
		switch {
		case strings.HasPrefix(l, "diff --git "):
			file = ""
			continue
		case strings.HasPrefix(l, "index "), strings.HasPrefix(l, "--- "), strings.HasPrefix(l, "new file mode"), strings.HasPrefix(l, "old mode"), strings.HasPrefix(l, "new mode"), strings.HasPrefix(l, "similarity index"):
			continue
		case strings.HasPrefix(l, "deleted file mode"):
			b.WriteString("(the file was deleted)\n")
			continue
		case strings.HasPrefix(l, "+++ "):
			file = strings.TrimPrefix(strings.TrimPrefix(l, "+++ "), "b/")
			numbered = file != "/dev/null" && s.known[file] && same(file)
			switch {
			case file == "/dev/null":
				continue
			case numbered:
				fmt.Fprintf(&b, "\n=== %s (numbered as the file is now: these lines count as read)\n", file)
			default:
				fmt.Fprintf(&b, "\n=== %s (the file has changed since, or is gone: read it before linking it)\n", file)
			}
			shown++
			continue
		}
		if m := hunk.FindStringSubmatch(l); m != nil {
			n, _ = strconv.Atoi(m[1])
			count := 1
			if m[2] != "" {
				count, _ = strconv.Atoi(m[2])
			}
			if numbered && count > 0 {
				s.mu.Lock()
				s.read[file] = append(s.read[file], [2]int{n, n + count - 1})
				s.mu.Unlock()
			}
			b.WriteString("  ⋮\n")
			continue
		}
		if file == "" || l == "" {
			if l != "" {
				b.WriteString(l + "\n")
				shown++
			}
			continue
		}
		switch l[0] {
		case '-':
			fmt.Fprintf(&b, "       - %s\n", l[1:])
		case '+', ' ':
			if numbered {
				fmt.Fprintf(&b, "%5d  %c %s\n", n, l[0], l[1:])
			} else {
				fmt.Fprintf(&b, "       %c %s\n", l[0], l[1:])
			}
			n++
		default:
			b.WriteString(l + "\n")
		}
		shown++
	}
	return b.String()
}
