// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The arms of E, F, L and T and the crossbar of H.
//
// An arm is a straight bar which leaves a stem and ends in a vertical serif: a
// vertical cut at the end of the arm, which is the end of the ink, and an inner
// edge which leaves the flat end of the serif at a small angle from the
// vertical and turns into the bar in one curve.  The serif is the beak of z
// (zBeak) at the size of the capitals: Height is the length of the cut, the
// thickness of the bar and the depth of the serif beyond it, and Thick its
// width halfway along that length.  An arm at the top of a letter has its serif
// hanging down, an arm on the baseline has it rising, and the middle arm of E
// and F has it both ways, one beak above and one below, which share the cut.
//
// The parts return chains, walked counter-clockwise:
//
//	capArmTop       from the right-hand edge of a stem, along the underside of
//	                the arm, round the serif, up the cut to the top; the arm of
//	                T on the right of the stem, of E and F
//	capArmTopLeft   from the top at the left end of an arm, down the cut, round
//	                the serif and along the underside to the left-hand edge of a
//	                stem; the arm of T on the left of the stem
//	capArmBottom    from the baseline at the end of the arm, up the cut, round
//	                the serif and along the top of the arm to the right-hand edge
//	                of the stem; the bottom arm of E and L
//	capArmMiddle    from the right-hand edge of a stem along the underside of
//	                the arm, round the lower and the upper serif, back along the
//	                top of the arm to the edge; the middle arm of E and F
//
// The crossbar of H is a plain bar between two stems (capCrossbar).

// A capArm describes one arm: the edge of the stem it leaves, the position of
// the cut at its end, the heights of its two edges and the depth of its serif
// beyond each edge (zero for no serif on that side).
type capArm struct {
	Edge, End    float64 // the stem edge and the cut
	Lower, Upper float64 // the heights of the underside and of the upper edge
	Drop, Rise   float64 // how far the serif reaches below Lower and above Upper
	SerifThick   float64 // the width of the serif halfway along each beak
	MinTurn      float64 // the least height of the curve of the inner edge of a beak, or zero
}

// beak returns the beak of z for a serif which reaches depth beyond the far edge
// of a bar of the thickness bar, in the frame of zBeak.
func beak(depth, thick, bar, minTurn float64) chain { return zBeakTurn(bar+depth, thick, bar, minTurn) }

// capArmTop returns an arm to the right of a stem whose serif hangs down from
// its upper edge, at the top of a letter: from (Edge, Lower) to (End, Upper).
func capArmTop(a capArm) chain {
	bar := a.Upper - a.Lower
	// The beak of z is on the left of its bar: mirror it.
	s := beak(a.Drop, a.SerifThick, bar, a.MinTurn).mapPoints(func(q outline.Point) outline.Point {
		return outline.Point{X: a.End - q.X, Y: a.Upper + q.Y}
	}).reversed()
	return join(lineChain(outline.Point{X: a.Edge, Y: a.Lower}, s.start()), s)
}

// capArmTopLeft returns an arm to the left of a stem whose serif hangs down
// from its upper edge: from (End, Upper) to (Edge, Lower).  It is the beak of z
// as it stands.
func capArmTopLeft(a capArm) chain {
	s := beak(a.Drop, a.SerifThick, a.Upper-a.Lower, a.MinTurn).shifted(a.End, a.Upper)
	return join(s, lineChain(s.end(), outline.Point{X: a.Edge, Y: a.Lower}))
}

// capArmBottom returns an arm on the baseline to the right of a stem whose serif
// rises from its upper edge: from (End, Lower) to (Edge, Upper).
func capArmBottom(a capArm) chain {
	bar := a.Upper - a.Lower
	s := beak(a.Rise, a.SerifThick, bar, a.MinTurn).halfTurn(a.End/2, a.Lower/2)
	return join(s, lineChain(s.end(), outline.Point{X: a.Edge, Y: a.Upper}))
}

// capArmMiddle returns an arm to the right of a stem with a serif above and
// below: from (Edge, Lower) to (Edge, Upper).  The lower serif is the serif of a
// top arm and the upper one that of a bottom arm; they share the cut.
func capArmMiddle(a capArm) chain {
	bar := a.Upper - a.Lower
	lower := beak(a.Drop, a.SerifThick, bar, a.MinTurn).mapPoints(func(q outline.Point) outline.Point {
		return outline.Point{X: a.End - q.X, Y: a.Upper + q.Y}
	}).reversed()
	upper := beak(a.Rise, a.SerifThick, bar, a.MinTurn).halfTurn(a.End/2, a.Lower/2)
	lower, upper = lower[:len(lower)-1], upper[1:]
	return join(
		lineChain(outline.Point{X: a.Edge, Y: a.Lower}, lower.start()),
		lower,
		lineChain(lower.end(), upper.start()),
		upper,
		lineChain(upper.end(), outline.Point{X: a.Edge, Y: a.Upper}),
	)
}

// capArmSpec returns the arm of a capital with the bar thickness of the arms,
// whose upper edge is at top, from the stem edge to the cut at end.
func capArmSpec(p design.Params, edge, end, top float64) capArm {
	return capArm{Edge: edge, End: end, Lower: top - p.Caps.Bar, Upper: top, SerifThick: p.Caps.ArmSerifThick}
}

// capCrossbar returns one edge of the crossbar of H at the height y, from the
// stem edge at x0 to the stem edge at x1: the underside from the left stem to
// the right one, and the upper edge walked back.
func capCrossbar(x0, x1, y float64) chain {
	return lineChain(outline.Point{X: x0, Y: y}, outline.Point{X: x1, Y: y})
}

// figBeakMinTurn is the least height of the curve by which the inner edge of the
// beak of 2 and 7 turns into the bar: a shorter curve loses its direction in the
// rounding to whole units and leaves a kink where the straight edge ends.
const figBeakMinTurn = 30.0
