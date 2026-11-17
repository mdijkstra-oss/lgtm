# What Charm already has

Every repository under `github.com/charmbracelet`, what it is, and which parts
can be used here. Look before building: most pieces of a TUI already exist.

Listed on 2026-11-17 (57 repositories). The 29 marked ✓ were cloned and read;
the rest are described from their GitHub description only. Import paths are
from each repo's `go.mod` at that date.

## Need a thing? Start here

| I need | Use | Import |
|---|---|---|
| The app loop, keys, mouse, resize, alt screen, cursor | Bubble Tea | `charm.land/bubbletea/v2` |
| Styles, padding, borders, joining blocks, measuring width | Lip Gloss | `charm.land/lipgloss/v2` |
| A static table / list / tree (rendered, not interactive) | Lip Gloss | `…/lipgloss/v2/table`, `/list`, `/tree` |
| Colours | this repo's `internal/kit/theme` (over Tone) | not `charm.land/tone` directly |
| A gradient across text or a border | Lip Gloss `Blend1D`, `Blend2D`, `BorderForegroundBlend` | `charm.land/lipgloss/v2` |
| Scrollable list with filter, pagination, help | Bubbles `list` | `charm.land/bubbles/v2/list` |
| Interactive table with a selected row | Bubbles `table` | `…/bubbles/v2/table` |
| Interactive tree | Bubbles `tree` | `…/bubbles/v2/tree` |
| One-line text input, with suggestions | Bubbles `textinput` | `…/bubbles/v2/textinput` |
| Multi-line editor with line numbers | Bubbles `textarea` | `…/bubbles/v2/textarea` |
| Scrolling a block of text | Bubbles `viewport` | `…/bubbles/v2/viewport` |
| Key bindings and the help line built from them | Bubbles `key` + `help` | `…/bubbles/v2/key`, `/help` |
| Spinner | Bubbles `spinner` (12 frame sets) | `…/bubbles/v2/spinner` |
| Progress bar, animated | Bubbles `progress` | `…/bubbles/v2/progress` |
| Pagination dots or `1/5` | Bubbles `paginator` | `…/bubbles/v2/paginator` |
| File picker | Bubbles `filepicker` | `…/bubbles/v2/filepicker` |
| Countdown / elapsed time | Bubbles `timer`, `stopwatch` | `…/bubbles/v2/timer`, `/stopwatch` |
| A form: input, text, select, multi-select, confirm, note, file | Huh | `charm.land/huh/v2` |
| "Run this and show a spinner meanwhile", outside a TUI | Huh `spinner` | `charm.land/huh/v2/spinner` |
| Render markdown | Glamour | `charm.land/glamour/v2` |
| Springs and projectile motion | Harmonica | `github.com/charmbracelet/harmonica` |
| Styled `--help`, errors, man pages, completions for Cobra | Fang | `charm.land/fang/v2` |
| Log lines with levels and key/values | Log | `charm.land/log/v2` |
| Truncate, wrap, cut or measure a string that has ANSI in it | x/ansi | `github.com/charmbracelet/x/ansi` |
| Know the colour depth; strip colour when piped; `NO_COLOR` | Colorprofile | `github.com/charmbracelet/colorprofile` |
| A cell buffer, rectangle layout with constraints, cell-diff rendering | Ultraviolet | `github.com/charmbracelet/ultraviolet` |
| Snapshot-test a rendered frame | x/exp/golden | `github.com/charmbracelet/x/exp/golden` |
| Drive a Bubble Tea model in a test | x/exp/teatest | `github.com/charmbracelet/x/exp/teatest/v2` |
| Test against a real terminal emulator | x/vt, x/vttest | `github.com/charmbracelet/x/vt`, `/vttest` |
| Show an image as half-block characters | x/mosaic | `github.com/charmbracelet/x/mosaic` |
| Open the user's `$EDITOR` | x/editor | `github.com/charmbracelet/x/editor` |
| Serve the TUI over SSH | Wish | `charm.land/wish/v2` |
| A GIF or screenshot of the TUI for docs | VHS, Freeze | command-line tools |

Already in `go.mod`: bubbletea, bubbles, glamour, lipgloss, fang, harmonica,
tone, x/ansi.

**Not available anywhere in Charm:** charts, sparklines, gauges (nothing in
`x` or Bubbles); a dialog/overlay stack, scrollbar, diff view, toast/status
line, tabs and buttons exist only inside Crush's `internal/` packages, which
cannot be imported. They can be read and rewritten; see Crush below.

