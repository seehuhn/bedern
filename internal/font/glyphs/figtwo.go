// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"errors"
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
// The left half of the bowl is that of the pair of super-ellipses with the
// squareness of the o of an oval from TwoBowlLeft to the ink right, whose top is
// the round top of the figures and whose centre is TwoBowlMid above the
// baseline; its top is Figures.Thin thick.  The ball sits on the upper left
// quarter as on c, mirrored, with the leftmost ink of the part TwoBallLeft from
// the ink left; the outer edge winds from the top into the ball (windSpiral),
// and the inner edge runs from the tip of the wedge over the top of the counter
// in one arch (archToTip).  The right extreme of the outer edge is at the ink
// right, TwoOuterDrop below TwoShoulder above the baseline; that of the inner
// edge is TwoNeckThick to its left, TwoShoulder above the baseline.  The outer edge runs from the top to its right extreme in one curve
// with the curvature of the other side at the top and the handle of the oval,
// for the height of this quarter, at the extreme (superHandle); the inner edge
// runs from the top of the counter to its right extreme in one curve whose
// handles are the same fraction of the width and of the height of the quarter,
// with the curvature of the arch at the top (quarterWithCurvature).  Below the
// right extreme each edge of the stroke is one cubic curve into the diagonal,
// which leaves the extreme with the curvature of the edge above: the outer edge runs down to the top of the base, which it meets
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
	ob := 2*f.TwoBowlMid - top // the bottom of the outer oval
	outer := newSuperShape(x0+f.TwoBowlLeft, ob, right, top, p.SupernessOuter)
	inner := newSuperShape(x0+f.TwoBowlLeft+f.TwoNeckThick, ob+f.Thin, right-f.TwoNeckThick, top-f.Thin, p.SupernessInner)

	ball := upperLeftBall("two", figBall(p), outer.cubic(1, true), inner.cubic(2, false), x0+f.TwoBallLeft, figStroke{Arch: true})
	rx := outline.Point{X: right, Y: f.TwoShoulder - f.TwoOuterDrop}
	ri := outline.Point{X: right - f.TwoNeckThick, Y: f.TwoShoulder}

	// The outer edge from the right extreme up to the top: the handle at the
	// extreme is that of the oval for this height, and the curvature at the
	// top is that of the edge on the other side.
	ot := outline.Point{X: outer.ex[1][0], Y: outer.ex[1][1]}
	kTop := ball[0].cubic.curvature(0)
	outerRight, err := g2End(rx, outline.Point{Y: 1}, superHandle(p.SupernessOuter)*(ot.Y-rx.Y), ot, outline.Point{X: -1}, kTop)
	if err != nil || !oneWay(outerRight, 1) {
		panic("glyphs: glyph two: the outer edge of the bowl cannot reach the top; glyphs.Check reports such parameters")
	}

	// The inner edge from the top of the counter down to the right extreme:
	// its handles are the same fraction of the width and of the height of the
	// quarter, such that its curvature at the top is that of the arch.
	last := ball[len(ball)-1].cubic
	innerRight, err := quarterWithCurvature(last.P3, ri, last.curvature(1))
	if err != nil || !oneWay(innerRight, -1) {
		panic("glyphs: glyph two: the inner edge of the bowl cannot leave the arch; glyphs.Check reports such parameters")
	}

	a := degrees(f.TwoDiagAngle)
	down := outline.Point{X: -math.Sin(a), Y: -math.Cos(a)} // along the diagonal, down to the left
	foot := outline.Point{X: x0 + f.TwoDiagFoot, Y: f.Heavy}
	kOuter := outerRight.curvature(0)
	kInner := innerRight.curvature(1)
	outerDiag := figCurve("two", foot, neg(down), 0, rx, outline.Point{Y: 1}, kOuter)
	innerDiag := fairStartCurve("two", ri, outline.Point{Y: -1}, kInner, outline.Point{X: x0, Y: f.TwoBaseEnd}, down)

	arm := capArm{Edge: foot.X, End: x0 + f.TwoBase, Lower: 0, Upper: f.Heavy, Rise: f.TwoRise, SerifThick: f.TwoSerifThick, MinTurn: figBeakMinTurn}
	base := capArmBottom(arm)
	c := join(
		lineChain(outline.Point{X: x0}, base.start()),
		base,
		curveChain(outerDiag),
		curveChain(outerRight),
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

// superHandle returns the length of the handles of a quarter of the
// super-ellipse with the superness s, as a fraction of the half axis along
// which each leaves its extreme.
func superHandle(s float64) float64 { return (8*s - 4) / 3 }

// quarterWithCurvature returns the quarter from q0, leaving horizontally, to
// q3, arriving vertically, whose handles are the same fraction l of the
// horizontal and of the vertical distance between them, such that its curvature
// at q0 is k: for the width w and the height h, |k| = 2/3 h (1-l) / (l w)^2.
func quarterWithCurvature(q0, q3 outline.Point, k float64) (cubic, error) {
	w, h := q3.X-q0.X, q3.Y-q0.Y
	aw, ah, ak := math.Abs(w), math.Abs(h), math.Abs(k)
	// ak w^2 l^2 + 2/3 h l - 2/3 h = 0
	qa, qb := ak*aw*aw, 2.0/3*ah
	l := (-qb + math.Sqrt(qb*qb+4*qa*qb)) / (2 * qa)
	if !(l > 0 && l < 1) {
		return cubic{}, errors.New("no quarter with this curvature")
	}
	return cubic{
		P0: q0,
		P1: outline.Point{X: q0.X + l*w, Y: q0.Y},
		P2: outline.Point{X: q3.X, Y: q3.Y - l*h},
		P3: q3,
	}, nil
}
