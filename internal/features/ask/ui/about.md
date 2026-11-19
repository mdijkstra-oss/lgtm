# ask

Where a document is asked for, opened and read.

- It opens covered in slop with LGTM stamped across it, and a blade
  scrapes it clean (`kit/scrape`). What is under it is already the page,
  and already answers the keys.
- The opening page is the history of the repository. The repository and
  its branch are named at the top. Under them is a line to type on, and
  under that the list: what is not committed yet, when there is any, then
  the commits, the newest first, under the name of the day they were made
  on. A commit is what it says, with its hash to the right; the kind of
  change it starts with ("feat(viewer):") is quieter than the rest. A
  commit that was explained before has a tick.
- A day with more than four commits lists four and folds the rest into
  one row, which enter opens. Typing lists every match.
- Typing narrows the list to the commits that have every word typed, in
  their hash, their subject or their author. When no commit has, what was
  typed is a question, and the list is the one row that offers to ask it.
- Up and down pick a row, which is one fill from edge to edge. Enter
  explains the change picked, or asks the question. A commit with a tick
  opens what was written about it: a commit does not change.
- At the foot are the questions answered before, at most five, numbered. A
  number opens one while nothing is typed; once something is typed a number
  is part of it. Changes explained before are not listed there: they are in
  the history.
- The line to type on and the earlier questions keep their rows whatever
  the list does. The list scrolls only when the picked row would leave it.
- While the model works the page shows it. Under what was asked is one
  line saying what the model is at, with the time it has taken and how many
  calls it has made. Under that, three rows for the end of what it is
  thinking, kept whether it thinks or not. Then the calls it has made, a
  row each, the latest last: a done one quiet, a refused one with a red
  cross, the one it waits on lit. When there are more than fit, the oldest
  go.
- Esc stops the model and goes back to the history. If the model cannot
  be asked, the line says why.
- When the document is accepted it opens on a tab. A tab a document open,
  along the top, the open one joined to what is below it. `[` and `]` go
  from tab to tab, and ctrl+k goes back to the history, where esc returns
  to the document.
- Under the tabs is the document, drawn by `kit/viewer`, which takes
  every other key.
- The last line says which keys work now.
- A window under 60 by 16 says to make it larger.
