package ask

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"charm.land/fantasy"

	"lgtm/internal/kit/markdown"
)

//go:embed prompt.md
var prompt string

// budget is how many tool calls a document may cost.
const budget = 40

// Event is one thing that happened while a document was being written: the
// model thought some more, called a tool, or the call came back.
type Event struct {
	Think string // more of what the model is thinking; the rest is empty

	ID     int    // the call, so what came of it can be put beside it
	Tool   string // list, grep, read, git or write_document
	What   string // what it was called on
	Done   bool
	Failed bool
	Note   string // what came of it, in a word or two
	Said   string // what the model was told, when the call was refused
}

// Written is a document a model wrote.
type Written struct {
	Doc    markdown.Doc
	File   string // where it was saved
	Source string
}

// Write has the model write a document on a subject about the repository
// at root, and keeps it. with is what every call to the model carries. on
// is told what happens as it happens, from another goroutine than the
// caller's.
func Write(ctx context.Context, model fantasy.LanguageModel, with fantasy.ProviderOptions, root string, sub Subject, on func(Event)) (Written, error) {
	if on == nil {
		on = func(Event) {}
	}
	s := open(root, sub, on)
	// What every question starts by reading is handed over with it, which
	// saves the model the turns it would spend asking.
	tree, _, _ := s.list(".", 2)
	say := fmt.Sprintf("%s\n\nThe repository is at commit %s. Its files, two levels down:\n\n%s\n", s.asking(), s.at, tree)
	for _, name := range []string{"README.md", "AGENTS.md"} {
		if text, _, err := s.reading(readIn{Path: name}); err == nil {
			say += "You have read this already:\n\n" + text + "\n"
		}
	}
	say += fmt.Sprintf("You have %d tool calls.", budget)

	agent := fantasy.NewAgent(model,
		fantasy.WithSystemPrompt(prompt),
		fantasy.WithTools(s.tools()...),
		fantasy.WithStopConditions(fantasy.StepCountIs(budget+10)),
	)
	think := func(_, text string) error {
		on(Event{Think: text})
		return nil
	}
	var said []fantasy.Message
	// A model may stop with words instead of the document; it is asked again.
	for try := 0; try < 3; try++ {
		res, err := agent.Stream(ctx, fantasy.AgentStreamCall{
			Prompt:           say,
			Messages:         said,
			ProviderOptions:  with,
			OnReasoningDelta: think,
			OnTextDelta:      think,
		})
		if s.out != nil {
			return *s.out, nil
		}
		if err != nil {
			return Written{}, err
		}
		if ctx.Err() != nil {
			return Written{}, ctx.Err()
		}
		said = append(said, fantasy.NewUserMessage(say))
		for _, step := range res.Steps {
			said = append(said, step.Messages...)
		}
		say = "You stopped without handing in the document. Call write_document now with the whole document."
		if strings.TrimSpace(res.Response.Content.Text()) == "" && len(res.Steps) == 0 {
			break
		}
	}
	if s.settle() {
		return *s.out, nil
	}
	return Written{}, errors.New("the model did not hand in a document")
}
