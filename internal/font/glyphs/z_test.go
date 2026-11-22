// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sort"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// zGlyph returns the outline and the advance of z.  The tests check a top bar and
// a bottom bar of 36, joined by a thick diagonal of 79 which runs from the upper
// right to the lower left; the top at the x-height and the bottom on the baseline,
// both flat; the top narrower than the bottom (0.888 to 0.971); both bars end in a
// beak, the upper one at the left hanging 118 high and 36 thick, the lower one at
// the right rising 115 high and 40 thick; an advance of 429.

func zGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, adv, ok := outlineOf("z", p)
	if !ok {
		t.Fatal("glyph z not found")
	}
	if len(path) != 1 {
		t.Fatalf("z has %d contours, want 1", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the contour of z is not counter-clockwise")
	}
	return path, adv
}

func TestZAdvanceAndInk(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := zGlyph(t, p)
	tight(t, "advance", adv, 426)
	tight(t, "advance from side spaces and body", adv, p.ZSideLeft+p.ZInk+p.ZSideRight)
	if r := adv / design.NAdvance(p); r < 0.727 || r > 0.888 {
		t.Errorf("z is %.3f of n wide, want 0.727 to 0.888", r)
	}
	// The width of 426 gives 0.789 of n.
	if r := adv / design.NAdvance(p); math.Abs(r-0.789) > 0.01 {
		t.Errorf("z is %.3f of n wide, want 0.789 within 0.01", r)
	}
	x0, y0, x1, y1 := path.BBox()
	tight(t, "left space", x0, p.ZSideLeft)
	tight(t, "right space", adv-x1, p.ZSideRight)
	tight(t, "ink width", x1-x0, p.ZInk)
	tight(t, "top is the x-height, flat", y1, p.XHeight)
	tight(t, "bottom is the baseline, flat", y0, 0)
	if e, ok := lookup("z"); !ok || e.left == sideUnknown || e.right == sideUnknown {
		t.Error("z has no side kinds")
	}
}

// The bars are 36 thick (the hairline) over 0.2 to 0.45 and 0.55 to 0.8 of the ink
// width.
func TestZBars(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := zGlyph(t, p)
	x0, _, x1, _ := path.BBox()
	w := x1 - x0
	// The median of the thickness over the stretch.
	median := func(vs []float64) float64 {
		sort.Float64s(vs)
		return vs[len(vs)/2]
	}
	var top, bottom []float64
	for f := 0.2; f <= 0.45; f += 0.01 {
		ys := path.CrossingsY(x0 + f*w)
		top = append(top, ys[len(ys)-1]-ys[len(ys)-2])
		tight(t, "top of the bar", ys[len(ys)-1], p.XHeight)
	}
	for f := 0.55; f <= 0.8; f += 0.01 {
		ys := path.CrossingsY(x0 + f*w)
		bottom = append(bottom, ys[1]-ys[0])
		tight(t, "bottom of the bar", ys[0], 0)
	}
	tight(t, "top bar", median(top), 36)
	tight(t, "top bar is the hairline", median(top), p.Hair)
	tight(t, "bottom bar", median(bottom), 36)
	if r := p.Hair / 32; r < 1.03 || r > 1.31 {
		t.Errorf("the bars are %.2f times the thin part of o, want 1.03 to 1.31", r)
	}
}

func TestZDiagonal(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := zGlyph(t, p)
	rs := readDiagonals(t, path, 150, 300, 1)
	d := rs[0]
	tight(t, "thickness of the diagonal", d.thick(), p.ZDiag)
	tight(t, "diagonal is 79", d.thick(), 79)
	tight(t, "angle of the diagonal", d.angle(), design.ZAngle(p))
	if d.lean <= 0 {
		t.Errorf("the diagonal leans %.3f, want it to the right going up", d.lean)
	}
	if a := d.angle(); a < 30 || a > 39 {
		t.Errorf("angle %.2f degrees, want 30 to 39", a)
	}
	if r := d.thick() / p.Stem; r < 0.87 || r > 1.03 {
		t.Errorf("the diagonal is %.3f of the stem, want 0.87 to 1.03", r)
	}
	for _, e := range edgeSamples2(path, 150, 300, 12) {
		if e > 1e-6 {
			t.Errorf("an edge of the diagonal of z deviates from a straight line by %v", e)
		}
	}
}

