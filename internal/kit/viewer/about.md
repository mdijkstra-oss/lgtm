# viewer

A document on screen. It is given a document and a room, and knows nothing
of where either came from.

- A document is set in one column, in the middle of the room. Its
  title and intro come first, and under them the sections, each under its
  heading.
- The cursor is one line, and goes up and down a line at a time, through
  text and through blocks alike. It stops on rows that have words on them,
  and passes over blank rows and rows that are only the lines of a drawing.
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
- Beside the document is the index: the title and then the sections,
  down a rail. Each has the
  mark its subsections came to, and its number.
  A number goes to its section. The section the cursor is in is a filled row there.
- The title or a heading with the cursor on it is lit across the column,
  fading out to the right, so going to one by number shows where it is.
- Nothing a document says is syntax of the viewer's. It recognises ordinary
  markdown and shows it as what it means:
  - the third-level headings of a section are its subsections, one shown at
    a time. They are picked from a row of buttons, or from a Mermaid drawing
    that names them: a sequence, a row a message in the lane it goes to, or
    a flowchart, a row a node under the first node that points at it. A ✓ or
    ✗ before a name is a mark. A subsection holds text and blocks, as a
    section does; on a line of a block in a subsection, left and right are
    that block's.
  - a link to lines of a file, alone on its line, is those lines.
  - a table with a column of `+n` and one of `-n` is bars: a bar a row,
    lines out to the left and in to the right.
  - a colour in a code span, as hex, `rgb()` or `hsl()`, has a cell of that
    colour before it.
  - code among the words, fenced or set in by four, has a rail down its
    side. It is coloured as its language when the fence names one, and
    not at all otherwise.
- A block shows the code of what is picked. Enter gives the code the whole
  view, with more of the file either side.
