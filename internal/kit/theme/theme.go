// Package theme is every colour strigil draws with: Crush's default theme,
// "Charmtone Pantera", over the Charmtone palette.
//
// A colour on screen is one of these, or a blend between two of them. The
// test in this package fails on a colour made any other way. What each tier
// is for is in docs/ux/README.md.
package theme

import "charm.land/tone"

// Brand.
const (
	Primary   = tone.Charple // the active frame, the chosen row
	Secondary = tone.Dolly   // the live point: cursor, focused control
	Accent    = tone.Bok     // the input prompt
	OnPrimary = tone.Butter  // text on a Primary or Secondary fill
)

// Text, from what is read first to what is there if you look.
const (
	Text1 = tone.Sash   // names, titles, values
	Text2 = tone.Smoke  // long-form body
	Text3 = tone.Squid  // metadata, a key in the help line
	Text4 = tone.Oyster // what a key does, placeholders
	Line  = tone.Char   // rules and separators
)

// Surfaces, from the app's own background outward.
const (
	Surface0 = tone.Pepper
	Surface1 = tone.BBQ
	Surface2 = tone.Char
	Surface3 = tone.Iron
)

// Code, told apart by what each word is. Nowhere but in code.
const (
	CodeKeyword  = tone.Orchid
	CodeType     = tone.Lichen
	CodeFunction = tone.Anchovy
	CodeString   = tone.Mojito
	CodeNumber   = tone.Yam
)

// States. On an icon or a small tag, never as decoration.
const (
	Error   = tone.Sriracha
	Warning = tone.Mustard
	Info    = tone.Malibu
	Success = tone.Julep
)
