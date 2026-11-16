# internal

Everything behind the binary.

- `features/`: one folder per thing strigil can do.
- `kit/`: interface pieces shared by features; knows no feature.
- `shell/`: the interactive interface. Shows what the features hand it.

Who may import whom is written in `.go-arch-lint.yml` at the repo root and
checked with `go run github.com/fe3dback/go-arch-lint@v1.19.1 check`.
