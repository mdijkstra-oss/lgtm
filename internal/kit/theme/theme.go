// Package theme is every colour lgtm draws with: a dark and a light
// theme over the Charmtone palette. Dark is Crush's "Charmtone Pantera".
//
// A colour the interface is drawn in is one of these, or a blend between two
// of them. The test in this package fails on a colour made any other way.
// The one exception is Literal: a colour a document names, shown as itself. What each tier
// is for is in docs/ux/README.md.
package theme

import (
	"image/color"
	"math"
	"strconv"
	"strings"

	"charm.land/tone"
)

// Theme names each colour for its job.
type Theme struct {
	// Brand.
	Primary   tone.Key // the active frame, the chosen row
	Secondary tone.Key // the live point: cursor, focused control
	Accent    tone.Key // the input prompt
	OnPrimary tone.Key // text on a Primary or Secondary fill

	// Text, from what is read first to what is there if you look.
	Text1 tone.Key // names, titles, values
	Text2 tone.Key // long-form body
	Text3 tone.Key // metadata, a key in the help line
	Text4 tone.Key // what a key does, placeholders
	Line  tone.Key // rules and separators

	// Surfaces, from the app's own background outward.
	Surface0, Surface1, Surface2, Surface3 tone.Key

	// Code, told apart by what each word is. Nowhere but in code.
	CodeKeyword, CodeType, CodeFunction, CodeString, CodeNumber tone.Key

	// States. On an icon or a small tag, never as decoration.
	Error, Warning, Info, Success tone.Key
}

var Dark = Theme{
	Primary:   tone.Charple,
	Secondary: tone.Dolly,
	Accent:    tone.Bok,
	OnPrimary: tone.Butter,

	Text1: tone.Sash,
	Text2: tone.Smoke,
	Text3: tone.Squid,
	Text4: tone.Oyster,
	Line:  tone.Char,

	Surface0: tone.Pepper,
	Surface1: tone.BBQ,
	Surface2: tone.Char,
	Surface3: tone.Iron,

	CodeKeyword:  tone.Orchid,
	CodeType:     tone.Lichen,
	CodeFunction: tone.Anchovy,
	CodeString:   tone.Mojito,
	CodeNumber:   tone.Yam,

	Error:   tone.Sriracha,
	Warning: tone.Mustard,
	Info:    tone.Malibu,
	Success: tone.Julep,
}

// Light keeps Dark's ladder of contrasts on a pale ground. A hue that is
// bright there is the darkest step Charmtone has of it here: the bright
// steps vanish on Salt.
var Light = Theme{
	Primary:   tone.Charple,
	Secondary: tone.Macaron,
	Accent:    tone.Pickle,
	OnPrimary: tone.Butter,

	Text1: tone.Pepper,
	Text2: tone.Char,
	Text3: tone.Oyster,
	Text4: tone.Squid,
	Line:  tone.Steep,

	Surface0: tone.Salt,
	Surface1: tone.Sash,
	Surface2: tone.Steep,
	Surface3: tone.Smoke,

	CodeKeyword:  tone.Grape,
	CodeType:     tone.Damson,
	CodeFunction: tone.Thunder,
	CodeString:   tone.Pickle,
	CodeNumber:   tone.Tahini,

	Error:   tone.Sriracha,
	Warning: tone.Tandoor,
	Info:    tone.Damson,
	Success: tone.Pickle,
}

// For is the theme for a terminal whose background is dark, or is not.
func For(dark bool) Theme {
	if dark {
		return Dark
	}
	return Light
}

// Literal is a colour a document names by its value, so that it can be
// shown as itself: #rgb, #rrggbb, #rrggbbaa, rgb(r, g, b) or hsl(h, s%, l%),
// as CSS writes them. How see-through it is, if it says, is left out. It is
// what the document is about, and never a colour to draw the interface in:
// the test lets only the viewer ask.
func Literal(s string) (color.Color, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if hex, ok := strings.CutPrefix(s, "#"); ok {
		switch len(hex) {
		case 3:
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		case 8:
			hex = hex[:6]
		}
		n, err := strconv.ParseUint(hex, 16, 32)
		return literal(n), err == nil && len(hex) == 6
	}
	kind, rest, ok := strings.Cut(s, "(")
	if !ok || !strings.HasSuffix(rest, ")") {
		return nil, false
	}
	// Commas, spaces and the slash before an alpha all part the numbers.
	parts := strings.FieldsFunc(rest[:len(rest)-1], func(r rune) bool { return r == ',' || r == ' ' || r == '/' })
	if len(parts) < 3 {
		return nil, false
	}
	var n [3]float64
	for i := range n {
		v, err := strconv.ParseFloat(strings.TrimRight(parts[i], "%deg"), 64)
		if err != nil || v < 0 {
			return nil, false
		}
		n[i] = v
	}
	channel := func(v float64) literal { return literal(min(255, v+0.5)) }
	switch kind {
	case "rgb", "rgba":
		return channel(n[0])<<16 | channel(n[1])<<8 | channel(n[2]), true
	case "hsl", "hsla":
		h, sat, l := n[0]/30, min(1, n[1]/100), min(1, n[2]/100)
		a := sat * min(l, 1-l)
		// https://www.w3.org/TR/css-color-4/#hsl-to-rgb
		f := func(k float64) literal {
			k = math.Mod(k+h, 12)
			return channel(255 * (l - a*max(-1, min(k-3, 9-k, 1))))
		}
		return f(0)<<16 | f(8)<<8 | f(4), true
	}
	return nil, false
}

// literal is 0xrrggbb.
type literal uint32

func (l literal) RGBA() (r, g, b, a uint32) {
	// 8 bits a channel to 16, as color.Color wants them.
	return uint32(l>>16&0xff) * 0x101, uint32(l>>8&0xff) * 0x101, uint32(l&0xff) * 0x101, 0xffff
}
