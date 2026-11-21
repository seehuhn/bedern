// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capZ draws the capital Z: a top bar and a bottom bar of the thickness of the bars
// of the capitals joined by a thick diagonal (ZDiag) from the upper right to the lower
// left.  The top is flat on the cap height and the bottom flat on the baseline, and the
// outer corners at the upper right and the lower left are square.  The bars end in the
// vertical serifs of the arms of E, F and T (caparm.go): the upper bar at the left in
// one which hangs down (capArmTopLeft, TopDrop), the lower bar at the right in one
// which rises (capArmBottom, BottomRise).  The width of the ink is ZInk, from the corner
// of the diagonal at the baseline to the end of the lower bar, which lies under the
// corner of the diagonal at the cap height; the upper bar is ZInkTop long.  The contour
// is the one of z:
//
//	the top right corner, along the top to the upper serif, round it, along the
//	underside of the bar to the diagonal, down its left-hand edge to the bottom left
//	corner, along the baseline to the lower serif, round it, along the top of the bar
//	to the diagonal, up its right-hand edge to the top right corner.
func capZ(p design.Params) (outline.Path, float64) {
	d, c := p.CapDiag, p.Caps
	xr := d.SideZ.Left + d.ZInk
	diag := diagFromEdge(edgeRight, xr, p.CapHeight, design.CapZLean(p), d.ZDiag)
	topRight := diag.edgeAt(edgeRight, p.CapHeight)
	bottomLeft := diag.edgeAt(edgeLeft, 0)

	upperArm := capArmSpec(p, diag.edgeAt(edgeLeft, p.CapHeight-c.Bar).X, xr-d.ZInkTop, p.CapHeight)
	upperArm.Drop = c.TopDrop
	lowerArm := capArmSpec(p, diag.edgeAt(edgeRight, c.Bar).X, xr, c.Bar)
	lowerArm.Rise = c.BottomRise
	upper := capArmTopLeft(upperArm)
	lower := capArmBottom(lowerArm)
	ch := join(
		lineChain(topRight, outline.Point{X: xr - d.ZInkTop, Y: p.CapHeight}),
		upper,
		lineChain(upper.end(), bottomLeft),
		lineChain(bottomLeft, outline.Point{X: xr}),
		lower,
		lineChain(lower.end(), topRight),
	)
	return capContour(ch), design.CapZAdvance(p)
}
