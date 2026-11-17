---
tab: intro commit
at: 584cb7b
---

# What did the intro commit add, and is it sound?

Commit `0e30354` replaced the whole first interface with one screen. It deleted the search command, the shell with its palette and tabs, the gallery viewer and five shared pieces under `kit`, and added the intro page with a theme built on Charmtone. 102 files changed.

```strigil-block
{
  "kind": "change",
  "parts": [
    {
      "name": "features/intro",
      "files": 14,
      "added": 838,
      "removed": 0,
      "note": "The command, the page, four snapshots and slop.txt."
    },
    {
      "name": "kit/theme",
      "files": 2,
      "added": 162,
      "removed": 112,
      "note": "Rewritten onto tone, with the test that polices it."
    },
    {
      "name": ".go-arch-lint.yml",
      "files": 1,
      "added": 24,
      "removed": 41,
      "note": "Rows for the removed packages went, rows for intro and theme came."
    },
    {
      "name": "cmd/strigil",
      "files": 3,
      "added": 6,
      "removed": 75,
      "note": "gallery.go and its test went; main lists one command."
    },
    {
      "name": "go.mod",
      "files": 1,
      "added": 3,
      "removed": 4,
      "note": "bubbles and catppuccin went, tone and harmonica came. x/ansi became indirect."
    },
    {
      "name": "features/search",
      "files": 20,
      "added": 0,
      "removed": 1188,
      "note": "The search command, its logic and its page."
    },
    {
      "name": "kit/palette",
      "files": 13,
      "added": 0,
      "removed": 787,
      "note": "The ctrl+k palette."
    },
    {
      "name": "shell",
      "files": 5,
      "added": 0,
      "removed": 562,
      "note": "The frame that held tabs and the palette."
    },
    {
      "name": "samples",
      "files": 6,
      "added": 0,
      "removed": 291,
      "note": "Text files the search was tried on."
    },
    {
      "name": "features/gallery",
      "files": 4,
      "added": 0,
      "removed": 191,
      "note": "The viewer that stepped through stories."
    },
    {
      "name": "kit/tabs",
      "files": 10,
      "added": 0,
      "removed": 191
    },
    {
      "name": "kit/text",
      "files": 2,
      "added": 0,
      "removed": 167
    },
    {
      "name": "kit/frame",
      "files": 8,
      "added": 0,
      "removed": 140
    },
    {
      "name": "kit/hints",
      "files": 7,
      "added": 0,
      "removed": 74
    },
    {
      "name": "kit/kit.go",
      "files": 1,
      "added": 0,
      "removed": 46
    }
  ]
}
```

## Does it hold together

At that commit it builds, passes vet, passes both test runs and breaks no import rule. `r` runs any of them again on the tree as it is now.

```strigil-block
{
  "kind": "run",
  "at": "0e30354",
  "cmds": [
    {
      "line": "go build ./...",
      "ok": true,
      "took": "1.0s",
      "out": [
        "(no output)"
      ]
    },
    {
      "line": "go vet ./...",
      "ok": true,
      "took": "0.3s",
      "out": [
        "(no output)"
      ]
    },
    {
      "line": "go test ./...",
      "ok": true,
      "took": "0.7s",
      "out": [
        "ok    strigil/cmd/strigil  0.260s",
        "?     strigil/internal/features/intro/cli  [no test files]",
        "?     strigil/internal/features/intro/ui  [no test files]",
        "ok    strigil/internal/kit/story  0.255s",
        "?     strigil/internal/kit/story/storytest  [no test files]",
        "ok    strigil/internal/kit/theme  0.260s"
      ]
    },
    {
      "line": "go test -tags gallery ./...",
      "ok": true,
      "took": "0.6s",
      "out": [
        "ok    strigil/cmd/strigil  0.193s",
        "?     strigil/internal/features/intro/cli  [no test files]",
        "ok    strigil/internal/features/intro/ui  0.188s",
        "ok    strigil/internal/kit/story  0.170s",
        "?     strigil/internal/kit/story/storytest  [no test files]",
        "ok    strigil/internal/kit/theme  0.176s"
      ]
    },
    {
      "line": "go run github.com/fe3dback/go-arch-lint@v1.19.1 check",
      "ok": true,
      "took": "3.1s",
      "out": [
        "module: strigil",
        "linters:",
        "   On | Base: component imports # always on",
        "   On | Advanced: vendor imports # switch 'allow.depOnAnyVendor = false' (or delete) to on",
        "   On | Advanced: method calls and dependency injections # switch 'allow.deepScan = true' (or delete) to on",
        "",
        "OK - No warnings found"
      ]
    }
  ]
}
```

## What it left behind

`cmd/README.md` still says `gallery.go` is compiled only with `-tags gallery`. The file went in this commit and the README was not touched. The folder test stayed green because it only looks for the names of subfolders.

```strigil-block
{
  "kind": "check",
  "claim": "every README still matches its folder",
  "cases": [
    {
      "name": "internal",
      "holds": true,
      "why": "Names `features/` and `kit/`, which are what is there."
    },
    {
      "name": "features",
      "holds": true,
      "why": "Names `intro/`, the one feature left."
    },
    {
      "name": "intro",
      "holds": true,
      "why": "Names `cli/` and `ui/`."
    },
    {
      "name": "kit",
      "holds": true,
      "why": "Names `story/` and `theme/`."
    },
    {
      "name": "cmd",
      "holds": false,
      "why": "`gallery.go` and `gallery_test.go` were deleted in this commit.",
      "code": {
        "file": "cmd/README.md",
        "line": 1,
        "show": [
          1,
          5
        ],
        "hot": [
          4,
          5
        ],
        "lang": "markdown",
        "lines": [
          "# cmd",
          "",
          "`strigil/` is the one binary. It only lists commands; each command is built by",
          "a `cli` package under `internal`. Its `gallery.go` is compiled only with",
          "`-tags gallery`."
        ]
      }
    }
  ]
}
```

The `gallery` build tag now gates nothing but tests. No command shows a story any more, and the snapshot tests behind the tag are the only tests the intro has. The pre-commit hook runs `go test ./...` without the tag, so a commit can change how the intro draws and the hook passes.

```strigil-block
{
  "kind": "check",
  "claim": "every test runs before a commit",
  "cases": [
    {
      "name": "cmd/strigil",
      "holds": true,
      "why": "The folder README test has no build tag."
    },
    {
      "name": "kit/story",
      "holds": true,
      "why": "The test that every piece has stories has no build tag."
    },
    {
      "name": "kit/theme",
      "holds": true,
      "why": "The colour test has no build tag."
    },
    {
      "name": "intro/ui",
      "holds": false,
      "why": "`stories_test.go` starts with `//go:build gallery`. `make test` runs with and without the tag; the hook runs without.",
      "code": {
        "file": "lefthook.yml",
        "line": 1,
        "show": [
          1,
          4
        ],
        "hot": [
          4,
          4
        ],
        "lang": "yaml",
        "lines": [
          "pre-commit:",
          "  jobs:",
          "    - name: test",
          "      run: go test ./..."
        ]
      }
    }
  ]
}
```

## What I would send back

- Fix the sentence in `cmd/README.md`.
- Give the hook `-tags gallery`, or have it call `make test`.
- The intro brought its own colour type that the theme test cannot see. The answer on colours covers it.
