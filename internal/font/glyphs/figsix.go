// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// figSix draws the figure 6: a round bowl at the bottom, closed, and a curved
// stem which rises from its left side, comes over the top and ends at the upper
// right in the ball terminal.  The figure 9 is the 6 turned through half a turn
// (figNine).
//
// The bowl is the pair of super-ellipses with the squareness of the o of an oval
// the full width of the ink, from the round bottom to SixBowlTop, with its thick
// part Figures.Bowl on the right, Figures.Thin at the bottom and SixBowlTopThin at
// the top; its counter has its left extreme SixCounterLeft from the ink left.
// The stem is the left half of an oval SixArcWidth wide whose left extreme lies at
// the ink left SixLeftY above the baseline and whose top is the round top of the
// figures; its inner edge is the oval SixStemThick inside it at the left extreme
// and SixTopThin at the top.  Below the left extreme the outer edge runs in one
// cubic curve, with the curvature of the stem's oval there and that of the bowl
// at its bottom, to the bottom of the bowl; the inner edge of the stem comes down
// into the top of the bowl in a notch.  From the top the stroke runs on to the
// right as a hook.  Its outer edge leaves the stem's oval at the top with the
// curvature there and opens evenly, the curvature falling with the fourth power
// of the length still to run (hookFall) towards that of a circle of
// SixHookRadius, which it would reach after a turn of SixHookAngle, and winds
// from where it falls at sixWindFall into the ball, which is the curled
// terminal of 9 (curlTerminal) with its outer extreme SixBallIn inside
// the ink right.  The inner edge is one arch from the top of the counter to the
// tip of the wedge (archToTip), and the inner edge of the stem runs from
// the stem's oval at its left extreme into the top with the curvature of the
// arch there.  The hook opens and carries the ball out to the right and up,
// with the white between ball and bowl open at 8 pt.
//
// The outer contour runs from the top down the left side, round the bottom and
// the right side of the bowl and over its top to the notch, up the inner edge of
// the stem to the ball, round it and back along the outer edge to the top; the
// counter of the bowl is a second contour, clockwise.
func figSix(p design.Params) (outline.Path, float64) {
	outer, counter := sixChains(p, design.FigureLeft(p, 6))
	return pathOfChains(outer, counter), p.Figures.Advance
}

// sixChains returns the outer contour and the counter of 6 with the ink left at
// x0.
func sixChains(p design.Params, x0 float64) (chain, chain) {
	f := p.Figures
	x1 := x0 + f.Ink[6]
	top, bot := figTop(p), figBottom(p)
	bo := newSuperShape(x0, bot, x1, f.SixBowlTop, p.SupernessOuter)
	bi := [4]float64{x0 + f.SixCounterLeft, bot + f.Thin, x1 - f.Bowl, f.SixBowlTop - f.SixBowlTopThin}
	so := newSuperShape(x0, 2*f.SixLeftY-top, x0+f.SixArcWidth, top, p.SupernessOuter)
	si := newSuperShape(x0+f.SixStemThick, 2*f.SixLeftY-(top-f.SixTopThin), x0+f.SixArcWidth-f.SixTopThin, top-f.SixTopThin, p.SupernessInner)

	// The hook, see above.
	turnTo := f.SixHookAngle * math.Pi / 180
	outTop := so.cubic(1, true)
	ho := openingCurve("six", outTop.P0, outTop.curvature(0), 1/f.SixHookRadius, turnTo).reversed()
	hi := si.cubic(3, false) // of which the arch takes the top and the direction there
	centre := outline.Point{X: math.Round(x1 - f.SixBallIn - f.Ball/2), Y: math.Round(top - f.CurlHeight)}
	ball := curlTerminal("six", p, ho, hi, centre, figStroke{Arch: true, Hook: true, Fall: sixWindFall})

	upLeft := so.cubic(1, true) // top to the left extreme
	k1 := upLeft.curvature(1)
	bBottom := bo.cubic(3, true) // bottom to right
	k2 := bBottom.curvature(0)
	lowLeft := figCurve("six", upLeft.P3, outline.Point{Y: -1}, k1, bBottom.P0, outline.Point{X: 1}, k2)

	bTop := bo.cubic(1, true) // top to left
	// The inner edge of the stem, from its left extreme to the top, takes the
	// curvature of the oval at the extreme and that of the arch at the top
	// (archToTip).
	sUp := si.cubic(2, false) // left to top
	first := ball[0].cubic
	sUp, err := g2Cubic(sUp.P0, sUp.tangent(0), sUp.curvature(0), first.P0, first.tangent(0), first.curvature(0))
	if err != nil || !oneWay(sUp, -1) {
		panic("glyphs: glyph six: the inner edge of the stem cannot run into the arch; glyphs.Check reports such parameters")
	}
	ta, tb := crossAt("six", bTop, sUp)
	toNotch, fromNotch := bTop.sub(0, ta), sUp.sub(tb, 1)
	fromNotch.P0 = toNotch.P3

	outer := join(
		chain{curveSeg(upLeft)},
		curveChain(lowLeft),
		chain{curveSeg(bBottom), curveSeg(bo.cubic(0, true))},
		curveChain(toNotch),
		curveChain(fromNotch),
		ball,
	)
	return outer, ellipseChain(bi[0], bi[1], bi[2], bi[3], p.SupernessInner, false)
}

// sixWindFall is the angle below the horizontal (degrees) at which the outer
// edge of the hook of 6 stops opening and winds into the ball (windSpiral):
// the spiral leaves it with a curvature which rises as the square of the
// length run.
const sixWindFall = 30.0

// hookFall is the power of the length still to run by which the curvature of
// the hook of 6 falls (openingCurve).
const hookFall = 4.0

// openingCurve returns the curve which leaves p0 to the right (in the direction
// of the positive x axis), turns clockwise through the angle turn (radians), and
// whose curvature falls from k0 to k1 (both given positive) with the power
// hookFall of the length still to run: the end point and its direction are
// those of that spiral, found by summing small steps, and the curve is the
// cubic between them with the two curvatures (figCurve).
func openingCurve(name string, p0 outline.Point, k0, k1, turn float64) cubic {
	length := turn / (k1 + (k0-k1)/(hookFall+1))
	const steps = 400
	ds := length / steps
	q, phi := p0, 0.0
	for i := 0; i < steps; i++ {
		s := (float64(i) + 0.5) * ds
		k := k1 + (k0-k1)*math.Pow(1-s/length, hookFall)
		mid := phi - k*ds/2
		q.X += ds * math.Cos(mid)
		q.Y += ds * math.Sin(mid)
		phi -= k * ds
	}
	return figCurve(name, p0, outline.Point{X: 1}, -k0, q, outline.Point{X: math.Cos(phi), Y: math.Sin(phi)}, -k1)
}

// figNine draws the figure 9 as the 6 turned through half a turn about the
// middle of its ink.
func figNine(p design.Params) (outline.Path, float64) {
	f := p.Figures
	x0 := design.FigureLeft(p, 9)
	outer, counter := sixChains(p, x0)
	cx, cy := x0+f.Ink[6]/2, (figTop(p)+figBottom(p))/2
	return pathOfChains(outer.halfTurn(cx, cy), counter.halfTurn(cx, cy)), f.Advance
}
