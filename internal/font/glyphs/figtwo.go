// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// figTwo draws the figure 2: a round upper bowl, open on the lower left, whose
// stroke ends at the upper left in the ball terminal and runs from its right
// side down to the lower left as a diagonal, and a heavy base, Figures.Heavy
// thick, which ends on the right in the vertical serif of the arms of E, F and T,
// rising TwoRise above it.
//
// The bowl is the pair of super-ellipses with the squareness of the o of an oval
// from TwoBowlLeft to the ink right, whose top is the round top of the figures
// and whose right extreme, the rightmost ink, is TwoShoulder above the baseline;
// its right side is TwoNeckThick thick and its top Figures.Thin.  The ball sits on
// the upper left quarter as on c, mirrored, with the leftmost ink of the part
// TwoBallLeft from the ink left; the outer edge winds from the top into the ball
// (windSpiral), and the inner edge runs from the tip of the wedge over the top of
// the counter in one arch (archToTip), whose curvature at the top the inner edge
// of the right side takes on as it runs down to the right extreme.  Below the
// right extreme each edge of the stroke is one cubic
// curve into the diagonal, which leaves the extreme with the curvature of the
// bowl: the outer edge runs down to the top of the base, which it meets
// TwoDiagFoot from the ink left at TwoDiagAngle from the vertical, and the inner
// edge to the left end of the base, at the ink left TwoBaseEnd above the
// baseline, which it reaches at the same angle; the neck is curved, the
// diagonal arrives flatter than it leaves the bowl, and the stroke thins from the
// side of the bowl into the diagonal and stands on a short vertical end.
//
// The contour runs from the lower left corner along the baseline to the end of
// the base, round the serif, along the top of the base to the diagonal, up the
// outer edge round the bowl to the ball, round it, back along the inner edge down
// to the left end of the base and down that to the lower left corner.
func figTwo(p design.Params) (outline.Path, float64) {
	f := p.Figures
	x0 := design.FigureLeft(p, 2)
	top := figTop(p)
	right := x0 + f.Ink[2]
	ob := 2*f.TwoShoulder - top // the bottom of the outer oval
	outer := newSuperShape(x0+f.TwoBowlLeft, ob, right, top, p.SupernessOuter)
	inner := newSuperShape(x0+f.TwoBowlLeft+f.TwoNeckThick, ob+f.Thin, right-f.TwoNeckThick, top-f.Thin, p.SupernessInner)

	ball := upperLeftBall("two", figBall(p), outer.cubic(1, true), inner.cubic(2, false), x0+f.TwoBallLeft, figStroke{Arch: true})
	// The inner edge from the top of the counter down to the right extreme
	// takes the curvature of the arch at the top (archToTip) and that of the
	// oval at the extreme.
	last := ball[len(ball)-1].cubic
	innerRight := inner.cubic(3, false)
	innerRight, err := g2Cubic(last.P3, last.tangent(1), last.curvature(1), innerRight.P3, innerRight.tangent(1), innerRight.curvature(1))
	if err != nil || !oneWay(innerRight, -1) {
		panic("glyphs: glyph two: the inner edge of the bowl cannot leave the arch; glyphs.Check reports such parameters")
	}

	a := degrees(f.TwoDiagAngle)
	down := outline.Point{X: -math.Sin(a), Y: -math.Cos(a)} // along the diagonal, down to the left
	foot := outline.Point{X: x0 + f.TwoDiagFoot, Y: f.Heavy}
	rx := outline.Point{X: outer.ex[0][0], Y: outer.ex[0][1]}
	ri := outline.Point{X: inner.ex[0][0], Y: inner.ex[0][1]}
	kOuter := outer.cubic(0, true).curvature(0)
	kInner := inner.cubic(0, false).curvature(0)
	outerDiag := figCurve("two", foot, neg(down), 0, rx, outline.Point{Y: 1}, kOuter)
	innerDiag := fairStartCurve("two", ri, outline.Point{Y: -1}, kInner, outline.Point{X: x0, Y: f.TwoBaseEnd}, down)

	arm := capArm{Edge: foot.X, End: x0 + f.TwoBase, Lower: 0, Upper: f.Heavy, Rise: f.TwoRise, SerifThick: f.TwoSerifThick}
	base := capArmBottom(arm)
	c := join(
		lineChain(outline.Point{X: x0}, base.start()),
		base,
		curveChain(outerDiag),
		chain{curveSeg(outer.cubic(0, true))},
		ball,
		curveChain(innerRight),
		curveChain(innerDiag),
		lineChain(innerDiag.P3, outline.Point{X: x0}),
	)
	return contourPath(mergeLines(c)), f.Advance
}

// fairStartCurve returns the curve from p0, leaving in the direction t0 with the
// curvature k0, to p3, arriving in the direction t3 and straight, or with its
// curvature free if no such curve exists, for the glyph name.
func fairStartCurve(name string, p0, t0 outline.Point, k0 float64, p3, t3 outline.Point) cubic {
	if c, err := g2Cubic(p0, t0, k0, p3, t3, 0); err == nil && oneWay(c, math.Copysign(1, k0)) {
		return c
	}
	c, err := fairStart(p0, t0, k0, p3, t3)
	if err != nil {
		panic("glyphs: glyph " + name + ": " + err.Error())
	}
	return c
}
