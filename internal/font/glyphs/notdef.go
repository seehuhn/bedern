// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The glyph .notdef, which is shown for a missing character.  It is a frame
// from the baseline to the height of the figures with a diagonal cross inside
// it, on the advance of the figures with NotdefSide on each side.  The frame
// and the two diagonals, which run from corner to corner, are NotdefThick
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
	q := p.Punct
	w, h, f, t := p.Figures.Advance-2*q.NotdefSide, p.Figures.Height, q.NotdefThick, q.NotdefThick

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
	return placeInk(path, design.PunctSides{Left: q.NotdefSide, Right: q.NotdefSide})
}
