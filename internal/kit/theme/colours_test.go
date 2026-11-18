package theme

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const (
	lipglossPath = "charm.land/lipgloss/v2"
	themeDir     = "internal/kit/theme"
)

// Ways lipgloss makes a colour out of nothing, or out of one colour and a
// number. Blend1D, Blend2D and LightDark only choose between colours given.
var lipglossMakers = map[string]bool{
	"Color": true, "RGBColor": true, "ANSIColor": true, "Complete": true,
	"Alpha": true, "Complementary": true, "Darken": true, "Lighten": true,
	"Black": true, "Red": true, "Green": true, "Yellow": true, "Blue": true,
	"Magenta": true, "Cyan": true, "White": true,
	"BrightBlack": true, "BrightRed": true, "BrightGreen": true, "BrightYellow": true,
	"BrightBlue": true, "BrightMagenta": true, "BrightCyan": true, "BrightWhite": true,
}

var (
	hexColour = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
	// An escape sequence with the colour already in it: a named one (31, 44,
	// 92...) or an indexed or RGB one with its first number written out.
	sgrColour = regexp.MustCompile(`\x1b\[[0-9;]*((3[0-7]|4[0-7]|9[0-7]|10[0-7])[;m]|[34]8;[25];[0-9])`)
)

// Every colour comes from this package. See the package comment.
func TestNoColourOutsideTheTheme(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != root || d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if !strings.HasSuffix(rel, ".go") || rel == themeDir+"/colours_test.go" {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		inTheme := filepath.ToSlash(filepath.Dir(rel)) == themeDir
		bad := func(pos token.Pos, why string) {
			t.Errorf("%s:%d: %s; use a colour from %s", rel, fset.Position(pos).Line, why, themeDir)
		}

		// The name each package goes by in this file.
		names := map[string]string{}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			name := p[strings.LastIndex(p, "/")+1:]
			if p == lipglossPath {
				name = "lipgloss"
			}
			if imp.Name != nil {
				name = imp.Name.Name
			}
			names[name] = p
			switch {
			case p == "charm.land/tone" || strings.HasSuffix(p, "/charmtone"):
				if !inTheme {
					bad(imp.Pos(), "the palette is imported directly")
				}
			case p == lipglossPath+"/compat":
				bad(imp.Pos(), "lipgloss/compat makes colours")
			}
		}

		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.SelectorExpr:
				pkg, ok := n.X.(*ast.Ident)
				if !ok {
					return true
				}
				switch names[pkg.Name] {
				case "strigil/" + themeDir:
					// A colour from a document is content. Only what shows
					// documents has a use for one.
					if n.Sel.Name == "Literal" && filepath.ToSlash(filepath.Dir(rel)) != "internal/kit/viewer" {
						bad(n.Pos(), "theme.Literal is for showing a colour a document names")
					}
				case "image/color":
					if n.Sel.Name != "Color" {
						bad(n.Pos(), "color."+n.Sel.Name+" makes a colour")
					}
				case lipglossPath:
					if lipglossMakers[n.Sel.Name] {
						bad(n.Pos(), "lipgloss."+n.Sel.Name+" makes a colour")
					}
				}
			case *ast.BasicLit:
				if n.Kind != token.STRING {
					return true
				}
				s, err := strconv.Unquote(n.Value)
				if err != nil {
					return true
				}
				if hexColour.MatchString(s) {
					bad(n.Pos(), "a colour is written as "+s)
				}
				if sgrColour.MatchString(s) {
					bad(n.Pos(), "a colour is written into an escape sequence")
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
