// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/outline"
)

// A superShape is the curve code of the bowls: the four extreme points of a
// super-ellipse and the directions in which a counter-clockwise curve leaves
// them.  The o draws all four segments of its two shapes; a bowl which hangs
// on a stem draws the two on the free side from the same shape.
type superShape struct {
	ex, dir [4][2]float64
}

// newSuperShape returns the shape which touches the given rectangle at the
// midpoints of its sides.  For superness 1/sqrt(2) the curve is close to an
// ellipse; larger values give squarer shapes.
func newSuperShape(x0, y0, x1, y1, superness float64) superShape {
	cx, cy := (x0+x1)/2, (y0+y1)/2
	a, b := (x1-x0)/2, (y1-y0)/2
	// The middle of a quarter arc with control points at distance k*a and
	// k*b from the corners lies at (4+3k)/8 times the half axes.
	k := (8*superness - 4) / 3

	var q superShape
	// The four extreme points: right, top, left, bottom.
	q.ex = [4][2]float64{{cx + a, cy}, {cx, cy + b}, {cx - a, cy}, {cx, cy - b}}
	// The direction in which the curve leaves each extreme point when
	// running counter-clockwise, scaled by the half axes.
	q.dir = [4][2]float64{{0, k * b}, {-k * a, 0}, {0, -k * b}, {k * a, 0}}
	return q
}

// segment adds the i-th of the four cubic segments of the curve, for i from 0
// to 3.  Counter-clockwise, segment i runs from extreme i to extreme i+1:
// right to top, top to left, left to bottom, bottom to right.  Clockwise, the
// extremes are visited in reverse order, so segment i runs from extreme
// (4-i)%4 to extreme (3-i)%4: right to bottom, bottom to left, left to top,
// top to right.  The sign reverses the direction of the control point offsets.
func (q superShape) segment(path *outline.Path, i int, ccw bool) {
	path.CurveTo(q.cubic(i, ccw).args())
}

// cubic returns the i-th of the four segments as a curve, in the order and
// the direction which segment describes.
func (q superShape) cubic(i int, ccw bool) cubic {
	var from, to int
	var sign float64
	if ccw {
		from, to, sign = i, (i+1)%4, 1
	} else {
		from, to, sign = (4-i)%4, (3-i+4)%4, -1
	}
	return cubic{
		P0: outline.Point{X: q.ex[from][0], Y: q.ex[from][1]},
		P1: outline.Point{X: q.ex[from][0] + sign*q.dir[from][0], Y: q.ex[from][1] + sign*q.dir[from][1]},
		P2: outline.Point{X: q.ex[to][0] - sign*q.dir[to][0], Y: q.ex[to][1] - sign*q.dir[to][1]},
		P3: outline.Point{X: q.ex[to][0], Y: q.ex[to][1]},
	}
}

// superEllipse adds a closed curve of four cubic segments which touches the
// given rectangle at the midpoints of its sides.  For superness 1/sqrt(2)
// the curve is close to an ellipse; larger values give squarer shapes.
func superEllipse(path *outline.Path, x0, y0, x1, y1, superness float64, ccw bool) {
	q := newSuperShape(x0, y0, x1, y1, superness)
	path.MoveTo(q.ex[0][0], q.ex[0][1])
	for i := 0; i < 4; i++ {
		q.segment(path, i, ccw)
	}
}