// The widths 0.4 bar below the top and above the baseline: the top is narrower
// (0.888 to 0.971); the beaks and the rounding of the shoulders take some units
// from the ink of the nominal widths.
func TestZWidths(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := zGlyph(t, p)
	y := 0.4 * p.Hair
	tx := path.CrossingsX(p.XHeight - y)
	bx := path.CrossingsX(y)
	top := tx[len(tx)-1] - tx[0]
	bottom := bx[len(bx)-1] - bx[0]
	if top < 320 || top > 390 {
		t.Errorf("width at the top %.1f, want 320 to 390", top)
	}
	if bottom < 341 || bottom > 435 {
		t.Errorf("width at the bottom %.1f, want 341 to 435", bottom)
	}
	if r := top / bottom; r < 0.888 || r > 0.971 {
		t.Errorf("the top is %.3f of the bottom, want 0.888 to 0.971", r)
	}
	// The nominal widths: the upper beak ends ZInk - ZInkTop right of the corner of
	// the diagonal at the baseline.
	x0, _, _, _ := path.BBox()
	lt := path.CrossingsX(p.XHeight - p.ZUpperBeakHeight/2)[0]
	tight(t, "left end of the upper beak", lt-x0, p.ZInk-p.ZInkTop)
}

// beak reads one beak: its height as the vertical thickness of the ink 20 in from
// its end, the drop below the bar, its thickness halfway up, and the straightness
// of the cut.
func TestZBeaks(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := zGlyph(t, p)
	// The upper beak, at the left of the top bar.
	lt := path.CrossingsX(p.XHeight - p.ZUpperBeakHeight/2)[0]
	ys := path.CrossingsY(lt + 20)
	top, bottom := ys[len(ys)-1], ys[len(ys)-2]
	if h := top - bottom; h < 84 || h > 198 || math.Abs(h-p.ZUpperBeakHeight) > 10 {
		t.Errorf("upper beak %.1f high, want about %v (84 to 198)", h, p.ZUpperBeakHeight)
	}
	tight(t, "the end of the upper beak", bottom, p.XHeight-p.ZUpperBeakHeight)
	drop := (p.XHeight - p.Hair) - bottom
	if drop < 43 || drop > 138 || math.Abs(drop-(p.ZUpperBeakHeight-p.Hair)) > 0.01 {
		t.Errorf("upper beak drops %.1f below the bar, want %v (43 to 138)", drop, p.ZUpperBeakHeight-p.Hair)
	}
	mid := path.CrossingsX(p.XHeight - p.ZUpperBeakHeight/2)
	tight(t, "thickness of the upper beak halfway", mid[1]-mid[0], p.ZUpperBeakThick)
	if th := mid[1] - mid[0]; th < 32 || th > 68 {
		t.Errorf("upper beak %.1f thick, want 32 to 68", th)
	}
	// The cut is a vertical edge over the lower part of the beak.
	for d := 50.0; d < p.ZUpperBeakHeight-2; d += 8 {
		tight(t, "cut of the upper beak", path.CrossingsX(p.XHeight - d)[0], lt)
	}
	// The end is flat.
	flat := path.CrossingsX(bottom + 1)
	if flat[1]-flat[0] < 20 || flat[1]-flat[0] > 60 {
		t.Errorf("the end of the upper beak is %.1f wide, want 20 to 60", flat[1]-flat[0])
	}

	// The lower beak, at the right of the bottom bar.
	rx := path.CrossingsX(p.ZLowerBeakHeight / 2)
	rt := rx[len(rx)-1]
	_, _, x1, _ := path.BBox()
	tight(t, "the lower beak is the right end of the ink", rt, x1)
	ys = path.CrossingsY(rt - 20)
	lo, hi := ys[0], ys[1]
	if h := hi - lo; h < 79 || h > 223 || math.Abs(h-p.ZLowerBeakHeight) > 10 {
		t.Errorf("lower beak %.1f high, want about %v (79 to 223)", h, p.ZLowerBeakHeight)
	}
	tight(t, "the end of the lower beak", hi, p.ZLowerBeakHeight)
	rise := hi - p.Hair
	if rise < 48 || rise > 161 || math.Abs(rise-(p.ZLowerBeakHeight-p.Hair)) > 0.01 {
		t.Errorf("lower beak rises %.1f above the bar, want %v (48 to 161)", rise, p.ZLowerBeakHeight-p.Hair)
	}
	tight(t, "thickness of the lower beak halfway", rx[len(rx)-1]-rx[len(rx)-2], p.ZLowerBeakThick)
	if th := p.ZLowerBeakThick; th < 37 || th > 75 {
		t.Errorf("lower beak %.1f thick, want 37 to 75", th)
	}
	for d := 50.0; d < p.ZLowerBeakHeight-2; d += 8 {
		got := path.CrossingsX(d)
		tight(t, "cut of the lower beak", got[len(got)-1], rt)
	}
}

