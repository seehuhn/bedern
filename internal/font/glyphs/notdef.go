// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The glyph .notdef, which is shown for a missing character.  It is a frame
// from the baseline to the height Notdef.Height with a diagonal cross inside
// it, on the advance Notdef.Advance with Notdef.Side on each side.  The frame
// and the two diagonals, which run from corner to corner, are Notdef.Thick
// thick: the glyph is one outer contour, the rectangle, with the four triangles
// between the frame and the diagonals as counters.

// notdefLine is a straight line through a point in a direction.
type notdefLine struct{ at, dir outline.Point }

// notdefTriangle returns the triangle between three lines as a clockwise contour.
func notdefTriangle(a, b, c notdefLine) []outline.Point {
	pts := []outline.Point{
		meetLines(a.at, a.dir, b.at, b.dir),
		meetLines(b.at, b.dir, c.at, c.dir),
		meetLines(c.at, c.dir, a.at, a.dir),
	}
	area := (pts[1].X-pts[0].X)*(pts[2].Y-pts[0].Y) - (pts[2].X-pts[0].X)*(pts[1].Y-pts[0].Y)
	if area > 0 {
		pts[1], pts[2] = pts[2], pts[1]
	}
	return pts
}

// notdef draws .notdef.
func notdef(p design.Params) (outline.Path, float64) {
	q := p.Notdef
	w, h, f, t := q.Advance-2*q.Side, q.Height, q.Thick, q.Thick

	var path outline.Path
	rect(&path, 0, 0, w, h, true)

	c := outline.Point{X: w / 2, Y: h / 2}
	l := math.Hypot(w, h)
	d1 := outline.Point{X: w / l, Y: h / l}  // bottom left to top right
	d2 := outline.Point{X: w / l, Y: -h / l} // top left to bottom right
	n1 := outline.Point{X: -d1.Y, Y: d1.X}   // to the upper left of the first diagonal
	n2 := outline.Point{X: -d2.Y, Y: d2.X}   // to the upper right of the second
	off := func(n outline.Point, s float64) outline.Point {
		return outline.Point{X: c.X + s*t/2*n.X, Y: c.Y + s*t/2*n.Y}
	}
	horiz := func(y float64) notdefLine { return notdefLine{outline.Point{Y: y}, outline.Point{X: 1}} }
	vert := func(x float64) notdefLine { return notdefLine{outline.Point{X: x}, outline.Point{Y: 1}} }
	line1 := func(s float64) notdefLine { return notdefLine{off(n1, s), d1} }
	line2 := func(s float64) notdefLine { return notdefLine{off(n2, s), d2} }
	for _, tri := range [][]outline.Point{
		notdefTriangle(horiz(h-f), line1(1), line2(1)), // top
		notdefTriangle(horiz(f), line1(-1), line2(-1)), // bottom
		notdefTriangle(vert(f), line1(1), line2(-1)),   // left
		notdefTriangle(vert(w-f), line1(-1), line2(1)), // right
	} {
		addPolygon(&path, tri...)
	}
	return placeInk(path, q.Side, q.Side)
}

// placeInk moves the path sideways so that its leftmost ink lies left from the
// left edge of the advance, and returns it with the advance, which leaves right
// between the rightmost ink and the right edge.  The shift and the advance are
// whole numbers, and no side space is less than it is asked to be.
func placeInk(path outline.Path, left, right float64) (outline.Path, float64) {
	x0, _, x1, _ := path.BBox()
	dx := math.Ceil(left - x0 - 1e-6)
	var out outline.Path
	for _, ct := range path {
		c := outline.Contour{Start: outline.Point{X: ct.Start.X + dx, Y: ct.Start.Y}}
		for _, s := range ct.Segs {
			c.Segs = append(c.Segs, outline.Segment{
				IsCurve: s.IsCurve,
				P1:      outline.Point{X: s.P1.X + dx, Y: s.P1.Y},
				P2:      outline.Point{X: s.P2.X + dx, Y: s.P2.Y},
				P3:      outline.Point{X: s.P3.X + dx, Y: s.P3.Y},
			})
		}
		out = append(out, c)
	}
	return out, math.Ceil(x1 + dx + right - 1e-6)
}

// addPolygon adds the closed polygon through the points as a contour of the path.
// The points run counter-clockwise for an outer contour and clockwise for a
// counter.
func addPolygon(path *outline.Path, pts ...outline.Point) {
	path.MoveTo(pts[0].X, pts[0].Y)
	for _, q := range pts[1:] {
		path.LineTo(q.X, q.Y)
	}
	path.LineTo(pts[0].X, pts[0].Y)
}

// meetLines returns the point where the line through a with the direction
// da meets the line through b with the direction db.
func meetLines(a, da, b, db outline.Point) outline.Point {
	den := da.X*db.Y - da.Y*db.X
	t := ((b.X-a.X)*db.Y - (b.Y-a.Y)*db.X) / den
	return outline.Point{X: a.X + t*da.X, Y: a.Y + t*da.Y}
}
