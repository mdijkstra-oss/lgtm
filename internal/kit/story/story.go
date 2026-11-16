// Package story is the format a component uses to show itself in isolation:
// named fixtures at fixed sizes, plus a note on what the component is for.
package story

// Story is one state of a component, drawn at exactly W by H cells.
type Story struct {
	Name   string
	Note   string
	W, H   int
	Render func(w, h int) string
}

// Component is what a package's Stories function returns.
type Component struct {
	Name    string
	About   string // the package's about.md
	Stories []Story
}
