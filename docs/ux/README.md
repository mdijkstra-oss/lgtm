# Designing a beautiful TUI

What makes a terminal interface look good, and how to make one. Distilled from
reading the source of 29 Charm repositories (Crush, Bubbles, Lip Gloss, Huh,
Gum, Glow, Glamour, Soft Serve, Fang, Log, Pop, Mods, Tone and others) and from
general interface design. Every rule says where it comes from:

- **[Charm]** observed in their code; the reference is `repo path:line`.
- **[UX]** general interface design, not specific to terminals.
- **[Derived]** my conclusion from the two; treat as a strong default, not a fact.

Charm's concrete values (hex colours, paddings, frame rates) are given as worked
examples. Copy the *relationships* (how far apart the tiers are, what gets a
hue).

In this repo the palette is fixed: every colour is a constant from
`internal/kit/theme` (Crush's default theme over Charmtone), named after the
tiers in section 3.3. `go test ./...` fails on a colour made any other way.
What Charm already ships, so it is not rebuilt, is in `docs/charm/README.md`.

## 0. The short version

1. **Grey first.** A screen must read correctly with no hue at all: hierarchy
   comes from 4 brightness tiers of text, blank space and alignment. Hue is
   added last and to very little.
2. **Colour the token, not the sentence.** One icon, one word, one bar. The
   text the user came to read stays the default foreground.
3. **One accent means one thing:** "you are here" (cursor, selection, focus).
   Status hues (red, yellow, green) appear only on one-cell icons and small
   tags.
4. **Labels are quieter than values, structure is quieter than content, help
   is the quietest text on screen.**
5. **Space separates, borders float.** Regions are divided by blank rows and
   columns. A box is for something that sits on top (dialog, popup).
6. **Nothing moves when state changes.** Focus, selection, spinners, errors and
   scrollbars all occupy space that was reserved for them.
7. **Every state is designed:** empty, loading, error, long, narrow, short,
   unfocused, no-colour. Not only the happy path at 120×40.
8. **When space runs out, drop whole elements in a planned order.** Never show
   half of something.
9. **Motion only while real work is happening,** slow for glyphs (≈10 fps),
   never bouncy.
10. **Know whether you own the canvas** (section 2). It decides how you dim,
    how you select and how many colours you can trust.

## 1. The medium

What a terminal gives you, and what follows from it.

| Constraint | Consequence |
|---|---|
| One font, one size, one grid. No font-size hierarchy. | Hierarchy has only these tools: position, space, brightness, weight (bold), hue, fill, glyph. Use them in that order (section 3). |
| A cell is about twice as tall as it is wide. | 2 columns look like 1 row. Horizontal padding is double the vertical: `Padding(1, 2)`, `Margin(1, 2)`. |
| The smallest unit of space is a whole cell. A border costs 2 columns and 2 rows. | Borders are expensive. Whitespace is the default separator. |
| Background, foreground and the 16 ANSI colours belong to the user's theme unless you paint over them. | You do not know what "grey" or "red" looks like until you decide to own the canvas (section 2). |
| Colour depth varies: truecolor, 256, 16, none. Pipes and `NO_COLOR` remove it. | No meaning may be carried by colour alone (section 12). |
| Glyph width is not certain (emoji, some symbols, CJK). | Stick to single-width symbols; measure with a grapheme-aware width function, never `len`. |
| The window is resized at any time, down to tiny. | Layout is a function of width and height with explicit breakpoints (section 6). |
| Keyboard first; mouse is optional; there is no hover by default. | Every action has a key and the key is shown. Discoverability lives in the help line. |
| Output can stay in scrollback or vanish with the alt screen. | Decide which; design the last frame the user is left with. |

## 2. First decision: do you own the canvas?

Charm's code splits cleanly in two, and most of its apparent contradictions
(Faint or not, fill or gutter bar) disappear once you see which side a program
is on. **[Derived]**

| | **You own it** | **You are a guest** |
|---|---|---|
| What it is | Full-screen app that paints its own background colour | Inline output, prompts, CLI help, logs; anything on the user's background |
| Example | Crush sets `v.BackgroundColor` and ships whole themes. No light/dark detection anywhere in its UI (0 uses of `LightDark`/`HasDarkBackground`). | Log, Fang, Gum, Pop, Glow's list |
| Dimming | Explicit greys from your own ramp. Crush's UI uses `Faint` 0 times. | The terminal's default foreground plus `Faint`, or light/dark colour pairs. Log: keys, separators and caller are Faint, message and value unstyled `log styles.go:46-52`. |
| Surfaces | Several close background tiers are safe (section 3). | Avoid background fills except small chips; you cannot predict the neighbour colour. |
| Selection | Full-width fill with a dedicated ink `crush quickstyle.go:1185` | Lists: gutter mark plus foreground change, no fill `bubbles list/defaultitem.go:46-49`. Table rows: an accent fill is fine, since columns need one continuous band `bubbletea examples/table/main.go:173-177`. What a guest cannot do is rely on subtle surface tiers. |
| Light/dark | A separate theme per background | Every colour is a (light, dark) pair resolved from the terminal's reply; default to dark until it arrives `lipgloss query.go:86-92` |

Pop's comment states the guest's problem exactly: some embedded terminals
report the application theme, not the visible background, so it unsets
foregrounds and uses "faint labels and faint italic placeholders [to] keep
hierarchy on any background without dimming real values" `pop theme.go:22-34`.

When Crush prints its exit banner, the background has already been restored,
so it drops to default foreground plus bold `crush quickstyle.go:238-241`.
The moment you stop owning the canvas, you stop using its colours.

## 3. Hierarchy

### 3.1 Rank first, style second **[UX]**

Before choosing any style, list what is on the screen and give each item a
rank. A screen has one primary thing (what the user came for), a few secondary
things, and chrome. If two things both look primary, neither is. Test: blur
your eyes, or view it at arm's length; the primary item should be the first
thing you see and the help line should nearly vanish.

### 3.2 The order of tools **[Derived]**

Reach for these in order. Stop as soon as the hierarchy reads.

1. **Position.** Top-left is read first. Primary content gets the largest
   region.
2. **Space.** A blank row before a group makes it a group. Related things sit
   close; unrelated things are separated (proximity).
3. **Alignment.** Shared left edges and aligned columns create order without
   ink.
4. **Brightness tier.** The main tool. See 3.3.
5. **Bold.** For the few things that are titles, badges or the selection.
6. **Hue.** For "you are here" and for state.
7. **Fill.** A background chip is the loudest thing a terminal can draw. One
   per screen for identity, plus error badges.
8. **Glyph.** A one-cell symbol that carries state without a word.

### 3.3 Text tiers and surface tiers **[Charm]**

Crush builds everything from one tinted neutral ramp: four text tiers and four
surface tiers `crush themes.go:62-72`. Contrast ratios are mine,
computed with the WCAG formula against its background `#201F26`.

| Tier | Crush token | Colour | Contrast | What goes here |
|---|---|---|---|---|
| Text 1 | fgBase | `#ECEBF0` | 13.8 | Names, titles, list items, what the user typed, values |
| Text 2 | fgSubtle | `#BFBCC8` | 8.7 | Long-form body text (markdown), error message text |
| Text 3 | fgMoreSubtle | `#858392` | 4.4 | Metadata: paths, branch, model; help **keys**; blurred text |
| Text 4 | fgMostSubtle | `#605F6B` | 2.6 | Help **descriptions**, placeholders, section headings, timers, parameters |
| Line | separator | `#3A3943` | 1.4 | Rules, the `•` between help items, scrollbar track, ellipsis |
| Surface 0 | bgBase | `#201F26` | 1.0 | The app |
| Surface 1 | bgLeastVisible | `#2D2C36` | 1.2 | Quiet panels: tool output, "thinking" box |
| Surface 2 | bgLessVisible | `#3A3943` | 1.4 | Inline code, blurred button |
| Surface 3 | bgMostVisible | `#4D4C57` | 1.9 | Inactive button, pill border |

What to take from this:

- **The steps are roughly 14 → 9 → 4.5 → 2.5 → 1.4.** Each tier is clearly a
  different thing, not a slightly different grey.
- **Long body text is not the brightest tier.** Markdown body is Text 2
  `crush quickstyle.go:304`; Glamour's body is `252` on dark
  and `234` on light, never pure white or black
  `glamour styles/styles.go:152-154,362`. Pure white is reserved so
  that titles can still be brighter than prose.
- **Text 4 sits below the 4.5:1 that WCAG asks for body text.** That is
  deliberate and only acceptable for text the user does not need to read to
  operate the screen (help descriptions, placeholders). Anything the user must
  read is Text 3 or above. **[UX]**
- **Neutrals are tinted** toward the brand hue (blue channel highest in every
  Charmtone neutral, `x exp/charmtone/charmtone.go:264-275`), which is
  why the greys and the purple look like one family.
- **Surfaces differ by 1.2–1.9.** Enough to see a panel edge, not enough to
  compete with text. Only possible when you own the canvas.

The older, guest-side libraries use the same ladder with plain greys. Bubbles
on dark: title `#dddddd`, description `#777777`, help key
`#626262`, help description `#4A4A4A`, separator
`#3C3C3C` `bubbles list/defaultitem.go:39,43`,
`bubbles help/help.go:51-53`. Glow reuses the same five as light/dark
pairs `glow ui/styles.go:60-64`.

### 3.4 Three quiet rules **[Charm]**

- **Labels are quieter than values.** Log keys are Faint, values are plain
  `log styles.go:46-52`. Crush's sidebar section title is Text 4
  followed by a rule in the Line colour `crush common/elements.go:229-248`.
  The eye should land on the data, and find the label only when it looks.
- **Structure is quieter than content.** Tree connectors, table borders, list
  bullets and gutters are drawn at `238`–`241` in nine Lip Gloss
  examples while the items are brighter (`lipgloss examples/tree/files/main.go:52`,
  `examples/table/pokemon/main.go:80`).
- **Help is the quietest text, and inside it the key is one step brighter than
  its description.** Bubbles `#626262` vs `#4A4A4A`; Soft Serve
  `241` vs `239` `soft-serve pkg/ui/styles/styles.go:278-286`;
  Crush Text 3 vs Text 4 `crush quickstyle.go:714-722`. The user scans
  for the key; the description is there if they need it.

## 4. Colour

### 4.1 What gets a hue and what never does

| Gets a hue | Stays neutral |
|---|---|
| The cursor, the selected row, the focused control | Body text and values. Bubbles sets `Text: lipgloss.NewStyle()` for inputs, file names and table cells `bubbles textinput/styles.go:21` |
| A one-cell status icon (`✓ × ●`) | The sentence next to the icon |
| A small padded tag (`ERROR`, `WARN`, a mode badge) | The explanation after the tag. Fang's error body has no foreground set `fang theme.go:187-190` |
| The name of the kind of thing (tool name, command, flag, heading) | Its arguments and description. Fang colours only the flag name `fang help.go:408,432` |
| The level in a log line | Timestamp, message, values `log styles.go:43-82` |
| The spinner | The label beside it (`bubbletea examples/spinners/main.go:26-28`) |
| The identity chip (logo) and the one focused frame | All other borders, rules, help, metadata, timers, line numbers |

### 4.2 A hue has one job **[Charm]**

Crush puts five hues on screen and each has a single meaning
`crush quickstyle.go` (lines in brackets):

| Hue | Meaning | Where |
|---|---|---|
| Primary (purple `#6B50FF`) | "This is the active frame / the chosen row" | Dialog border [1173], selected row fill [1185], user-message bar [1074-1077] |
| Secondary (pink `#FF60FF`) | "The live point" | Cursor [265], focused button, scrollbar thumb [1196], spinner |
| Green | "Good / ready / go" | Success icon, online dot, focused prompt `>` [920], additions |
| Blue `#00A4FF` | "Name of a thing" | Tool names [818], headings [319], links [384] |
| Red, yellow, orange | State only | `ERROR` tag, deletions, `WARN` tag, needs-auth dot |

The older apps do the same with fewer: pink `212`/`#EE6FF8` is "where
you are" across table, file picker, list and 13 Gum defaults; indigo
`99` is "what this is" (headers, titles, directories); green
`#04B575` is "positive" in three files across two repos.

Rules:

- **Budget:** two brand hues plus neutrals for chrome. Status hues are extra
  but small. If a third decorative hue appears, something is being coloured
  that should be grey. **[Derived]**
- **A status hue never doubles as decoration.** Crush colours inline code in
  its "destructive" red and had to override it in two themes
  `crush themes.go:249-255,574-583`. Red means wrong; nothing else is red.
- **Each status hue comes in strengths.** Full strength fills a badge; a
  quieter shade of the same hue is used for icons and text. Crush success:
  `#00FFB2` (icon), `#12C78F` (pending dot, focus bar)
  `crush quickstyle.go:813-814,1019`.
- **Inside a selected item, keep its hierarchy** by shifting each tier into a
  tint of the accent, not flattening it to one colour. Glow: title
  `#EE6FF8`, date `#99519E`, third line `#7B4380`
  `glow ui/styles.go:68-71`.
- **Charm's red leans pink** (`#FF4672`, `#FF5F87`,
  `#FF577D`, `#EB4268`). A status palette should sit in the
  same family as the brand, not be pulled from pure `#FF0000`.

