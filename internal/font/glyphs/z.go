// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter z.
//
// The z is a top bar and a bottom bar, both of the hairline thickness, joined by
// a thick diagonal from the upper right to the lower left.  The top is flat at
// the x-height and the bottom flat on the baseline, without overshoot.  The width
// of the ink is the parameter ZInk, from the corner of the diagonal at the baseline
// (the left end of the ink) to the end of the lower bar, which lies under the
// corner of the diagonal at the x-height; the lean of the diagonal follows
// (design.ZLean).  The upper bar starts ZInk - ZInkTop to the right of the left
// end.  The corners where the diagonal meets the bars are sharp: the outer corners
// at the top right and the bottom left, and the inner ones under the upper bar and
// on the lower bar.
//
// The bars end in beaks, like those of s: the upper bar at the left in a
// beak which hangs down, the lower bar at the right in one which rises.  The lower
// beak is the upper one turned through half a turn.  In the frame of the upper
// beak, with the origin at the top of the bar over the cut, a beak is one chain:
//
//	along the top of the bar to the left, round the square outer corner, down the
//	cut, which is vertical, to the end of the beak, across the flat end, and up the
//	inner edge, which leaves the end at the angle design.ZBeakTilt from the vertical and
//	turns into the underside of the bar, running east, in one curve.
//
// The beak is Height high from the top of the bar to its end, and Thick wide halfway
// along its height; the width of the flat end follows from that (zBeakEnd).  The
// contour is one counter-clockwise outline:
//
//	the top right corner, along the top to the upper beak, round it, along the
//	underside of the bar to the diagonal, down its left-hand edge to the bottom left
//	corner, along the baseline to the lower beak, round it, along the top of the
//	bar to the diagonal, up its right-hand edge to the top right corner.
//
// Side kinds: the left side is measured to the bottom left corner and the right
// side to the end of the lower bar, which is the right end of the ink.

// zBeak returns the upper beak in its frame, in which the origin is the top of the
// bar over the cut and y points up.  The chain starts at the origin, which is the
// square outer corner, and ends on the underside of the bar, which is bar below the
// top, design.ZBeakRun to the right of the inner edge at half the height.  The inner
// edge leaves the flat end at the angle design.ZBeakTilt from the vertical and runs
// straight to half the height of the beak, where it turns into the underside of the
// bar in one curve.  The beak is thick wide at half the height.
func zBeak(height, thick, bar float64) chain { return zBeakTurn(height, thick, bar, 0) }

// zBeakTurn is zBeak with the inner edge turning into the bar by a curve of at least
// minTurn in height.
func zBeakTurn(height, thick, bar, minTurn float64) chain {
	tilt := math.Tan(design.ZBeakTilt)
	end := thick - height/2*tilt
	// The straight inner edge ends at half the height, or, if that leaves a curve
	// shorter than minTurn above the underside of the bar, that much above it.
	half := height / 2
	if minTurn > 0 {
		half = math.Min(math.Max(half, bar+minTurn), 0.9*height)
	}
	s1 := outline.Point{X: 0, Y: 0}
	s2 := outline.Point{X: 0, Y: -height}
	s3 := outline.Point{X: end, Y: -height}
	x4 := thick
	if half != height/2 {
		x4 = end + (height-half)*tilt
	}
	s4 := outline.Point{X: x4, Y: -half}
	s5 := outline.Point{X: thick + design.ZBeakRun, Y: -bar}
	return chain{
		lineSeg(s1, s2),
		lineSeg(s2, s3),
		lineSeg(s3, s4),
		curveSeg(quarter(s4, s5, outline.Point{X: math.Sin(design.ZBeakTilt), Y: math.Cos(design.ZBeakTilt)}, outline.Point{X: 1}, sShape.Quarter)),
	}
}

// z draws the lowercase letter z, see the comment above.
func z(p design.Params) (outline.Path, float64) {
	l := design.ZLean(p)
	xr := p.ZSideLeft + p.ZInk // the corner at the top and the end of the lower bar
	d := diagFromEdge(edgeRight, xr, p.XHeight, l, p.ZDiag)
	topRight := d.edgeAt(edgeRight, p.XHeight)
	under := d.edgeAt(edgeLeft, p.XHeight-p.Hair)
	bottomLeft := d.edgeAt(edgeLeft, 0)
	barTop := d.edgeAt(edgeRight, p.Hair)

	upper := zBeak(p.ZUpperBeakHeight, p.ZUpperBeakThick, p.Hair).
		mapPoints(func(q outline.Point) outline.Point {
			return outline.Point{X: xr - p.ZInkTop + q.X, Y: p.XHeight + q.Y}
		})
	lower := zBeak(p.ZLowerBeakHeight, p.ZLowerBeakThick, p.Hair).
		mapPoints(func(q outline.Point) outline.Point {
			return outline.Point{X: xr - q.X, Y: -q.Y}
		})
	c := join(
		lineChain(topRight, upper.start()),
		upper,
		lineChain(upper.end(), under),
		lineChain(under, bottomLeft),
		lineChain(bottomLeft, lower.start()),
		lower,
		lineChain(lower.end(), barTop),
		lineChain(barTop, topRight),
	)
	return contourPath(c), design.ZAdvance(p)
}
