// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// xGlyph returns the outline and the advance of x.  The tests check a thick
// diagonal from the upper left to the lower right and a thin one from the lower left
// to the upper right, which cross slightly above the middle; the thin stroke in two
// halves, the lower one 5 to the left of the line of the upper one at the crossing;
// four serifs, 40 outside and 70 inside, 33 thick at the tip; the ink slightly
// wider at the bottom than at the top (472 and 450); an advance of 496.

func xGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, adv, ok := outlineOf("x", p)
	if !ok {
		t.Fatal("glyph x not found")
	}
	if len(path) != 1 {
		t.Fatalf("x has %d contours, want 1", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the contour of x is not counter-clockwise")
	}
	return path, adv
}

// xReadings returns the diagonals of x read at heights above and below the
// crossing: above, the thick diagonal and the thin one; below, the thin one and
// the thick one.
func xReadings(t *testing.T, path outline.Path) (upper, lower []diagReading) {
	t.Helper()
	return readDiagonals(t, path, 300, 328, 2), readDiagonals(t, path, 120, 165, 2)
}

// centreAt returns the position of the centre line of the reading, taken at y0, at
// the height y.
func (r diagReading) centreAt(y0, y float64) float64 {
	return (r.left+r.right)/2 + r.lean*(y-y0)
}

func TestXAdvanceAndInk(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := xGlyph(t, p)
	tight(t, "advance", adv, 500)
	tight(t, "advance from side spaces and body", adv, p.DiagSideLeft+p.XInk+p.DiagSideRight)
	if r := adv / design.NAdvance(p); r < 0.785 || r > 1.0 {
		t.Errorf("x is %.3f of n wide, want 0.785 to 1.000", r)
	}
	x0, y0, x1, y1 := path.BBox()
	tight(t, "left space", x0, p.DiagSideLeft)
	tight(t, "right space", adv-x1, p.DiagSideRight)
	tight(t, "ink width", x1-x0, 472)
	tight(t, "top is the x-height", y1, p.XHeight)
	tight(t, "bottom is the baseline", y0, 0)
	if e, ok := lookup("x"); !ok || e.left != sideDiag || e.right != sideDiag {
		t.Errorf("the side kinds of x are %v and %v, want the diagonal kinds", e.left, e.right)
	}
}

// extent returns the greatest extent of the ink at the heights from y0 to y1.
func extent(path outline.Path, y0, y1 float64) (lo, hi float64) {
	lo, hi = math.Inf(1), math.Inf(-1)
	for y := y0; y <= y1; y += 0.25 {
		xs := path.CrossingsX(y)
		if len(xs) == 0 {
			continue
		}
		lo, hi = math.Min(lo, xs[0]), math.Max(hi, xs[len(xs)-1])
	}
	return lo, hi
}

func TestXWidthsAtTopAndBottom(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := xGlyph(t, p)
	tl, th := extent(path, 0.92*p.XHeight, p.XHeight-0.1)
	bl, bh := extent(path, 0.1, 0.08*p.XHeight)
	top, bottom := th-tl, bh-bl
	if math.Abs(top-design.XInkTop(p)) > 0.5 {
		t.Errorf("width at the top %.2f, want %.2f", top, design.XInkTop(p))
	}
	if math.Abs(bottom-design.XInkBottom(p)) > 0.5 {
		t.Errorf("width at the bottom %.2f, want %.2f", bottom, design.XInkBottom(p))
	}
	if math.Abs(bottom-p.XInk) > 1 {
		t.Errorf("width at the bottom %.2f, want XInk = %v within a unit", bottom, p.XInk)
	}
	if r := top / bottom; r < 0.885 || r > 0.974 {
		t.Errorf("the top is %.3f of the bottom, want 0.885 to 0.974", r)
	}
	if top >= bottom {
		t.Errorf("x is not wider at the bottom (%.1f) than at the top (%.1f)", bottom, top)
	}
	if r := bottom / p.XHeight; r < 1.015 || r > 1.221 {
		t.Errorf("the ink is %.3f x-heights wide, want 1.015 to 1.221", r)
	}
}

