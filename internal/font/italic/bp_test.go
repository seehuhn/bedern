// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// checkEntryOf checks the entry of the stem whose left edge stands at edge,
// up above the x-height, as TestEntry does: the opening under the hook is at
// least 40 at the height of its start, and the hook rises to the top of o.
func checkEntryOf(t *testing.T, p Params, name string, edge, up float64) {
	t.Helper()
	g := glyph(t, name)
	u := unsheared(p, g.Path)
	y := p.EntryLow + up + 0.5
	rs := runs(u, y)
	gap := -1.0
	for i := 0; i+1 < len(rs); i++ {
		if rs[i][1] < edge && rs[i+1][0] > edge-p.Stem/2 {
			gap = rs[i+1][0] - rs[i][1]
		}
	}
	if gap < 40 {
		t.Errorf("%s: the opening under the entry is %.1f at %.0f, want at least 40", name, gap, y)
	}
	y1 := math.Inf(-1)
	for _, c := range g.Path {
		if _, _, _, top := bbox(outline.Path{c}); top < p.XHeight+up+50 {
			y1 = math.Max(y1, top)
		}
	}
	if top := p.XHeight + up + p.Overshoot; math.Abs(y1-top) > 1 {
		t.Errorf("%s: the entry rises to %.1f, want the top of o, %v", name, y1, top)
	}
}

// TestBPEntries asks the entries of b, at the ascender, and of p, at the
// x-height, for the opening and the height of every entry.
func TestBPEntries(t *testing.T) {
	p := Default()
	checkEntryOf(t, p, "b", p.SideStemLeft-p.BLeftLess, p.AscenderEntryRise)
	checkEntryOf(t, p, "p", p.SideStemLeft, 0)
}

// TestBBranch asks that the bowl of b closes at the top as the arch of n
// leaves its stem: its underside meets the stem at 0.53 to 0.62 of the
// x-height and the notch above it opens at more than 23 degrees.
func TestBBranch(t *testing.T) {
	p := Default()
	checkBranch(t, p, "b", p.SideStemLeft-p.BLeftLess+p.Stem/2)
}

// bowlCentre returns the middle of the box of the counter of the glyph,
// the contour which lies inside another.
func bowlCentre(t *testing.T, g *Glyph) Point {
	t.Helper()
	for _, c := range g.Path {
		x0, y0, x1, y1 := bbox(outline.Path{c})
		if y0 > -1 && y1 < 461 && x1-x0 < 300 {
			return Point{X: (x0 + x1) / 2, Y: (y0 + y1) / 2}
		}
	}
	t.Fatalf("%s: no counter", g.Name)
	return Point{}
}

// bowlStress returns the line of the stress of the bowl whose counter has
// its middle at c: on every vertical line through the middle three fifths
// of the counter, away from the joins to the stem, the
// thickness of the ink just above and just below the counter; the middle of
// the abscissae where the thickness lies within 1 unit of its least, at the
// top and at the bottom, with the middle of the ink there, and the lean of
// the line through the two from the vertical, in degrees.
func bowlStress(g *Glyph, c Point) (lean float64, top, bottom Point) {
	var hole outline.Path
	for _, ct := range g.Path {
		x0, y0, x1, y1 := bbox(outline.Path{ct})
		if x0 < c.X && c.X < x1 && y0 < c.Y && c.Y < y1 && y1 < 461 && x1-x0 < 300 {
			hole = outline.Path{ct}
		}
	}
	x0, _, x1, _ := bbox(hole)
	x0, x1 = x0+(x1-x0)/5, x1-(x1-x0)/5
	side := func(up bool) Point {
		type cut struct{ x, w, m float64 }
		var cs []cut
		for x := math.Ceil(x0) + 0.5; x < x1; x++ {
			ys := crossings(g.Path, Point{X: x, Y: c.Y}, Point{Y: 1})
			// The crossings nearest to the counter on the side: the edge of
			// the counter and the outer edge beyond it.
			var a, b float64
			ok := false
			for i := 0; i+1 < len(ys); i++ {
				if up && ys[i] > 0 && (!ok) {
					a, b, ok = ys[i], ys[i+1], true
				}
				if !up && ys[i+1] < 0 {
					a, b, ok = ys[i], ys[i+1], true
				}
			}
			if ok {
				cs = append(cs, cut{x, b - a, c.Y + (a+b)/2})
			}
		}
		least := math.Inf(1)
		for _, q := range cs {
			least = math.Min(least, q.w)
		}
		lo, hi := math.Inf(1), math.Inf(-1)
		var ys []float64
		for _, q := range cs {
			if q.w <= least+1 {
				lo, hi = math.Min(lo, q.x), math.Max(hi, q.x)
				ys = append(ys, q.m)
			}
		}
		return Point{X: (lo + hi) / 2, Y: ys[len(ys)/2]}
	}
	top, bottom = side(true), side(false)
	return math.Atan2(top.X-bottom.X, top.Y-bottom.Y) * 180 / math.Pi, top, bottom
}

