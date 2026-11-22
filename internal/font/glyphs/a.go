// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter a.
//
// The a is the two-storey form: a stem on the right, a bowl on its left in the
// lower half, and above the bowl a thin arc which leaves the stem, runs over to
// the left and curls down into the ball terminal (ball.go).  It is one outer
// contour and one counter, that of the bowl.  The outer contour runs
// counter-clockwise from the lower notch of the bowl on the left-hand edge of the
// stem:
//
//	the tail on the right (aTailChain), up the right-hand edge of the stem,
//	straight as far as aOuterDrop below ShoulderStart
//	the arc (aArcChain): the outer edge from the stem, which it leaves vertically,
//	over the crest at the top of the o to the section of the terminal, the
//	terminal, the mirror image of that of r, and the underside back to the
//	left-hand edge of the stem, which it meets vertically aInnerRaise above
//	ShoulderStartInner
//	down the left-hand edge of the stem to the upper notch of the bowl
//	the outer edge of the bowl (bowlOuter), to the lower notch, where the contour
//	closes on its start
//
// and the counter is the counter of the bowl (bowlCounter), bounded by the
// left-hand edge of the stem between the joins.
//
// The arc.  From the crest to the stem it is the shoulder of n stretched to the
// width and the height of a: both edges are shoulder splines (arch.go), which
// leave the stem vertically and with zero curvature, the outer edge aOuterDrop
// below ShoulderStart and the underside aInnerRaise above ShoulderStartInner
// (below it, since aInnerRaise is negative).  From the crest to the section of
// the terminal each edge is one cubic curve which takes the curvature of the
// spline at the crest and that of the stalk at the section.  The stroke keeps
// the thickness of the stalk, BallStalk, over the top and thickens into the stem
// as the shoulder of n does (aArcChain).  The terminal is turned by aBallAngle
// and placed so that its leftmost ink lies AArcReach inside the free extreme of
// the bowl and its centre aBallDrop below the crest, which lies at the top of
// the o.
//
// The bowl is an oval ABowl.Width wide, with its top at ABowl.Top.  Its free
// side has the thin parts and the overshoot of the o, the squareness
// aSuperness, and the thickest part ABowl.Thick at ABowl.ThickPos of the
// height.  The extremes at top and bottom lie ABowl.Run from the stem.  Both
// edges of the upper stroke fall from the extreme at the top to the stem in one
// convex curve (aBowlUpper, the parabolic connection of d), the counter by 17
// to the upper join and the outer edge by 15 to the upper notch, which lies 34
// above the join.  The lower stroke is one curve on each edge from the extreme
// at the bottom up into the stem (aBowlOuter, aBowlCounter): it bends one way
// only and goes on from the free side at the extreme with its curvature.  The
// two halves of each edge meet at the free extreme with the same curvature,
// aFreeMeet of the way from that of the upper half to that of the lower, and
// the counter leaves the extreme at the top into the connection with the
// curvature of the free side there.
//
// The left side space is that of the bowl: ASideLeft is measured to the free
// extreme of the bowl, and the ball stands AArcReach further in.  The right
// side is the end of the tail, with the space ATailSideRight.
//
// The tail.  The stem has no foot serif: the contour starts at the lower
// notch of the bowl, runs down the left-hand edge of the stem, which is the
// outer edge of the curve of the tail and leaves the notch vertically, round
// the low point (OvershootLower below the baseline) and up to the tip of the
// tail, and back along the upper edge and the inner curve to the right-hand
// edge of the stem, which it meets 75 above the baseline.  It is the foot of
// t (footChain) with the values ATailReach, ATailTipHeight and ATailTipThick
// and the constants of aTailShape, and its low point lies 42 beyond the stem
// (sideTailA).