func TestXDiagonals(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := xGlyph(t, p)
	up, lo := xReadings(t, path)
	thickU, thinU := up[0], up[1]
	thinL, thickL := lo[0], lo[1]
	for _, r := range []diagReading{thickU, thickL} {
		tight(t, "thick diagonal", r.thick(), p.XThick)
		tight(t, "angle of the thick diagonal", r.angle(), design.XAngleThick(p))
		if r.lean >= 0 {
			t.Errorf("the thick diagonal leans %.3f, want it to the left going up (from the upper left to the lower right)", r.lean)
		}
	}
	for _, r := range []diagReading{thinU, thinL} {
		tight(t, "thin diagonal", r.thick(), p.Hair)
		tight(t, "angle of the thin diagonal", r.angle(), design.XAngleThin(p))
		if r.lean <= 0 {
			t.Errorf("the thin diagonal leans %.3f, want it to the right going up", r.lean)
		}
	}
	tight(t, "thick diagonal is 74", thickU.thick(), 74)
	tight(t, "thin diagonal is the hairline 36", thinU.thick(), 36)
	if r := thinU.thick() / thickU.thick(); r < 0.43 || r > 0.99 {
		t.Errorf("thin over thick %.3f, want 0.43 to 0.99", r)
	}
	if a := thickU.angle(); a < 32.3 || a > 39.0 {
		t.Errorf("angle of the thick diagonal %.2f, want 32.3 to 39.0", a)
	}
	if a := thinU.angle(); a < 34.1 || a > 41.7 {
		t.Errorf("angle of the thin diagonal %.2f, want 34.1 to 41.7", a)
	}
	// The thick diagonal is one straight stroke through the crossing: its two
	// halves lie on one line.
	tight(t, "thick halves, centre line at the crossing", thickL.centreAt(120, p.XCross), thickU.centreAt(300, p.XCross))
	tight(t, "thick halves, lean", thickL.lean, thickU.lean)
	// All edges are straight lines between the serifs and the crossing.
	for _, span := range [][2]float64{{300, 328}, {120, 165}} {
		for _, e := range edgeSamples2(path, span[0], span[1], 12) {
			if e > 1e-6 {
				t.Errorf("an edge of x deviates from a straight line by %v between %v and %v", e, span[0], span[1])
			}
		}
	}
}

// The two halves of the thin stroke are parallel, the lower one stands XThinOffset
// to the right of the line of the upper one at the crossing, and the crossing lies
// at XCross.
func TestXThinHalvesAndCrossing(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := xGlyph(t, p)
	up, lo := xReadings(t, path)
	thickU, thinU := up[0], up[1]
	thinL, thickL := lo[0], lo[1]
	tight(t, "the halves of the thin stroke are parallel", thinL.lean, thinU.lean)
	off := thinL.centreAt(120, p.XCross) - thinU.centreAt(300, p.XCross)
	tight(t, "offset of the lower half at the crossing", off, p.XThinOffset)
	tight(t, "offset is -5", off, -5)
	// The centre lines of the upper halves cross at XCross; those of the lower halves
	// a little above it.
	cross := func(a, b, la, lb float64) float64 {
		// the lines are x = a + la (y - 300) and x = b + lb (y - 300) (y in units from 300)
		return 300 + (b-a)/(la-lb)
	}
	yu := cross(thickU.centreAt(300, 300), thinU.centreAt(300, 300), thickU.lean, thinU.lean)
	tight(t, "the upper halves cross at", yu, p.XCross)
	yl := 120 + (thinL.centreAt(120, 120)-thickL.centreAt(120, 120))/(thickL.lean-thinL.lean)
	want := p.XCross - p.XThinOffset/(design.XLeanThick(p)+design.XLeanThin(p))
	tight(t, "the lower halves cross at", yl, want)
	if r := (yu + yl) / 2 / (p.XHeight / 2); r < 1.00 || r > 1.11 {
		t.Errorf("the crossing is at %.3f of half the x-height, want 1.002 to 1.107", r)
	}
	tight(t, "the crossing is at 232", p.XCross, 232)
	// The crossing is where design.XCentre says.
	tight(t, "the horizontal position of the crossing", thickU.centreAt(300, p.XCross), design.XCentre(p))
}

