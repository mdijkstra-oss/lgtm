# ask

Documents about the codebase, and a place to open them.

This slice does not render anything itself. The format a document is
written in is `kit/markdown`, and `kit/viewer` draws one on screen; `ask`
uses both. What is its own is which documents there are, finding one by its
title, and the frame they open in: a palette and a tab per document.

No model writes documents yet. One written by hand is built in, which shows
every block of the format and says what each is for.

- `docs/`: the document built in.
- `cli/`: `strigil ask`.
- `ui/`: the palette to open a document with, and a tab per document. See
  `ui/about.md`.