// TestBPStress asks that the bowls of b and p take their stress from the
// pen: the line through their thinnest places at the top and the bottom
// leans 7 to 11 degrees forward from the vertical, the thinnest place at
// the top lies left of the topmost point of the bowl, and the bowl is 28 to
// 34 thick at its thinnest and 80 to 90 at its thickest.
func TestBPStress(t *testing.T) {
	p := Default()
	for _, cs := range []struct {
		name string
		e    float64
	}{{"b", p.SideStemLeft - p.BLeftLess}, {"p", p.SideStemLeft}} {
		g := glyph(t, cs.name)
		c := bowlCentre(t, g)
		t.Logf("%s: centre %v", cs.name, c)
		lean, top, bottom := bowlStress(g, c)
		t.Logf("%s: thinnest at %v and %v", cs.name, top, bottom)
		t.Logf("%s: stress %.1f degrees", cs.name, lean)
		if lean < 7 || lean > 11 {
			t.Errorf("%s: the stress of the bowl leans %.1f degrees, want 7 to 11", cs.name, lean)
		}
		// The topmost point of the bowl, right of the stem.
		var topX, topY float64
		for _, pts := range polys(g.Path) {
			for _, q := range pts {
				if q.Y <= p.XHeight+p.Overshoot+1 && q.X > top.X-100 && q.Y > topY {
					topX, topY = q.X, q.Y
				}
			}
		}
		// The top node keeps its thickness unchanging, as the peak of the
		// arch of n does; the thinnest place lies at the topmost
		// point, never right of it.
		if top.X > topX+1 {
			t.Errorf("%s: the thinnest place at the top is at %.0f, right of the topmost point at %.0f", cs.name, top.X, topX)
		}
		least, most := math.Inf(1), 0.0
		for a := -45.0; a <= 45; a++ {
			for _, b := range []float64{a, 90 + a, 270 + a} {
				w, _ := rayThickness(g.Path, c, b)
				least = math.Min(least, w)
				if b == a {
					most = math.Max(most, w)
				}
			}
		}
		if least < 28 || least > 34 {
			t.Errorf("%s: the thinnest place of the bowl is %.1f, want 28 to 34", cs.name, least)
		}
		if most < 80 || most > 90 {
			t.Errorf("%s: the thickest place of the bowl is %.1f, want 80 to 90", cs.name, most)
		}
	}
}

// TestBPSides asks that at ShearHeight b and p have the side of the round
// letters on the right, and on the left the side of n less BLeftLess for b
// and that of n for p.
func TestBPSides(t *testing.T) {
	p := Default()
	for _, cs := range []struct {
		name string
		left float64
	}{{"b", p.SideStemLeft - p.BLeftLess}, {"p", p.SideStemLeft}} {
		g := glyph(t, cs.name)
		rs := runs(g.Path, p.ShearHeight+0.5)
		l, r := rs[0][0], g.Advance-rs[len(rs)-1][1]
		if math.Abs(l-cs.left) > 1.5 || math.Abs(r-p.SideRoundRight) > 2 {
			t.Errorf("%s: the sides are %.1f and %.1f, want %v and %v", cs.name, l, r, cs.left, p.SideRoundRight)
		}
	}
}

