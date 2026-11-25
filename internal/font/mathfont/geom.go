// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// The transformations of outlines by which the derived glyphs are made.  An
// outline keeps the orientation of the glyphs of the regular
// (bedern/internal/font/glyphs/doc.go): counter-clockwise outside, clockwise
// round a counter.  A mirror image reverses its contours to keep it.

// mapPoints returns the path with every point mapped by f.
func mapPoints(p outline.Path, f func(outline.Point) outline.Point) outline.Path {
	out := make(outline.Path, 0, len(p))
	for _, c := range p {
		n := outline.Contour{Start: f(c.Start), Segs: make([]outline.Segment, len(c.Segs))}
		for i, s := range c.Segs {
			n.Segs[i] = outline.Segment{IsCurve: s.IsCurve, P1: f(s.P1), P2: f(s.P2), P3: f(s.P3)}
		}
		out = append(out, n)
	}
	return out
}

// shifted returns the path moved by (dx, dy).
func shifted(p outline.Path, dx, dy float64) outline.Path {
	return mapPoints(p, func(q outline.Point) outline.Point { return outline.Point{X: q.X + dx, Y: q.Y + dy} })
}

// reversed returns the contour traversed the other way round.
func reversed(c outline.Contour) outline.Contour {
	segs := c.Segs
	if n := len(segs); n > 0 && segs[n-1].P3 != c.Start {
		segs = append(segs[:n:n], outline.Segment{P3: c.Start})
	}
	n := len(segs)
	r := outline.Contour{Start: segs[n-1].P3}
	for i := n - 1; i >= 0; i-- {
		from := c.Start
		if i > 0 {
			from = segs[i-1].P3
		}
		s := segs[i]
		r.Segs = append(r.Segs, outline.Segment{IsCurve: s.IsCurve, P1: s.P2, P2: s.P1, P3: from})
	}
	return r
}

// mirroredX returns the mirror image of the path in the vertical line at x,
// with its contours reversed.
func mirroredX(p outline.Path, x float64) outline.Path {
	m := mapPoints(p, func(q outline.Point) outline.Point { return outline.Point{X: 2*x - q.X, Y: q.Y} })
	for i := range m {
		m[i] = reversed(m[i])
	}
	return m
}

// mirroredY returns the mirror image of the path in the horizontal line at y,
// with its contours reversed.
func mirroredY(p outline.Path, y float64) outline.Path {
	m := mapPoints(p, func(q outline.Point) outline.Point { return outline.Point{X: q.X, Y: 2*y - q.Y} })
	for i := range m {
		m[i] = reversed(m[i])
	}
	return m
}

// turned returns the path turned through half a turn about (cx, cy).
func turned(p outline.Path, cx, cy float64) outline.Path {
	return mapPoints(p, func(q outline.Point) outline.Point { return outline.Point{X: 2*cx - q.X, Y: 2*cy - q.Y} })
}

// polygon returns the closed contour through the points, which must run
// counter-clockwise for ink.
func polygon(pts ...outline.Point) outline.Contour {
	c := outline.Contour{Start: pts[0]}
	for _, q := range pts[1:] {
		c.Segs = append(c.Segs, outline.Segment{P3: q})
	}
	c.Segs = append(c.Segs, outline.Segment{P3: pts[0]})
	return c
}

// box returns the counter-clockwise rectangle from (x0, y0) to (x1, y1).
func box(x0, y0, x1, y1 float64) outline.Contour {
	return polygon(pt(x0, y0), pt(x1, y0), pt(x1, y1), pt(x0, y1))
}

func pt(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }

// bbox returns the bounding box of the path.
func bbox(p outline.Path) (x0, y0, x1, y1 float64) { return p.BBox() }

// centredIn returns the path moved sideways to centre its ink in the
// advance adv, to a whole number of units.
func centredIn(p outline.Path, adv float64) outline.Path {
	x0, _, x1, _ := p.BBox()
	return shifted(p, math.Round((adv-(x0+x1))/2), 0)
}

// withSides returns the path moved to start its ink at left, and the
// advance which leaves right after it, both to whole units.
func withSides(p outline.Path, left, right float64) (outline.Path, float64) {
	x0, _, x1, _ := p.BBox()
	dx := math.Round(left - x0)
	return shifted(p, dx, 0), math.Round(x1 + dx + right)
}

// contours returns the path of the contours.
func contours(cs ...outline.Contour) outline.Path { return outline.Path(cs) }

// join returns the contours of the paths in one path.
func join(ps ...outline.Path) outline.Path {
	var out outline.Path
	for _, p := range ps {
		out = append(out, p...)
	}
	return out
}
