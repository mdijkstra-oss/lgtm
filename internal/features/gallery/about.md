# gallery

A debug-only viewer for component stories.

    go run -tags gallery ./cmd/strigil gallery

A component takes part by having, next to its code:

- `stories.go`, starting with `//go:build gallery`, with
  `func Stories(theme.Theme) story.Component`: named states, each with a note,
  a fixed size and a function that draws it from fixture data;
- `about.md`: what the component is and what its caller decides;
- `stories_test.go` calling `storytest.Check`, which checks every story fills
  its size exactly and matches its snapshot under `testdata`.

A new component also goes in the list in `cmd/strigil/gallery.go`.

    go test -tags gallery ./...                  check the snapshots
    STORY_UPDATE=1 go test -tags gallery ./...   accept a changed look
