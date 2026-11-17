---
tab: colours
at: 584cb7b
---

# How are colours kept to the theme?

Every colour is a constant in `internal/kit/theme`, named for its job and set to a Charmtone colour. The import rules stop any other package reaching the palette, and a test reads every Go file for a colour built by hand.

```strigil-block
{
  "kind": "colours",
  "groups": [
    {
      "name": "brand",
      "items": [
        {
          "name": "Primary",
          "tone": "Charple",
          "job": "the active frame, the chosen row",
          "uses": 1
        },
        {
          "name": "Secondary",
          "tone": "Dolly",
          "job": "the live point: cursor, focused control",
          "uses": 4
        },
        {
          "name": "Accent",
          "tone": "Bok",
          "job": "the input prompt",
          "uses": 2
        },
        {
          "name": "OnPrimary",
          "tone": "Butter",
          "job": "text on a Primary or Secondary fill",
          "uses": 1
        }
      ]
    },
    {
      "name": "text",
      "items": [
        {
          "name": "Text1",
          "tone": "Sash",
          "job": "names, titles, values",
          "uses": 1
        },
        {
          "name": "Text2",
          "tone": "Smoke",
          "job": "long-form body",
          "uses": 0
        },
        {
          "name": "Text3",
          "tone": "Squid",
          "job": "metadata, a key in the help line",
          "uses": 1
        },
        {
          "name": "Text4",
          "tone": "Oyster",
          "job": "what a key does, placeholders",
          "uses": 2
        },
        {
          "name": "Line",
          "tone": "Char",
          "job": "rules and separators",
          "uses": 1
        }
      ]
    },
    {
      "name": "surfaces",
      "items": [
        {
          "name": "Surface0",
          "tone": "Pepper",
          "job": "the app's own background",
          "uses": 2
        },
        {
          "name": "Surface1",
          "tone": "BBQ",
          "job": "quiet panels",
          "uses": 0
        },
        {
          "name": "Surface2",
          "tone": "Char",
          "job": "inline code",
          "uses": 0
        },
        {
          "name": "Surface3",
          "tone": "Iron",
          "job": "inactive controls",
          "uses": 1
        }
      ]
    },
    {
      "name": "states",
      "items": [
        {
          "name": "Error",
          "tone": "Sriracha",
          "job": "wrong, on an icon or a small tag",
          "uses": 0
        },
        {
          "name": "Warning",
          "tone": "Mustard",
          "job": "take care, on an icon or a small tag",
          "uses": 2
        },
        {
          "name": "Info",
          "tone": "Malibu",
          "job": "the name of a thing",
          "uses": 2
        },
        {
          "name": "Success",
          "tone": "Julep",
          "job": "good, on an icon or a small tag",
          "uses": 1
        }
      ]
    }
  ]
}
```

## The import

`charm.land/tone` is a vendor in `.go-arch-lint.yml` that only the `theme` component may use. The theme test checks the same import again, so the rule holds on a machine where the linter never ran.