## Libraries

### bubbletea ✓ — `charm.land/bubbletea/v2`
The TUI framework: model, update, view. In v2 the view declares terminal state
(`AltScreen`, `MouseMode`, `Cursor`, `WindowTitle`, `BackgroundColor`,
`ProgressBar`, `ReportFocus`, `KeyboardEnhancements`).

- Messages worth knowing: `WindowSizeMsg`, `KeyPressMsg`, `MouseMsg`,
  `PasteMsg`, `BackgroundColorMsg` (`.IsDark()`), `FocusMsg`/`BlurMsg`,
  `ResumeMsg`, `InterruptMsg`.
- `tea.Tick` for animation; `tea.Printf`/`tea.Println` to leave lines in
  scrollback from an inline program; `tea.Suspend`; `tea.Exec` to hand the
  terminal to another program.
- A real terminal cursor with position, colour, shape and blink.
- `View.OnMouse` with `lipgloss.NewCompositor(...).Hit(x, y)` for click
  targets.
- `examples/` has 62 small programs, the fastest way to see how a thing is
  done: tabs, pager, split-editors, composable-views, table-resize, chat,
  autocomplete, file-picker, prevent-quit, progress-download, package-manager
  (inline output with scrollback), clickable, realtime, debounce, exec,
  suspend, window-size, focus-blur.

### bubbles ✓ — `charm.land/bubbles/v2`
Ready-made components for Bubble Tea. One package each:

| Package | What it gives |
|---|---|
| `list` | Scrollable list with fuzzy filter, pagination, status bar, spinner, help; item rendering is a delegate you can replace |
| `table` | Rows and columns with a selected row and keyboard navigation |
| `tree` | Interactive tree |
| `textinput` | One line: prompt, placeholder, echo modes, validation, suggestions |
| `textarea` | Many lines: line numbers, cursor line, soft wrap, prompt |
| `viewport` | Scroll any string; mouse wheel, key map, soft wrap |
| `help` | Short and full help built from key bindings; drops items that do not fit |
| `key` | Bindings with help text, enable/disable, matching |
| `spinner` | Line, Dot, MiniDot, Jump, Pulse, Points, Globe, Moon, Monkey, Meter, Hamburger, Ellipsis |
| `progress` | Bar with solid or gradient fill, spring-animated |
| `paginator` | Page arithmetic plus dots or arabic rendering |
| `filepicker` | Directory browser with allowed types |
| `timer`, `stopwatch` | Count down, count up |
| `cursor` | The drawn cursor used by the inputs; blink modes |

