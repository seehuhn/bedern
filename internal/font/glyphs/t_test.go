// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of t check the stem of n (81 at 99), a top at 626 which is cut
// obliquely with a flat of 18, a crossbar of the hairline thickness 36 at the
// x-height with the extents 88 and 140, and a foot in which the stem curves to the
// right and ends in a tail whose tip is 110 high and reaches 172 beyond the stem.

// tGlyph returns the outline and the advance of t after checking its contour.
func tGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, advance, ok := outlineOf("t", p)
	if !ok {
		t.Fatal("glyph t not found")
	}
	if len(path) != 1 {
		t.Fatalf("t has %d contours, want one: the bar and the stem are one outline", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the contour of t is not counter-clockwise")
	}
	return path, advance
}

// TestTAdvanceAndStem checks the advance, the ink extents and the stem of t.
func TestTAdvanceAndStem(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, advance := tGlyph(t, p)
	if advance != design.TAdvance(p) {
		t.Errorf("advance %v, want %v", advance, design.TAdvance(p))
	}
	if r := advance / design.NAdvance(p); r < 0.547 || r > 0.700 {
		t.Errorf("the advance is %.3f of that of n, outside the range 0.547 to 0.700", r)
	}
	x0, y0, x1, y1 := path.BBox()
	near(t, "top", y1, p.XHeight+p.TRise)
	near(t, "lowest ink", y0, -p.OvershootLower)
	near(t, "left ink", x0, design.TInkLeft(p))
	near(t, "reach of the tail beyond the stem", x1-(p.SideStemLeft+p.Stem), p.TReach)
	near(t, "space right of the tail", advance-x1, p.TSideRight)
	xs := path.CrossingsX(300)
	if len(xs) != 2 {
		t.Fatalf("crossings at height 300: %v", xs)
	}
	near(t, "left edge of the stem", xs[0], p.SideStemLeft)
	near(t, "stem width", xs[1]-xs[0], p.Stem)
	// The stem is that of n above the foot and below the bar.
	n, _, _ := outlineOf("n", p)
	for y := 150.0; y < p.ArchJoin-5; y += 7 {
		ts, ns := path.CrossingsX(y), n.CrossingsX(y)
		if len(ts) != 2 || math.Abs(ts[0]-ns[0]) > 0.01 || math.Abs(ts[1]-ns[1]) > 0.01 {
			t.Errorf("at height %.0f the stem of t has the crossings %v, the left stem of n %v", y, ts, ns)
		}
	}
}

// TestTTop checks the top of the stem: a cut which rises to the right at a slope
// of 48 to 60 degrees, ending in a short horizontal flat at the top right corner.
func TestTTop(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := tGlyph(t, p)
	top := design.TTop(p)
	xl, xr := float64(p.SideStemLeft), p.SideStemLeft+p.Stem
	// The top of the stem is at the height 626, 178 above the x-height.
	near(t, "rise over the x-height", top-p.XHeight, 178)
	if r := (top - p.XHeight) / (p.Ascender - p.XHeight); r < 0.36 || r > 0.83 {
		t.Errorf("the top rises %.2f of the way to the ascender, want 0.36 to 0.83", r)
	}
	_, _, _, y1 := path.BBox()
	near(t, "top of the stem", y1, top)

	// The flat is horizontal and TObliqueFlat long: the ink reaches the top
	// along it and falls below the top just left of it.
	for _, x := range []float64{xr - 0.5, xr - p.TObliqueFlat/2, xr - p.TObliqueFlat + 0.5} {
		ys := path.CrossingsY(x)
		near(t, "height of the flat", ys[len(ys)-1], top)
	}
	ys := path.CrossingsY(xr - p.TObliqueFlat - 1)
	if got := ys[len(ys)-1]; got > top-0.5 {
		t.Errorf("the top is still at %.1f, 1 unit left of the flat", got)
	}

	// The cut is straight from the end of the flat down to the left-hand edge,
	// TObliqueDrop below the top.
	run := p.Stem - p.TObliqueFlat
	slope := math.Atan2(p.TObliqueDrop, run) * 180 / math.Pi
	if slope < 48 || slope > 60 {
		t.Errorf("slope of the cut %.1f degrees, want 48 to 60", slope)
	}
	for _, f := range []float64{0, 0.25, 0.5, 0.75} {
		ys := path.CrossingsY(xl + f*run)
		near(t, "height of the cut", ys[len(ys)-1], top-p.TObliqueDrop*(1-f))
	}
	// Above the bar the left-hand edge of the stem is the first ink: there is no spur.
	for y := p.XHeight + 5; y < top-p.TObliqueDrop-1; y += 9 {
		if xs := path.CrossingsX(y); xs[0] != xl {
			t.Errorf("at height %.0f the left ink is at %.1f, want the stem edge %.1f", y, xs[0], xl)
		}
	}
	// The right-hand edge is straight up to the flat.
	for y := p.XHeight + 1; y < top-1; y += 9 {
		xs := path.CrossingsX(y)
		near(t, "right edge of the stem", xs[len(xs)-1], xr)
	}
}