### 4.3 Fills and ink

- **Text on a filled background is cream, not white:** `#FFFAF1`,
  `#FFFDF5`, `#FFF7DB` across Fang, Huh, Glow, Wishlist and
  the Lip Gloss layout example. Give the fill its own "on-fill" ink token
  (Crush: `onPrimary`).
- **Check the contrast of ink on every fill.** Crush computes WCAG contrast and
  swaps the ink below 3.0, the floor for bold text
  `crush quickstyle.go:161-179`. Targets **[UX]**: 4.5:1 for text the
  user must read, 3:1 for bold tags and for the edge of a control.
- **A chip is** `Padding(0, 1)`, bold, often uppercase. The same shape is
  used for the logo, markdown H1, mode badges and the `ERROR` tag, so
  one shape reads as "label" everywhere (`fang theme.go:191-198`,
  `glow ui/styles.go:151-154`, `glamour styles/styles.go:171-178`).
- **One identity chip per screen.** It is the single loud element.

### 4.4 Gradients

Crush uses a gradient in exactly five places: logo, dialog-title filler,
spinner, queue markers and header wordmark; always primary→secondary, applied
per character as a foreground ramp `crush styles/grad.go:16-42`. Blending
is done in a perceptual space (CIELAB in Lip Gloss `blending.go:78`) so
the midpoint is not muddy. A gradient is brand texture and "alive" signalling.
It never carries information and never sits on body text. **[Derived]**