```strigil-block
{
  "kind": "check",
  "claim": "only theme reaches the palette",
  "cases": [
    {
      "name": "theme.go",
      "holds": true,
      "why": "The one import of the palette in the repository.",
      "code": {
        "file": "internal/kit/theme/theme.go",
        "line": 1,
        "show": [
          1,
          9
        ],
        "hot": [
          9,
          9
        ],
        "lang": "go",
        "lines": [
          "// Package theme is every colour strigil draws with: Crush's default theme,",
          "// \"Charmtone Pantera\", over the Charmtone palette.",
          "//",
          "// A colour on screen is one of these, or a blend between two of them. The",
          "// test in this package fails on a colour made any other way. What each tier",
          "// is for is in docs/ux/README.md.",
          "package theme",
          "",
          "import \"charm.land/tone\"",
          "",
          "// Brand.",
          "const (",
          "    Primary   = tone.Charple // the active frame, the chosen row",
          "    Secondary = tone.Dolly   // the live point: cursor, focused control",
          "    Accent    = tone.Bok     // the input prompt",
          "    OnPrimary = tone.Butter  // text on a Primary or Secondary fill",
          ")",
          "",
          "// Text, from what is read first to what is there if you look.",
          "const (",
          "    Text1 = tone.Sash   // names, titles, values",
          "    Text2 = tone.Smoke  // long-form body",
          "    Text3 = tone.Squid  // metadata, a key in the help line",
          "    Text4 = tone.Oyster // what a key does, placeholders",
          "    Line  = tone.Char   // rules and separators",
          ")",
          "",
          "// Surfaces, from the app's own background outward.",
          "const (",
          "    Surface0 = tone.Pepper",
          "    Surface1 = tone.BBQ",
          "    Surface2 = tone.Char",
          "    Surface3 = tone.Iron",
          ")"
        ]
      }
    },
    {
      "name": ".go-arch-lint.yml",
      "holds": true,
      "why": "`theme` is the only component allowed the `tone` vendor.",
      "code": {
        "file": ".go-arch-lint.yml",
        "line": 7,
        "show": [
          32,
          37
        ],
        "hot": [
          36,
          37
        ],
        "lang": "yaml",
        "lines": [
          "  kit:       { in: [internal/kit, internal/kit/story, internal/kit/story/**] }",
          "  theme:     { in: internal/kit/theme }",
          "",
          "vendors:",
          "  cobra:     { in: github.com/spf13/cobra }",
          "  fang:      { in: charm.land/fang/v2 }",
          "  bubbletea: { in: charm.land/bubbletea/v2 }",
          "  lipgloss:  { in: charm.land/lipgloss/v2 }",
          "  harmonica: { in: github.com/charmbracelet/harmonica }",
          "  tone:      { in: charm.land/tone }",
          "",
          "deps:",
          "  # main only lists commands.",
          "  main:",
          "    mayDependOn: [intro_cli]",
          "    canUse: [cobra, fang]",
          "",
          "  # A feature's command reaches its interface; the interface never reaches",
          "  # a command.",
          "  intro_cli:",
          "    mayDependOn: [intro_ui]",
          "    canUse: [cobra]",
          "  intro_ui:",
          "    mayDependOn: [kit, theme]",
          "    canUse: [bubbletea, lipgloss, harmonica]",
          "  kit:",
          "    mayDependOn: [kit]",
          "    canUse: [lipgloss]",
          "  # The only place the palette is named.",
          "  theme:",
          "    canUse: [tone]"
        ]
      }
    },
    {
      "name": "colours_test.go",
      "holds": true,
      "why": "The test fails on the import anywhere outside the theme.",
      "code": {
        "file": "internal/kit/theme/colours_test.go",
        "line": 53,
        "show": [
          78,
          84
        ],
        "hot": [
          80,
          81
        ],
        "lang": "go",
        "lines": [
          "        rel = filepath.ToSlash(rel)",
          "        if !strings.HasSuffix(rel, \".go\") || rel == themeDir+\"/colours_test.go\" {",
          "            return nil",
          "        }",
          "        f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)",
          "        if err != nil {",
          "            return err",
          "        }",
          "        inTheme := filepath.ToSlash(filepath.Dir(rel)) == themeDir",
          "        bad := func(pos token.Pos, why string) {",
          "            t.Errorf(\"%s:%d: %s; use a colour from %s\", rel, fset.Position(pos).Line, why, themeDir)",
          "        }",
          "",
          "        // The name each package goes by in this file.",
          "        names := map[string]string{}",
          "        for _, imp := range f.Imports {",
          "            p, _ := strconv.Unquote(imp.Path.Value)",
          "            name := p[strings.LastIndex(p, \"/\")+1:]",
          "            if p == lipglossPath {",
          "                name = \"lipgloss\"",
          "            }",
          "            if imp.Name != nil {",
          "                name = imp.Name.Name",
          "            }",
          "            names[name] = p",
          "            switch {",
          "            case p == \"charm.land/tone\" || strings.HasSuffix(p, \"/charmtone\"):",
          "                if !inTheme {",
          "                    bad(imp.Pos(), \"the palette is imported directly\")",
          "                }",
          "            case p == lipglossPath+\"/compat\":",
          "                bad(imp.Pos(), \"lipgloss/compat makes colours\")",
          "            }",
          "        }",
          "",
          "        ast.Inspect(f, func(n ast.Node) bool {",
          "            switch n := n.(type) {",
          "            case *ast.SelectorExpr:",
          "                pkg, ok := n.X.(*ast.Ident)",
          "                if !ok {",
          "                    return true",
          "                }",
          "                switch names[pkg.Name] {",
          "                case \"image/color\":",
          "                    if n.Sel.Name != \"Color\" {",
          "                        bad(n.Pos(), \"color.\"+n.Sel.Name+\" makes a colour\")",
          "                    }",
          "                case lipglossPath:",
          "                    if lipglossMakers[n.Sel.Name] {",
          "                        bad(n.Pos(), \"lipgloss.\"+n.Sel.Name+\" makes a colour\")",
          "                    }",
          "                }",
          "            case *ast.BasicLit:",
          "                if n.Kind != token.STRING {",
          "                    return true",
          "                }",
          "                s, err := strconv.Unquote(n.Value)"
        ]
      }
    }
  ]
}
```

