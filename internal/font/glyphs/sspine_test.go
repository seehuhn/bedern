// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// sCurvature returns the curvature of the cubic c at t.
func sCurvature(c cubic, t float64) float64 {
	u := 1 - t
	x1 := 3*u*u*(c.P1.X-c.P0.X) + 6*u*t*(c.P2.X-c.P1.X) + 3*t*t*(c.P3.X-c.P2.X)
	y1 := 3*u*u*(c.P1.Y-c.P0.Y) + 6*u*t*(c.P2.Y-c.P1.Y) + 3*t*t*(c.P3.Y-c.P2.Y)
	x2 := 6*u*(c.P2.X-2*c.P1.X+c.P0.X) + 6*t*(c.P3.X-2*c.P2.X+c.P1.X)
	y2 := 6*u*(c.P2.Y-2*c.P1.Y+c.P0.Y) + 6*t*(c.P3.Y-2*c.P2.Y+c.P1.Y)
	return (x1*y2 - y1*x2) / math.Pow(x1*x1+y1*y1, 1.5)
}

// sLongEdges returns the two long edges of s in the frame of the lower half
// (C at the origin), each as its six cubic segments from one beak to the
// other: the outer edge of the lower half followed by the inner edge of the
// upper, and the outer edge of the upper half followed by the inner edge of
// the lower.
func sLongEdges(f sFrame) [2][]cubic {
	turned := func(c cubic) cubic { // rotated through 180 degrees about the origin
		for _, q := range []*outline.Point{&c.P0, &c.P1, &c.P2, &c.P3} {
			q.X, q.Y = -q.X, -q.Y
		}
		return c
	}
	L, U := f.LowerE, f.UpperE
	return [2][]cubic{
		{L.Outer[2].reversed(), L.Outer[1].reversed(), L.Outer[0].reversed(), turned(U.Inner[0]), turned(U.Inner[1]), turned(U.Inner[2])},
		{turned(U.Outer[2]).reversed(), turned(U.Outer[1]).reversed(), turned(U.Outer[0]).reversed(), L.Inner[0], L.Inner[1], L.Inner[2]},
	}
}

// Each long edge is one smooth curve from beak to beak: its curvature
// changes sign once, no stretch of it is straight, and it has no elbow, where
// the curvature changes abruptly.  A straight stretch is one with a
// curvature below 0.002 per unit.
func TestSSpineEdgesAreOneFlowingCurve(t *testing.T) {
	t.Parallel()
	const (
		flatCurvature = 0.002 // 1/units: below this an edge counts as straight
		maxFlat       = 80.0  // units: the longest straight stretch allowed
		maxJump       = 0.006 // 1/units: the largest change of curvature between samples 1/40 of a segment apart (0.003 in the regular set)
	)
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			for k, edge := range sLongEdges(newSFrame(np.params)) {
				var s, kap []float64
				length, prev := 0.0, edge[0].at(0)
				for _, c := range edge {
					for i := 0; i <= 40; i++ {
						q := c.at(float64(i) / 40)
						length += math.Hypot(q.X-prev.X, q.Y-prev.Y)
						prev = q
						s = append(s, length)
						kap = append(kap, sCurvature(c, float64(i)/40))
					}
				}
				changes, flat, flatFrom, jump := 0, 0.0, -1.0, 0.0
				for i, v := range kap {
					if i > 0 {
						if v*kap[i-1] < 0 {
							changes++
						}
						jump = math.Max(jump, math.Abs(v-kap[i-1]))
					}
					if math.Abs(v) < flatCurvature {
						if flatFrom < 0 {
							flatFrom = s[i]
						}
						flat = math.Max(flat, s[i]-flatFrom)
					} else {
						flatFrom = -1
					}
				}
				if changes != 1 {
					t.Errorf("edge %d: the curvature changes sign %d times, want once", k, changes)
				}
				if flat > maxFlat {
					t.Errorf("edge %d: a straight stretch of %.0f units, want at most %v", k, flat, maxFlat)
				}
				if jump > maxJump {
					t.Errorf("edge %d: the curvature changes by %.4f between neighbouring samples, want at most %v", k, jump, maxJump)
				}
			}
		})
	}
}

// The bowls are round.  At the leftmost point of the upper bowl and the
// rightmost point of the lower bowl the outline runs vertically, and its
// radius of curvature there is at least 0.9 of the half-height of a bowl (a
// quarter of the height of the letter).  The extremes lie at
// the heights 0.69 to 0.76 and 0.26 to 0.32 of the
// x-height, and the counters have no tight corner: along the inner edges the
// radius stays above 0.4 of the half-height.
func TestSBowlsAreRound(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		if !boldApplies(t, np) {
			continue
		}
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			f := newSFrame(p)
			half := (p.XHeight + 2*p.OvershootLower) / 4
			for _, x := range []struct {
				name   string
				c      cubic // ends at the extreme
				y      float64
				lo, hi float64
			}{
				{"upper bowl, left", f.UpperE.Outer[0], f.C.Y - f.UpperE.Outer[0].P3.Y, 0.69, 0.76},
				{"lower bowl, right", f.LowerE.Outer[0], f.C.Y + f.LowerE.Outer[0].P3.Y, 0.26, 0.32},
			} {
				tg := x.c.tangent(1)
				if math.Abs(tg.X) > 1e-9*math.Hypot(tg.X, tg.Y) {
					t.Errorf("%s: the tangent at the extreme is %v, not vertical", x.name, tg)
				}
				if r := 1 / math.Abs(sCurvature(x.c, 1)); r < 0.9*half {
					t.Errorf("%s: radius %.0f at the extreme, want at least %.0f", x.name, r, 0.9*half)
				}
				if h := x.y / p.XHeight; h < x.lo || h > x.hi {
					t.Errorf("%s: the extreme lies at %.3f of the x-height, want %v to %v", x.name, h, x.lo, x.hi)
				}
			}
			for _, e := range []sEdges{f.LowerE, f.UpperE} {
				for _, c := range e.Inner[:2] {
					for i := 0; i <= 40; i++ {
						if r := 1 / math.Abs(sCurvature(c, float64(i)/40)); r < 0.4*half {
							t.Errorf("a counter has a radius of %.0f, want at least %.0f", r, 0.4*half)
							return
						}
					}
				}
			}
		})
	}
}

// The letter stands upright: the centre of its ink lies within 8 units of the
// middle of its ink box across and within 6 units of it up and down.
func TestSStandsUpright(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, _, _ := outlineOf("s", np.params)
		pts, _ := sPoints(path)
		var a, cx, cy float64
		for i, q := range pts {
			r := pts[(i+1)%len(pts)]
			cr := q.X*r.Y - r.X*q.Y
			a += cr / 2
			cx += (q.X + r.X) * cr / 6
			cy += (q.Y + r.Y) * cr / 6
		}
		cx, cy = cx/a, cy/a
		x0, y0, x1, y1 := path.BBox()
		if dx := cx - (x0+x1)/2; math.Abs(dx) > 8 {
			t.Errorf("%s: the centre of the ink lies %.1f units right of the middle", np.name, dx)
		}
		if dy := cy - (y0+y1)/2; math.Abs(dy) > 6 {
			t.Errorf("%s: the centre of the ink lies %.1f units above the middle", np.name, dy)
		}
	}
}