### 4.5 Light and dark (guests only)

- Each role is a pair from one hue: the darker step on a light background, the
  lighter step on a dark one. Fang: `#FF4FBF`/`#FF79D0`,
  `#FF577D`/`#FF7F90` `fang theme.go:44-62`. Charmtone is
  organised as hue × three lightnesses for this reason.
- Dim means "step toward the background" in both modes: dimmed text is
  `#858392` on light and `#605F6B` on dark.
- Ask the terminal, default to dark until it answers, rebuild styles when the
  reply arrives `bubbletea examples/textarea/main.go:34,52-54`.
- Test both. Charm's own slip: Huh's default theme has
  `normalFg = lightDark("252", "235")`, near-black text on dark
  backgrounds `huh theme.go:144`.

## 5. Text attributes

Counted with ripgrep over non-test Go files.

| Attribute | Use it for | Evidence |
|---|---|---|
| **Bold** | Chips and badges, titles, table headers, the selected item. Not for emphasis in running text. | 28 uses in Crush's UI, almost all on filled badges |
| **Faint** | Secondary text when you are a guest. Not when you own the canvas: terminals render it differently and it stacks unpredictably with your greys. | 0 in Crush, Bubbles, Huh, Glow, Fang; 4 in Log, 2 in Pop |
| **Italic** | Almost never. Placeholders at most. Many terminal fonts lack a good italic. | 2 in Crush, 0 in Bubbles, Huh, Glow, Fang, Log |
| **Underline** | "This character matters": fuzzy-match characters, a button's hotkey letter, links. | 5 in Crush; filter matches `bubbles list/defaultitem.go:61` |
| **Strikethrough** | Done items, with a check mark and grey `lipgloss examples/layout/main.go:155-165` | |
| **Reverse** | The drawn text cursor `bubbles cursor/cursor.go:240` | |
| **UPPERCASE** | Badge words (`ERROR`), table headers, CLI section titles `fang theme.go:128-133` | |

