// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"fmt"
	"math"
	"slices"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

var (
	builtOnce sync.Once
	built     map[string]*Glyph
	builtErr  error
)

// all returns the glyphs of the package, built once for all tests.
func all(t *testing.T) map[string]*Glyph {
	t.Helper()
	builtOnce.Do(func() {
		gs, err := Glyphs(Default())
		builtErr = err
		built = map[string]*Glyph{}
		for _, g := range gs {
			built[g.Name] = g
		}
	})
	if builtErr != nil {
		t.Fatal(builtErr)
	}
	return built
}

// glyph returns the built glyph of the given name, or skips the test.
func glyph(t *testing.T, name string) *Glyph {
	t.Helper()
	g, ok := all(t)[name]
	if !ok {
		t.Skipf("no glyph %s", name)
	}
	return g
}

// eachSegment calls f for every segment of the path with its start.
func eachSegment(p outline.Path, f func(p0 Point, s outline.Segment)) {
	for _, c := range p {
		p0 := c.Start
		for _, s := range c.Segs {
			f(p0, s)
			p0 = s.P3
		}
	}
}

// leaningStems are the glyphs whose stems are not vertical in the roman.
var leaningStems = []string{"dagger", "daggerdbl", "uni203D", "uni2E18"}

// TestSlantedStemsOnTheLattice: in the rounded outline every straight
// segment at least minStem high which runs within 2 degrees of the slant runs
// exactly at the slant between ends at heights which are multiples of 4,
// except an end at a corner on the cap height or the figure height (the
// outer edges of the beaks of E, F, T and Z and the flats of the figures),
// where the line keeps within half a unit of the lattice line through its
// other end, and
// the handle of a curve which leaves such a stem along it lies on its line at
// such a height.  The one exception is an end at a corner on the cap height
// or the figure height, which are not multiples of 4: there the end lies
// within half a unit of the line of the slant through the other end, which
// is on the lattice.
//
// The stems of the daggers taper from the arms to the ends, and the stems of
// the interrobangs lean two degrees in the roman: their edges are no
// vertical lines of the roman for the lattice to take; after the shear they
// run within 2 degrees of the slant and are not looked at.
func TestSlantedStemsOnTheLattice(t *testing.T) {
	for _, g := range all(t) {
		if sources[g.Name].upright || slices.Contains(leaningStems, g.Name) {
			continue
		}
		for _, msg := range latticeProblems(g) {
			t.Error(msg)
		}
	}
}

// latticeProblems returns the places of the glyph at which a straight
// segment of the slant is off the lattice, as TestSlantedStemsOnTheLattice
// describes them.
func latticeProblems(g *Glyph) []string {
	k := Default().slant()
	var out []string
	for _, c := range g.Path {
		n := len(c.Segs)
		for i, s := range c.Segs {
			p0 := c.Start
			if i > 0 {
				p0 = c.Segs[i-1].P3
			}
			d := sub(s.P3, p0)
			if s.IsCurve || math.Abs(d.Y) < minStem || math.Abs(math.Atan(d.X/d.Y)-math.Atan(k)) > 2*math.Pi/180 {
				continue
			}
			fixedEnd := func(q Point) bool { return q.Y == 683 || q.Y == 666 }
			switch {
			case math.Mod(p0.Y, 4) == 0 && math.Mod(s.P3.Y, 4) == 0 && d.X == k*d.Y:
			case math.Mod(p0.Y, 4) == 0 && fixedEnd(s.P3) && math.Abs(d.X-k*d.Y) <= 0.5,
				math.Mod(s.P3.Y, 4) == 0 && fixedEnd(p0) && math.Abs(d.X-k*d.Y) <= 0.5:
			default:
				out = append(out, fmt.Sprintf("%s: stem from %v to %v is not on the lattice", g.Name, p0, s.P3))
				continue
			}
			check := func(q, h Point) {
				e := sub(h, q)
				if e.Y == 0 || math.Abs(math.Atan(e.X/e.Y)-math.Atan(k)) > 2*math.Pi/180 {
					return
				}
				if e.X != k*e.Y || math.Mod(h.Y, 4) != 0 {
					out = append(out, fmt.Sprintf("%s: handle %v next to the stem at %v is off the lattice", g.Name, h, q))
				}
			}
			if nx := c.Segs[(i+1)%n]; nx.IsCurve {
				check(s.P3, nx.P1)
			}
			if pv := c.Segs[(i-1+n)%n]; pv.IsCurve {
				check(p0, pv.P2)
			}
		}
	}
	return out
}

// TestCapStemWidth: the stems of H and I are CapStem wide horizontally,
// within 1, along their straight part.
func TestCapStemWidth(t *testing.T) {
	p := Default()
	for _, name := range []string{"H", "I", "E", "L"} {
		g := glyph(t, name)
		for _, y := range []float64{150, 250, 450, 550} {
			xs := g.Path.CrossingsX(y)
			if len(xs) < 2 {
				t.Fatalf("%s: no ink at %g", name, y)
			}
			if w := xs[1] - xs[0]; math.Abs(w-p.CapStem) > 1 {
				t.Errorf("%s: stem at %g is %.1f wide, want %g", name, y, w, p.CapStem)
			}
		}
	}
}