// aBowlUpper is the shape of the connection of the bowl of a with its stem.  The
// lower half takes its curves for the free side and for the directions at the
// stem, and replaces the connection by one curve on each edge.
var aBowlUpper = bowlShape{
	OuterStemHandle:    1.0,
	OuterExtremeHandle: 1.0,
	InnerStemHandle:    1.0,
	InnerExtremeHandle: 1.0,
	Parabolic:          true,
}

// aBowlOuter returns the outer edge of the bowl of a, see bowlOuter: the upper
// half with the connection aBowlUpper, and the lower half with one curve from
// the extreme at the bottom, which it leaves with the curvature of the free side
// there, to the lower notch, which it reaches at the angle at which the stroke
// meets the stem (leaveAngle): the fairest such curve (fairStart).  The two
// quarters of the free side meet with the curvature aFreeMeet (g2OrKeep).
func aBowlOuter(b bowlSpec) chain {
	up, low := bowlOuter(aBowlUpper, b), bowlOuter(aBowlUpper, aLowerHalf(b))
	free := low[2].cubic
	join := low[3].cubic
	c, err := fairStart(join.P0, outline.Point{X: 1}, free.curvature(1), join.P3, join.tangent(1))
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph a: the bottom of the bowl does not reach the lower notch: %v; glyphs.Check reports such parameters", err))
	}
	k := up[1].curvature(1) + aFreeMeet*(low[2].curvature(0)-up[1].curvature(1))
	upper, lower := g2OrKeep(up[1].cubic, up[1].curvature(0), k), g2OrKeep(low[2].cubic, k, low[2].curvature(1))
	return chain{up[0], curveSeg(upper), curveSeg(lower), curveSeg(c)}
}

// g2OrKeep returns the curve c redrawn with the curvatures k0 at its start and
// k3 at its end, with its points and directions (g2Cubic), or c where there is
// no such curve.
func g2OrKeep(c cubic, k0, k3 float64) cubic {
	if r, err := g2Cubic(c.P0, c.tangent(0), k0, c.P3, c.tangent(1), k3); err == nil {
		return r
	}
	return c
}

// aBowlCounter returns the counter of the bowl of a, see bowlCounter: the upper
// half with the connection aBowlUpper, and below one curve from the lower join,
// which it leaves aJoinTurn more steeply than the outer edge meets the stem, to
// the extreme at the bottom, which it reaches with the curvature of the free
// side there: of the curvatures at the join, the one whose curve changes its
// curvature least abruptly (curvatureRate).  The two quarters of the free side
// meet with the curvature aFreeMeet, and the connection to the upper join
// leaves the extreme at the top with the curvature of the free side there and
// keeps its handle at the join.
func aBowlCounter(b bowlSpec) chain {
	up, low := bowlCounter(aBowlUpper, b), bowlCounter(aBowlUpper, aLowerHalf(b))
	join, free := low[1].cubic, low[2].cubic
	k3 := free.curvature(0)
	tj := join.tangent(0)
	a := math.Atan2(tj.Y, tj.X) + aJoinTurn*math.Pi/180
	dir := outline.Point{X: math.Cos(a), Y: math.Sin(a)}
	c, best := cubic{}, math.Inf(1)
	for i := 1; i <= 100; i++ {
		cv, err := g2Cubic(join.P0, dir, k3*float64(i)/100, join.P3, outline.Point{X: -1}, k3)
		if err != nil || !oneWay(cv, -1) {
			continue
		}
		if e := curvatureRate(cv); e < best {
			c, best = cv, e
		}
	}
	if math.IsInf(best, 1) {
		panic("glyphs: glyph a: the bottom of the counter does not reach the lower join; glyphs.Check reports such parameters")
	}
	k := up[3].curvature(0) + aFreeMeet*(low[2].curvature(1)-up[3].curvature(0))
	free, top := g2OrKeep(low[2].cubic, low[2].curvature(0), k), g2OrKeep(up[3].cubic, k, up[3].curvature(1))
	conn := up[4].cubic
	if cc, err := g2Start(conn.P0, conn.tangent(0), top.curvature(1), conn.P3, conn.tangent(1), math.Hypot(conn.P3.X-conn.P2.X, conn.P3.Y-conn.P2.Y)); err == nil {
		conn = cc
	}
	return chain{low[0], curveSeg(c), curveSeg(free), curveSeg(top), curveSeg(conn)}
}

