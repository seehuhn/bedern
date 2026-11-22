// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// figThree draws the figure 3: two bowls open on the left, the upper one
// narrower, which end in the ball terminal at the upper left and at the lower
// left and meet in a short middle stroke a little above the middle of the
// figure.
//
// Each bowl is the pair of super-ellipses with the squareness of the o.  The
// lower one reaches from ThreeLowerLeft to the ink right, from the round bottom up to
// ThreeOverlap above the waist (ThreeWaist); the upper one is ThreeUpperWidth
// wide with its right extreme ThreeUpperRight from the ink left, from
// ThreeOverlap below the waist to the round top.  Their thick parts, on the
// right, are Figures.Bowl and ThreeUpperBowl, their tops and bottoms
// Figures.Thin, and the inner edges lie ThreeOpen inside the outer ones on the
// open side.  The counters reach ThreeGap/2 above and below the waist; the
// middle stroke between them is ThreeGap thick and the outer edges of the
// bowls cross on the right in a notch.  The middle ends on the left in a
// vertical cut ThreeMiddle from the ink left.  The balls sit on the upper left
// and the lower left quarters as on c, with their leftmost ink ThreeUpperBall
// and ThreeLowerBall from the ink left.
//
// The contour runs from the bottom of the lower bowl round its right side to the
// notch, round the right side and the top of the upper bowl to its ball, round
// it, back along the inner edge of the upper bowl to the cut of the middle, down
// the cut, and along the inner edge of the lower bowl to its ball, round it, and
// back to the bottom.
func figThree(p design.Params) (outline.Path, float64) {
	f := p.Figures
	x0 := design.FigureLeft(p, 3)
	x1 := x0 + f.Ink[3]
	top, bot := figTop(p), figBottom(p)
	w, g, ov := f.ThreeWaist, f.ThreeGap, f.ThreeOverlap

	lo := newSuperShape(x0+f.ThreeLowerLeft, bot, x1, w+ov, p.SupernessOuter)
	li := newSuperShape(x0+f.ThreeLowerLeft+f.ThreeOpen, bot+f.Thin, x1-f.Bowl, w-g/2, p.SupernessInner)
	ur := x0 + f.ThreeUpperRight
	ul := ur - f.ThreeUpperWidth
	uo := newSuperShape(ul, w-ov, ur, top, p.SupernessOuter)
	ui := newSuperShape(ul+f.ThreeOpen, w+g/2, ur-f.ThreeUpperBowl, top-f.Thin, p.SupernessInner)

	b := figBall(p)
	upper := upperLeftBall("three", b, uo.cubic(1, true), ui.cubic(2, false), x0+f.ThreeUpperBall, figStroke{})
	lower := lowerLeftBall("three", b, lo.cubic(2, true), li.cubic(1, false), x0+f.ThreeLowerBall, figStroke{Rise: true, Before: li.cubic(0, false)})

	// The notch on the right, where the outer edges cross.
	loR, uoR := lo.cubic(0, true), uo.cubic(3, true) // right to top, bottom to right
	ta, tb := crossAt("three", loR, uoR)
	// The cut of the middle.
	xm := x0 + f.ThreeMiddle
	uiL, liL := ui.cubic(1, false), li.cubic(2, false) // bottom to left, left to top
	toCut, fromCut := uiL.sub(0, uiL.tAtX(xm)), liL.sub(liL.tAtX(xm), 1)
	toNotch, fromNotch := loR.sub(0, ta), uoR.sub(tb, 1)
	fromNotch.P0 = toNotch.P3 // the two curves are made to meet at the crossing

	c := join(
		chain{curveSeg(lo.cubic(3, true))},
		curveChain(toNotch),
		curveChain(fromNotch),
		chain{curveSeg(uo.cubic(0, true))},
		upper,
		chain{curveSeg(ui.cubic(3, false)), curveSeg(ui.cubic(0, false))},
		curveChain(toCut),
		lineChain(toCut.P3, fromCut.P0),
		curveChain(fromCut),
		chain{curveSeg(li.cubic(3, false))},
		lower,
	)
	return contourPath(c), f.Advance
}
