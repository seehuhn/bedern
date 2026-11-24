// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package fair

import (
	"math"
)

// thickStep is the arc length between two measurements of the thickness.
const thickStep = 3.0

// maxThick is the longest ray; a stroke is never thicker.
const maxThick = 450.0

// steepCos is the cosine of the smallest angle between the ray and the edge
// through which it leaves the ink, for the measurement to count as the width
// of a stroke: 50 degrees.  A ray which grazes the opposite edge measures
// along a stroke, not across it.
var steepCos = math.Cos(50 * math.Pi / 180)

type edge struct {
	a, b                   Point
	minX, maxX, minY, maxY float64
}

// flatPolys returns the edges of the flattened contours.
func flatPolys(cs [][]seg) []edge {
	var es []edge
	for _, ss := range cs {
		pts := flatten(ss, 48)
		for i, a := range pts {
			b := pts[(i+1)%len(pts)]
			if a == b {
				continue
			}
			es = append(es, edge{a, b, min(a.X, b.X), max(a.X, b.X), min(a.Y, b.Y), max(a.Y, b.Y)})
		}
	}
	return es
}

// thickness measures the ink along the inward normal at every thickStep of
// the samples.
func thickness(samples []Sample, es []edge) []Thick {
	var out []Thick
	next := 0.0
	for i, s := range samples {
		if s.S < next && i != len(samples)-1 {
			continue
		}
		next = s.S + thickStep
		if s.T == (Point{}) {
			continue
		}
		n := Point{X: -s.T.Y, Y: s.T.X} // the ink is on the left
		o := add(s.P, mul(n, 0.05))
		end := add(o, mul(n, maxThick))
		bx0, bx1 := min(o.X, end.X), max(o.X, end.X)
		by0, by1 := min(o.Y, end.Y), max(o.Y, end.Y)
		best := math.Inf(1)
		var hit edge
		for _, e := range es {
			if e.maxX < bx0 || e.minX > bx1 || e.maxY < by0 || e.minY > by1 {
				continue
			}
			d := sub(e.b, e.a)
			den := cross(n, d)
			if den == 0 {
				continue
			}
			w := sub(e.a, o)
			t := cross(w, d) / den // along the ray
			u := cross(w, n) / den // along the edge
			// The flattened edge cuts the chord of a convex curve up to
			// a fraction of a unit inside the ink, so that a crossing
			// close to the start is the edge itself.
			if t > 1.5 && t < best && u >= 0 && u <= 1 {
				best, hit = t, e
			}
		}
		if math.IsInf(best, 1) {
			continue
		}
		d := unit(sub(hit.b, hit.a))
		valid := math.Abs(cross(n, d)) >= steepCos
		out = append(out, Thick{S: s.S, P: s.P, Q: add(o, mul(n, best)), T: best, Valid: valid})
	}
	return out
}
