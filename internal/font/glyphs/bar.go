// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import "seehuhn.de/go/bedern/internal/outline"

// The crossbar of f and t.
//
// The bar is a part of the outer contour of the letter and not a contour of its
// own: it runs out of the right-hand edge of the stem to the right, and out of
// the left-hand edge to the left, with its top at one height.  The top is flat.
// The underside is straight from the stem, where the bar is Thick thick, to the
// end, where it is End thick: with End below Thick the bar thins a little
// towards its ends, and with End equal to Thick it is a plain strip.  The
// ends are cut vertically.  The contour runs counter-clockwise: out along
// the underside of the right-hand part, up its end and back along its top, and
// out along the top of the left-hand part, down its end and back along its
// underside.

// A barSpec describes the crossbar of a stem.
type barSpec struct {
	Top, Thick, End float64 // the height of the top, the thickness at the stem and at the ends
	LeftEdge        float64 // the left-hand edge of the stem
	RightEdge       float64 // the right-hand edge of the stem
	Left, Right     float64 // the distances by which the bar extends beyond the stem on each side
}

// barRight adds the part of the bar to the right of the stem.  The current
// point is on the right-hand edge of the stem at the height of the underside of
// the bar, Top - Thick.  The part ends on the same edge, at the height Top.
func barRight(path *outline.Path, b barSpec) {
	x := b.RightEdge + b.Right
	path.LineTo(x, b.Top-b.End)
	path.LineTo(x, b.Top)
	path.LineTo(b.RightEdge, b.Top)
}

// barLeft adds the part of the bar to the left of the stem.  The current point is
// on the left-hand edge of the stem at the height Top.  The part ends on the
// same edge, at the height of the underside of the bar, Top - Thick.
func barLeft(path *outline.Path, b barSpec) {
	x := b.LeftEdge - b.Left
	path.LineTo(x, b.Top)
	path.LineTo(x, b.Top-b.End)
	path.LineTo(b.LeftEdge, b.Top-b.Thick)
}