Underline for search matches is worth copying exactly: an attribute composes
with whatever colour the row already has (normal, selected, dimmed); a match
colour would collide with the selection colour.

## 6. Space and layout

### 6.1 Numbers that recur **[Charm]**

| What | Value | Evidence |
|---|---|---|
| Outer margin, full app | 1 row, 2 columns | `soft-serve styles.go:181-182`, `wishlist styles.go:19`, `bubbletea examples/list-default/main.go:12`. Crush uses 1/1, and 2 on landing `ui.go:4618-4627` |
| Left indent of everything in a guest view | 2 columns | Bubbles list: title, status, items, pagination, help all at 2 `list/style.go:50,71,85,87` |
| Between regions | 1 blank row | Glow's whole list view is one format string with `\n\n` between every block `glow ui/stash.go:725-734` |
| Between form fields | 1 blank row; title and description adjacent | `huh theme.go:104`, `field_select.go:727` |
| Before the help line | 1 blank row | `huh group.go:390`, `gum choose/choose.go:291` |
| Inside chips, cells, tabs | `Padding(0, 1)` | `bubbles table/table.go:117`, Crush tags |
| Inside panels and cards | `Padding(1, 2)` | `crush quickstyle.go:1134,1172,1193` |
| Buttons | `Padding(0, 2)` or `(0, 3)`, gap 2 | `crush common/button.go:44-45`, `huh theme.go:106-108` |
| Nested content indent | 2 | Tool body, code block margin, list indent `crush styles.go:84-85` |
| Prose width cap | 72–80 | Glow 78, Pop 80, Melt 72 |
| Document / chat width cap | 120 | `crush chat/messages.go:23`, `glow main.go:201-203` |
| Fixed sidebar | 32 columns | `crush ui.go:4601` |

Use a small fixed scale (1, 2, 4 columns; 0, 1 rows) and nothing else. Uneven
gaps are the most common reason a TUI looks amateur. **[UX]**

### 6.2 Rules

- **Content never touches the terminal edge.** Seven of seven Charm apps with
  a full view indent it.
- **Cap line length.** Text that spans a 200-column terminal is unreadable; 45
  to 80 characters is the comfortable range **[UX]**. Compute
  `min(available, cap)` and leave the rest empty.
- **Separate with space, not lines.** Crush's sidebar has no divider; it has 1
  column of padding `crush ui.go:4749`. The only box in all of Charm's
  older apps is one error box in VHS.
- **Borders are for floating things:** dialogs, popups, cards, the details
  overlay. Rounded, in the primary colour, no fill
  `crush quickstyle.go:1173`.
- **A section header is a dim title plus a rule** filling the width, then the
  rows `crush common/elements.go:229-248`.
- **Pin the footer.** Fill the gap with blank rows so help and status sit on
  the last line at any height `glow ui/stash.go:697-702`.
- **Align.** Help in full mode is two aligned sub-columns per group
  `bubbles help/help.go:211-216`; tree prefixes are right-aligned to
  the widest sibling `lipgloss tree/renderer.go:100-103`. Numbers
  right-align.
- **Fill emptiness with brand texture, dimly, or leave it empty.** Crush
  repeats `╱` in the logo, header and dialog titles.

### 6.3 Size is a breakpoint, not a hope **[Charm]**

Crush's breakpoints, as a model:

- Below 120 columns or 30 rows: compact mode. The sidebar collapses to a
  one-row header; details move to an overlay `crush ui.go:76-77`.
- Below 30 rows: the block logo becomes a one-line logo.
- Dialogs are a percentage of the window with a cap, and go full-screen at
  ≤77 columns or ≤20 rows `crush dialog/permissions.go:36-55`.
- Buttons stack vertically when they do not fit in a row.

Degrade by **dropping whole elements in priority order**:

- A list's right-hand info column is hidden for *all* rows at once when it
  would take more than 25–35% of the row `crush dialog/common.go:87-120`.
