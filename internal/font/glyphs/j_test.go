// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The conditions for j: the stem, head serif and dot of i, with the left edge of the
// stem at 125 and a right space of 90 (advance 296), a stem which continues to a
// depth of 200 and curves to the left into a tail which reaches 184 beyond the stem
// and ends in the ball terminal, the centre of the ball at -106, and a left side of
// -59.

func jGlyph(t *testing.T, p design.Params) (path outline.Path, stem, dot outline.Path, advance float64) {
	t.Helper()
	path, advance, ok := outlineOf("j", p)
	if !ok {
		t.Fatal("glyph j not found")
	}
	if len(path) != 2 {
		t.Fatalf("j has %d contours, want 2 (the stem with its tail, and the dot)", len(path))
	}
	for i, ct := range path {
		if ct.SignedArea() <= 0 {
			t.Errorf("contour %d of j is not counter-clockwise", i)
		}
	}
	dot, stem = dotOf(t, path)
	return path, stem, dot, advance
}

func TestJAdvanceAndStem(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, stem, _, advance := jGlyph(t, p)
	if advance != 296 || advance != design.JAdvance(p) {
		t.Errorf("advance %v, want 296", advance)
	}
	if r := advance / design.NAdvance(p); r < 0.402 || r > 0.556 {
		t.Errorf("j is %.3f of n wide, want 0.402 to 0.556", r)
	}
	x0, y0, _, _ := path.BBox()
	near(t, "depth", y0, -p.JDepth)
	near(t, "leftmost ink", x0, design.JInkLeft(p))
	near(t, "left ink is the reach of the tail left of the stem", p.JStemLeft-x0, p.JReach)
	_, _, sx1, _ := stem.BBox()
	near(t, "right edge of the stem to the advance", advance-sx1, design.NarrowSideRight(p))
	xs := stem.CrossingsX(100)
	if len(xs) != 2 {
		t.Fatalf("crossings at height 100: %v", xs)
	}
	near(t, "left edge of the stem", xs[0], p.JStemLeft)
	near(t, "stem width", xs[1]-xs[0], p.Stem)
	// The stem is that of i above the foot: the same edges, 11 to the right.
	i, _, _ := outlineOf("i", p)
	shift := p.JStemLeft - design.NarrowStemLeft(p)
	for y := p.BracketHeight + 5; y < p.XHeight-p.HeadSerifDrop-p.SerifThick-15; y += 7 {
		js, is := stem.CrossingsX(y), i.CrossingsX(y)
		if len(js) != 2 || math.Abs(js[0]-is[0]-shift) > 0.01 || math.Abs(js[1]-is[1]-shift) > 0.01 {
			t.Errorf("at height %.0f the stem of j has the crossings %v, that of i %v (shifted by %v)", y, js, is, shift)
		}
	}
	// The stem has the head serif of i on the left only, and the top at the
	// x-height.
	checkHeadSerif(t, p, stem, xs[0], xs[1], p.XHeight)
	_, _, _, top := stem.BBox()
	near(t, "top of the stem", top, p.XHeight)
	// No foot serif: at the baseline there is the stem and nothing to its sides.
	if bs := stem.CrossingsX(1); len(bs) != 2 || math.Abs(bs[0]-p.JStemLeft) > 0.01 {
		t.Errorf("crossings at the baseline: %v, want the bare stem from %v", bs, p.JStemLeft)
	}
}

// The dot has the parameters of the dot of i and stands over the stem of j as it
// does over that of i.
func TestJDot(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	_, stem, dot, _ := jGlyph(t, p)
	ipath, _, _ := outlineOf("i", p)
	idot, istem := dotOf(t, ipath)
	jx0, jy0, jx1, jy1 := dot.BBox()
	ix0, iy0, ix1, iy1 := idot.BBox()
	near(t, "width of the dot", jx1-jx0, ix1-ix0)
	near(t, "height of the dot", jy1-jy0, iy1-iy0)
	near(t, "height of the centre of the dot", (jy0+jy1)/2, (iy0+iy1)/2)
	sx := stem.CrossingsX(100)
	near(t, "offset of the dot from the centre of the stem", (jx0+jx1)/2-(sx[0]+sx[1])/2, design.DotOffset(p))
	ix := istem.CrossingsX(100)
	near(t, "the offset is that of i", (jx0+jx1)/2-(sx[0]+sx[1])/2, (ix0+ix1)/2-(ix[0]+ix[1])/2)
	if a, b := dot[0].SignedArea(), idot[0].SignedArea(); math.Abs(a-b) > 1 {
		t.Errorf("the dot has the area %.0f, that of i %.0f", a, b)
	}
}

