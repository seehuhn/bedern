// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// g2 returns the cubic from p0 in the direction t0 to p3 in the direction t3
// whose curvature is k0 at its start and k3 at its end (positive anticlockwise),
// found from the lengths a and b of its two handles.  The result is false if
// no such curve with positive handles exists near the chord.
func g2(p0, t0 Point, k0 float64, p3, t3 Point, k3 float64) (cubic, bool) {
	return g2Near(p0, t0, k0, p3, t3, k3, 0, 0)
}

// g2Near is g2 with the root of the equations taken which lies nearest to the
// handle lengths a and b (not used for a or b of 0).
func g2Near(p0, t0 Point, k0 float64, p3, t3 Point, k3 float64, a, b float64) (cubic, bool) {
	d := sub(p3, p0)
	c01 := cross(t0, t3)
	f := func(a, b float64) (float64, float64) {
		return 2.0/3*(cross(t0, d)-b*c01) - k0*a*a, 2.0/3*(-a*c01-cross(t3, d)) - k3*b*b
	}
	if a > 0 && b > 0 {
		if c, ok := g2From(p0, t0, p3, t3, f, a, b); ok {
			return c, true
		}
	}
	for _, start := range []float64{1.0 / 3, 0.15, 0.6, 0.05, 1.0, 0.3} {
		if c, ok := g2From(p0, t0, p3, t3, f, start*norm(d), start*norm(d)); ok {
			return c, true
		}
	}
	return cubic{}, false
}

// g2From runs the Newton iteration of g2 from handles of the given length.
func g2From(p0, t0, p3, t3 Point, f func(a, b float64) (float64, float64), a, b float64) (cubic, bool) {
	for iter := 0; iter < 200; iter++ {
		f0, f1 := f(a, b)
		if math.Abs(f0) < 1e-10 && math.Abs(f1) < 1e-10 {
			break
		}
		const h = 1e-6
		fa0, fa1 := f(a+h, b)
		fb0, fb1 := f(a, b+h)
		j00, j10 := (fa0-f0)/h, (fa1-f1)/h
		j01, j11 := (fb0-f0)/h, (fb1-f1)/h
		det := j00*j11 - j01*j10
		if math.Abs(det) < 1e-14 {
			return cubic{}, false
		}
		da := (f0*j11 - f1*j01) / det
		db := (j00*f1 - j10*f0) / det
		// Damp steps which leave the positive handles.
		s := 1.0
		for (a-s*da <= 1e-3 || b-s*db <= 1e-3) && s > 1e-6 {
			s /= 2
		}
		a -= s * da
		b -= s * db
	}
	f0, f1 := f(a, b)
	if math.Abs(f0) > 1e-6 || math.Abs(f1) > 1e-6 || a <= 0 || b <= 0 {
		return cubic{}, false
	}
	return cubic{p0, add(p0, mul(t0, a)), sub(p3, mul(t3, b)), p3}, true
}

// keep is the curvature which setEnds takes from the curve it replaces.
var keep = math.NaN()

// setEnds returns the path with the curve which ends nearest to q drawn again
// with the curvatures k0 at its start and k3 at its end (keep: those of the
// curve now), the nodes and the directions at them as they are.
func setEnds(path outline.Path, q Point, k0, k3 float64) outline.Path {
	ci, si := segAt(path, q)
	c := path[ci]
	old := curveAt(c, si)
	if math.IsNaN(k0) {
		k0 = old.curvature(0)
	}
	if math.IsNaN(k3) {
		k3 = old.curvature(1)
	}
	m, ok := g2Near(old[0], unit(sub(old[1], old[0])), k0, old[3], unit(sub(old[3], old[2])), k3, dist(old[0], old[1]), dist(old[3], old[2]))
	if !ok {
		panic("upper: no curve with the curvatures of a refit")
	}
	nc := outline.Contour{Start: c.Start, Segs: append([]outline.Segment(nil), c.Segs...)}
	setCurve(&nc, si, m)
	out := append(outline.Path(nil), path...)
	out[ci] = nc
	return out
}