// aLowerHalf returns the bowl from which the lower half of the bowl of a is
// drawn: the bowl with its extreme at the bottom aLowerRun from the stem.  The
// upper half has its extreme at the top ABowl.Run from the stem.  The two
// halves meet at the free extreme, where both are vertical.
func aLowerHalf(b bowlSpec) bowlSpec {
	b.Half = b.X - b.Left - aLowerRun
	return b
}

// aFreeMeet is where the curvature at which the two halves of each edge of the
// bowl of a meet at the free extreme lies between the curvatures of the
// upper and of the lower half there, as a fraction of the way from the upper.
const aFreeMeet = 0.4

// aJoinTurn is the angle in degrees by which the counter of a leaves the
// stem at the lower join more steeply than the outer edge leaves the notch.
const aJoinTurn = 8.0

// aLowerRun is the distance from the extreme of the bowl of a at the bottom to
// the stem.
const aLowerRun = 105.0

// aTailShape is the shape of the tail of a, the foot of t with its constants
// (tFoot) except for these: the outer edge starts at the lower notch of the
// bowl, a corner, the inner guide meets the right-hand edge of the stem 65
// above the baseline, the low point lies 42 beyond the stem and the inner low
// point 12 to the left of it, the radii are 25 at the low point, 20 at the
// inner low point and 15 at the highest point of the tip.
func aTailShape(bowl bowlSpec) tFootShape {
	sh := tFoot
	sh.OuterStart = bowl.NotchLower
	sh.InnerTop = 65
	sh.LowRight = 42
	sh.LowRadius = 25
	sh.InnerShift = 12
	sh.InnerRadius = 20
	sh.OuterCorner = true
	sh.TopRadius = 15
	return sh
}

// aTailChain returns the tail of a, see footChain: it starts at the lower
// notch of the bowl on the left-hand edge of the stem and ends on the
// right-hand edge.  The left-hand edge of the stem is not straight below the
// notch: it is the outer edge of the curve of the tail, which leaves the notch
// vertically.
func aTailChain(p design.Params, bowl bowlSpec, stem stemSpec) chain {
	return footChain(p, stem.Left, stem.right(), p.ATailReach, p.ATailTipHeight, p.ATailTipThick, aTailShape(bowl))
}

// aBowl returns the bowl of a.  Its free side has the thin parts, the
// overshoot at the bottom and the squareness of the o, and the thickest part
// below the middle.
func aBowl(p design.Params) bowlSpec {
	b := bowlFrom(p, p.ABowl, design.AStemLeft(p), p.ASideLeft)
	b.SupernessOuter, b.SupernessInner = aSuperness, aSuperness
	return b
}

// aSuperness is the squareness of both edges of the free side of the bowl of a.
const aSuperness = 0.70

// aArcChain returns the arc of a with the ball terminal tb.  From the crest
// to the stem it is the shoulder of n with the height and the width of a: the
// outer edge is a shoulder spline (shoulderSpline) from the right-hand edge of
// the stem, which it leaves aOuterDrop below ShoulderStart, to the crest, and
// the underside is one from the left-hand edge of the stem, which it meets
// aInnerRaise above ShoulderStartInner, to its own highest point, BallStalk
// below the crest.  The handles of the splines at the top are the fractions of
// the widths of the shoulders which those of n are (nShoulderHandles).  Left of
// the crest each edge is one cubic curve between its highest point, where it
// runs horizontally with the curvature of the spline, and the section of the
// ball, where it takes the direction and the curvature of the stalk (g2Cubic).
// The crest lies at aCrestX or, where no such curve joins it to the ball, as
// little to the left of it as allows one (aCrestSearch); where none does, the
// curvature at the crest is the greatest for which there is one.
func aArcChain(p design.Params, tb placedBall) chain {
	x, bend := aCrestAt(p, tb)
	c, ok := aArc(p, tb, x, bend)
	if !ok {
		panic("glyphs: glyph a: no edge joins the crest to the ball")
	}
	return c
}