## The test

`TestNoColourOutsideTheTheme` parses each file. It fails on a lipgloss function that makes a colour, on anything from `image/color` but the `Color` interface, on a string that is a hex colour, and on an escape sequence with a colour number written in it.

```strigil-block
{
  "kind": "run",
  "at": "584cb7b",
  "cmds": [
    {
      "line": "go test ./internal/kit/theme",
      "ok": true,
      "took": "0.5s",
      "out": [
        "ok    strigil/internal/kit/theme  0.480s"
      ]
    }
  ]
}
```

## Where it leaks

The intro page does not use lipgloss. It keeps a colour as three floats in its own `rgb` type and writes the escape sequences itself, putting the numbers in at run time. The test sees neither. I put each of these four in a scratch copy and ran the test.

```strigil-block
{
  "kind": "check",
  "claim": "a colour made by hand fails the test",
  "cases": [
    {
      "name": "hex string",
      "holds": true,
      "why": "Caught.",
      "code": {
        "file": "rogue.go, then go test ./internal/kit/theme",
        "show": [
          0,
          0
        ],
        "hot": [
          0,
          0
        ],
        "lines": [
          "const rogueHex = \"#ff0000\"",
          "",
          "--- FAIL: TestNoColourOutsideTheTheme (0.00s)",
          "    rogue.go:4: a colour is written as #ff0000;",
          "    use a colour from internal/kit/theme"
        ]
      }
    },
    {
      "name": "whole escape",
      "holds": true,
      "why": "Caught.",
      "code": {
        "file": "rogue.go, then go test ./internal/kit/theme",
        "show": [
          0,
          0
        ],
        "hot": [
          0,
          0
        ],
        "lines": [
          "const rogueInk = \"\\x1b[38;2;255;0;0m\"",
          "",
          "--- FAIL: TestNoColourOutsideTheTheme (0.00s)",
          "    rogue.go:3: a colour is written into an escape sequence;",
          "    use a colour from internal/kit/theme"
        ]
      }
    },
    {
      "name": "rgb literal",
      "holds": false,
      "why": "Passes. The test looks at lipgloss, `image/color` and string literals, and a struct of three floats is none of those.",
      "code": {
        "file": "rogue.go, then go test ./internal/kit/theme",
        "show": [
          0,
          0
        ],
        "hot": [
          0,
          0
        ],
        "lines": [
          "var rogue = rgb{1, 0, 0}",
          "",
          "ok    strigil/internal/kit/theme  0.201s"
        ]
      }
    },
    {
      "name": "split escape",
      "holds": false,
      "why": "Passes. The pattern wants a digit after `38;2;` inside one string. `intro.go:230` builds its sequences the same way, from numbers.",
      "code": {
        "file": "rogue.go, then go test ./internal/kit/theme",
        "show": [
          0,
          0
        ],
        "hot": [
          0,
          0
        ],
        "lines": [
          "const rogueInk = \"\\x1b[38;2;\" + \"255;0;0m\"",
          "",
          "ok    strigil/internal/kit/theme  0.201s"
        ]
      }
    }
  ]
}
```

Today every `rgb` in the intro comes from `of(theme.X)` or from mixing two of those, so nothing on screen is off the palette. Only habit keeps it so, and the next page an agent writes this way would get through.

## What I would send back

- Make the test fail on an `rgb` built from numbers anywhere but `of` and `mix`, or move the type and the escape writing into `kit` so a page cannot build one.
- `Text2`, `Surface1`, `Surface2` and `Error` have no users yet, so nobody has seen them on screen.
