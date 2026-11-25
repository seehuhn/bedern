// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"slices"

	"seehuhn.de/go/bedern/internal/outline"
)

// The circular arcs, circles, rings and rotations from which the arrows, the
// relations and the circled operators are built (arrows.go, relations.go).

// arcSeg returns the cubic curve which follows the circle about c of the
// radius r from the position angle t0 to t1 (radians, either way round).
func arcSeg(c outline.Point, r, t0, t1 float64) outline.Segment {
	k := 4.0 / 3 * math.Tan((t1-t0)/4) * r
	p0 := pt(c.X+r*math.Cos(t0), c.Y+r*math.Sin(t0))
	p3 := pt(c.X+r*math.Cos(t1), c.Y+r*math.Sin(t1))
	return outline.Segment{
		IsCurve: true,
		P1:      pt(p0.X-k*math.Sin(t0), p0.Y+k*math.Cos(t0)),
		P2:      pt(p3.X+k*math.Sin(t1), p3.Y-k*math.Cos(t1)),
		P3:      p3,
	}
}

// arcSegs returns the arc from t0 to t1 as cubic curves of at most a quarter
// turn each.
func arcSegs(c outline.Point, r, t0, t1 float64) []outline.Segment {
	n := int(math.Ceil(math.Abs(t1-t0)/(math.Pi/2) - 1e-9))
	n = max(n, 1)
	var out []outline.Segment
	for i := range n {
		a := t0 + (t1-t0)*float64(i)/float64(n)
		b := t0 + (t1-t0)*float64(i+1)/float64(n)
		out = append(out, arcSeg(c, r, a, b))
	}
	return out
}

// circle returns the circle about c of the radius r, counter-clockwise for
// ink, or clockwise round a counter if counter is true, starting at its right
// extreme.
func circle(c outline.Point, r float64, counter bool) outline.Contour {
	t1 := 2 * math.Pi
	if counter {
		t1 = -t1
	}
	segs := arcSegs(c, r, 0, t1)
	segs[len(segs)-1].P3 = pt(c.X+r, c.Y)
	return outline.Contour{Start: pt(c.X+r, c.Y), Segs: segs}
}

// ring returns the ring about c whose outer edge has the radius r and whose
// stroke is t thick.
func ring(c outline.Point, r, t float64) outline.Path {
	return outline.Path{circle(c, r, false), circle(c, r-t, true)}
}

// rotated returns the path turned counter-clockwise by the angle a (radians)
// about c.  A turn keeps the orientation of the contours.
func rotated(p outline.Path, c outline.Point, a float64) outline.Path {
	s, co := math.Sin(a), math.Cos(a)
	return mapPoints(p, func(q outline.Point) outline.Point {
		x, y := q.X-c.X, q.Y-c.Y
		return pt(c.X+co*x-s*y, c.Y+s*x+co*y)
	})
}

// slab returns the counter-clockwise rectangle of the thickness t about the
// segment from a to b, its ends cut at right angles to it.
func slab(a, b outline.Point, t float64) outline.Contour {
	l := math.Hypot(b.X-a.X, b.Y-a.Y)
	nx, ny := -(b.Y-a.Y)/l*t/2, (b.X-a.X)/l*t/2
	return polygon(pt(a.X-nx, a.Y-ny), pt(b.X-nx, b.Y-ny), pt(b.X+nx, b.Y+ny), pt(a.X+nx, a.Y+ny))
}

// runAt returns the runs of ink of the path along the horizontal line at y,
// in increasing order.
func runAt(p outline.Path, y float64) [][2]float64 {
	xs := p.CrossingsX(y)
	slices.Sort(xs)
	var out [][2]float64
	for i := 0; i+1 < len(xs); i += 2 {
		out = append(out, [2]float64{xs[i], xs[i+1]})
	}
	return out
}