// The tail of j is the hook of f turned round through half a turn: with the same
// thickness, hang and terminal constants, its outer edge on the right of the stem
// begins to curve JOuterDrop above the lowest point, and its left edge, which is the
// inner edge, JInnerDrop above it, each leaving the straight edge vertically.
func TestJEdgesLeaveTheStemSmoothly(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _, _, _ := jGlyph(t, p)
	ct := path[0]
	xl, xr := float64(p.JStemLeft), p.JStemLeft+p.Stem
	for _, c := range []struct {
		name string
		node outline.Point
	}{
		{"right edge (outer)", outline.Point{X: xr, Y: -(p.JDepth - p.JOuterDrop)}},
		{"left edge (inner)", outline.Point{X: xl, Y: -(p.JDepth - p.JInnerDrop)}},
	} {
		i := nodeNear(t, ct, c.node)
		before, after := segCubic(ct, i+len(ct.Segs)-1), segCubic(ct, i)
		if c.node.X == xl {
			// The edge runs down into the curve: straight above, vertical handle below.
			if before.P0.X != before.P3.X || after.P1.X != xl {
				t.Errorf("%s: want a vertical edge into a vertical handle, got %v then %v", c.name, before, after.P1)
			}
		} else if before.P2.X != xr || after.P0.X != after.P3.X {
			t.Errorf("%s: want a vertical handle into a vertical edge, got %v then %v", c.name, before.P2, after)
		}
	}
}

// The outer edge of the tail is smooth through its lowest point, as the crest of
// the hook of f is, with horizontal handles and continuous curvature.
func TestJLowestPointIsSmooth(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _, _, _ := jGlyph(t, p)
	ct := path[0]
	var low outline.Point
	for _, q := range contourNodes(ct) {
		if q.Y < low.Y {
			low = q
		}
	}
	near(t, "depth of the lowest point", low.Y, -p.JDepth)
	i := nodeNear(t, ct, low)
	before, after := segCubic(ct, i+len(ct.Segs)-1), segCubic(ct, i)
	if math.Abs(before.P2.Y-low.Y) > 1e-9 || math.Abs(after.P1.Y-low.Y) > 1e-9 {
		t.Errorf("the handles at the lowest point are not horizontal: %v and %v", before.P2, after.P1)
	}
	if !(before.P2.X < low.X && after.P1.X > low.X) {
		t.Errorf("the handles do not point away from each other: %v and %v on %v", before.P2, after.P1, low)
	}
	k0, k1, d0, d1 := curvaturesAround(t, ct, low, crestDistance)
	t.Logf("curvature before %.4f, after %.4f (ratio %.2f); %v away: %.4f, %.4f (ratio %.2f)", k0, k1, k1/k0, crestDistance, d0, d1, d1/d0)
	for _, r := range []float64{k1 / k0, d1 / d0} {
		if r < crestRatioLow || r > crestRatioHigh {
			t.Errorf("the curvature after the lowest point is %.2f times that before it, want %v to %v", r, crestRatioLow, crestRatioHigh)
		}
	}
}

// The left side of j is an overhang: the tail reaches 50 beyond the edge of the
// advance, which is less than MaxOverhang.
func TestJOverhang(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _, _, _ := jGlyph(t, p)
	x0, _, _, _ := path.BBox()
	if x0 >= 0 || -x0 > p.MaxOverhang {
		t.Errorf("the tail starts at %.1f, want it to overhang the advance by at most %v", x0, p.MaxOverhang)
	}
}

// jCounter measures the white which the tail of j encloses between the stem, the arm
// and the ball: the width of the gap between the ball and the left edge of the stem
// at the height of the centre of the ball (where the ball is widest), and the height
// of the white from the upper edge of the arm, in the middle of that gap, up to the
// same height.  The white is open above.
func jCounter(t *testing.T, p design.Params) (width, height float64) {
	t.Helper()
	_, stem, _, _ := jGlyph(t, p)
	cy := jTerminal(p).Centre.Y
	xs := stem.CrossingsX(cy)
	if len(xs) != 4 {
		t.Fatalf("crossings at the height of the centre of the ball, %.1f: %v, want the ball and the stem", cy, xs)
	}
	width = xs[2] - xs[1]
	mid := (xs[1] + xs[2]) / 2
	ys := stem.CrossingsY(mid)
	if len(ys) < 2 {
		t.Fatalf("crossings at x = %.1f: %v, want the arm", mid, ys)
	}
	return width, cy - ys[1]
}

// The white between the stem, the arm and the ball is at least jCounterWidth wide
// and jCounterHeight high.
const (
	jCounterWidth  = 75.0
	jCounterHeight = 36.0
)

func TestJHasACounterBetweenStemArmAndBall(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	w, h := jCounter(t, p)
	t.Logf("the counter of the tail of j is %.1f wide and %.1f high", w, h)
	if w < jCounterWidth || h < jCounterHeight {
		t.Errorf("the white between the stem, the arm and the ball is %.1f by %.1f, want at least %v by %v", w, h, jCounterWidth, jCounterHeight)
	}
}
