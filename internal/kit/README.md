# kit

Interface pieces shared by features and the shell. Nothing here knows a
feature. `kit.go` holds the contracts a feature fills in to appear in the
interface: `Source`, `Item`, `Page`.

- `theme/`: the colours, by role.
- `text/`: cutting and padding text to a width.
- `frame/`: a bordered box with head, body and foot.
- `hints/`: the line of keys that work right now.
- `tabs/`: the list of open pages.
- `palette/`: the ctrl+k overlay.
- `story/`: the format components use to show their states, and its checks.

Each piece that draws has an `about.md` and stories next to its code.
