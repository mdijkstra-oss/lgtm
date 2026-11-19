# ask

Having a model write a document about a repository, and the page to ask
for one and read it on.

A document answers a question, or explains a change: a commit, or what is
not committed yet. `lgtm "a question"` and `lgtm <commit>` ask for one
from the shell; plain `lgtm` opens the history to pick a change from.

The model is given read-only tools (list, grep, read, git) and a budget of
calls. A change is handed to it as its diff, numbered as the files are
now, so it starts from the change and reads only round it. A document
about a change has a shape of its own, sized to the change: what it did
before and what it does now, how the new code works, and a question the
reader should be able to answer before accepting it. It explains and does
not review. `docs/eval` at the root is how that shape is judged. What it writes
goes through a check: it has to parse, every link has to be to lines the
model read, and no code of the repository may be pasted in. What is
refused goes back to the model with what is wrong. A document is kept in
the system's temporary folder, in a folder per repository, so nothing is
written into the repository asked about.

This slice does not render anything itself. The format a document is
written in is `kit/markdown`, and `kit/viewer` draws one on screen; `ask`
uses both.

- `prompt.md`: what the model is told: who reads, how to work, what a
  document is made of and what a good one is. Change it here.
- `docs/`: a document written by hand that shows everything the viewer
  recognises. The stories and tests draw it; the app does not list it.
- `cli/`: `lgtm ask`, and plain `lgtm`.
- `ui/`: the history to pick from, the page while the model works, and a
  tab per document. See `ui/about.md`.
