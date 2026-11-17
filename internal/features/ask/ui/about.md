# ask

Where a question is asked and its answer read.

- It opens on the palette. What is typed narrows the questions there is an
  answer for; enter opens the chosen one on a tab of its own. ctrl+k brings
  the palette back, and `[` and `]` go from tab to tab.
- An answer is a notebook in one column, in the middle of the window. Its
  lead is always there: a paragraph and the figure that carries the answer.
  Under it are the sections, each under its heading.
- The cursor is one line, and goes up and down the page a line at a time,
  through text and through blocks alike. It stops on rows that have words on them, and
  passes over blank rows and rows that are only the lines of a drawing.
- A block has a grey outline so it can be told from text. While the cursor
  is on one of its lines the outline is purple and the keys work on it. The
  outline is always there, so nothing moves as the cursor comes and goes.
- What can be picked in a block is a button: on its own surface, lifted when
  it is the one picked, filled while the cursor is in its block. Left and
  right pick the one before and after, coming round at either end. Up and
  down never change the pick.
- Picking changes what the block says below its buttons. A cursor that was
  down there goes to the first line of it, and the page shows the block from
  its top.
- The page scrolls to keep the cursor between the third row and a third of
  the way down, so most of the window shows what is ahead. In a block it
  holds back, keeping the block's top in view until the cursor is two thirds
  of the way down. One line up or down moves the page one row more than the
  cursor at most, so it catches up over a few lines and never jumps.
- Beside the notebook is the index: the question and then the sections,
  down a rail. Each has the
  mark its checks came to, its number, and the claim it stands or falls on.
  A number goes to its section. The section the cursor is in is lit there.
- The question or a heading with the cursor on it is lit across the page,
  fading out to the right, so going to one by number shows where it is.
- The blocks:
  - a flow: the way through the packages, a row a hop, and what the hop
    picked does.
  - a check: a claim and the cases it was tried on, with why the one picked
    holds or does not, and what it rests on.
  - colours: the palette as itself, with the name, job and use of the one
    picked.
  - a run: commands and how each ended. `r` runs the one picked again, here.
  - a change: a bar a part, lines out to the left and in to the right.
- A block shows the code of what is picked. Enter gives the code the whole
  view, with more of the file either side.
