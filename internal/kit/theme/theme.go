// Package theme is every colour strigil draws with: a dark and a light
// theme over the Charmtone palette. Dark is Crush's "Charmtone Pantera".
//
// A colour on screen is one of these, or a blend between two of them. The
// test in this package fails on a colour made any other way. What each tier
// is for is in docs/ux/README.md.
package theme

import "charm.land/tone"

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