- Title info is dropped whole, not truncated `dialog/common.go:336-343`.
- Help drops whole `key desc` items and shows `…` only if it fits;
  it never ends on a dangling separator `bubbles help/help.go:234-243`.
- Pagination dots become `1/5` when the dots are too wide
  `bubbles list/list.go:1188-1193`.
- A "Updated…" column disappears before the name is truncated
  `soft-serve pages/selection/item.go:178-180`.

Truncate single values with `…` (one character, not three dots). Clamp
every computed size with `max(0, …)`. Render nothing until the first
window-size message arrives.

## 7. Focus and selection

**Nothing moves.** This is the rule Charm follows most consistently.

- Selected list item: 2 columns of padding become a 1-column bar plus 1 column
  of padding `bubbles list/defaultitem.go:40,46-49`.
- Unselected rows are padded by the cursor's width
  `huh field_select.go:707`, `gum choose/choose.go:259`.
- A blurred form field keeps its border as `HiddenBorder()`, which draws
  spaces `huh theme.go:128-129`.
- Focusing a chat item swaps `│` for `▌` in the same column and
  colour `crush quickstyle.go:1069-1080`.
- A scrollbar's column is reserved whenever content overflows.

**Selection** (which row):

| Situation | Idiom |
|---|---|
| Guest, or a quiet list | Gutter mark (`│`, `┃`, `>`) in a reserved column plus accent foreground. No fill. |
| A table row, anywhere | Accent fill across all columns with an on-fill ink |
| You own the canvas; a modal list or picker | Full-width fill in primary with the on-fill ink `crush quickstyle.go:1185` |
| Either | Add a non-colour carrier (the gutter mark, or bold) so it survives `NO_COLOR` **[Derived]** |

**Focus** (which pane or control has the keyboard):

- One structural mark changes; content keeps its colours. Three of four
  coloured Huh themes change only the border between focused and blurred.
- Blurred input: prompt and text drop one tier, placeholder stays
  `crush quickstyle.go:251-263`.
- Blurred badge: same shape, hue removed (background-colour ink on a grey
  fill) `crush quickstyle.go:928-953`.
- Buttons differ by background only, no borders: focused, blurred, inactive,
  hovered `crush quickstyle.go:913-917`.
- **A mode recolours the whole frame.** Delete mode turns the dialog's border,
  title, filler and selected row red; rename turns them yellow
  `crush quickstyle.go:1228-1240`. The user cannot miss which mode they
  are in, and no extra text is needed.
- While the user is typing a filter, every item drops one tier until there is
  a result `bubbles list/defaultitem.go:54-59`.

## 8. Components

**Help line.** Bottom row. `key desc • key desc`. Key one tier brighter
than description; separator dimmer than both; nothing bold or coloured.
Lowercase verbs, arrows as glyphs, aliases joined with a slash: `↑/k up`,
`? more` `bubbles list/keys.go:39-87`. `?` expands to
aligned columns. Disabled bindings are left out, not greyed
`bubbles help/help.go:138`. Show only what works *right now*.

**Status bar.** Identity chip on the left, a flexible middle that truncates
with `…`, fixed-width facts on the right, `? Help` rightmost
`glow ui/pager.go:291-297`, `soft-serve statusbar/statusbar.go:67-91`.
Cells are told apart by background steps, not separators. Transient messages
take over the bar instead of opening a popup `glow ui/styles.go:123-136`.

**Lists.** Two-line items (title in Text 1, description two tiers down), 1
blank row between items `bubbles list/defaultitem.go:98-99`. Pad a
short last page so the chrome below does not jump `list.go:1228-1238`.
Filter status reads as a sentence: `“query” 12 items • 3 filtered`.

**Tables.** Header bold, no colour; cells `Padding(0, 1)`; no row
separators; border in the Line tier or none. Headers never wrap; cells truncate
with `…` `lipgloss table/table.go:81-93,321`. Zebra rows, if any,
use two adjacent greys.

**Inputs.** Prompt glyph in the accent when focused, typed text in the default
foreground, placeholder in Text 4, short and imperative ("Type to filter").
A real blinking block cursor in the secondary colour
`crush quickstyle.go:251-269`.

**Forms.** Title, description directly under it, control, blank row. The
focused field has a thick left bar in dark grey, not the accent
`huh theme.go:111,152`. A validation error is a red `*` after the
title, and the message **replaces the help line**; the field itself is not
recoloured `huh theme.go:113,158`, `group.go:412-418`.

**Buttons.** Filled blocks, no brackets, no border. Focused is the loud fill;
the others are grey. A destructive confirm turns the focused button red
`crush dialog/permissions.go:829-833`. Underline the hotkey letter.

**Dialogs.** Centred, rounded border in primary, no fill, no backdrop dimming.
Anatomy: title, filter input, list, help footer. Default max 70×20
`crush dialog/dialog.go:16-18`. Size content to the *inner* width;
inset with padding, never margin `crush internal/ui/AGENTS.md`.

**Tabs.** Active and inactive differ by tier or by underline; a dim `│`
between `glow ui/styles.go:101-102,149`.

**Scrollbar.** One column: thumb `┃` in secondary on track `│` in
Line. Hide it 2 s after the last scroll unless the pane is focused
`crush model/chat.go:168`.

