# tabs

The list of open pages, as a column (`Sidebar`) or as one line (`Strip`) for
terminals too narrow for the column.

- The caller says which entry is active and which one is lit. Lit is the tab
  holding the result currently hovered in the palette; -1 for none.
- An entry is a title and a short meta text (a count, for a search).
- Long titles keep their start, since that is the name of the action.
