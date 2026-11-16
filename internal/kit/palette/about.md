# palette

The ctrl+k overlay: a text input, results arriving while you type, and a
preview of the one under the cursor.

- It is given sources and knows nothing about what they search.
- The owner sizes it (`Resize`) and places it; `Cursor` is relative to its own corner.
- `Update` returns an `Outcome`: the highlight moved, something was chosen, or it closed.
- Results of an earlier query stay up until the new query has something to show.
- Under 76 cells of inner width the preview is dropped.