**Progress.** Width `min(available, 80)`. Filled cells in the accent,
empty cells `░` in dark grey, percentage after the bar
`bubbles progress/progress.go:46-58`.

**Long output.** Collapse to 10 lines with a hint that says how many are
hidden and how to expand; if collapsing would hide only one line, show it
`crush chat/tools.go:28-43`.

**Trees.** `├──` and `╰──` in a dim tier, items plain.

**Markdown.** One chip (H1), one heading colour for H2–H5 told apart by their
`##` prefixes, inline code as a small chip on Surface 2, links
underlined, rule and blockquote bar in the Line tier, bullets `•`
`glamour styles/styles.go:164-272`.

**Diff.** Three bands: line-number gutter, sign, code. Backgrounds are the
insert/delete hue blended over the app background at 20% (code) and 13%
(gutter) for inserts, 15% and 8% for deletes. The code keeps its syntax
colours; only the sign and line numbers take the diff hue
`crush quickstyle.go:64-69,119-124`. Tint the background, do not recolour
the text.

**Log lines.** One coloured token (the level, bold, fixed width 5). Keys and
`=` dim. Message and values plain. Multi-line values behind a dim
`│` gutter `log styles.go:43-82`, `text.go:249-253`.

**CLI help.** Uppercase bold section titles; flag and command names coloured,
descriptions plain; defaults appended dim as `(value)` and omitted when
empty, false or zero; placeholders like `[flags]` dimmed; width capped at
120 `fang help.go:24-39,436-438`.

## 9. States

Each state says what it is with a glyph or word, not only a colour.

| State | How Charm draws it |
|---|---|
| **Loading** | Spinner in the accent plus a lowercase label and ellipsis ("loading…"). Delay showing it (25 ms in Huh `eval.go:26`, 100 ms in Soft Serve) so fast work never flashes a spinner. Reserve the spinner's cell when idle `bubbles list/list.go:1091-1094`. |
| **Working, long** | Spinner, label, and a quiet elapsed timer in Text 4. Different animation for different kinds of work: Crush scrambles characters only for model "thinking", because scrambling "implies thinking rather than running" `crush anim/anim.go:106-110`. |
| **Empty** | A short grey sentence with a full stop, where the content would be: "No items.", "Nothing found." `glow ui/stash.go:752,801`. Keep the region's height. |
| **Success** | `✓` in green; the text stays neutral. Do not let a fast result flash past: Crush enforces a 750 ms minimum on quick async steps `crush dialog/oauth_hyper.go:35`, and the download example pauses 750 ms at 100% before quitting. Prefer changing the thing in place ("(copied to clipboard)") over a popup. |
| **Error** | Three parts: an `ERROR` chip; one plain sentence in the default foreground saying what happened; then the raw detail and the way out in grey `mods main.go:465-474`, `glow ui/ui.go:340-352`, `fang theme.go:187-198`. Inline errors are red foreground only. |
| **Warning** | Same shape as error, yellow or orange tag. A refusal by the user is `WARN`, not `ERROR` `crush chat/tools.go:728-733`. |
| **Disabled** | One or two tiers down; removed from help. |
| **Destructive** | The frame, title and selected row turn red; the confirm button turns red; a second key press is required. |
| **Resource status** | One glyph `●`, only its colour changes: grey offline, yellow busy, red error, green online `crush quickstyle.go:1016-1025`. Pair with a word where there is room. |
| **Transient message** | A two-part toast on the help line: bold uppercase badge on the strong hue, message on the quieter shade of the same hue; gone after 5 s `crush quickstyle.go:1259-1268`, `model/status.go:18`. |

Response-time limits **[UX]**: under 0.1 s feels instant, so show nothing;
under 1 s keep the screen as is; beyond 1 s show a spinner; beyond about 10 s
show progress or elapsed time, and a way to cancel.

Error wording: say what happened in plain words, then what to do. Mods:
"You haven't provided any prompt input." followed by the command to run, shown
as inline code. The raw `err.Error()` is always demoted to grey.

## 10. Motion

| What | Rate | Evidence |
|---|---|---|
| Glyph spinners | 3–12 fps; 10 for dots and lines, slower for detailed glyphs | `bubbles spinner/spinner.go:28-83` |
| Crush working animation | 20 fps; ellipsis every 400 ms | `crush anim/anim.go:19-31` |
| Continuous movement (progress fill) | 60 fps | `bubbles progress/progress.go:48` |
| Cursor blink | 530 ms | `bubbles cursor/cursor.go:14` |
| Progress spring | frequency 18, damping 1.0 (critically damped, no overshoot) | `bubbles progress/progress.go:50-51` |

- **Animate only to say "work is happening" or "this value changed".** No
  decorative loops on an idle screen.
- **No bounce in real UI.** Underdamped springs (0.15–0.2) appear only in
  Charm's demos. The one shipped spring is critically damped.
- **One clock.** All spinners advance from a single tick, only while a
  spinning item is on screen, and stop when work stops
  `crush model/chat.go:536-654`. Independent tickers drift and waste CPU.
- **Stop at rest.** The progress bar stops ticking at equilibrium; the renderer
  parks after about half a second with nothing to draw.
