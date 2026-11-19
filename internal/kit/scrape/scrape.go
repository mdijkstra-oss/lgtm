// Package scrape is a page covered in slop and the blade that scrapes it
// off: what a page of lgtm's opens from. It is given the page and draws the
// slop over it; it knows nothing of what the page is.
package scrape

import (
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/x/ansi"
)

const (
	// FPS is how often Step is to be called while Busy.
	FPS = 30
	// The blade's edge bulges forward by this many cells at mid-height.
	bulge = 7.0
)

type Model struct {
	w, h int

	// The blade: where its edge is at mid-height, and how fast it moves.
	blade, bladeV float64
	sweep         harmonica.Spring
	scraping      bool
	wait          float64 // seconds the slop sits before the blade comes

	// clean says which cells the blade has been over.
	clean []bool
	// What the blade knocked loose, falling.
	bits []bit
	// The word stamped over the junk.
	lines []run

	rng *rand.Rand
}

type bit struct {
	ch   rune
	from rgb // the colour it came off as
	age  int // frames since it was knocked loose
	body *harmonica.Projectile
}

// New covers a page.
func New() Model { return Seeded(uint64(time.Now().UnixNano())) }

// Seeded covers a page so that what falls falls the same way every time,
// for showing it.
func Seeded(seed uint64) Model {
	return Model{
		sweep: harmonica.NewSpring(harmonica.FPS(FPS), 1.1, 1),
		rng:   rand.New(rand.NewPCG(seed, 11)),
	}.Soil()
}

// Soil covers the page again and sends the blade back to the left.
func (m Model) Soil() Model {
	m.clean = make([]bool, m.w*m.h)
	m.blade, m.bladeV, m.scraping, m.wait = -bulge-4, 0, false, 0.9
	m.bits = nil
	m.lines = stamp(m.w, m.h)
	return m
}

// Size is the page's size. One already scraped stays scraped.
func (m Model) Size(w, h int) Model {
	// The zero Model is a page nothing was ever put over.
	if m.rng == nil {
		return m
	}
	covered := m.scraping || m.wait > 0
	m.w, m.h = w, h
	m.clean = make([]bool, w*h)
	m.lines = stamp(w, h)
	if !covered {
		for i := range m.clean {
			m.clean[i] = true
		}
	}
	return m
}

// Busy says whether anything is still moving.
func (m Model) Busy() bool { return m.wait > 0 || m.scraping || len(m.bits) > 0 }

// edge is the blade's edge on a row: furthest forward at mid-height.
func (m Model) edge(row int) float64 {
	mid := float64(m.h) / 2
	n := (float64(row) + 0.5 - mid) / mid
	return m.blade - bulge*n*n
}

// Step moves everything on by one frame.
func (m Model) Step() Model {
	if m.w == 0 {
		return m
	}
	switch {
	case m.wait > 0:
		if m.wait -= 1.0 / FPS; m.wait <= 0 {
			m.scraping = true
		}
	case m.scraping:
		// The spring pulls toward a point well past the right edge, so the
		// blade is still moving when it leaves.
		m.blade, m.bladeV = m.sweep.Update(m.blade, m.bladeV, float64(m.w)*1.45)
		for row := range m.h {
			to := int(math.Floor(m.edge(row)))
			for x := 0; x <= to && x < m.w; x++ {
				i := row*m.w + x
				if m.clean[i] {
					continue
				}
				m.clean[i] = true
				// Some of what is scraped off flies ahead and falls.
				if ch, mark, _ := slopAt(m.lines, x, row); ch != ' ' && m.rng.IntN(9) == 0 && len(m.bits) < 300 {
					from := rim
					if mark != nil {
						from = mark.ink(x)
					}
					m.bits = append(m.bits, bit{ch: ch, from: from, body: harmonica.NewProjectile(
						harmonica.FPS(FPS),
						harmonica.Point{X: float64(x), Y: float64(row)},
						harmonica.Vector{X: m.bladeV*0.6 + m.rng.Float64()*14, Y: -3 - m.rng.Float64()*9},
						harmonica.Vector{Y: 46},
					)})
				}
			}
		}
		if m.blade-bulge > float64(m.w)+2 {
			m.scraping = false
		}
	}
	alive := m.bits[:0:0]
	for _, b := range m.bits {
		if p := b.body.Update(); p.Y < float64(m.h) && p.X < float64(m.w)+1 {
			b.age++
			alive = append(alive, b)
		}
	}
	m.bits = alive
	return m
}

