# ask

Documents about the codebase: asking for one, and a place to open them.

`lgtm "a question"` has a model read the repository and answer with a
document. The model is given read-only tools (list, grep, read, git) and a
budget of calls, and hands the document in through a check: it has to
parse, every link has to be to lines the model read, and no code of the
repository may be pasted in. What is refused goes back to the model with
what is wrong. An accepted document is kept in the system's temporary folder,
in a folder per repository, so nothing is written into the repository
asked about.

This slice does not render anything itself. The format a document is
written in is `kit/markdown`, and `kit/viewer` draws one on screen; `ask`
uses both. What is its own is the writing of a document, which documents
there are, finding one by its title, and the frame they open in.

- `prompt.md`: what the model is told: who reads, how to work, what a
  document is made of and what a good one is. Change it here.
- `docs/`: the document built in, which shows everything the viewer
  recognises and says what each is for.
- `cli/`: `lgtm ask`, and plain `lgtm`.
- `ui/`: the page while the model works, the palette to open a document
  with, and a tab per document. See `ui/about.md`.