// TestTTopHasNoSharpCorner checks that no corner of the top of t is sharper than
// minTopAngle degrees.  Every vertex above the crossbar is a corner of straight
// edges; its interior angle is the angle between the two edges at it.
func TestTTopHasNoSharpCorner(t *testing.T) {
	t.Parallel()
	const minTopAngle = 85.0
	p := design.Regular()
	path, _ := tGlyph(t, p)
	c := path[0]
	pts := []outline.Point{c.Start}
	for _, s := range c.Segs {
		pts = append(pts, s.P3)
	}
	if last := pts[len(pts)-1]; last == c.Start {
		pts = pts[:len(pts)-1]
	}
	n := len(pts)
	checked := 0
	for i, q := range pts {
		if q.Y < p.XHeight+p.Hair {
			continue
		}
		a, b := pts[(i+n-1)%n], pts[(i+1)%n]
		v1 := outline.Point{X: a.X - q.X, Y: a.Y - q.Y}
		v2 := outline.Point{X: b.X - q.X, Y: b.Y - q.Y}
		cos := (v1.X*v2.X + v1.Y*v2.Y) / (math.Hypot(v1.X, v1.Y) * math.Hypot(v2.X, v2.Y))
		angle := math.Acos(cos) * 180 / math.Pi
		checked++
		if angle < minTopAngle {
			t.Errorf("the corner at (%.0f, %.0f) is %.1f degrees, want at least %v", q.X, q.Y, angle, minTopAngle)
		}
	}
	if checked < 3 {
		t.Errorf("checked %d corners at the top, want the three of the top", checked)
	}
}

// TestTCrossbar checks the bar: a plain strip, the hairline thick throughout like
// that of f, with its top at the x-height and the extents 88 and 140.
func TestTCrossbar(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := tGlyph(t, p)
	xl, xr := float64(p.SideStemLeft), p.SideStemLeft+p.Stem
	scan := func(x float64) (bottom, topY float64) {
		var ys []float64
		for _, y := range path.CrossingsY(x) {
			if y > 150 && y < p.XHeight+30 {
				ys = append(ys, y)
			}
		}
		if len(ys) != 2 {
			t.Fatalf("crossings of the bar at x = %.1f: %v", x, ys)
		}
		return ys[0], ys[1]
	}
	// At the stem the bar is Hair thick, 1 unit in from the end just as thick, and
	// its top is level at the x-height.
	for _, c := range []struct {
		name string
		x    float64
		want float64
	}{
		{"left, at the stem", xl - 1, p.Hair}, {"left, at the end", xl - p.TBarLeft + 1, p.Hair},
		{"right, at the stem", xr + 1, p.Hair}, {"right, at the end", xr + p.TBarRight - 1, p.Hair},
	} {
		b, tp := scan(c.x)
		near(t, c.name+": top", tp, p.XHeight)
		if math.Abs((tp-b)-c.want) > 1 {
			t.Errorf("%s: thickness %.1f, want %.1f", c.name, tp-b, c.want)
		}
	}
	xs := path.CrossingsX(p.XHeight - p.Hair/2)
	near(t, "extent to the left", xl-xs[0], p.TBarLeft)
	near(t, "extent to the right", xs[len(xs)-1]-xr, p.TBarRight)
	// The bar of t is that of f except for the extents: the same top and the same
	// thickness at the stem.
	fp, _ := fGlyph(t, p)
	fy := fp.CrossingsY(design.FStemLeft(p) - 1)
	b, tp := scan(xl - 1)
	if math.Abs((tp-b)-(fy[len(fy)-1]-fy[len(fy)-2])) > 0.5 {
		t.Errorf("the bars of f and t are %.1f and %.1f thick at the stem", fy[len(fy)-1]-fy[len(fy)-2], tp-b)
	}
}

