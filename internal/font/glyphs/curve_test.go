// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// TestCubicCutAtHeight checks fromY and toY on the upper right quarter of the
// o: the cut starts (or ends) exactly at the height, lies on the original
// curve, and the two halves together run through the same point.
func TestCubicCutAtHeight(t *testing.T) {
	t.Parallel()
	f := newEFrame(design.Regular())
	c := f.outerU.cubic(0, true) // right extreme to the top
	const y = 330.0
	up, low := c.fromY(y), c.toY(y)
	if up.P0.Y != y || low.P3.Y != y {
		t.Errorf("the cuts are at heights %v and %v, want exactly %v", up.P0.Y, low.P3.Y, y)
	}
	if up.P3 != c.P3 || low.P0 != c.P0 {
		t.Errorf("the other ends moved: %v and %v", up.P3, low.P0)
	}
	tc := c.tAtY(y)
	if got := c.at(tc); math.Abs(got.X-up.P0.X) > 1e-6 || math.Abs(got.X-low.P3.X) > 1e-6 {
		t.Errorf("the cuts lie at x = %v and %v, the curve at %v", up.P0.X, low.P3.X, got.X)
	}
	// A point of the upper part is a point of the curve.
	mid := up.at(0.5)
	d := math.Inf(1)
	for k := 0; k <= 4000; k++ {
		q := c.at(float64(k) / 4000)
		d = math.Min(d, math.Hypot(q.X-mid.X, q.Y-mid.Y))
	}
	if d > 0.05 {
		t.Errorf("the middle of the upper part lies %v from the curve", d)
	}
}

// testCubic is a curve with no symmetry, for the tests of the helpers of curves.
var testCubic = cubic{
	outline.Point{X: 10, Y: 20}, outline.Point{X: 120, Y: -40},
	outline.Point{X: 260, Y: 180}, outline.Point{X: 330, Y: 60},
}

func closePoint(a, b outline.Point, tol float64) bool {
	return math.Hypot(a.X-b.X, a.Y-b.Y) <= tol
}

// A part of a curve at u is the whole curve at the mapped parameter, for every
// pair of parameters, including those with t1 = 0, with t0 above t1 and with
// t0 = t1.
func TestSubIsTheCurveAtTheMappedParameter(t *testing.T) {
	t.Parallel()
	c := testCubic
	params := []float64{0, 0.1, 0.25, 0.5, 0.75, 1}
	for _, t0 := range params {
		for _, t1 := range params {
			s := c.sub(t0, t1)
			for _, u := range []float64{0, 0.2, 0.5, 0.9, 1} {
				if got, want := s.at(u), c.at(t0+u*(t1-t0)); !closePoint(got, want, 1e-9) {
					t.Errorf("sub(%v, %v).at(%v) = %v, want %v", t0, t1, u, got, want)
				}
			}
		}
	}
	// Parameters outside the curve continue it.
	for _, pr := range [][2]float64{{-0.5, 0}, {-0.5, -0.1}, {-0.3, 0.4}, {0.6, 1.4}, {1, 1.5}, {1.2, 1.6}, {1.5, 0.5}} {
		s := c.sub(pr[0], pr[1])
		for _, u := range []float64{0, 0.5, 1} {
			if got, want := s.at(u), c.at(pr[0]+u*(pr[1]-pr[0])); !closePoint(got, want, 1e-9) {
				t.Errorf("sub(%v, %v).at(%v) = %v, want %v", pr[0], pr[1], u, got, want)
			}
		}
	}
}

func TestSubOfTheWholeAndReversed(t *testing.T) {
	t.Parallel()
	c := testCubic
	if s := c.sub(0, 1); s != c {
		t.Errorf("sub(0, 1) = %v, want the curve %v", s, c)
	}
	r := c.sub(0.7, 0.2)
	f := c.sub(0.2, 0.7)
	if !closePoint(r.P0, f.P3, 1e-12) || !closePoint(r.P1, f.P2, 1e-12) || !closePoint(r.P2, f.P1, 1e-12) || !closePoint(r.P3, f.P0, 1e-12) {
		t.Errorf("sub(0.7, 0.2) = %v is not sub(0.2, 0.7) = %v reversed", r, f)
	}
	if z := c.sub(0, 0); !closePoint(z.P0, c.P0, 1e-12) || !closePoint(z.P3, c.P0, 1e-12) {
		t.Errorf("sub(0, 0) = %v, want the point %v", z, c.P0)
	}
}

