// Package ui is the intro page: the screen starts covered in slop, a blade
// scrapes it off, and what is under it is the app.
package ui

import (
	"context"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/harmonica"

	"strigil/internal/kit/theme"
)

const (
	fps = 30
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
	// The sentences laid over the junk this time.
	lines []said

	rng *rand.Rand
}

type bit struct {
	ch   rune
	from rgb // the colour it came off as
	age  int // frames since it was knocked loose
	body *harmonica.Projectile
}

type tickMsg struct{}

// New makes the page with sentences picked afresh.
func New() Model { return seeded(uint64(time.Now().UnixNano())) }

// seeded makes the page with the same pick every time, for showing it.
func seeded(seed uint64) Model {
	return Model{
		sweep: harmonica.NewSpring(harmonica.FPS(fps), 1.1, 1),
		rng:   rand.New(rand.NewPCG(seed, 11)),
	}.soil()
}

func Run(ctx context.Context) error {
	use(theme.For(lipgloss.HasDarkBackground(os.Stdin, os.Stdout)))
	_, err := tea.NewProgram(New(), tea.WithContext(ctx)).Run()
	return err
}

// soil covers the screen again and sends the blade back to the left.
func (m Model) soil() Model {
	m.clean = make([]bool, m.w*m.h)
	m.blade, m.bladeV, m.scraping, m.wait = -bulge-4, 0, false, 0.9
	m.bits = nil
	m.lines = scatter(m.rng, m.w, m.h)
	return m
}

// busy says whether anything is still moving.
func (m Model) busy() bool { return m.wait > 0 || m.scraping || len(m.bits) > 0 }

func (m Model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	return tea.Tick(time.Second/fps, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		covered := m.scraping || m.wait > 0
		m.w, m.h = msg.Width, msg.Height
		m.clean = make([]bool, m.w*m.h)
		m.lines = scatter(m.rng, m.w, m.h)
		if !covered {
			// Already scraped: stay scraped at the new size.
			for i := range m.clean {
				m.clean[i] = true
			}
		}
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "space":
			was := m.busy()
			m = m.soil()
			if !was {
				return m, tick()
			}
		}
	case tickMsg:
		// Once the page is clean nothing moves, so the clock stops.
		if m = m.step(); m.busy() {
			return m, tick()
		}
	}
	return m, nil
}

// edge is the blade's edge on a row: furthest forward at mid-height.
func (m Model) edge(row int) float64 {
	mid := float64(m.h) / 2
	n := (float64(row) + 0.5 - mid) / mid
	return m.blade - bulge*n*n
}

// step moves everything on by one frame.
func (m Model) step() Model {
	if m.w == 0 {
		return m
	}
	switch {
	case m.wait > 0:
		if m.wait -= 1.0 / fps; m.wait <= 0 {
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
						harmonica.FPS(fps),
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

func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.WindowTitle = "strigil"
	v.BackgroundColor = pal.Surface0
	if m.w >= 60 && m.h >= 12 {
		v.SetContent(m.render())
	}
	return v
}

// render draws the frame: the app where the blade has been, slop where it
// has not, and over both the blade and the falling bits.
func (m Model) render() string {
	page := app(m.w, m.h)

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
		edge := m.edge(row)
		var last [2]rgb
		lastAttr := ""
		first := true
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
			under := page.cells[i]
			ahead := float64(x) - edge
			wake := ground
			if m.scraping && ahead < 0 && -ahead <= tail {
				wake = trail(-ahead / tail)
			}
			chip, falling := fall[i]
			switch {
			// The bar is the app's own edge: slop never covers it.
			case row == m.h-1:
				attr(under.attr)
				put(under.fg, under.bg, under.ch)
			// The rim is half a cell wide on the side the edge is on, so
			// the curve is not a staircase.
			case m.scraping && ahead >= 0 && ahead < 1:
				attr("")
				if edge-math.Floor(edge) < 0.5 {
					put(rim, ground, "▌")
				} else {
					put(rim, ground, "█")
				}
			case !m.clean[i]:
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
			// What flies off keeps the colour it had, and cools.
			case falling && under.ch == " ":
				attr("")
				put(chip.from.mix(slopInk, float64(chip.age)/(fps*0.8)), wake, string(chip.ch))
			case wake != ground:
				attr("")
				put(wake, wake, " ")
			default:
				attr(under.attr)
				put(under.fg, under.bg, under.ch)
			}
		}
	}
	out.WriteString("\x1b[m")
	return out.String()
}