Each takes a `Styles` struct: set it from `internal/kit/theme`, never keep the
defaults (they are Charm's older colours, and `list` is hardcoded dark).

### lipgloss ✓ — `charm.land/lipgloss/v2`
Styling and layout for strings.

- `Style`: foreground, background, bold/faint/italic/underline/strikethrough/
  reverse, padding, margin, width/height, alignment, borders, `Transform`,
  `Inherit`, `SetString`.
- Borders: normal, rounded, thick, double, block, half-block, hidden, ASCII,
  markdown; per-side on/off and colour; gradient borders.
- Layout: `JoinHorizontal`, `JoinVertical`, `Place`, `PlaceHorizontal`,
  `PlaceVertical`; `Width`, `Height`, `Size` to measure.
- Layers and compositing (`NewLayer`, `NewCompositor`, hit testing) for
  overlays.
- `table`: static tables with borders, per-cell style function, wrapping,
  column shrinking. `list`: bullets and enumerators, nested. `tree`: tree
  rendering with enumerators.
- Colour: `LightDark`, `HasDarkBackground`, `BackgroundColor`, `Blend1D`,
  `Blend2D`. Colour *makers* (`Color`, `Lighten`, `Darken`…) are forbidden
  here by the theme test.
- `lipgloss.Println`/`Sprint`/`Writer`: print with colour downsampled to what
  the terminal supports.

### huh ✓ — `charm.land/huh/v2`
Forms and prompts. Fields: `Input`, `Text`, `Select`, `MultiSelect`,
`Confirm`, `Note`, `FilePicker`. Groups are pages; fields can be hidden,
skipped or computed from earlier answers (`*Func` variants). Validation,
themes, layouts (columns, grid, stack), an accessible line-prompt mode. A form
is a Bubble Tea model, so it embeds in an app or runs alone with `.Run()`.
`huh/spinner` runs an action behind a spinner.

### glamour ✓ — `charm.land/glamour/v2`
Markdown to styled terminal text, driven by a stylesheet (`ansi.StyleConfig`):
headings, code blocks with syntax highlighting, tables, lists, links, word
wrap. Built-in styles: dark, light, dracula, tokyo-night, pink, ascii, notty.
Crush builds its stylesheet from its theme tokens; do the same.

### harmonica ✓ — `github.com/charmbracelet/harmonica`
`Spring` (frequency, damping) and `Projectile` (position, velocity,
acceleration), framework-agnostic. Used by the intro.

### fang ✓ — `charm.land/fang/v2`
Wraps a Cobra command: styled help and usage, styled errors, `--version`,
man page, shell completions, a colour scheme hook. Used by `cmd/strigil`.

### log ✓ — `charm.land/log/v2`
Levelled logger with key/values; text, JSON and logfmt formatters; styles per
level; `slog` handler and standard-library adapter.

### tone ✓ — `charm.land/tone`
The Charmtone palette as named constants (`tone.Charple`, `tone.Sash`…),
grouped: spectrum, neutrals, charples, additions, deletions. Each is a
`color.Color`. Imported only by `internal/kit/theme`. Ships a `charmtone`
command that prints the palette and exports CSS/SCSS/Vim.

### colorprofile ✓ — `github.com/charmbracelet/colorprofile`
Detects the terminal's colour depth (truecolor, 256, 16, none, not a TTY)
from the environment, and a `Writer` that downsamples or strips colour on the
way out. Honours `NO_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`.

### ultraviolet ✓ — `github.com/charmbracelet/ultraviolet`
The low-level layer under Bubble Tea v2 and Crush: a cell buffer that
understands wide characters, `StyledString` drawn into a rectangle, a
cell-diffing renderer, input decoding. `ultraviolet/layout` is a constraint
solver for splitting rectangles (`Len`, `Min`, `Max`, `Percent`, `Ratio`,
`Fill`, with spacing and padding). Reach for it when string joining stops
being enough: overlays, absolute positioning, big screens.

### x ✓ — `github.com/charmbracelet/x/…`
A monorepo of small modules, each imported separately.

| Module | What it gives |
|---|---|
| `ansi` | Every escape sequence as a function or constant; `StringWidth`, `Truncate`, `TruncateLeft`, `Cut`, `Strip`, `Wrap`, `Wordwrap`, `Hardwrap`; a sequence parser |
| `cellbuf` | The older cell buffer (superseded by ultraviolet) |
| `exp/golden` | `RequireEqual`: compare output to `testdata/*.golden`, `-update` to rewrite |
| `exp/teatest` (`/v2`) | Run a model in a test, send keys, wait for output |
| `vt`, `vttest` | A terminal emulator in Go, and snapshot helpers on top of it |
| `mosaic` | Images as Unicode blocks |
| `nerdfont` (`charm.land/x/nerdfont`) | Detect Nerd Font support; pick glyph or fallback |
| `exp/charmtone` | The palette's old home; use `tone` |
| `colors` | Older named colour definitions |
| `term`, `termios`, `conpty`, `xpty`, `windows` | Terminal size, raw mode, pseudo-terminals, per platform |
| `input` | Input event decoding (now inside ultraviolet) |
| `wcwidth` | Character width |
| `editor` | Build the command that opens `$EDITOR` at a file and line |
| `exp/open` | Open a file or URL with the system handler |
| `gitignore` | Gitignore matching |
| `powernap` | Language Server Protocol client and server configuration |
| `pony` | Declarative markup for TUIs; its README calls it experimental and mostly AI-generated |
| `exp/strings`, `exp/slice`, `exp/maps`, `exp/ordered`, `exp/higherorder` | Small generic helpers |
| `errors`, `json`, `etag`, `vcr`, `sshkey` | Error joining, JSON helpers, HTTP ETags, HTTP record/replay, SSH key parsing |

### wish ✓ — `charm.land/wish/v2`
SSH server with middleware. `bubbletea` middleware serves a TUI per
connection; others: `logging`, `activeterm`, `accesscontrol`, `ratelimiter`,
`recover`, `elapsed`, `comment`, `git`, `scp`; `testsession` for tests.

### Others
- **fantasy** — build AI agents in Go across providers. Not UI.
- **catwalk** — a database of LLM providers and models. Not UI.
- **a2tea** — renders A2UI-over-MCP in a terminal; a read-only mirror.
- **keygen** — generate SSH key pairs.
- **hotdiva2000** — human-readable random names.
- **promwish** — Prometheus metrics middleware for Wish.
- **xunicode** — experimental Unicode algorithm packages.
- **git-lfs-transfer** — server side of Git LFS over SSH.

## Applications: read, do not import

Their UI code is `internal` or `package main`. They are the reference for how
a whole screen is put together.

### crush ✓ — the coding agent
The most recent and most complete Charm TUI; `internal/ui` is the model for
anything larger than a prompt.

| Path under `internal/ui` | What to take from it |
|---|---|
| `styles/` | One `Styles` struct built from ~40 tokens by `quickStyle`; themes as token sets plus overrides; WCAG contrast check for ink on fills; gradient helper |
| `dialog/` | Dialog interface, an overlay stack, shared render helpers for title, input, list plus scrollbar, help footer |
| `list/` | Lazily rendered scrollable list that only draws visible items |
| `common/` | Buttons, section header (title plus rule), scrollbar, status rows, markdown and syntax-highlight setup, ANSI-16 remapping |
| `diffview/` | Unified and split diffs with syntax highlighting and line numbers |
| `anim/` | The gradient "working" animation driven by one shared clock |
| `completions/` | Autocomplete popup |
| `model/` | Layout by rectangles with compact-mode breakpoints; status line with toasts; header; sidebar; key map |
| `chat/` | Per-item renderers that cache their output; collapsed long output |
| `logo/`, `exitbanner/` | Wordmark with gradient; what is printed after exit |
| `image/` | Images through the Kitty graphics protocol |
| `AGENTS.md` | Their own rules for writing UI code |

### glow ✓ — markdown reader
`ui/`: a file list with tabs, filter and pagination; a pager with a status
bar (identity chip, truncating note, scroll percent, help cell); short and
full help.

### gum ✓ — components for shell scripts
One command per component: `choose`, `filter`, `input`, `write`, `confirm`,
`file`, `pager`, `spin`, `table`, `style`, `join`, `format`, `log`. Each is
a small complete Bubble Tea program: good examples of a single-purpose prompt
that also works when not attached to a terminal.

### soft-serve ✓ (archived) — Git server with a TUI
`pkg/ui`: tabs, status bar, footer help, selectable repo list, code viewer
with line numbers, all from one styles struct.

### Others
- **freeze** ✓ — renders code or ANSI output to PNG/SVG. Used here to look at
  story snapshots: `freeze --language ansi -o out.png < file.golden`.
- **vhs** ✓ — records a terminal session to GIF from a script (`.tape`).
- **pop** ✓ — send email; a form with focused/blurred fields and attachments.
- **mods** ✓ (archived) — LLM on the command line; streaming output rendered
  through Glamour, error presentation.
- **sequin** ✓ — explains ANSI escape sequences; pipe TUI output through it to
  debug rendering.
- **wishlist** ✓ — SSH directory; a list app served over Wish.
- **skate** ✓ — key-value store CLI.
- **melt** ✓ — SSH key backup as seed words; styled CLI output with explicit
  truecolor/256/16 fallbacks.
- **charm** ✓ (archived) — the old Charm cloud tool; menus and buttons in
  `ui/common`.
- **confettysh** ✓ — confetti over SSH; particle animation.
- **bubbletea-app-template** ✓ — repository skeleton with CI and release config.
- **wizard-tutorial** — a step-by-step wizard, from a tutorial video.
- **inspo** (archived) — a gallery of projects built with Charm.

## Not code to use

- **meta** ✓ — shared CI workflows and lint configuration.
- **vhs-action**, **soft-serve-action** — GitHub Actions.
- **tree-sitter-vhs** — syntax highlighting for `.tape` files.
- **homebrew-tap**, **scoop-bucket**, **nur**, **winget-pkgs** — package
  manager manifests.
- **runway** — 3D models.
- **pi-hyper-provider** — extension for another tool.
- **.github** — organisation defaults.
- Forks kept for patches: **ssh**, **sh**, **console**, **anthropic-sdk-go**,
  **openai-go**, **go-genai**, **markscribe**, **readme-scribe**.