// TestBarsStayHorizontal: the bar of H keeps both edges horizontal and is
// between 34 and 38 thick.
func TestBarsStayHorizontal(t *testing.T) {
	g := glyph(t, "H")
	var flats []float64
	eachSegment(g.Path, func(p0 Point, s outline.Segment) {
		if !s.IsCurve && p0.Y == s.P3.Y && p0.Y > 200 && p0.Y < 500 && math.Abs(s.P3.X-p0.X) > 100 {
			flats = append(flats, p0.Y)
		}
	})
	if len(flats) != 2 {
		t.Fatalf("H: %d horizontal edges of the bar, want 2", len(flats))
	}
	if d := math.Abs(flats[0] - flats[1]); d < 34 || d > 38 {
		t.Errorf("H: bar %.0f thick, want 34 to 38", d)
	}
}

// TestRoundStress: the stress of O leans 7 to 11 degrees forward, with the
// thinnest place at the top left of the topmost point of the outline.
func TestRoundStress(t *testing.T) {
	for _, name := range []string{"O"} {
		g := glyph(t, name)
		x0, y0, x1, y1 := g.Path.BBox()
		c := pt((x0+x1)/2, (y0+y1)/2)
		a, top, _, _, _ := Stress(g.Path, c)
		if a < 7 || a > 11 {
			t.Errorf("%s: stress %.1f degrees, want 7 to 11", name, a)
		}
		xs := g.Path.CrossingsY(top.X)
		_ = xs
		var topX, topY float64
		for _, ct := range g.Path {
			for _, q := range ct.Flatten() {
				if q.Y > topY {
					topX, topY = q.X, q.Y
				}
			}
		}
		if top.X >= topX {
			t.Errorf("%s: thinnest place at the top %.0f is not left of the top %.0f", name, top.X, topX)
		}
	}
}

// TestFiguresShareTheAdvance: the ten figures have one advance, and each
// is centred in it by its ink unsheared about half the figure height, within
// 16, the largest optical shift of the roman (3 and 5).
func TestFiguresShareTheAdvance(t *testing.T) {
	var adv float64
	for i, n := range []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"} {
		g := glyph(t, n)
		if i == 0 {
			adv = g.Advance
		}
		if g.Advance != adv {
			t.Errorf("%s: advance %g, zero has %g", n, g.Advance, adv)
		}
		// The ink, unsheared about half the figure height.
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, c := range g.Path {
			for _, q := range c.Flatten() {
				u := q.X - (q.Y-333)/4
				lo, hi = math.Min(lo, u), math.Max(hi, u)
			}
		}
		if m := (lo + hi) / 2; math.Abs(m-adv/2) > 16 {
			t.Errorf("%s: ink centred at %.0f, the cell at %.0f", n, m, adv/2)
		}
	}
}

// TestDotsStayRound: the dots of the period, the colon and the question
// mark are circles in the italic: their width and height agree within 2.
func TestDotsStayRound(t *testing.T) {
	for _, n := range []string{"period", "colon", "question", "exclam"} {
		g := glyph(t, n)
		for _, c := range g.Path {
			x0, y0, x1, y1 := outline.Path{c}.BBox()
			if y1 > 150 {
				continue
			}
			if d := (x1 - x0) - (y1 - y0); math.Abs(d) > 2 {
				t.Errorf("%s: dot %.1f wide and %.1f high", n, x1-x0, y1-y0)
			}
		}
	}
}

// TestSignDotsStayRound: the dots of the ellipsis, the middle dot and the
// bullet are circles in the italic, which the shear alone makes ellipses:
// their width and height agree within 2.
func TestSignDotsStayRound(t *testing.T) {
	for _, n := range []string{"ellipsis", "periodcentered", "bullet"} {
		g := glyph(t, n)
		if len(g.Path) == 0 {
			t.Errorf("%s: no dots", n)
		}
		for _, c := range g.Path {
			x0, y0, x1, y1 := outline.Path{c}.BBox()
			if d := (x1 - x0) - (y1 - y0); math.Abs(d) > 2 {
				t.Errorf("%s: dot %.1f wide and %.1f high", n, x1-x0, y1-y0)
			}
		}
	}
}

// TestSmooth: the italic profile of the smoothness check reports nothing
// which is not marked as intended.
func TestSmooth(t *testing.T) {
	for _, g := range all(t) {
		for _, f := range Findings(g) {
			t.Errorf("%s: %s at %.0f,%.0f (dir %.0f): %s", g.Name, f.Kind, f.X, f.Y, f.Dir, f.Detail)
		}
	}
}
