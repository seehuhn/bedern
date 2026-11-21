// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The bracket at the foot of the left-hand edge of the stroke of 5: it ends
// fiveBracketDrop above where the edge meets the inner edge of the bowl and runs
// into that edge fiveBracketRun to the right of the stroke.
const (
	fiveBracketRun  = 25
	fiveBracketDrop = 34
)

// figFive draws the figure 5: a heavy flag at the height of the figures,
// Figures.Heavy thick, cut square at its right end; a thin vertical stroke,
// FiveStroke wide, down from its left end; and a bowl open on the left, as the
// lower bowl of 3, which leaves the stroke and ends at the lower left in the
// ball terminal.
//
// The bowl is the pair of super-ellipses with the squareness of the o of an oval
// from FiveOuterLeft to the ink right and from the round bottom to FiveBowlTop,
// with its thick part Figures.Bowl on the right and its top and bottom
// Figures.Thin; the inner edge has its left extreme at FiveInnerLeft.  The stroke
// cuts both edges: its right-hand edge meets the outer edge in a notch, where the
// bowl begins, and its left-hand edge meets the inner edge in a corner.  The ball
// sits on the lower left quarter as on c, turned through half a turn, with its
// leftmost ink FiveBallLeft from the ink left.
//
// The contour runs from the top left corner down the left-hand edge of the
// stroke, along the inner edge of the bowl round its right side and bottom to
// the ball, round it, back along the outer edge to the notch, up the right-hand
// edge of the stroke to the flag, along its underside, up the cut and along the
// top.
func figFive(p design.Params) (outline.Path, float64) {
	f := p.Figures
	x0 := design.FigureLeft(p, 5)
	x1 := x0 + f.Ink[5]
	h, bot := f.Height, figBottom(p)
	o := newSuperShape(x0+f.FiveOuterLeft, bot, x1, f.FiveBowlTop, p.SupernessOuter)
	i := newSuperShape(x0+f.FiveInnerLeft, bot+f.Thin, x1-f.Bowl, f.FiveBowlTop-f.Thin, p.SupernessInner)
	ball := lowerLeftBall("five", figBall(p), o.cubic(2, true), i.cubic(1, false), x0+f.FiveBallLeft, figStroke{})

	sl := x0 + f.FiveStrokeLeft
	sr := sl + f.FiveStroke
	end := sl + f.FiveFlag
	iTop := i.cubic(2, false) // left to top
	// The left-hand edge ends above the inner edge in a short bracket which runs
	// tangentially into the inner edge, fiveBracketRun to the right of the stroke.
	corner := iTop.at(iTop.tAtX(sl))
	fromCorner := iTop.sub(iTop.tAtX(sl+fiveBracketRun), 1)
	kneeY := corner.Y + fiveBracketDrop
	knee := outline.Point{X: sl, Y: kneeY}
	tn := fromCorner.tangent(0)
	tl := math.Hypot(tn.X, tn.Y)
	tn = outline.Point{X: tn.X / tl, Y: tn.Y / tl}
	d := math.Hypot(fromCorner.P0.X-knee.X, fromCorner.P0.Y-knee.Y)
	// The curve leaves the knee downwards and arrives with the direction of the
	// inner edge and a curvature which is up to twice that of the inner edge; of
	// the curves which do, the one with the gentlest bend is taken.
	k0 := fromCorner.curvature(0)
	var bracket cubic
	best := math.Inf(1)
	for a := 0.1 * d; a <= 1.2*d; a += 0.05 * d {
		for b := 0.1 * d; b <= 1.2*d; b += 0.05 * d {
			c := cubic{knee,
				outline.Point{X: knee.X, Y: knee.Y - a},
				outline.Point{X: fromCorner.P0.X - b*tn.X, Y: fromCorner.P0.Y - b*tn.Y},
				fromCorner.P0}
			k := c.curvature(1)
			if k*k0 <= 0 || math.Abs(k) < math.Abs(k0) || math.Abs(k) > 1.9*math.Abs(k0) {
				continue
			}
			if m := math.Abs(c.curvature(0)) + math.Abs(k); m < best {
				best, bracket = m, c
			}
		}
	}
	if math.IsInf(best, 1) {
		panic("glyphs: the bracket of five: no curve")
	}
	oTop := o.cubic(1, true) // top to left
	toNotch := oTop.sub(0, oTop.tAtX(sr))
	c := join(
		lineChain(outline.Point{X: sl, Y: h}, knee),
		splitAtExtrema(bracket),
		curveChain(fromCorner),
		chain{curveSeg(i.cubic(3, false)), curveSeg(i.cubic(0, false))},
		ball,
		chain{curveSeg(o.cubic(3, true)), curveSeg(o.cubic(0, true))},
		curveChain(toNotch),
		lineChain(toNotch.P3, outline.Point{X: sr, Y: h - f.Heavy}),
		lineChain(outline.Point{X: sr, Y: h - f.Heavy}, outline.Point{X: end, Y: h - f.Heavy}),
		lineChain(outline.Point{X: end, Y: h - f.Heavy}, outline.Point{X: end, Y: h}),
		lineChain(outline.Point{X: end, Y: h}, outline.Point{X: sl, Y: h}),
	)
	return contourPath(mergeLines(c)), f.Advance
}
