// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// wGlyph returns the outline and the advance of w.  The tests check four
// diagonals, thick, thin, thick and thin from the left, meeting in two vertices at
// the baseline and an apex in the middle; the thick ones 0.91 to 1.00 stem and the
// thin ones the hairline; the same angle for all four, less than that of v; a flat
// apex at 0.87 to 1.00 of the x-height; head serifs on the two outer diagonals; the
// half width of w 0.693 to 0.762 that of v; an advance of 708.

func wGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, adv, ok := outlineOf("w", p)
	if !ok {
		t.Fatal("glyph w not found")
	}
	if len(path) != 1 {
		t.Fatalf("w has %d contours, want 1", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the contour of w is not counter-clockwise")
	}
	return path, adv
}

func TestWAdvanceAndInk(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := wGlyph(t, p)
	tight(t, "advance", adv, 728)
	tight(t, "advance from side spaces and body", adv, p.DiagSideLeft+p.WInk+p.DiagSideRight)
	if r := adv / design.NAdvance(p); r < 1.182 || r > 1.444 {
		t.Errorf("w is %.3f of n wide, want 1.182 to 1.444", r)
	}
	v, _ := vGlyph(t, p)
	_, _, vx1, _ := v.BBox()
	vx0, _, _, _ := v.BBox()
	x0, y0, x1, y1 := path.BBox()
	tight(t, "left space", x0, p.DiagSideLeft)
	tight(t, "right space", adv-x1, p.DiagSideRight)
	tight(t, "ink width", x1-x0, 700)
	if r := (x1 - x0) / p.XHeight; r < 1.50 || r > 1.71 {
		t.Errorf("the ink is %.3f x-heights wide, want 1.50 to 1.71", r)
	}
	// Half the width of w against the width of v.
	if r := (x1 - x0) / 2 / (vx1 - vx0); r < 0.693 || r > 0.762 {
		t.Errorf("half the width of w is %.3f of that of v, want 0.693 to 0.762", r)
	}
	tight(t, "bottom", y0, -p.DiagOvershoot)
	tight(t, "top", y1, p.XHeight)
}

func TestWDiagonals(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := wGlyph(t, p)
	rs := readDiagonals(t, path, 150, 290, 4)
	want := []float64{design.WThick(p), p.Hair, design.WThick(p), p.Hair}
	for i, r := range rs {
		tight(t, "thickness of a diagonal", r.thick(), want[i])
		tight(t, "angle of a diagonal", r.angle(), design.WAngle(p))
		if a := r.angle(); a < 15.6 || a > 24.8 {
			t.Errorf("angle %.2f degrees, want 15.6 to 24.8", a)
		}
	}
	tight(t, "thick diagonals", want[0], 76)
	tight(t, "thin diagonals", want[1], 36)
	for _, i := range []int{0, 2} {
		if r := rs[i].thick() / p.Stem; r < 0.91 || r > 1.0 {
			t.Errorf("the thick diagonal is %.3f stems, want 0.91 to 1.00", r)
		}
	}
	// Parallel in pairs: the thick ones lean the same way and so do the thin ones.
	tight(t, "the two thick diagonals are parallel", rs[0].lean, rs[2].lean)
	tight(t, "the two thin diagonals are parallel", rs[1].lean, rs[3].lean)
	tight(t, "thick and thin lean by the same amount", rs[0].lean, -rs[1].lean)
	if rs[0].lean >= 0 {
		t.Error("the first diagonal does not lean to the right going down")
	}
	// The angle of w is less than that of v.
	if design.WAngle(p) >= design.VAngle(p) {
		t.Errorf("the angle of w, %.1f, is not below that of v, %.1f", design.WAngle(p), design.VAngle(p))
	}
	// Straight edges.
	for _, span := range [][2]float64{{140, 300}, {-8, 120}} {
		for _, e := range edgeSamples2(path, span[0], span[1], 12) {
			if e > 1e-6 {
				t.Errorf("an edge of w deviates from a straight line by %v between %v and %v", e, span[0], span[1])
			}
		}
	}
}

// The structure in height: two runs below the crotches, four between the crotches
// and the notch under the apex, three above it.
func TestWRunsAndCrotches(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := wGlyph(t, p)
	count := func(y float64) int { return len(path.CrossingsX(y)) / 2 }
	for _, c := range []struct {
		y    float64
		runs int
	}{{60, 2}, {250, 4}, {400, 3}} {
		if got := count(c.y); got != c.runs {
			t.Errorf("%d runs at height %v, want %d", got, c.y, c.runs)
		}
	}
	first := func(from, to, step float64, runs int) float64 {
		for y := from; (step > 0 && y < to) || (step < 0 && y > to); y += step {
			if count(y) == runs {
				return y
			}
		}
		return math.NaN()
	}
	if got := first(5, 300, 0.02, 4); math.Abs(got-design.WCrotch(p)) > 0.1 {
		t.Errorf("the diagonals part at %.2f, want the crotch at %.2f", got, design.WCrotch(p))
	}
	if got := first(300, 440, 0.02, 3); math.Abs(got-design.WApexNotch(p)) > 0.1 {
		t.Errorf("the middle diagonals join at %.2f, want the notch at %.2f", got, design.WApexNotch(p))
	}
	// Both crotches are at the height of those of v, up to the thicker diagonal.
	if c := design.WCrotch(p); c < 40 || c > 0.5*p.XHeight {
		t.Errorf("the crotch is at %.1f", c)
	}
}