- **Things that should not jitter are computed once:** Crush picks the logo's
  stretched letter once so it does not change on resize.

## 11. Interaction

- `ctrl+c` always quits. `q` quits only when no text input is on
  screen. `esc` backs out one level (close dialog, blur, cancel) before it
  ever quits. In Crush esc never quits.
- **Interrupting needs a second press inside a visible window:** esc once arms
  cancel and the help text changes to "press again to cancel"; it disarms
  after 2 s `crush model/ui.go:4066-4070`.
- **Unsaved state turns quit into a confirmation**, defaulting to no.
- **Guard against stray keys:** a dialog that opens by itself ignores input
  until the keyboard has been quiet for 425 ms
  `crush dialog/dialog.go:63-67`.
- **Keep input responsive:** no I/O in the update loop; coalesce mouse-move and
  wheel events (one per 16 ms in Crush) so key presses never queue behind them.
- **Alt screen only for a full-canvas app.** 14 of Bubble Tea's 64 examples
  use it. Otherwise stay inline, print finished items into scrollback, and
  design the final frame (empty, a one-line summary, or a goodbye).
- Recognition over recall **[UX]**: the keys that work now are on screen now.
- Same key, same meaning, on every screen. **[UX]**

## 12. Robustness

- **Colour may be gone.** With `NO_COLOR` on a terminal, colour is dropped
  but bold, faint, italic and underline survive; when output is not a
  terminal, all styling is stripped `colorprofile writer.go:40-44,98-167`,
  `env.go:86-91`. So every state needs a glyph, word or attribute.
  Check: done = `✓` + strikethrough + grey; selected = bar + bold.
- **Padding that only makes sense around a fill must go when the fill goes.**
  Fang removes code-block padding when there is no background
  `fang help.go:55-59`.
- **16 colours:** the neutral ramp collapses to black, bright black, white. If
  that matters, author a second scheme by role, as Fang
  (`theme.go:66-83`) and Huh's Base16 theme do. Downsampling will not
  design it for you.
- **Do not trust the user's ANSI 16** for output you display inside your own
  canvas: Crush remaps subprocess colours to its palette because the defaults
  are "often illegible on our dark background" `crush common/ansi16.go:11-15`.
- **Not a terminal:** print plain text, no spinner, no prompts
  `fang help.go:143-149`, `gum spin/spin.go:162-164`.
- **Width:** measure, cut and wrap with ANSI- and grapheme-aware functions
  (`x/ansi`: `StringWidth`, `Truncate`, `Cut`, `Wrap`). Never slice a
  styled string by bytes.
- **Style segments separately and concatenate.** Wrapping an already-styled
  string in an outer style loses the outer colour after the inner reset
  `crush internal/ui/AGENTS.md`.
- **Colour-vision deficiency [UX]:** about 1 in 12 men cannot tell some reds
  from greens. Insert/delete, pass/fail and online/error must differ by glyph
  (`+`/`-`, `✓`/`×`) as well as hue.
- **Snapshot every state.** Golden files of rendered frames at fixed sizes
  catch layout regressions (`x/exp/golden`, `x/exp/teatest` at
  80×24).

## 13. Glyphs

Single-width, in the basic plane, no emoji, no Nerd Font requirement. Crush's
whole set `crush styles/styles.go:23-80`:

| Meaning | Glyph |
|---|---|
| Done / success | `✓` |
| Failed | `×` |
| Pending, status dot | `●` |
| In progress | `→` |
| List bullet, help separator, pagination | `•` |
| Radio | `◉` `○` |
| Focus / selection bar | `│` `┃` `▌` |
| Rule | `─` |
| Scrollbar | `┃` on `│` |
| Truncation | `…` |
| Prompt | `>` |
| Tree | `├──` `╰──` |
| Disclosure | `▶` `▼` |
| Brand filler | `╱` |
| Diagnostics | `E` `W` `I` `H` |

Pick one glyph per meaning and use it everywhere. Rounded corners
(`╭╮╰╯`) for boxes. Half blocks (`▀`, `▌`) with separate foreground
and background double the resolution for bars and pixel art.

## 14. Words

- Help: lowercase verb, no punctuation: `enter select`, `esc back`.
- Empty and status text: a short sentence with a full stop.
- Loading: lowercase with an ellipsis.
- Badges: one uppercase word.
- Placeholders: a short imperative.
- Errors: what happened, then what to do. No stack traces in the primary line.
- Personality is welcome in small doses and only in low-stakes places (empty
  states, placeholders: "Find your fave", "Bummer"). Never in errors that the
  user must act on. **[Derived]**

## 14a. Personality

Charm's own line is "We make software glamorous." Their website
(charm.land, viewed 2026-11-17) is loud: the whole page sits on the brand
purple, buttons are pink, cards are cream, there are sparkles and small 3D
mascots. The terminal screenshots *inside* those cards are dark, grey and
restrained. The glamour surrounds the work; it is not in the work.
**[Derived]**

Where the personality goes in a TUI, in order of how much it buys:

1. **Words.** The site labels things "Flavor: Taro", "Glossiness: Very",
   "* Self-awarded", "haters > /dev/null". In the apps: `OKAY!` and `HEY!`
   toasts, "Find your fave", "Bummer". One short human line where a dry one
   would do, in places where nothing is at stake.
