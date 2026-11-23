// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// figSeven draws the figure 7: a heavy bar, Figures.Heavy thick, at the height of
// the figures across the whole ink, whose left end carries the vertical serif of
// the arms of E, F and T hanging SevenDrop below it, and a diagonal which leaves
// the bar at its right end and comes down to the baseline, where it ends in a
// flat cut SevenFoot wide with its middle SevenFootX from the ink left.  Each
// edge of the diagonal is one cubic curve: it leans SevenTopAngle from the
// vertical at the bar and SevenFootAngle at the foot; it is SevenTopThick wide
// across under the bar and the foot is wider.  The right
// edge meets the top of the bar in a sharp corner at the ink right, and the left
// edge meets the underside of the bar in a square-cut inner corner.
//
// The contour runs from the left end of the foot along the baseline, up the right
// edge of the diagonal to the top right corner, along the top to the left end,
// round the serif, along the underside of the bar to the diagonal and down its
// left edge to the foot.
func figSeven(p design.Params) (outline.Path, float64) {
	f := p.Figures
	x0 := design.FigureLeft(p, 7)
	x1 := x0 + f.Ink[7]
	h := f.Height
	lower := h - f.Heavy
	top, foot := degrees(f.SevenTopAngle), degrees(f.SevenFootAngle)
	dirUp := func(a float64) outline.Point { return outline.Point{X: math.Sin(a), Y: math.Cos(a)} }

	fl := outline.Point{X: x0 + f.SevenFootX - f.SevenFoot/2}
	fr := outline.Point{X: x0 + f.SevenFootX + f.SevenFoot/2}
	tr := outline.Point{X: x1, Y: h}
	// The right edge at the underside of the bar, and the left edge there.
	rightEdge := quarter(fr, tr, dirUp(foot), dirUp(top), 0.36)
	xr := rightEdge.at(rightEdge.tAtY(lower)).X
	tl := outline.Point{X: xr - f.SevenTopThick, Y: lower}
	leftEdge := quarter(tl, fl, neg(dirUp(top)), neg(dirUp(foot)), 0.36)

	arm := capArm{Edge: tl.X, End: x0, Lower: lower, Upper: h, Drop: f.SevenDrop, SerifThick: f.SevenSerifThick, MinTurn: figBeakMinTurn}
	serif := capArmTopLeft(arm)
	c := join(
		lineChain(fl, fr),
		curveChain(rightEdge),
		lineChain(tr, serif.start()),
		serif,
		curveChain(leftEdge),
	)
	return contourPath(mergeLines(c)), f.Advance
}
