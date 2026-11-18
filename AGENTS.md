# lgtm

## Setup

    make setup   once per clone: installs the pre-commit hook (lefthook.yml)
    make test    tests, with and without the gallery
    make lint    vet and the import boundaries
    make run     runs lgtm with what .env names; ARGS='"a question"' asks one

## Folder READMEs

`cmd`, `internal`, every folder directly under `internal` and every slice under
`internal/features` has a `README.md`: what the folder is for and what each
subfolder holds. The big picture only, so it survives a refactor.

When a change alters what a folder is for, or adds, removes or renames one of
its subfolders, update that README in the same commit. `go test ./...` fails
when a README is missing or does not name a subfolder.

## Seeing the interface

To see the interface in colour, record it with `vhs`: a `.tape` script runs
the program, sends the keys and writes a PNG or a GIF. Never render terminal
output to HTML, and never open a browser for it.
