# internal

Everything behind the binary.

- `features/`: one folder per thing lgtm can do.
- `kit/`: what features share; knows no feature.

Who may import whom is written in `.go-arch-lint.yml` at the repo root and
checked with `go run github.com/fe3dback/go-arch-lint@v1.19.1 check`.
