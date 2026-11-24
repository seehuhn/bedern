// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// TestLettersAreSmooth asks every letter for a clean build: no warning of
// the stroke tool and no finding of the italic profile of the smoothness
// check which is not marked as intended.
func TestLettersAreSmooth(t *testing.T) {
	for _, name := range Names() {
		g := glyph(t, name)
		for _, w := range g.Warnings {
			t.Errorf("%s: warning: %s", name, w)
		}
		for _, f := range g.Findings {
			t.Errorf("%s: %s at %.0f,%.0f: %s", name, f.Kind, f.X, f.Y, f.Detail)
		}
	}
}

// TestStemsOnTheLattice asks that every straight segment of the slant of a
// stem has its ends at heights which are multiples of SlantDen and on the
// line exactly, and that the handles of the curves next to it lie on its
// line: the rounding leaves the stems and the joins of curves to them as
// drawn.
func TestStemsOnTheLattice(t *testing.T) {
	p := Default()
	for _, name := range Names() {
		checkLattice(t, p, name, glyph(t, name).Path)
	}
}

// checkLattice reports the stems of the outline whose ends are not at
// heights which are multiples of SlantDen or whose neighbouring handles are
// off their line.
func checkLattice(t *testing.T, p Params, name string, path outline.Path) {
	t.Helper()
	k := p.slant()
	den := float64(p.SlantDen)
	for ci, c := range path {
		n := len(c.Segs)
		start := func(i int) Point {
			if i == 0 {
				return c.Start
			}
			return c.Segs[i-1].P3
		}
		for i, s := range c.Segs {
			a, b := start(i), s.P3
			if s.IsCurve || a.Y == b.Y || math.Abs((b.X-a.X)/(b.Y-a.Y)-k) > 1e-9 {
				continue
			}
			for _, q := range []Point{a, b} {
				if math.Mod(q.Y, den) != 0 {
					t.Errorf("%s: contour %d: the stem ends at %v, not at a multiple of %v", name, ci, q, den)
				}
			}
			// The handle of a curve which joins the stem smoothly, before
			// it and after it, lies on its line; a corner (the crotch of
			// a join) is left out.
			on := func(q Point) bool { return math.Abs((q.X-a.X)-k*(q.Y-a.Y)) < 1e-9 }
			smooth := func(q, end Point) bool {
				d, e := Point{X: b.X - a.X, Y: b.Y - a.Y}, Point{X: q.X - end.X, Y: q.Y - end.Y}
				cos := (d.X*e.X + d.Y*e.Y) / math.Hypot(d.X, d.Y) / math.Hypot(e.X, e.Y)
				return math.Abs(cos) > math.Cos(15*math.Pi/180)
			}
			if prev := c.Segs[(i+n-1)%n]; prev.IsCurve && smooth(prev.P2, a) && !on(prev.P2) {
				t.Errorf("%s: contour %d: the handle %v before the stem at %v is off its line", name, ci, prev.P2, a)
			}
			if next := c.Segs[(i+1)%n]; next.IsCurve && smooth(next.P1, b) && !on(next.P1) {
				t.Errorf("%s: contour %d: the handle %v after the stem at %v is off its line", name, ci, next.P1, b)
			}
		}
	}
}

// TestColour asks that "nnonoonn" in the italic has 0.92 to 0.96 of the
// colour of the roman.
func TestColour(t *testing.T) {
	ci, cr, err := Colour(Default(), "nnonoonn")
	if err != nil {
		t.Fatal(err)
	}
	if r := ci / cr; r < 0.92 || r > 0.96 {
		t.Errorf("the colour of nnonoonn is %.3f of the roman, want 0.92 to 0.96", r)
	}
}

// TestStemWeight asks that every stem is Stem wide within 1 unit,
// horizontally, along its straight part.
func TestStemWeight(t *testing.T) {
	p := Default()
	e1 := p.SideStemLeft
	c1 := e1 + p.Stem/2
	c2 := c1 + p.Stem + p.NCounter
	cases := []struct {
		name string
		c    float64
		y0   float64
		y1   float64
	}{
		{"n", c1, 20, 230},
		{"n", c2, p.ExitTurn + 4, p.Shoulder - 4},
		{"i", c1, p.ExitTurn + 4, p.XHeight - p.EntryJoin - 4},
		{"u", c1, p.XHeight - p.Shoulder + 4, p.XHeight - p.EntryJoin - 4},
		{"u", c2, p.XHeight - p.Branch + 24, p.XHeight - 4},
	}
	for _, cs := range cases {
		g := glyph(t, cs.name)
		for y := cs.y0 + 0.5; y <= cs.y1; y += 8 {
			r := stemRun(t, p, g.Path, cs.c, y)
			if w := r[1] - r[0]; math.Abs(w-p.Stem) > 1 {
				t.Errorf("%s: the stem at %.0f is %.1f wide at %.0f, want %v", cs.name, cs.c, w, y, p.Stem)
			}
		}
	}
}
