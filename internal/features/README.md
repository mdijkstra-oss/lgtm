# features

One folder per capability. A feature owns its logic, its command and its
interface pieces, and does not import another feature.

- `ask/`: a question answered, or a change explained, with a document, and
  the page to ask for one and read it on.
- `intro/`: the slop and the blade, on a page of their own.

The shape of a feature, leaving out what it does not have:

- `<name>.go`: the logic. No printing, no interface.
- `cli/`: the command.
- `ui/`: what it shows in the interactive interface.

`cli/` and `ui/` both call the logic; neither reimplements it.

A new feature also gets its rows in `.go-arch-lint.yml` and its command in
`cmd/lgtm/main.go`.
