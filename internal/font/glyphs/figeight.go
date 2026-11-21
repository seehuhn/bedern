// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// figEight draws the figure 8: two closed bowls, the upper one smaller, one above
// the other on one vertical.
//
// Each bowl is a super-ellipse with the squareness of the o; the outer contour of
// the figure is the outline of the two ovals together, which cross on each side
// of the waist in a notch.  The lower oval is the full width of the ink, from the
// round bottom up to EightWaistThin above its counter, and the upper one is
// EightUpperWidth wide, from EightWaistThin below its counter to the round top.
// The counters are EightGap apart, centred on EightWaist, and Figures.Thin from
// the round top and bottom.  The sides of the bowls are EightUL, EightUR, EightLL
// and EightLR thick, heavier on the upper left and the lower right: the
// heavy strokes run from the upper left through the waist to the lower right as
// one S-shaped stroke, and the light ones the other way.
//
// The outer contour runs from the bottom round the right side of the lower bowl
// to the right notch, round the upper bowl to the left notch and round the left
// side of the lower bowl back to the bottom; the counters are two contours,
// clockwise.
func figEight(p design.Params) (outline.Path, float64) {
	f := p.Figures
	x0 := design.FigureLeft(p, 8)
	x1 := x0 + f.Ink[8]
	top, bot := figTop(p), figBottom(p)
	lTop := f.EightWaist - f.EightGap/2 // the top of the lower counter
	uBot := f.EightWaist + f.EightGap/2 // the bottom of the upper counter
	ul := x0 + (f.Ink[8]-f.EightUpperWidth)/2
	ur := ul + f.EightUpperWidth
	lo := newSuperShape(x0, bot, x1, lTop+f.EightWaistThin, p.SupernessOuter)
	uo := newSuperShape(ul, uBot-f.EightWaistThin, ur, top, p.SupernessOuter)

	// The notches, where the outer edges of the ovals cross.
	loR, uoR := lo.cubic(0, true), uo.cubic(3, true) // right to top, bottom to right
	loL, uoL := lo.cubic(1, true), uo.cubic(2, true) // top to left, left to bottom
	ra, rb := crossAt("eight", loR, uoR)
	la, lb := crossAt("eight", uoL, loL)
	toR, fromR := loR.sub(0, ra), uoR.sub(rb, 1)
	fromR.P0 = toR.P3
	toL, fromL := uoL.sub(0, la), loL.sub(lb, 1)
	fromL.P0 = toL.P3
	outer := join(
		chain{curveSeg(lo.cubic(3, true))},
		curveChain(toR), curveChain(fromR),
		chain{curveSeg(uo.cubic(0, true)), curveSeg(uo.cubic(1, true))},
		curveChain(toL), curveChain(fromL),
		chain{curveSeg(lo.cubic(2, true))},
	)
	upper := ellipseChain(ul+f.EightUL, uBot, ur-f.EightUR, top-f.Thin, p.SupernessInner, false)
	lower := ellipseChain(x0+f.EightLL, bot+f.Thin, x1-f.EightLR, lTop, p.SupernessInner, false)
	return pathOfChains(outer, upper, lower), f.Advance
}
