# search

How a search shows up in the interface.

- `Source` feeds the palette: one item per hit, with a preview of the file around it.
- `Page` is a tab for one query: every hit, grouped by file. It runs its own
  search when opened, so it shows "Searching…" first.
- Enter on a hit opens its file over the list, in a frame; esc returns.
- Two queries that differ only in letter case are the same page.