// TestTFoot checks the foot: the stem curves to the right, its left edge is the
// outer edge of the curve with the low point OvershootLower below the baseline,
// and the tail rises to its tip, which is TReach beyond the stem.
func TestTFoot(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := tGlyph(t, p)
	xr := float64(p.SideStemLeft + p.Stem)
	// The low point of the outer edge, and the thickness there.
	var low outline.Point
	for x := 0.0; x < 400; x += 0.25 {
		if ys := path.CrossingsY(x); len(ys) > 0 && ys[0] < low.Y {
			low = outline.Point{X: x, Y: ys[0]}
		}
	}
	near(t, "low point", low.Y, -p.OvershootLower)
	ys := path.CrossingsY(low.X)
	near(t, "thickness of the foot at the low point", ys[1]-ys[0], p.ThinLower)
	if low.X < xr-20 || low.X > xr+60 {
		t.Errorf("the low point of the foot is at %.1f, the stem ends at %.1f", low.X, xr)
	}
	// There are no foot serifs: at the baseline the stem has no ink to either side
	// but the curve.
	if xs := path.CrossingsX(1); len(xs) != 2 || xs[0] < p.SideStemLeft {
		t.Errorf("crossings at the baseline: %v, want ink from the stem edge %v only", xs, p.SideStemLeft)
	}
	// The tail: the rightmost ink.
	x0, _, x1, _ := path.BBox()
	_ = x0
	near(t, "reach of the tail", x1-xr, p.TReach)
	// Halfway between the stem and the tip the foot is at most 62 thick and
	// thicker than the tip.
	mid := path.CrossingsY(xr + p.TReach/2)
	if th := mid[1] - mid[0]; th < p.TTipThick || th > 62 {
		t.Errorf("the foot halfway to the tip is %.1f thick, want %v to 62", th, p.TTipThick)
	} else {
		t.Logf("the foot is %.1f thick halfway between the stem and the tip", th)
	}
}

// TestTFootEdges checks that the left edge of the stem is the outer edge of the
// curve: it leaves the straight edge vertically, and the low point has horizontal
// handles.
func TestTFootEdges(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := tGlyph(t, p)
	ct := path[0]
	xl := float64(p.SideStemLeft)
	// The first curve after the left edge starts on it.
	var found bool
	for i, s := range ct.Segs {
		if !s.IsCurve {
			continue
		}
		a := ct.Start
		if i > 0 {
			a = ct.Segs[i-1].P3
		}
		if a.X == xl && a.Y < p.XHeight-p.Hair && s.P1.X == xl && s.P1.Y < a.Y {
			found = true
			// This curve ends at the low point, whose handle is horizontal.
			if s.P3.Y != -p.OvershootLower || s.P2.Y != s.P3.Y {
				t.Errorf("the outer curve of the foot ends at %v with the handle %v, want a horizontal tangent at the low point", s.P3, s.P2)
			}
		}
	}
	if !found {
		t.Error("no curve leaves the left edge of the stem downwards")
	}
}

// TestTTBarsKeepApart checks that the right end of the bar of one t and the left
// end of the bar of the next are at least minBarGap apart.
func TestTTBarsKeepApart(t *testing.T) {
	t.Parallel()
	const minBarGap = 40.0
	for _, np := range allParams() {
		path, advance := cachedOutline(np, "t")
		x0, _, _, _ := path.BBox()
		p := np.params
		gap := advance + x0 - (p.SideStemLeft + p.Stem + p.TBarRight)
		if gap < minBarGap {
			t.Errorf("parameter set %s: the bars of tt are %.1f apart, want at least %v", np.name, gap, minBarGap)
		}
	}
}