// The corners where the diagonal meets the bars are sharp: a node at each of the
// four, with a straight segment on both sides, the outer ones convex and the inner
// ones concave.
func TestZCornersAreSharp(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := zGlyph(t, p)
	x0, _, x1, _ := path.BBox()
	l := design.ZLean(p)
	// The diagonal's right edge passes (x1, top); its left edge passes (x0, 0).
	right := func(y float64) float64 { return x1 - l*(p.XHeight-y) }
	left := func(y float64) float64 { return x0 + l*y }
	want := []struct {
		name   string
		q      outline.Point
		convex bool
	}{
		{"top right", outline.Point{X: x1, Y: p.XHeight}, true},
		{"under the top bar", outline.Point{X: left(p.XHeight - p.Hair), Y: p.XHeight - p.Hair}, false},
		{"bottom left", outline.Point{X: x0, Y: 0}, true},
		{"on the bottom bar", outline.Point{X: right(p.Hair), Y: p.Hair}, false},
	}
	ct := path[0]
	nodes := contourNodes(ct)
	n := len(ct.Segs)
	for _, w := range want {
		i := nodeNear(t, ct, w.q)
		if ct.Segs[(i+n-1)%n].IsCurve || ct.Segs[i%n].IsCurve {
			t.Errorf("the corner %q has a curve next to it", w.name)
		}
		a, b, c := nodes[(i+n-1)%n], nodes[i], nodes[i+1]
		cross := (b.X-a.X)*(c.Y-b.Y) - (b.Y-a.Y)*(c.X-b.X)
		if (cross > 0) != w.convex {
			t.Errorf("the corner %q turns %.1f, want a %s corner", w.name, cross, map[bool]string{true: "convex", false: "concave"}[w.convex])
		}
	}
}

// The beaks are like those of s except for the outer corner, which is
// square: the cut is vertical and runs up to the top of the bar, the end is flat
// (checked above), and the inner edge leaves the end of the beak at a small angle
// from the vertical.
func TestZBeaksAreOfTheFamilyOfS(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := zGlyph(t, p)
	lt := path.CrossingsX(p.XHeight - p.ZUpperBeakHeight/2)[0]
	// The outer corner is square: 1 unit below the top the ink already starts at the cut.
	if x := path.CrossingsX(p.XHeight - 1)[0]; math.Abs(x-lt) > 0.01 {
		t.Errorf("the outer corner of the upper beak is not square: 1 below the top the ink starts at %.2f, the cut is at %.2f", x, lt)
	}
	// The inner edge leans away from the vertical by less than 20 degrees at the
	// end and the beak is not thinner at the end than halfway.
	y0 := p.XHeight - p.ZUpperBeakHeight
	a, b := path.CrossingsX(y0+2), path.CrossingsX(y0+22)
	slope := math.Atan(((b[1] - a[1]) / 20)) * 180 / math.Pi
	if slope < 0 || slope > 20 {
		t.Errorf("the inner edge leaves the end of the beak at %.1f degrees from the vertical, want 0 to 20", slope)
	}
}
