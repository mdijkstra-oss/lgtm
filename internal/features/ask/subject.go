package ask

import (
	"strconv"
	"strings"
	"time"
)

// Subject is what a document is asked for: a question, a commit to
// explain, or the changes that are not committed yet. One of them is set.
type Subject struct {
	Question string
	Commit   string // its short hash
	About    string // that commit's subject line
	Pending  bool
}

// Title is the subject as a page names it.
func (s Subject) Title() string {
	switch {
	case s.Pending:
		return "Uncommitted changes"
	case s.Commit != "":
		return strings.TrimSpace("Commit " + s.Commit + ": " + s.About)
	}
	return s.Question
}

// Change names the change a subject is, as a kept document records it;
// empty for a question.
func (s Subject) Change() string {
	if s.Pending {
		return "uncommitted"
	}
	return s.Commit
}

// Named reads what was typed on the command line: one word that git takes
// for a commit is that commit, and anything else is a question.
func Named(root, text string) Subject {
	text = strings.TrimSpace(text)
	if text != "" && !strings.ContainsAny(text, " \t") && !strings.HasPrefix(text, "-") {
		if _, err := git(root, "rev-parse", "--verify", "--quiet", text+"^{commit}"); err == nil {
			if out, err := git(root, "log", "-1", "--format=%h%x1f%s", text); err == nil {
				hash, about, _ := strings.Cut(strings.TrimSpace(out), "\x1f")
				return Subject{Commit: hash, About: about}
			}
		}
	}
	return Subject{Question: text}
}

// Commit is one commit of the history, as a list shows it.
type Commit struct {
	Hash, Subject, Author string
	Day                   string // the day it was made, as a list heads it: Today, Friday, 3 November
}

// day names a date for the heading over its commits: the nearer it is,
// the less it takes to say.
func day(date string, now time.Time) string {
	t, err := time.ParseInLocation("2006-01-02", date, now.Location())
	if err != nil {
		return date
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch days := int(today.Sub(t).Hours()/24 + 0.5); {
	case days <= 0:
		return "Today"
	case days == 1:
		return "Yesterday"
	case days < 7:
		return t.Format("Monday")
	case t.Year() == now.Year():
		return t.Format("2 January")
	}
	return t.Format("2 January 2006")
}

// History gives the latest commits of the repository at root, the newest
// first, and none where there is no repository.
func History(root string, most int) []Commit {
	out, err := git(root, "log", "-n", strconv.Itoa(most), "--date=format:%Y-%m-%d", "--format=%h%x1f%s%x1f%an%x1f%cd")
	if err != nil {
		return nil
	}
	var all []Commit
	now := time.Now()
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if p := strings.Split(l, "\x1f"); len(p) == 4 {
			all = append(all, Commit{Hash: p[0], Subject: p[1], Author: p[2], Day: day(p[3], now)})
		}
	}
	return all
}

// Uncommitted counts the files that differ from the last commit, new ones
// included.
func Uncommitted(root string) int {
	out, err := git(root, "status", "--porcelain")
	if err != nil || strings.TrimSpace(out) == "" {
		return 0
	}
	return len(strings.Split(strings.TrimRight(out, "\n"), "\n"))
}

// Branch names the branch the repository at root is on.
func Branch(root string) string {
	out, err := git(root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// Sizes of a change, by the lines it adds and removes. A question has none.
const (
	noChange = iota
	tiny
	small
	medium
	large
)

// limits is what a change of each size may have: sections before the
// practice, and words of text. The model is told a little less than the
// checker allows, so that a document written to the number passes.
var limits = map[int][2]int{tiny: {3, 190}, small: {3, 260}, medium: {3, 420}, large: {4, 600}}

// sized reads how big a change is from git's numstat.
func sized(numstat string) (size, lines, files int) {
	for _, l := range strings.Split(strings.TrimSpace(numstat), "\n") {
		f := strings.Fields(l)
		if len(f) < 3 {
			continue
		}
		a, _ := strconv.Atoi(f[0])
		r, _ := strconv.Atoi(f[1])
		lines, files = lines+a+r, files+1
	}
	switch {
	case lines < 25:
		return tiny, lines, files
	case lines < 60:
		return small, lines, files
	case lines < 400:
		return medium, lines, files
	}
	return large, lines, files
}

// told says the size of a change as the model is told it.
func told(size, lines, files int) string {
	name := map[int]string{tiny: "small change, under 25 lines", small: "small change", medium: "medium change", large: "large change"}[size]
	limit := limits[size]
	return "This is a " + name + ": " + strconv.Itoa(lines) + " lines added and removed, in " + strconv.Itoa(files) + " files. " +
		"Its document has at most " + strconv.Itoa(limit[0]) + " sections before the questions, and at most " + strconv.Itoa(limit[1]-30) + " words of text in them, code not counted: " +
		"an opening of three sentences under 60 words, and steps under 70 words each. Count before you hand it in."
}

// asking is the first thing the model is told about a subject: what is
// wanted, and for a change the size of it, so that it starts from the
// change and not from the repository.
func (s *session) asking() string {
	capped := func(out string, most int) string {
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) > most {
			lines = append(lines[:most], "…")
		}
		return strings.Join(lines, "\n")
	}
	switch {
	case s.sub.Pending:
		status, _ := git(s.root, "status", "--short")
		stat, _ := git(s.root, "diff", "HEAD", "--numstat")
		diff, _ := git(s.root, "diff", "HEAD", "--no-ext-diff", "--no-color")
		var lines, files int
		s.size, lines, files = sized(stat)
		return "Explain the changes in this repository that are not committed yet, so that the reader understands them quickly, before reading the diff: what the code did, what it does now, and how the new code works. Explain; do not review. This is a question about a change. " + told(s.size, lines, files) + "\n\n" +
			"git status --short:\n\n" + capped(status, 80) + "\n\n" +
			"git diff HEAD --numstat (added, removed, file):\n\n" + capped(stat, 80) + "\n\n" +
			"A file marked ?? is new and in no diff: read it.\n\n" +
			"The change itself. A line with - was removed; a line with + was added. Read round it only where you need to know what the changed code is called from or what it calls.\n" +
			s.patch(diff, func(string) bool { return true })
	case s.sub.Commit != "":
		head, _ := git(s.root, "show", "--no-patch", "--format=%h %an, %ad%n%n%B", "--date=short", s.sub.Commit)
		stat, _ := git(s.root, "show", "--numstat", "--format=", s.sub.Commit)
		diff, _ := git(s.root, "show", "--format=", "--no-ext-diff", "--no-color", s.sub.Commit)
		// A file is as the commit left it when nothing differs between
		// the commit and what is on disk.
		same := func(file string) bool {
			_, err := git(s.root, "diff", "--quiet", s.sub.Commit, "--", file)
			return err == nil
		}
		var lines, files int
		s.size, lines, files = sized(stat)
		return "Explain commit " + s.sub.Commit + ", so that the reader understands it quickly, before reading the diff: what the code did, what it does now, and how the new code works. Explain; do not review. This is a question about a change. " + told(s.size, lines, files) + "\n\n" +
			capped(head, 40) + "\n\n" +
			"git show --numstat (added, removed, file):\n\n" + capped(stat, 80) + "\n\n" +
			"The change itself. A line with - was removed; a line with + was added. Links are to the files as they are now. Read round the change only where you need to know what the changed code is called from or what it calls.\n" +
			s.patch(diff, same)
	}
	return "The question:\n\n" + s.sub.Question
}
