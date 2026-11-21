// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

// The hook of f, the tail of j and the tail of y.
//
// A stem which ends in a hook: both edges of the stem curve over and become the two
// edges of a stalk which ends in the ball terminal (ball.go).  A hookSpec holds the
// edges of the stem and the heights at which they begin to curve, the height of the
// crest and the reach of the terminal; the hook of f opens downwards with the ball
// on the right, the tails of j and y are the same in the frame turned through half
// a turn (for y about the centre line of the thin diagonal, see yHook).

// A hookSpec describes the hook of a stem.
type hookSpec struct {
	XL, XR    float64 // the edges of the stem
	Cell      float64 // the right-hand edge of the cell of the stem, which Reach is measured from
	Apex      float64 // the height of the crest
	Reach     float64 // the distance from Cell to the rightmost ink of the terminal
	OuterJoin float64 // the height at which the left-hand edge begins to curve
	InnerJoin float64 // the height at which the right-hand edge begins to curve
}
