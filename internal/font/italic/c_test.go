// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// roundStress returns the lean of the line through the thinnest places of a
// round at its top and its bottom, in degrees forward from the vertical in
// the italic, measured along rays from the point c every degree within the
// directions top and bottom, and the thinnest place at the top.
func roundStress(path outline.Path, c Point, top, bottom [2]float64) (float64, Point) {
	_, qt := thinnest(path, c, top[0], top[1])
	_, qb := thinnest(path, c, bottom[0], bottom[1])
	return math.Atan2(qt.X-qb.X, qt.Y-qb.Y) * 180 / math.Pi, qt
}

// italicPoint returns the point of the italic of the upright point q.
func italicPoint(p Params, q Point) Point {
	return pt(q.X+p.slant()*q.Y-p.shift(), q.Y)
}

// inkRange returns the least and the greatest abscissa of the ink of the
// path between the heights y0 and y1, every unit.
func inkRange(path outline.Path, y0, y1 float64) (float64, float64) {
	lo, hi := math.Inf(1), math.Inf(-1)
	for y := y0 + 0.5; y < y1; y++ {
		for _, r := range runs(path, y) {
			lo, hi = math.Min(lo, r[0]), math.Max(hi, r[1])
		}
	}
	return lo, hi
}

// TestCStress asks that the round of c has the stress of the pen of o: the
// thickest place of the left side lies below its left extreme and the
// thinnest place of the bottom right of its lowest point, and the round is
// 28 to 34 thick at its thinnest and 80 to 90 at its thickest.  The top of c
// runs into the stalk of the ball, which is its thinnest place there.
func TestCStress(t *testing.T) {
	p := Default()
	g := glyph(t, "c")
	ctr := italicPoint(p, pt(p.SideRoundLeft+p.CBowlWidth/2, p.XHeight/2))
	thick, thickAt := 0.0, Point{}
	for a := 150.0; a <= 250; a++ {
		if w, q := rayThickness(g.Path, ctr, a); w > thick {
			thick, thickAt = w, q
		}
	}
	// The left extreme of the ink.
	x0, y0, _, _ := bbox(g.Path)
	var leftY float64
	for y := y0; y < p.XHeight; y++ {
		if r := runs(g.Path, y+0.5); len(r) > 0 && r[0][0] < x0+0.5 {
			leftY = y
		}
	}
	if thickAt.Y >= leftY {
		t.Errorf("the thickest place of the left side of c is at the height %.0f, not below the left extreme at %.0f", thickAt.Y, leftY)
	}
	_, bottomAt := thinnest(g.Path, ctr, 240, 300)
	lowX, _ := inkRange(g.Path, y0, y0+1)
	if bottomAt.X <= lowX {
		t.Errorf("the thinnest place of the bottom of c is at %.0f, not right of the lowest point at %.0f", bottomAt.X, lowX)
	}
	least, most := math.Inf(1), 0.0
	for a := 120.0; a <= 280; a++ {
		w, _ := rayThickness(g.Path, ctr, a)
		least, most = math.Min(least, w), math.Max(most, w)
	}
	if least < 28 || least > 34 || most < 80 || most > 90 {
		t.Errorf("the round of c is %.1f to %.1f thick, want its thinnest 28 to 34 and its thickest 80 to 90", least, most)
	}
}

// TestCTerminals asks of the two ends of c: the ball hangs from the top with
// its lowest ink 1.55 to 1.9 ball diameters below the top of the letter and its rightmost
// ink, upright, 10 to 35 inside that of the lower stroke; the wedge between
// the ball and the stalk stays open (at least 8 high on a vertical line);
// and the lower stroke ends as an exit, its highest ink 90 to 120 above the
// baseline, its bottom overshooting the baseline as o does.
func TestCTerminals(t *testing.T) {
	p := Default()
	g := glyph(t, "c")
	up := unsheared(p, g.Path)
	top := p.XHeight + p.Overshoot
	mid := p.XHeight / 2
	// The ball: the ink right of the left side above the middle.
	ballLow := math.Inf(1)
	for y := mid; y < top; y++ {
		if r := runs(up, y+0.5); len(r) == 2 {
			ballLow = math.Min(ballLow, y)
		}
	}
	if d := (top - ballLow) / p.Ball; d < 1.55 || d > 1.9 {
		t.Errorf("the lowest ink of the ball of c lies %.2f ball diameters below the top, want 1.55 to 1.9", d)
	}
	_, ballRight := inkRange(up, mid, top)
	_, tipRight := inkRange(up, 0, mid)
	if d := tipRight - ballRight; d < 10 || d > 35 {
		t.Errorf("the ball of c stays %.0f inside the end of the lower stroke, want 10 to 35", d)
	}
	_, y0, _, _ := bbox(g.Path)
	if math.Abs(-y0-p.Overshoot) > 1 {
		t.Errorf("the bottom of c lies %.1f below the baseline, want %v", -y0, p.Overshoot)
	}
	// The highest ink of the exit: of the ink right of the middle of the
	// bowl, below the middle of the x-height.
	cx := p.SideRoundLeft + p.CBowlWidth/2
	high := 0.0
	for y := 0.0; y < mid; y++ {
		for _, r := range runs(up, y+0.5) {
			if r[0] > cx {
				high = y + 1
			}
		}
	}
	if high < 90 || high > 120 {
		t.Errorf("the end of the lower stroke of c reaches %.0f, want 90 to 120", high)
	}
	best := 0.0
	for x := cx; x < ballRight; x++ {
		var ys []float64
		for _, s := range crossings(up, Point{X: x, Y: 0}, Point{Y: 1}) {
			if s > mid {
				ys = append(ys, s)
			}
		}
		if len(ys) == 4 && ys[2]-ys[1] > best {
			best = ys[2] - ys[1]
		}
	}
	if best < 8 {
		t.Errorf("the wedge of c is at most %.1f high, want at least 8", best)
	}
}

// TestCOpen asks that the aperture of c, the white between the ball and the
// end of the lower stroke, is at least 120 high.
func TestCOpen(t *testing.T) {
	p := Default()
	up := unsheared(p, glyph(t, "c").Path)
	_, ballRight := inkRange(up, p.XHeight/2, p.XHeight)
	x := ballRight - 15
	ss := crossings(up, Point{X: x, Y: -50}, Point{Y: 1})
	if len(ss) < 4 {
		t.Fatalf("the vertical line at %.0f crosses c %d times, want 4", x, len(ss))
	}
	if gap := ss[2] - ss[1]; gap < 120 {
		t.Errorf("the aperture of c is %.0f high at %.0f, want at least 120", gap, x)
	}
}