// aCrestAt returns the abscissa of the crest of a with the ball terminal tb,
// and whether the curvature at the top of the splines must be lowered, see
// aArcChain.
func aCrestAt(p design.Params, tb placedBall) (float64, bool) {
	x0 := aCrestX(p)
	for dx := 0.0; dx <= aCrestSearch; dx++ {
		if _, ok := aArc(p, tb, x0-dx, false); ok {
			return x0 - dx, false
		}
	}
	return x0, true
}

// aCrestSearch is how far, in units, the crest of a may move to the left of
// aCrestX.
const aCrestSearch = 40.0

// aArc returns the arc of a with its crest at the abscissa x, see aArcChain,
// and whether the edges of the stalk join the splines with their curvatures.
// If bend is set, a curvature at the top of a spline which no edge of the
// stalk can take is lowered until one can.
func aArc(p design.Params, tb placedBall, x float64, bend bool) (chain, bool) {
	left := design.AStemLeft(p)
	right := left + p.Stem
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	crest := pt(x, p.XHeight+p.OvershootLower)
	inner := pt(crest.X, crest.Y-p.BallStalk)
	so := pt(right, p.ShoulderStart-aOuterDrop)
	si := pt(left, p.ShoulderStartInner+aInnerRaise)
	fo, fi := nShoulderHandles(p)
	ko := shoulderK(so, crest, fo*(so.X-crest.X), nArchShape.ShoulderHigh)
	ki := shoulderK(si, inner, fi*(si.X-inner.X), nArchShape.UnderShoulderHigh)
	outer, ko, ok1 := aStalkEdge(ko, bend, func(k float64) (cubic, error) {
		return g2Cubic(crest, pt(-1, 0), k, tb.In.At, tb.In.Dir, tb.In.Kappa)
	})
	under, ki, ok2 := aStalkEdge(ki, bend, func(k float64) (cubic, error) {
		return g2Cubic(tb.Out.At, tb.Out.Dir, tb.Out.Kappa, inner, pt(1, 0), -k)
	})
	if !ok1 || !ok2 {
		return nil, false
	}
	outerShoulder := shoulderSpline(so, crest, shoulderHandle(so, crest, ko, nArchShape.ShoulderHigh), nArchShape.ShoulderLow, nArchShape.ShoulderHigh)
	innerShoulder := shoulderSpline(si, inner, shoulderHandle(si, inner, ki, nArchShape.UnderShoulderHigh), nArchShape.UnderShoulderLow, nArchShape.UnderShoulderHigh)
	return join(outerShoulder, splitAtExtrema(outer), tb.Chain, splitAtExtrema(under), innerShoulder.reversed()), true
}

// aStalkEdge returns the edge of the stalk of a between the ball and the
// highest point of the edge, drawn by edge for the curvature k at the highest
// point, if it turns one way only.  If bend is set and there is no such curve,
// it lowers k step by step to a tenth of its value and returns the first curve
// it finds, with its curvature.
func aStalkEdge(k float64, bend bool, edge func(k float64) (cubic, error)) (cubic, float64, bool) {
	steps := 1
	if bend {
		steps = 90
	}
	for i := 0; i < steps; i++ {
		kk := k * (1 - 0.01*float64(i))
		c, err := edge(kk)
		if err == nil && oneWay(c, math.Copysign(1, c.curvature(0.5))) {
			return c, kk, true
		}
	}
	return cubic{}, 0, false
}