// The four inner corners at the crossing are sharp: the outline has a node at each
// of them, with a straight segment on both sides, and it turns clockwise there.
func TestXInnerCornersAreSharp(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := xGlyph(t, p)
	up, lo := xReadings(t, path)
	// An edge as the line x = x0 + lean y.
	e := func(r diagReading, y0 float64, right bool) [2]float64 {
		x := r.left
		if right {
			x = r.right
		}
		return [2]float64{x - r.lean*y0, r.lean}
	}
	meet := func(a, b [2]float64) outline.Point {
		y := (b[0] - a[0]) / (a[1] - b[1])
		return outline.Point{X: a[0] + a[1]*y, Y: y}
	}
	corners := map[string]outline.Point{
		// The thin upper half meets the right-hand edge of the thick diagonal below
		// (right notch) and above (top notch); the lower half meets the left-hand edge.
		"top":    meet(e(up[0], 300, true), e(up[1], 300, false)),
		"right":  meet(e(up[0], 300, true), e(up[1], 300, true)),
		"bottom": meet(e(lo[1], 120, false), e(lo[0], 120, true)),
		"left":   meet(e(lo[1], 120, false), e(lo[0], 120, false)),
	}
	ct := path[0]
	nodes := contourNodes(ct)
	n := len(ct.Segs)
	for name, q := range corners {
		i := nodeNear(t, ct, q)
		prev, next := ct.Segs[(i+n-1)%n], ct.Segs[i%n]
		if prev.IsCurve || next.IsCurve {
			t.Errorf("the %s corner of the crossing has a curve next to it", name)
		}
		a, b, c := nodes[(i+n-1)%n], nodes[i], nodes[i+1]
		if cross := (b.X-a.X)*(c.Y-b.Y) - (b.Y-a.Y)*(c.X-b.X); cross >= 0 {
			t.Errorf("the %s corner of the crossing is not an inner corner (turn %.1f)", name, cross)
		}
	}
	// Exactly four inner corners between two straight segments lie in the middle
	// of the letter.
	inner := 0
	for i := 0; i < n; i++ {
		a, b, c := nodes[(i+n-1)%n], nodes[i], nodes[i+1]
		if ct.Segs[(i+n-1)%n].IsCurve || ct.Segs[i%n].IsCurve {
			continue
		}
		if cross := (b.X-a.X)*(c.Y-b.Y) - (b.Y-a.Y)*(c.X-b.X); cross < 0 && b.Y > 100 && b.Y < 350 {
			inner++
		}
	}
	if inner != 4 {
		t.Errorf("%d inner corners between 100 and 350, want 4", inner)
	}
}

// The feet of x are SerifThick thick at the tip, flat on the baseline, and the eight
// brackets of the serifs are the only curves.
func TestXSerifs(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := xGlyph(t, p)
	xb := path.CrossingsX(footScanY)
	if len(xb) != 4 {
		t.Fatalf("crossings just above the baseline: %v, want 4", xb)
	}
	for _, tip := range []struct {
		x, side float64
	}{{xb[0], 1}, {xb[3], -1}} {
		ys := path.CrossingsY(tip.x + tip.side*tipInset)
		if len(ys) < 2 {
			t.Fatalf("no foot at %v: %v", tip.x, ys)
		}
		tight(t, "foot serif bottom", ys[0], 0)
		if th := ys[1] - ys[0]; th < p.SerifThick-0.01 || th > p.SerifThick+1.5 {
			t.Errorf("foot serif thickness five units from the tip %.2f, want about %v", th, p.SerifThick)
		}
	}
	// Four brackets at the head and four at the feet are the only curves.
	curves := 0
	for _, s := range path[0].Segs {
		if s.IsCurve {
			curves++
		}
	}
	if curves < 8 || curves > 16 {
		t.Errorf("%d curved segments, want one or two for each of the eight brackets", curves)
	}
}

// The crossing is not a blob: the run through the crossing is the width of the thick
// diagonal plus what the thin one adds on each side, and no more.
func TestXCrossingIsNotBlobby(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := xGlyph(t, p)
	xs := path.CrossingsX(p.XCross)
	if len(xs) != 2 {
		t.Fatalf("crossings at the crossing: %v, want one run", xs)
	}
	wt := design.DiagWidth(p.XThick, design.XLeanThick(p))
	wn := design.DiagWidth(p.Hair, design.XLeanThin(p))
	if w := xs[1] - xs[0]; w < wt-0.01 || w > wt+2*wn {
		t.Errorf("the run at the crossing is %.1f wide, want between %.1f and %.1f", w, wt, wt+2*wn)
	}
}
