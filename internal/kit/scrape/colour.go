package scrape

import (
	"image/color"

	"lgtm/internal/kit/theme"
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
	pal theme.Theme

	ground, plain rgb

	// The blade: a pale rim, then the two brand colours cooling into the
	// ground.
	hot, deep, rim rgb

	slopInk, slopBack rgb

	// The marker the word is gone over with: it runs from one brand colour
	// to the other along the word, and the code is punched out of it.
	marker [2]rgb
)

func init() { Use(theme.Dark) }

// Use draws everything from here on in that theme.
func Use(t theme.Theme) {
	pal = t
	ground, plain = of(t.Surface0), of(t.Text1)
	hot, deep = of(t.Secondary), of(t.Primary)
	rim = hot.mix(of(t.OnPrimary), 0.6)
	slopInk, slopBack = of(t.Text4), of(t.Surface3)
	marker = [2]rgb{of(t.Primary), of(t.Secondary)}
}

// ink is the marker's colour over column x of the word.
func (s *run) ink(x int) rgb {
	return marker[0].mix(marker[1], float64(x-s.from)/float64(max(1, s.span)))
}

// trail is the colour left t of the way from the blade to the end of its wake.
func trail(t float64) rgb {
	if t < 0.2 {
		return hot.mix(deep, t/0.2)
	}
	u := (t - 0.2) / 0.8
	return deep.mix(ground, u*(2-u))
}