2. **One hero moment.** A logo, an intro, a landing screen: the single place
   that gets the gradient, the brand texture and the full palette. Crush's is
   its wordmark; lgtm's is the blade. Everything after it is quiet.
3. **Motion with weight.** Things move as if they had mass (Harmonica springs
   and projectiles), settle, and stop. Physical, not decorative.
4. **Small sparkle.** A two-part badge (dark label, pink value) is the site's
   signature shape and the same shape as Crush's toast. One gradient on the
   thing that is alive.

What does not carry over: the saturated background, colour on every surface,
mascots. A terminal is where someone is working; the site is where they are
being sold to.

## 15. How to design a screen

1. **Inventory and rank.** List everything the screen shows. Mark one primary,
   a few secondary, the rest chrome.
2. **Lay it out in one colour.** Regions, margins, blank rows, alignment, width
   caps. Decide what is fixed and what flexes.
3. **Assign tiers.** Text 1–4 and Line. Labels below values, structure below
   content, help at the bottom. It must already look finished.
4. **Add the accent** to the single "you are here".
5. **Add state:** glyph first, then hue on the glyph or a tag.
6. **Design every state:** empty, loading, error, one item, a thousand items,
   the longest plausible string, unfocused.
7. **Design the sizes:** 80×24, each breakpoint, very wide. Write down what is
   dropped at each step.
8. **Strip the colour** and confirm every state is still readable.
9. **Snapshot** each state at each size.

Review checklist:

- [ ] Can I name the primary element, and is it the first thing I see?
- [ ] Is every hue on screen accounted for by a single meaning?
- [ ] Is anything coloured that is not a token, an icon, a tag or the selection?
- [ ] Are all gaps on the scale (1/2/4 columns, 0/1 rows)?
- [ ] Does anything touch the terminal edge or run wider than its cap?
- [ ] Does anything shift when focus or selection changes, or when a spinner,
      error or scrollbar appears?
- [ ] Is every box a floating thing? Could the others be space?
- [ ] Is every key that works shown, and nothing that does not?
- [ ] Does each state have a non-colour carrier?
- [ ] Is ink on every fill at least 3:1 (bold) or 4.5:1?
- [ ] Is there one truncation glyph, one bullet, one check mark?
- [ ] Does anything animate while idle?

## 16. What not to copy from Charm

Their code has slips; a rule here was not taken on trust where the source
contradicted it.

- **Low-contrast ink on a fill.** Crush's default focused button is
  `#FFFAF1` on `#FF60FF`: 2.4:1 by my calculation, below the 3.0
  floor its own badge code enforces elsewhere.
- **A status hue used as decoration** (red inline code), patched per theme.
- **Hardcoded dark.** Bubbles' list is built with `DefaultStyles(true)`
  behind an `XXX` comment `bubbles list/list.go:208-210`; Soft
  Serve has no light variant at all; Log's level colours are pale tints with a
  `TODO` for light backgrounds `log styles.go:44`.
- **Hiding things with ANSI black** (`0`), which depends on the user's
  palette `bubbles textarea/textarea.go:441-443`.
- **Glyph and wording drift:** selected is `[•]`, `✓` or `◉`;
  truncation is `…` in 41 places and `...` in 2; the error badge
  says `ERROR` in three apps and `Bummer` in one; quit hints are
  worded seven ways across the examples.
- **Three generations of one brand colour:** `#5A56E0`, ANSI `57/62/63`,
  `#6B50FF` (and `#6C50FF`). One token, one value.
- **Unstyled raw errors** in most Bubble Tea examples.
- **No "terminal too small" handling** in any example; they only clamp.
- **Confusing token names:** Crush's `fgSubtle` is brighter than
  `fgMoreSubtle`, and the local `muted` and `subtle` are the
  other way round from what the names suggest. Number the tiers.

## 17. Sources

Shallow clones of `github.com/charmbracelet/*` read on 2026-11-17:
bubbletea, crush, glow, gum, lipgloss, bubbles, huh, freeze, glamour, log, pop,
fang, skate, wishlist, harmonica, sequin, melt, ultraviolet, x, colorprofile,
tone, soft-serve, mods, vhs, wish, meta, confettysh, bubbletea-app-template,
charm. Nothing was built or run: findings come from style definitions and
render code. The only rendered output checked is the set of GIFs in the Bubbles README, viewed afterwards; the table-row rule in sections 2 and 7 was corrected from them.

References are `repo path:line` at that date. Crush paths without a
directory are under `internal/ui/styles/` (`quickstyle.go`,
`themes.go`, `styles.go`) or `internal/ui/` otherwise. The
references were collected by five parallel readers; I re-read about thirty of
them against the source (the tier tokens, breakpoints, help colours, selection
and focus styles, spinner and spring numbers, colour-profile behaviour) and
all held. Line numbers will drift as those repos change.

Contrast ratios and the attribute counts are my own calculations. The **[UX]**
items are standard interface-design guidance (WCAG 2 contrast thresholds,
Nielsen's response-time limits, Gestalt grouping, line-length research) and
are not derived from Charm.