// nShoulderHandles returns the handles at the top of the shoulder splines of
// the outer edge and of the underside of the arch of n, as fractions of the
// widths of their shoulders.
func nShoulderHandles(p design.Params) (outer, under float64) {
	left := stemSpec{Left: p.SideStemLeft, Width: p.Stem}
	right := stemSpec{Left: left.right() + p.NCounter, Width: p.Stem}
	a := archBetween(p, left, right)
	o, u := archOuter(nArchShape, a), archUnder(nArchShape, a)
	sh := o[1].cubic // the upper curve of the outer shoulder, ending at the peak
	under = u[len(u)-2].cubic.P1.X - u[len(u)-2].cubic.P0.X
	return (sh.P2.X - sh.P3.X) / (a.X3 - a.XPeak), under / (a.X2 - u[0].P3.X)
}

// aCrestX returns the abscissa of the crest of a, the highest point of its top:
// the fraction aCrest of the way from the leftmost ink of the ball to the
// right-hand edge of the stem.
func aCrestX(p design.Params) float64 {
	x0 := p.ASideLeft + p.AArcReach
	return x0 + aCrest*(design.AStemLeft(p)+p.Stem-x0)
}

// aCrest is the place of the crest of a between the leftmost ink and the
// right-hand edge of the stem, as a fraction of that distance.
const aCrest = 0.40

// aInnerRaise is the height by which the underside of the arc of a meets the
// left-hand edge of the stem above ShoulderStartInner, the height at which the
// underside of the arch of n meets its right stem.
const aInnerRaise = -10.0

// aOuterDrop is the distance by which the outer edge of the arc of a leaves the
// right-hand edge of the stem below ShoulderStart, where the outer edge of the
// arch of n leaves its right stem.
const aOuterDrop = 25.0

// aBallDrop is the distance of the centre of the ball of a below the crest.
const aBallDrop = 101.0

// aTerminal returns the ball terminal of a, the mirror image of that of r turned
// by aBallAngle: its leftmost ink AArcReach inside the free extreme of the bowl,
// and its centre aBallDrop below the crest.
func aTerminal(p design.Params) placedBall {
	b := newBallPart(p)
	pl := ballPlace{Angle: aBallAngle, Mirror: true}
	x0, _, _, _ := chainBBox(b.place(pl).Chain)
	pl.Centre = outline.Point{X: p.ASideLeft + p.AArcReach - x0, Y: p.XHeight + p.OvershootLower - aBallDrop}
	return b.place(pl).snapped()
}

// aBallAngle is the angle (radians) by which the mirrored ball terminal of a is
// turned: a little clockwise.
var aBallAngle = -10 * math.Pi / 180

// a draws the lowercase letter a, see the comment above.
func a(p design.Params) (outline.Path, float64) {
	return aOutline(p), design.AAdvance(p)
}

// aOutline draws the letter a.
func aOutline(p design.Params) outline.Path {
	stem := stemSpec{
		Left: design.AStemLeft(p), Width: p.Stem,
	}
	bowl := aBowl(p)
	right := stem.right()

	// The tail: the contour starts at the lower notch of the bowl and runs down
	// the left-hand edge of the stem, round the foot and up the right-hand edge,
	// see aTailChain.
	var path outline.Path
	tail := aTailChain(p, bowl, stem)
	path.MoveTo(tail.start().X, tail.start().Y)
	tail.appendTo(&path)
	path.LineTo(right, p.ShoulderStart-aOuterDrop)

	// The arc: the outer edge from the stem, the ball and the underside, down
	// into the left-hand edge of the stem.
	aArcChain(p, aTerminal(p)).appendTo(&path)

	// The stem, with the bowl in its left-hand edge.
	path.LineTo(stem.Left, bowl.NotchUpper)
	aBowlOuter(bowl).appendTo(&path)
	aBowlCounter(bowl).addContour(&path)
	return path
}