// TestBRoundTurn asks that b has no foot: below the turn of the stem into
// the bowl no ink lies left of the line of the stem's left edge, and 20
// above the baseline the left edge lies at least 10 inside that line, so
// that the turn is round and not a corner.
func TestBRoundTurn(t *testing.T) {
	p := Default()
	g := glyph(t, "b")
	e := p.SideStemLeft - p.BLeftLess
	for y := 0.5; y < p.XHeight-p.Shoulder; y += 4 {
		line := italicX(p, e, y)
		rs := runs(g.Path, y)
		if len(rs) == 0 {
			continue
		}
		in := rs[0][0] - line
		if in < -0.5 {
			t.Errorf("b: ink %.1f left of the stem at %.0f", -in, y)
		}
		if y == 20.5 && in < 10 {
			t.Errorf("b: the left edge lies only %.1f inside the stem's line at %.0f, want at least 10", in, y)
		}
	}
}

// TestPBowlJoin asks that the counter of p meets the stem at the bottom at
// 0.10 to 0.20 of the x-height, and that both b and p have a closed counter.
func TestPBowlJoin(t *testing.T) {
	p := Default()
	for _, name := range []string{"b", "p"} {
		if n := len(glyph(t, name).Path); n != 2 {
			t.Errorf("%s: %d contours, want 2 (the outline and the counter)", name, n)
		}
	}
	g := glyph(t, "p")
	var hole outline.Path
	for _, c := range g.Path {
		_, y0, _, _ := bbox(outline.Path{c})
		if y0 > 0 {
			hole = append(hole, c)
		}
	}
	if len(hole) != 1 {
		t.Fatalf("p: no counter")
	}
	// The join: the lowest point of the counter within 15 units of the line
	// of the stem's right edge.
	y0 := math.Inf(1)
	for _, q := range polys(hole)[0] {
		if q.X < italicX(p, p.SideStemLeft+p.Stem, q.Y)+15 {
			y0 = math.Min(y0, q.Y)
		}
	}
	if f := y0 / p.XHeight; f < 0.10 || f > 0.20 {
		t.Errorf("p: the counter meets the stem at %.0f, %.2f of the x-height, want 0.10 to 0.20", y0, f)
	}
}

// TestPFoot asks that p stands on the descender line with a level foot
// serif: its lowest ink lies on the descender line along the whole serif,
// the top of the serif is level, and the serif on the obtuse side, the left,
// is 0.85 to 0.99 as long as the other, measured from the lines of the
// stem's edges at the middle of the serif.
func TestPFoot(t *testing.T) {
	p := Default()
	g := glyph(t, "p")
	_, y0, _, _ := bbox(g.Path)
	if math.Abs(y0+p.Descender) > 0.5 {
		t.Errorf("p: the lowest ink is at %.1f, want %v", y0, -p.Descender)
	}
	th := p.FootThick * p.Hair
	mid := -p.Descender + th/2
	rs := runs(g.Path, mid)
	if len(rs) != 1 {
		t.Fatalf("p: %d runs in the middle of the serif, want 1", len(rs))
	}
	e := p.SideStemLeft
	left := italicX(p, e, mid) - rs[0][0]
	right := rs[0][1] - italicX(p, e+p.Stem, mid)
	if f := left / right; f < 0.85 || f > 0.99 {
		t.Errorf("p: the left jut over the right is %.2f (%.0f, %.0f), want 0.85 to 0.99", f, left, right)
	}
	// Just below the top of the serif the run spans the serif; just above it
	// only the stem.
	below, above := runs(g.Path, -p.Descender+th-0.5), runs(g.Path, -p.Descender+th+0.5)
	if len(below) != 1 || math.Abs((below[0][1]-below[0][0])-(rs[0][1]-rs[0][0])) > 1 {
		t.Errorf("p: the serif is not level at its top: %v against %v", below, rs)
	}
	if len(above) != 1 || above[0][1]-above[0][0] > p.Stem+1 {
		t.Errorf("p: the stem above the serif is %v, want the stem alone", above)
	}
}