func TestSplitJoinsTheWhole(t *testing.T) {
	t.Parallel()
	c := testCubic
	for _, u := range []float64{0.1, 0.5, 0.8} {
		head, tail := c.split(u, true), c.split(u, false)
		if head.P3 != tail.P0 || head.P0 != c.P0 || tail.P3 != c.P3 {
			t.Errorf("split at %v: the parts do not meet or do not keep the ends", u)
		}
		for _, v := range []float64{0, 0.3, 1} {
			if got, want := head.at(v), c.at(u*v); !closePoint(got, want, 1e-9) {
				t.Errorf("head at %v of the split at %v = %v, want %v", v, u, got, want)
			}
			if got, want := tail.at(v), c.at(u+(1-u)*v); !closePoint(got, want, 1e-9) {
				t.Errorf("tail at %v of the split at %v = %v, want %v", v, u, got, want)
			}
		}
	}
}

// The tangent of a straight line has the direction of the line, and that of a
// quarter circle turns by a quarter turn.
func TestTangent(t *testing.T) {
	t.Parallel()
	line := cubic{outline.Point{X: 0, Y: 0}, outline.Point{X: 10, Y: 20}, outline.Point{X: 20, Y: 40}, outline.Point{X: 30, Y: 60}}
	for _, u := range []float64{0, 0.3, 1} {
		tg := line.tangent(u)
		if n := math.Hypot(tg.X, tg.Y); math.Abs(n-1) > 1e-9 || math.Abs(tg.X*2-tg.Y) > 1e-9 {
			t.Errorf("tangent of the line at %v = %v, want the unit vector (1, 2)/sqrt 5", u, tg)
		}
	}
	q := arcPieces(outline.Point{}, 100, 0, math.Pi/2)[0].cubic
	if a, b := q.tangent(0), q.tangent(1); !closePoint(a, outline.Point{X: 0, Y: 1}, 1e-9) || !closePoint(b, outline.Point{X: -1, Y: 0}, 1e-9) {
		t.Errorf("tangents of the quarter circle: %v and %v, want (0, 1) and (-1, 0)", a, b)
	}
}

// The curvature of a circle of the radius r is 1/r, within 3 per cent (the error
// of four cubic pieces, which is largest at their ends), and that of a straight line is 0.
func TestCurvature(t *testing.T) {
	t.Parallel()
	for _, r := range []float64{30, 100, 450} {
		for _, pc := range arcPieces(outline.Point{X: 7, Y: -3}, r, 0, 2*math.Pi) {
			for _, u := range []float64{0, 0.25, 0.5, 0.75, 1} {
				if k := pc.cubic.curvature(u); math.Abs(k*r-1) > 0.03 {
					t.Errorf("radius %v: the curvature at %v is %v, want %v", r, u, k, 1/r)
				}
			}
		}
	}
	line := cubic{outline.Point{X: 0, Y: 0}, outline.Point{X: 10, Y: 20}, outline.Point{X: 20, Y: 40}, outline.Point{X: 30, Y: 60}}
	for _, u := range []float64{0, 0.5, 1} {
		if k := line.curvature(u); math.Abs(k) > 1e-12 {
			t.Errorf("the curvature of a line at %v is %v", u, k)
		}
	}
	// The curvature of the reversed curve is the negative.
	if a, b := testCubic.curvature(0.3), testCubic.reversed().curvature(0.7); math.Abs(a+b) > 1e-12 {
		t.Errorf("the curvature %v of a curve and %v of its reverse do not differ in sign", a, b)
	}
}