func TestWVertices(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := wGlyph(t, p)
	_, y0, _, _ := path.BBox()
	xs := path.CrossingsX(y0 + 0.01)
	if len(xs) != 4 {
		t.Fatalf("crossings above the flats: %v, want two runs", xs)
	}
	lean := design.WLean(p)
	tight(t, "width of the first flat", xs[1]-xs[0], p.DiagVertexFlat+2*lean*0.01)
	tight(t, "width of the second flat", xs[3]-xs[2], p.DiagVertexFlat+2*lean*0.01)
	tight(t, "the vertices stand a period apart", (xs[2]+xs[3])/2-(xs[0]+xs[1])/2, design.WPeriod(p))
	// The vertices of w are those of v: the same flat, depth and overshoot.
	if p.DiagVertexFlat != 24 || p.DiagOvershoot != 12 {
		t.Errorf("flat %v, overshoot %v, want 24 and 12", p.DiagVertexFlat, p.DiagOvershoot)
	}
	// 3 and 10 above the lowest point the vertex has a width of 10 to 74 and
	// 20 to 80.
	for _, c := range []struct{ up, lo, hi float64 }{{3, 10, 74}, {10, 20, 80}} {
		xs := path.CrossingsX(y0 + c.up)
		if w := xs[1] - xs[0]; w < c.lo || w > c.hi {
			t.Errorf("the first vertex is %.1f wide %v above its lowest point, want %v to %v", w, c.up, c.lo, c.hi)
		}
	}
}

func TestWApex(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := wGlyph(t, p)
	// The apex: the highest ink in the middle between the outer serifs, at the
	// given height, cut flat.
	mid := p.DiagSideLeft + p.WInk/2
	var top float64
	var topX float64
	for x := mid - 60; x <= mid+60; x += 0.05 {
		if ys := path.CrossingsY(x); len(ys) > 0 && ys[len(ys)-1] > top {
			top, topX = ys[len(ys)-1], x
		}
	}
	_ = topX
	tight(t, "height of the apex", top, p.WApexHeight)
	if r := top / p.XHeight; r < 0.87 || r > 1.0 {
		t.Errorf("the apex is at %.3f x-heights, want 0.87 to 1.00", r)
	}
	// A flat of WApexFlat; 3 and 15 units below the top the width is 25 to 73 and
	// 34 to 80.
	lean := design.WLean(p)
	for _, c := range []struct{ down, lo, hi float64 }{{0.01, 28, 32}, {3, 25, 73}, {15, 34, 90}} {
		xs := path.CrossingsX(top - c.down)
		var run []float64
		for i := 0; i+1 < len(xs); i += 2 {
			if xs[i] > mid-120 && xs[i+1] < mid+120 {
				run = xs[i : i+2]
			}
		}
		if len(run) != 2 {
			t.Fatalf("no apex run at %v below the top: %v", c.down, xs)
		}
		w := run[1] - run[0]
		tight(t, "width of the apex", w, p.WApexFlat+2*lean*c.down)
		if w < c.lo || w > c.hi {
			t.Errorf("the apex is %.1f wide %v below the top, want %v to %v", w, c.down, c.lo, c.hi)
		}
	}
	// No serif at the apex: just below the tips of the outer serifs the runs of the
	// middle are the diagonals.
	xs := path.CrossingsX(p.XHeight - p.SerifThick/2)
	if len(xs) != 6 {
		t.Fatalf("crossings just below the tips: %v, want three runs", xs)
	}
	if w := xs[3] - xs[2]; w > 80 {
		t.Errorf("the apex is %.1f wide at the height of the serifs: it has a serif", w)
	}
}

// The inner serif of the first diagonal keeps its distance from the diagonal next to
// it, and that of the last one from the one before.
func TestWSerifsKeepTheirDistance(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := wGlyph(t, p)
	xs := path.CrossingsX(p.XHeight - p.SerifThick/2)
	if len(xs) != 6 {
		t.Fatalf("crossings at the level of the serifs: %v, want three runs", xs)
	}
	for _, gap := range []float64{xs[2] - xs[1], xs[4] - xs[3]} {
		if gap < 3*design.ClearanceFloor(p) {
			t.Errorf("an inner serif of w ends %.1f from the diagonal next to it, want at least %v", gap, 3*design.ClearanceFloor(p))
		}
	}
}

// TestWCountersAgainstV holds the counters of w, the white between the first and the
// second diagonal and between the third and the fourth, at 0.85 of the counter of v
// at a level below the serifs, and the counters of the two halves equal.
func TestWCountersAgainstV(t *testing.T) {
	t.Parallel()
	for _, p := range []design.Params{design.Regular(), design.Bold()} {
		wp, _ := wGlyph(t, p)
		vp, _ := vGlyph(t, p)
		y := 0.75 * p.XHeight
		vx := vp.CrossingsX(y)
		wx := wp.CrossingsX(y)
		if len(vx) != 4 || len(wx) != 6 {
			t.Fatalf("the scan at %.0f has %d edges in v and %d in w, want 4 and 6", y, len(vx), len(wx))
		}
		vc := vx[2] - vx[1]
		c1, c2 := wx[2]-wx[1], wx[4]-wx[3]
		if c1 < 0.85*vc || c2 < 0.85*vc {
			t.Errorf("the counters of w are %.1f and %.1f, want at least 0.85 of %.1f in v", c1, c2, vc)
		}
		if math.Abs(c1-c2) > 1 {
			t.Errorf("the two counters of w differ: %.1f and %.1f", c1, c2)
		}
	}
}
