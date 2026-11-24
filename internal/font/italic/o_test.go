// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// rayThickness returns the thickness of the ink along the ray from c in the
// direction a (degrees), and the middle of the ink there: the distance
// between the first two crossings of the outline.
func rayThickness(path outline.Path, c Point, a float64) (float64, Point) {
	d := Point{X: math.Cos(a * math.Pi / 180), Y: math.Sin(a * math.Pi / 180)}
	var ss []float64
	for _, s := range crossings(path, c, d) {
		if s > 0 {
			ss = append(ss, s)
		}
	}
	if len(ss) < 2 {
		return math.NaN(), Point{}
	}
	m := (ss[0] + ss[1]) / 2
	return ss[1] - ss[0], Point{X: c.X + m*d.X, Y: c.Y + m*d.Y}
}

// thinnest returns the direction of the least thickness along the rays from
// c between the directions a0 and a1 (the middle of the directions within 1
// unit of the least) and the middle of the ink there.
func thinnest(path outline.Path, c Point, a0, a1 float64) (float64, Point) {
	least := math.Inf(1)
	for a := a0; a <= a1; a++ {
		if w, _ := rayThickness(path, c, a); w < least {
			least = w
		}
	}
	var lo, hi = math.Inf(1), math.Inf(-1)
	for a := a0; a <= a1; a++ {
		if w, _ := rayThickness(path, c, a); w <= least+1 {
			lo, hi = math.Min(lo, a), math.Max(hi, a)
		}
	}
	a := (lo + hi) / 2
	_, q := rayThickness(path, c, a)
	return a, q
}

// TestOStress asks that the line through the thinnest places of o at its
// top and bottom leans 7 to 11 degrees forward from the vertical, less than
// the stems (14), with the thin place at the top left of the topmost point,
// and that the thinnest places are 28 to 34 thick and the thickest 80 to 90
// (the pen).
func TestOStress(t *testing.T) {
	g := glyph(t, "o")
	x0, y0, x1, y1 := bbox(g.Path)
	c := Point{X: (x0 + x1) / 2, Y: (y0 + y1) / 2}
	_, top := thinnest(g.Path, c, 1, 179)
	_, bottom := thinnest(g.Path, c, 181, 359)
	lean := math.Atan2(top.X-bottom.X, top.Y-bottom.Y) * 180 / math.Pi
	t.Logf("stress of o %.1f degrees", lean)
	if lean < 7 || lean > 11 {
		t.Errorf("the stress of o leans %.1f degrees, want 7 to 11", lean)
	}
	// The topmost point of the outer contour.
	var topX float64
	for _, pts := range polys(g.Path) {
		for _, q := range pts {
			if q.Y == y1 {
				topX = q.X
			}
		}
	}
	if top.X >= topX {
		t.Errorf("the thinnest place at the top of o is at %.0f, not left of the topmost point at %.0f", top.X, topX)
	}
	least, most := math.Inf(1), 0.0
	for a := 0.0; a < 360; a++ {
		w, _ := rayThickness(g.Path, c, a)
		least, most = math.Min(least, w), math.Max(most, w)
	}
	if least < 28 || least > 34 {
		t.Errorf("the thinnest place of o is %.1f, want 28 to 34", least)
	}
	if most < 80 || most > 90 {
		t.Errorf("the thickest place of o is %.1f, want 80 to 90", most)
	}
}

// TestOShape asks that the outer edge of o, upright, has a squareness of 0.70
// to 0.74 (the point of the edge on the diagonal of each quarter between the
// extremes, as a fraction of the way to the corner of their box), and that the o has its sides at ShearHeight.
func TestOShape(t *testing.T) {
	p := Default()
	g := glyph(t, "o")
	up := unsheared(p, g.Path)
	x0, y0, x1, y1 := bbox(up)
	c := Point{X: (x0 + x1) / 2, Y: (y0 + y1) / 2}
	for _, q := range [][2]float64{{1, 1}, {-1, 1}, {-1, -1}, {1, -1}} {
		corner := Point{X: c.X + q[0]*(x1-x0)/2, Y: c.Y + q[1]*(y1-y0)/2}
		d := Point{X: corner.X - c.X, Y: corner.Y - c.Y}
		ss := crossings(up, c, d)
		s := ss[len(ss)-1]
		if s < 0.70 || s > 0.74 {
			t.Errorf("the squareness of o in the quarter %v is %.3f, want 0.70 to 0.74", q, s)
		}
	}
	r := runs(g.Path, p.ShearHeight+0.5)
	if l, rr := r[0][0], g.Advance-r[len(r)-1][1]; math.Abs(l-p.SideRoundLeft) > 1.5 || math.Abs(rr-p.SideRoundRight) > 1.5 {
		t.Errorf("the sides of o are %.1f and %.1f, want %v and %v", l, rr, p.SideRoundLeft, p.SideRoundRight)
	}
}
