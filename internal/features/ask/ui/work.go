package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"lgtm/internal/features/ask"
	"lgtm/internal/kit/markdown"
	"lgtm/internal/kit/scrape"
)

// work is a question being answered: what the model has done so far, and
// what it is thinking.
type work struct {
	question string
	steps    []step
	think    string // the end of what the model is thinking
	fresh    bool   // a call came since: what it thinks next starts over
	frame    int    // of the spinner
	began    time.Time
	took     time.Duration
	err      error

	halt context.CancelFunc
	feed chan ask.Event
}

// step is one call the model made.
type step struct {
	id         int
	tool, what string
	note       string
	done, bad  bool
}

type (
	scrapeMsg struct{}
	workMsg   struct{} // the clock of the page while the model works
	eventsMsg []ask.Event
	doneMsg   struct {
		doc ask.Written
		err error
	}
)

func scrapeTick() tea.Cmd {
	return tea.Tick(time.Second/scrape.FPS, func(time.Time) tea.Msg { return scrapeMsg{} })
}

// Ten frames a second: docs/ux/README.md, section 10.
func workTick() tea.Cmd {
	return tea.Tick(time.Second/10, func(time.Time) tea.Msg { return workMsg{} })
}

// ask sets the model to work on a question.
func (m Model) ask(ctx context.Context, asked Asked) (Model, tea.Cmd) {
	ctx, halt := context.WithCancel(ctx)
	feed := make(chan ask.Event, 256)
	m.choosing = false
	m.work = &work{question: asked.Question, began: time.Now(), halt: halt, feed: feed}
	run := func() tea.Msg {
		doc, err := asked.Write(ctx, func(e ask.Event) {
			select {
			case feed <- e:
			case <-ctx.Done():
			}
		})
		close(feed)
		return doneMsg{doc, err}
	}
	return m, tea.Batch(run, listen(feed), workTick())
}

// listen waits for what the model does next, and takes with it whatever
// else has happened by then: thoughts come many to the frame.
func listen(feed chan ask.Event) tea.Cmd {
	return func() tea.Msg {
		e, ok := <-feed
		if !ok {
			return nil
		}
		all := eventsMsg{e}
		for len(all) < 128 {
			select {
			case e, ok := <-feed:
				if !ok {
					return all
				}
				all = append(all, e)
			default:
				return all
			}
		}
		return all
	}
}

// stop gives up on the question being answered, if one is.
func (m Model) stop() Model {
	if m.work == nil {
		return m
	}
	m.work.halt()
	m.work = nil
	m.choosing, m.pick = len(m.tabs) == 0, 0
	return m
}

// working is a message about the question being answered.
func (m Model) working(msg tea.Msg) (Model, tea.Cmd) {
	if m.work == nil {
		return m, nil
	}
	w := *m.work
	m.work = &w
	switch msg := msg.(type) {
	case workMsg:
		if w.err != nil {
			return m, nil
		}
		w.frame++
		w.took = time.Since(w.began)
		return m, workTick()
	case eventsMsg:
		w.steps = append([]step(nil), w.steps...)
		for _, e := range msg {
			w.take(e)
		}
		return m, listen(w.feed)
	case doneMsg:
		if msg.err != nil {
			w.err = msg.err
			return m, nil
		}
		m.work = nil
		m.known = append([]markdown.Doc{msg.doc.Doc}, m.known...)
		return m.open(msg.doc.Doc), nil
	}
	return m, nil
}

func (w *work) take(e ask.Event) {
	switch {
	case e.Think != "":
		if w.fresh {
			w.think, w.fresh = "", false
		}
		w.think += e.Think
		// Only its end is ever shown.
		if r := []rune(w.think); len(r) > 2000 {
			w.think = string(r[len(r)-1000:])
		}
	case !e.Done:
		w.steps, w.fresh = append(w.steps, step{id: e.ID, tool: e.Tool, what: e.What}), true
	default:
		for i := range w.steps {
			if w.steps[i].id == e.ID {
				w.steps[i].done, w.steps[i].bad, w.steps[i].note = true, e.Failed, e.Note
			}
		}
	}
}

// doing says what the model is at, from the call it is waiting on.
func (w work) doing() string {
	if n := len(w.steps); n > 0 && !w.steps[n-1].done {
		switch w.steps[n-1].tool {
		case "list":
			return "finding its way…"
		case "grep":
			return "searching…"
		case "read":
			return "reading…"
		case "git":
			return "reading the history…"
		case "write_document", "fix_document":
			return "checking the document…"
		}
	}
	if len(w.steps) == 0 {
		return "starting…"
	}
	return "thinking…"
}

// thoughts is how many rows the model's thinking is given, always, so
// that nothing under it moves as it comes and goes.
const thoughts = 3

// rows draws the page: the question, what the model is at and for how
// long, the end of what it is thinking, and the calls it has made, the
// latest last.
func (w work) rows(width, height int) []string {
	rows := strings.Split(ansi.Wordwrap(w.question, width, ""), "\n")
	if len(rows) > 2 {
		rows = append(rows[:1], ansi.Truncate(strings.Join(rows[1:], " "), width, "…"))
	}
	for i := range rows {
		rows[i] = sHead.Render(rows[i])
	}
	rows = append(rows, "")

	if w.err != nil {
		said := strings.Split(ansi.Wordwrap(strings.TrimSpace(w.err.Error()), width-2, ""), "\n")
		rows = append(rows, sBad.Render("✗")+" "+sName.Render(said[0]))
		for _, l := range said[1:min(len(said), 4)] {
			rows = append(rows, "  "+sBody.Render(l))
		}
	} else {
		frames := spinner.MiniDot.Frames
		took := int(w.took.Seconds())
		right := fmt.Sprintf("%d:%02d", took/60, took%60)
		if n := len(w.steps); n == 1 {
			right += " • 1 step"
		} else if n > 1 {
			right += fmt.Sprintf(" • %d steps", n)
		}
		left := sLive.Render(frames[w.frame%len(frames)]) + " " + sBody.Render(w.doing())
		rows = append(rows, left+strings.Repeat(" ", max(1, width-wid(left)-wid(right)))+sDim.Render(right))
	}
	rows = append(rows, "")

	think := strings.Split(ansi.Wordwrap(strings.Join(strings.Fields(w.think), " "), width-2, ""), "\n")
	think = think[max(0, len(think)-thoughts):]
	for i := range thoughts {
		if i < len(think) && think[i] != "" && w.err == nil {
			rows = append(rows, sRule.Render("│")+" "+sDim.Render(think[i]))
		} else {
			rows = append(rows, "")
		}
	}
	rows = append(rows, "")

	room := max(0, height-len(rows))
	for _, s := range w.steps[max(0, len(w.steps)-room):] {
		mark, what := sDim.Render("✓"), sMeta
		switch {
		case s.bad:
			mark = sBad.Render("✗")
		case !s.done:
			mark, what = sLive.Render("→"), sName
		}
		tool := s.tool
		switch tool {
		case "write_document":
			tool, s.what = "write", "the document"
		case "fix_document":
			tool, s.what = "fix", "the document"
		}
		head := mark + " " + sDim.Render(fit(tool, 6))
		note := sDim.Render(s.note)
		rows = append(rows, head+what.Render(fit(s.what, max(0, width-wid(head)-wid(note)-2)))+"  "+note)
	}
	if len(rows) > height {
		rows = rows[:height]
	}
	return rows
}
