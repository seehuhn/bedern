// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// yGlyph returns the outline and the advance of y.  The tests check the upper part
// of v, the thin diagonal continued below the baseline into a tail which curves to
// the left and ends in the terminal of j; the diagonals meeting at 0.18 to 0.27 of
// the x-height (118, 0.263); a tail 200 to 283 deep which reaches 193 to 285 to the
// left of the centre line of the thin diagonal at the baseline; an advance equal to
// that of v, 504.

func yGlyph(t *testing.T, p design.Params) (outline.Path, float64) {
	t.Helper()
	path, adv, ok := outlineOf("y", p)
	if !ok {
		t.Fatal("glyph y not found")
	}
	if len(path) != 1 {
		t.Fatalf("y has %d contours, want 1", len(path))
	}
	if path[0].SignedArea() <= 0 {
		t.Error("the contour of y is not counter-clockwise")
	}
	return path, adv
}

func TestYAdvanceAndUpperPart(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, adv := yGlyph(t, p)
	v, vadv := vGlyph(t, p)
	tight(t, "advance", adv, 505)
	tight(t, "advance of y is that of v", adv, vadv)
	if r := adv / vadv; r < 0.98 || r > 1.04 {
		t.Errorf("y is %.3f of v wide, want 0.98 to 1.04", r)
	}
	// The upper part is that of v: above the junction the runs are the same.
	for _, y := range []float64{130, 200, 300, 380, 425, 440, 447} {
		a, b := path.CrossingsX(y), v.CrossingsX(y)
		if len(a) != len(b) {
			t.Fatalf("at height %v y has the crossings %v, v has %v", y, a, b)
		}
		for i := range a {
			if math.Abs(a[i]-b[i]) > 1e-9 {
				t.Errorf("at height %v y has the crossings %v, v has %v", y, a, b)
				break
			}
		}
	}
	_, _, x1, y1 := path.BBox()
	tight(t, "right space", adv-x1, p.DiagSideRight)
	tight(t, "top", y1, p.XHeight)
}

func TestYJunction(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := yGlyph(t, p)
	got := math.NaN()
	for y := 300.0; y > 0; y -= 0.02 {
		if len(path.CrossingsX(y)) == 2 {
			got = y
			break
		}
	}
	near(t, "junction of the diagonals", got, design.VCrotch(p))
	if r := got / p.XHeight; r < 0.183 || r > 0.272 {
		t.Errorf("the diagonals meet at %.3f x-heights, want 0.183 to 0.272", r)
	}
	// Below the junction there is one run to the lowest point of the tail.
	for _, y := range []float64{100, 50, 0, -40, -50} {
		if n := len(path.CrossingsX(y)); n != 2 {
			t.Errorf("%d crossings at height %v, want one run", n, y)
		}
	}
	// The thin diagonal goes on below the baseline in a straight line: the right-hand
	// edge, the outer edge of the tail, is the same line from the serif down to where
	// the tail begins to curve.
	rs := readDiagonals(t, path, 150, 330, 2)
	for _, y := range []float64{100, 50, 0, -20} {
		xs := path.CrossingsX(y)
		want := rs[1].right + rs[1].lean*(y-150)
		tight(t, "right-hand edge of the tail", xs[len(xs)-1], want)
	}
	// The left-hand edge of the thin diagonal is one line from the crotch to where
	// the tail begins to curve, past the foot of the thick diagonal.
	for _, y := range []float64{5, 0, -20} {
		xs := path.CrossingsX(y)
		want := rs[1].left + rs[1].lean*(y-150)
		tight(t, "left-hand edge of the tail", xs[0], want)
	}
}

// The tail of y keeps its distance from the preceding glyph.
func TestYTailKeepsAwayFromWhatPrecedes(t *testing.T) {
	t.Parallel()
	for _, v := range design.Variants() {
		y, _ := yGlyph(t, v.Params)
		for _, name := range followers {
			prev, adv, ok := outlineOf(name, v.Params)
			if !ok {
				continue
			}
			d, a, b := minDistanceBetween(prev, shiftedPath(y, adv))
			if v.Name == "regular" {
				t.Logf("%s followed by y: nearest approach %.1f, at (%.0f, %.0f) and (%.0f, %.0f)", name, d, a.X, a.Y, b.X, b.Y)
			}
			if d < pairClearance {
				t.Errorf("%s: %s followed by y: nearest approach %.1f, want at least %v", v.Name, name, d, pairClearance)
			}
		}
	}
}

// yCounter measures the white which the tail of y encloses between the thin
// diagonal, the arm and the ball: the width of the gap between the ball and the
// left-hand edge of the diagonal at the height of the centre of the ball (where the
// ball is widest), and the height of the white from the upper edge of the arm, in the
// middle of that gap, up to the highest ink of the ball.
func yCounter(t *testing.T, p design.Params) (width, height float64) {
	t.Helper()
	path, _ := yGlyph(t, p)
	c := yTerminal(p).Centre
	xs := path.CrossingsX(c.Y)
	if len(xs) != 4 {
		t.Fatalf("crossings at the height of the centre of the ball, %.1f: %v, want the ball and the diagonal", c.Y, xs)
	}
	width = xs[2] - xs[1]
	mid := (xs[1] + xs[2]) / 2
	ys := path.CrossingsY(mid)
	if len(ys) < 2 {
		t.Fatalf("crossings at x = %.1f: %v, want the arm", mid, ys)
	}
	return width, (c.Y + p.BallDiameter/2) - ys[1]
}

// The white between the diagonal, the arm and the ball is at least 55 wide and 40
// high.
func TestYHasACounterBetweenDiagonalArmAndBall(t *testing.T) {
	t.Parallel()
	w, h := yCounter(t, design.Regular())
	t.Logf("the counter of the tail of y is %.1f wide and %.1f high", w, h)
	if w < 55 || h < 40 {
		t.Errorf("the white between the diagonal, the arm and the ball is %.1f by %.1f, want at least 55 by 40", w, h)
	}
}

// The outer edge of the tail is smooth through its lowest point:
// horizontal handles which point away from each other, and the curvature after the
// lowest point is 0.8 to 1.25 times that before it, at the point and 5 units away.
func TestYLowestPointIsSmooth(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	path, _ := yGlyph(t, p)
	ct := path[0]
	var low outline.Point
	for _, q := range contourNodes(ct) {
		if q.Y < low.Y {
			low = q
		}
	}
	i := nodeNear(t, ct, low)
	before, after := segCubic(ct, i+len(ct.Segs)-1), segCubic(ct, i)
	if before.P2.Y != low.Y || after.P1.Y != low.Y {
		t.Errorf("the handles at the lowest point are not horizontal: %v and %v", before.P2, after.P1)
	}
	if !(before.P2.X < low.X && after.P1.X > low.X) {
		t.Errorf("the handles do not point away from each other: %v and %v on %v", before.P2, after.P1, low)
	}
	k0, k1, d0, d1 := curvaturesAround(t, ct, low, crestDistance)
	t.Logf("curvature before %.4f, after %.4f (ratio %.2f); %v away: %.4f, %.4f (ratio %.2f)", k0, k1, k1/k0, crestDistance, d0, d1, d1/d0)
	// The lowest point lies inside one cubic curve of the outline.
	for _, r := range []float64{k1 / k0} {
		if r < crestRatioLow || r > crestRatioHigh {
			t.Errorf("the curvature after the lowest point is %.2f times that before it, want %v to %v", r, crestRatioLow, crestRatioHigh)
		}
	}
}
