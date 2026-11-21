// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/outline"
)

// An eShapeSpec holds the constants which shape the lower terminal of e: the
// lengths of the handles of the Bézier curve which is the upper edge of the
// terminal, as fractions of the distance from the tip to the lowest point of
// that edge.
type eShapeSpec struct {
	InnerTipHandle    float64 // upper edge: handle at the tip, along the direction of the stroke
	InnerBottomHandle float64 // upper edge: handle at the lowest point, which is horizontal
}

// eShape is the shape of the terminal of e.
var eShape = eShapeSpec{
	InnerTipHandle:    0.30,
	InnerBottomHandle: 0.38,
}

// lowerTerminal returns the terminal of a stroke which ends below the right
// extreme of a bowl, as e and c end: a chain of the outer edge, the tip and the
// upper edge.  The outer edge is the part of the curve outer, which runs from
// the bottom of the bowl up to the right extreme, as far as the height yTip.
// The tip is a straight cut at right angles to the direction of the stroke
// there, tip long, towards the inside of the bowl.  From its upper corner the
// upper edge of the terminal runs back in one curve, whose handle at the tip
// lies along the direction of the stroke, to the point low, the lowest point of
// the inner edge, where it is horizontal.  The chain starts at the start of
// outer and ends at low.
func lowerTerminal(outer cubic, yTip, tip float64, low outline.Point, sh eShapeSpec) chain {
	tTip := outer.tAtY(yTip)
	tipOuter := outer.at(tTip)
	dir := outer.tangent(tTip)
	tipInner := outline.Point{X: tipOuter.X - tip*dir.Y, Y: tipOuter.Y + tip*dir.X}
	chord := math.Hypot(tipInner.X-low.X, tipInner.Y-low.Y)
	cut := outer.sub(0, tTip)
	return chain{
		curveSeg(cut),
		lineSeg(cut.P3, tipInner),
		curveSeg(cubic{
			P0: tipInner,
			P1: outline.Point{X: tipInner.X - sh.InnerTipHandle*chord*dir.X, Y: tipInner.Y - sh.InnerTipHandle*chord*dir.Y},
			P2: outline.Point{X: low.X + sh.InnerBottomHandle*chord, Y: low.Y},
			P3: low,
		}),
	}
}