// Over draws the frame: the page where the blade has been, slop where it
// has not, and over both the blade and the falling bits. The page is a row
// a string, each as wide as the page.
func (m Model) Over(page []string) string {
	fall := map[int]bit{}
	for _, b := range m.bits {
		p := b.body.Position()
		if x, y := int(p.X), int(p.Y); x >= 0 && x < m.w && y >= 0 && y < m.h-1 {
			fall[y*m.w+x] = b
		}
	}

	// The faster the blade, the longer its wake.
	tail := min(14, max(3, m.bladeV/5))

	var out strings.Builder
	out.Grow(m.w * m.h * 24)
	for row := range m.h {
		if row > 0 {
			out.WriteString("\x1b[m\n")
		}
		under := ""
		if row < len(page) {
			under = page[row]
		}
		// What is on the page with its colours off, to find its blanks.
		bare := []rune(ansi.Strip(under))
		edge := m.edge(row)
		var last [2]rgb
		lastAttr := ""
		first := true
		// The page shows through in runs: run is where this one began.
		run := -1
		shown := func(to int) {
			if run < 0 {
				return
			}
			out.WriteString("\x1b[m" + ansi.Cut(under, run, to) + "\x1b[m")
			run, lastAttr, first = -1, "", true
		}
		attr := func(a string) {
			if a != lastAttr {
				// Attributes are reset, then the colours set again.
				out.WriteString("\x1b[m")
				if a != "" {
					out.WriteString("\x1b[" + a + "m")
				}
				lastAttr, first = a, true
			}
		}
		put := func(fg, bg rgb, s string) {
			if first || fg != last[0] {
				r, g, b := fg.bytes()
				out.WriteString("\x1b[38;2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b) + "m")
			}
			if first || bg != last[1] {
				r, g, b := bg.bytes()
				out.WriteString("\x1b[48;2;" + strconv.Itoa(r) + ";" + strconv.Itoa(g) + ";" + strconv.Itoa(b) + "m")
			}
			last, first = [2]rgb{fg, bg}, false
			out.WriteString(s)
		}
		for x := range m.w {
			i := row*m.w + x
			ahead := float64(x) - edge
			wake := ground
			if m.scraping && ahead < 0 && -ahead <= tail {
				wake = trail(-ahead / tail)
			}
			chip, falling := fall[i]
			switch {
			// The last row is the page's own edge: slop never covers it.
			case row == m.h-1:
				if run < 0 {
					run = x
				}
			// The rim is half a cell wide on the side the edge is on, so
			// the curve is not a staircase.
			case m.scraping && ahead >= 0 && ahead < 1:
				shown(x)
				attr("")
				if edge-math.Floor(edge) < 0.5 {
					put(rim, ground, "▌")
				} else {
					put(rim, ground, "█")
				}
			case !m.clean[i]:
				shown(x)
				ch, mark, faint := slopAt(m.lines, x, row)
				// Slop lights up just ahead of the blade.
				near := 0.0
				if m.scraping && ahead < 7 {
					near = 0.6 * (1 - ahead/7)
				}
				switch {
				case mark != nil:
					attr("1")
					put(ground, mark.ink(x).mix(plain, near), string(ch))
				case faint:
					attr("")
					put(slopBack.mix(plain, near), ground, string(ch))
				default:
					attr("")
					put(slopInk.mix(plain, near), ground, string(ch))
				}
			// What flies off keeps the colour it had, and cools. It
			// falls through the blanks of the page, never over its words.
			case falling && (x >= len(bare) || bare[x] == ' '):
				shown(x)
				attr("")
				put(chip.from.mix(slopInk, float64(chip.age)/(FPS*0.8)), wake, string(chip.ch))
			case wake != ground:
				shown(x)
				attr("")
				put(wake, wake, " ")
			case run < 0:
				run = x
			}
		}
		shown(m.w)
	}
	out.WriteString("\x1b[m")
	return out.String()
}
