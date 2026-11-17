# ask

Answers to questions about a codebase, as notebooks: markdown, and in it
blocks that show something and respond (a flow, a check, colours, a run, a
change).

An answer is one markdown document. A block is a fenced `strigil-block`
holding its JSON; a second-level heading starts a section. `parse.go` reads
that, and says how it is written.

No model is asked yet. Three answers written by hand about this repository
are built in, to see what reading one is like.

- `answers/`: the three answers.
- `cli/`: `strigil ask`.
- `ui/`: the palette to ask with, a tab per question, and the notebook. See
  `ui/about.md`.
