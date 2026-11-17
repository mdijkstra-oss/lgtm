package ui

import (
	"image/color"

	"strigil/internal/kit/theme"
)

// rgb is a colour as three numbers from 0 to 1, for doing sums with.
type rgb struct{ r, g, b float64 }

func of(c color.Color) rgb {
	r, g, b, _ := c.RGBA()
	return rgb{float64(r) / 0xffff, float64(g) / 0xffff, float64(b) / 0xffff}
}

func (a rgb) mix(b rgb, t float64) rgb {
	t = min(1, max(0, t))
	return rgb{a.r + (b.r-a.r)*t, a.g + (b.g-a.g)*t, a.b + (b.b-a.b)*t}
}

func (a rgb) bytes() (r, g, b int) {
	return int(a.r*255 + 0.5), int(a.g*255 + 0.5), int(a.b*255 + 0.5)
}

var (
	ground = of(theme.Surface0)
	plain  = of(theme.Text1)
	calm   = of(theme.Text3) // a key
	hush   = of(theme.Text4) // what a key does
	rule   = of(theme.Line)

	// The blade: a pale rim, then the two brand colours cooling into the dark.
	hot  = of(theme.Secondary)
	deep = of(theme.Primary)
	rim  = hot.mix(of(theme.OnPrimary), 0.6)

	slopInk  = of(theme.Text4)
	slopBack = of(theme.Surface3)
)

// The markers the sentences are gone over with: each runs from one colour
// to another along the sentence, and the words are punched out of it.
var markers = [][2]rgb{
	{of(theme.Secondary), of(theme.Warning)},
	{of(theme.Accent), of(theme.Info)},
	{of(theme.Warning), of(theme.Success)},
	{of(theme.Info), of(theme.Secondary)},
	{of(theme.Secondary), of(theme.Accent)},
}

// ink is the marker's colour over column x of a sentence.
func (s *said) ink(x int) rgb {
	m := markers[s.hue%len(markers)]
	return m[0].mix(m[1], float64(x-s.x+1)/float64(len(s.text)+1))
}

// trail is the colour left t of the way from the blade to the end of its wake.
func trail(t float64) rgb {
	if t < 0.2 {
		return hot.mix(deep, t/0.2)
	}
	u := (t - 0.2) / 0.8
	return deep.mix(ground, u*(2-u))
}
