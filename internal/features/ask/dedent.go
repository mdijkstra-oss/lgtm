package ask

import "strings"

// dedent takes the indentation the lines of a diff fence share off them.
// Removed code is copied out of the middle of a function, and seven tabs
// in front of one line push it off the page.
func dedent(src string) string {
	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "```diff" {
			continue
		}
		first := i + 1
		last := first
		for last < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[last]), "```") {
			last++
		}
		// What each line has after its + or -, and the least of it.
		least := -1
		body := func(l string) (mark, rest string) {
			if l != "" && strings.ContainsRune("+- ", rune(l[0])) {
				return l[:1], l[1:]
			}
			return "", l
		}
		for _, l := range lines[first:last] {
			_, rest := body(l)
			if strings.TrimSpace(rest) == "" {
				continue
			}
			rest = strings.ReplaceAll(rest, "\t", "    ")
			if n := len(rest) - len(strings.TrimLeft(rest, " ")); least < 0 || n < least {
				least = n
			}
		}
		for j := first; j < last && least > 0; j++ {
			mark, rest := body(lines[j])
			rest = strings.ReplaceAll(rest, "\t", "    ")
			if len(rest) >= least {
				rest = rest[least:]
			}
			// One space after the mark, as a diff has it.
			lines[j] = strings.TrimRight(mark+" "+rest, " ")
		}
		i = last
	}
	return strings.Join(lines, "\n")
}
