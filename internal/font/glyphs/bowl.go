// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The bowl.
//
// A bowl hangs on the left-hand edge of a stem, as in d and a.  Its free side,
// to the left of the vertical through the extremes at top and bottom, is the
// left half of the o: the same four extreme points and the same curve code
// (superShape), with the same thickest part, thin parts and squareness.
// Between the extremes at top and bottom and the stem the bowl has a part of
// its own, the connection:
//
//	outer edge   leaves the extreme at the top horizontally and falls to the
//	             stem edge, which it meets at the upper notch at an angle,
//	             running steeply; the lower half mirrors it
//	counter      leaves its extreme at the top horizontally and falls to the
//	             stem edge, which it meets lower down, at the upper join, at
//	             the same angle; the lower half mirrors it
//
// At the extremes the stroke is as thin as the thin part; towards the stem
// the vertical gap between the two edges grows, with the notch above
// the join, and the stroke runs into the stem at the angle whose cosine is
// JoinThick divided by the gap, which makes the thickness at right angles to
// the stroke equal to JoinThick where the stroke meets the stem.  This is the
// construction of the arch of n at its notch and join, applied above and
// below.
//
// The outer contour of the letter is the stem with the bowl inserted in its
// left-hand edge: the caller draws the stem up its right-hand edge and along
// its top, goes down the left-hand edge to the upper notch, calls bowlOuter,
// which runs from the upper notch round the free side to the lower notch,
// and goes down the left-hand edge to the foot.  The edge between the notches
// is not part of the outer contour.
//
// The counter is a second, clockwise contour, added by bowlCounter after the
// outer contour is closed.  The stem edge bounds it on the right: the
// contour starts at the upper join, runs straight down the stem edge to the
// lower join, round the connection and the free side and back to the upper
// join, where it closes.  The counter has two corners, one at each
// join, as the notch and the join of n have.
//
// Parameters and constants.  The extent and the heights of the bowl are the
// fields of a bowlSpec, which the letter fills from the design parameters;
// the handle lengths of the connection are the fields of a bowlShape, which
// only shape this curve.

// A bowlShape holds the constants which shape the connection between the
// extremes of a bowl and the stem: the lengths of the handles of its Bézier
// curves as fractions of the horizontal distance from the extreme to the
// stem.  The outer edge and the counter have constants of their own.  A handle
// at the stem is shortened where it would carry the curve beyond the extreme.
type bowlShape struct {
	OuterStemHandle    float64 // outer edge: handle at the notch
	OuterExtremeHandle float64 // outer edge: handle at the extreme at top and bottom
	InnerStemHandle    float64 // counter: handle at the join
	InnerExtremeHandle float64 // counter: handle at the extreme at top and bottom
	// Parabolic selects the construction of the connection of b, d, p and q: each
	// handle is the given fraction of the handle of the parabola with the same
	// end points and end directions, instead of a fraction of the run.
	Parabolic bool
}

// joinCurve returns the connection from the stem to an extreme: the cubic from
// (x0, y0) on the stem edge, which it leaves at the angle with cosine c and sine s
// to the horizontal towards the extreme (xt, y1) at the left, where it arrives
// horizontally.  Its handles are the fractions fs and fe of those of the
// parabola with the same ends and tangents, whose control point is where the
// ray from the stem meets the horizontal through the extreme.
func joinCurve(x0, y0, c, s, xt, y1, fs, fe float64) cubic {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	run := x0 - xt
	rise := math.Abs(y1 - y0)
	qx := xt + 0.1*run
	if s > 1e-9 {
		qx = math.Max(qx, x0-rise*c/s)
	}

	d := math.Hypot(x0-qx, rise)
	hs := fs * 2 / 3 * d
	he := fe * 2 / 3 * (qx - xt)
	return cubic{pt(x0, y0), pt(x0+hs*(qx-x0)/d, y0+hs*(y1-y0)/d), pt(xt+he, y1), pt(xt, y1)}
}

// A bowlSpec describes a bowl on the left of a stem, or on the right of it if
// OnRight is set: the mirror image of the bowl on the left in the edge of the
// stem.  The heights are measured from the baseline.  The
// extremes at top and bottom lie at the height Top and Bottom, which include
// the overshoot.  The horizontal values are real positions: a bowl on
// the right has X the right-hand edge of the stem and Left, the extreme of the
// free side, to the right of it.  The comments below which speak of the left
// and the right read the other way round for it.
type bowlSpec struct {
	OnRight        bool    // the bowl stands on the right of the stem
	X              float64 // the edge of the stem which the bowl hangs on: the left-hand edge, or the right-hand edge with OnRight
	Left           float64 // the extreme of the free side (the leftmost point, or the rightmost with OnRight)
	Half           float64 // distance from Left to the vertical through the extremes at top and bottom
	Bottom, Top    float64 // outer extremes, with overshoot
	ThinBottom     float64 // thickness at the bottom extreme
	ThinTop        float64 // thickness at the top extreme
	Thick          float64 // thickness at the free extreme
	NotchUpper     float64 // height at which the outer edge of the upper stroke meets the stem
	JoinUpper      float64 // height at which the counter edge of the upper stroke meets the stem
	NotchLower     float64 // height at which the outer edge of the lower stroke meets the stem
	JoinLower      float64 // height at which the counter edge of the lower stroke meets the stem
	JoinThick      float64 // thickness at right angles to the stroke where it meets the stem
	SupernessOuter float64 // squareness of the outer contour on the free side
	SupernessInner float64 // squareness of the counter on the free side
	ThickPos       float64 // height of the free extreme as a fraction of the height of the bowl; the middle (0.5) if zero
}

// leftForm returns the bowl on the left of its stem from which the bowl on the
// right is made: the mirror image in the vertical line x = 0.  A bowl on the left is
// returned unchanged.
func (b bowlSpec) leftForm() bowlSpec {
	if b.OnRight {
		b.OnRight = false
		b.X, b.Left = -b.X, -b.Left
	}
	return b
}

// xExtreme returns the position of the extremes at top and bottom.
func (b bowlSpec) xExtreme() float64 {
	if b.OnRight {
		return b.Left - b.Half
	}
	return b.Left + b.Half
}

// thickY returns the height of the free extreme, where the stroke is thickest.
func (b bowlSpec) thickY() float64 {
	if b.ThickPos == 0 {
		return (b.Bottom + b.Top) / 2
	}
	return b.Bottom + b.ThickPos*(b.Top-b.Bottom)
}

// centred reports whether the free extreme lies at the middle of the bowl.
func (b bowlSpec) centred() bool { return b.ThickPos == 0 || b.ThickPos == 0.5 }

// outerShapes and innerShapes return the super-ellipses of which the free side
// is the left half: the one which gives the upper quarter, from the extreme at
// the top to the free extreme, and the one which gives the lower quarter.  The
// outer one is the rectangle of the bowl as a whole o with the same width.  If
// the free extreme lies at the middle of the bowl the two are one shape, that of
// the o.  Otherwise each is the shape of a bowl whose free extreme lies at the
// height of the thickest part, and the halves differ in height.
func (b bowlSpec) outerShapes() (up, down superShape) {
	right := b.Left + 2*b.Half
	if b.centred() {
		q := newSuperShape(b.Left, b.Bottom, right, b.Top, b.SupernessOuter)
		return q, q
	}
	y := b.thickY()
	return newSuperShape(b.Left, y-(b.Top-y), right, b.Top, b.SupernessOuter),
		newSuperShape(b.Left, b.Bottom, right, y+(y-b.Bottom), b.SupernessOuter)
}

func (b bowlSpec) innerShapes() (up, down superShape) {
	left, right := b.Left+b.Thick, b.Left+2*b.Half-b.Thick
	bottom, top := b.Bottom+b.ThinBottom, b.Top-b.ThinTop
	if b.centred() {
		q := newSuperShape(left, bottom, right, top, b.SupernessInner)
		return q, q
	}
	y := b.thickY()
	return newSuperShape(left, y-(top-y), right, top, b.SupernessInner),
		newSuperShape(left, bottom, right, y+(y-bottom), b.SupernessInner)
}

// leaveAngle returns the cosine and the sine of the angle at which both edges
// of a stroke meet the stem, for the notch above or below the join.
func (b bowlSpec) leaveAngle(notch, join float64) (cos, sin float64) {
	cos = math.Min(1, b.JoinThick/math.Abs(notch-join))
	return cos, math.Sqrt(1 - cos*cos)
}

// stemHandle returns the length of the handle at the stem of a connection
// which rises by rise, at the angle with sine sin to the horizontal: the given
// fraction of the run, but at most 0.95 of the length at which the handle
// reaches the height of the extreme.
func stemHandle(fraction, run, rise, sin float64) float64 {
	return math.Min(fraction*run, 0.95*rise/sin)
}

// bowlOuterCurves returns the four curves of the outer edge of the bowl, from
// the upper notch: up and left to the extreme at the top, round the free side
// to the free extreme, on to the extreme at the bottom, and right and up to the
// lower notch.
func bowlOuterCurves(sh bowlShape, b bowlSpec) [4]cubic {
	xt := b.xExtreme()
	run := b.X - xt
	cu, su := b.leaveAngle(b.NotchUpper, b.JoinUpper)
	cl, sl := b.leaveAngle(b.NotchLower, b.JoinLower)
	he := sh.OuterExtremeHandle * run
	hu := stemHandle(sh.OuterStemHandle, run, b.Top-b.NotchUpper, su)
	hl := stemHandle(sh.OuterStemHandle, run, b.NotchLower-b.Bottom, sl)
	up, down := b.outerShapes()

	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	if sh.Parabolic {
		return [4]cubic{
			joinCurve(b.X, b.NotchUpper, cu, su, xt, b.Top, sh.OuterStemHandle, sh.OuterExtremeHandle),
			up.cubic(1, true),
			down.cubic(2, true),
			joinCurve(b.X, b.NotchLower, cl, sl, xt, b.Bottom, sh.OuterStemHandle, sh.OuterExtremeHandle).reversed(),
		}
	}
	return [4]cubic{
		{pt(b.X, b.NotchUpper), pt(b.X-hu*cu, b.NotchUpper+hu*su), pt(xt+he, b.Top), pt(xt, b.Top)},
		up.cubic(1, true),   // top to the free extreme
		down.cubic(2, true), // free extreme to the bottom
		{pt(xt, b.Bottom), pt(xt+he, b.Bottom), pt(b.X-hl*cl, b.NotchLower-hl*sl), pt(b.X, b.NotchLower)},
	}
}

// bowlOuter returns the outer edge of the bowl as a chain.  It starts on the
// stem edge, at the upper notch.  The chain runs up and left to the extreme at
// the top, round the free side through the extreme on the left to the extreme
// at the bottom, and right and up to the stem edge.  It ends on the stem edge,
// at the lower notch.
//
// A bowl on the right of its stem is the mirror image of that in the edge of
// the stem, and the chain is the one which keeps the contour counter-clockwise:
// it starts at the lower notch, runs down and right to the extreme at the
// bottom, round the free side through the extreme on the right, and back to the
// stem edge at the upper notch, where it ends.  The caller goes up the stem
// edge to the lower notch before it and on up the stem edge after it.
func bowlOuter(sh bowlShape, b bowlSpec) chain {
	if b.OnRight {
		return bowlOuter(sh, b.leftForm()).mirrorX(0)
	}
	var c chain
	for _, q := range bowlOuterCurves(sh, b) {
		c = append(c, curveSeg(q))
	}
	return c
}

// bowlCounter returns the counter of the bowl as the chain of a clockwise
// contour of its own, which closes on itself, to be added with addContour.  The
// chain starts at the upper join, runs down the stem edge to the lower join,
// left and down to the extreme at the bottom, round the free side to the extreme
// at the top and right and down to the stem edge.  It ends at the upper join,
// where it started.
//
// A bowl on the right gives the mirror image, again clockwise: the chain starts
// at the upper join, runs round the free side, which is now on the right, and
// down the stem edge from the lower join to the upper join.
func bowlCounter(sh bowlShape, b bowlSpec) chain {
	if b.OnRight {
		return bowlCounter(sh, b.leftForm()).mirrorX(0)
	}
	xt := b.xExtreme()
	run := b.X - xt
	cu, su := b.leaveAngle(b.NotchUpper, b.JoinUpper)
	cl, sl := b.leaveAngle(b.NotchLower, b.JoinLower)
	bottom, top := b.Bottom+b.ThinBottom, b.Top-b.ThinTop
	he := sh.InnerExtremeHandle * run
	hu := stemHandle(sh.InnerStemHandle, run, top-b.JoinUpper, su)
	hl := stemHandle(sh.InnerStemHandle, run, b.JoinLower-bottom, sl)
	up, down := b.innerShapes()
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }

	if sh.Parabolic {
		return chain{
			lineSeg(pt(b.X, b.JoinUpper), pt(b.X, b.JoinLower)),
			curveSeg(joinCurve(b.X, b.JoinLower, cl, sl, xt, bottom, sh.InnerStemHandle, sh.InnerExtremeHandle)),
			curveSeg(down.cubic(1, false)),
			curveSeg(up.cubic(2, false)),
			curveSeg(joinCurve(b.X, b.JoinUpper, cu, su, xt, top, sh.InnerStemHandle, sh.InnerExtremeHandle).reversed()),
		}
	}
	return chain{
		lineSeg(pt(b.X, b.JoinUpper), pt(b.X, b.JoinLower)),
		curveSeg(cubic{pt(b.X, b.JoinLower),
			pt(b.X-hl*cl, b.JoinLower-hl*sl),
			pt(xt+he, bottom),
			pt(xt, bottom)}),
		curveSeg(down.cubic(1, false)), // bottom to the free extreme
		curveSeg(up.cubic(2, false)),   // free extreme to the top
		curveSeg(cubic{pt(xt, top),
			pt(xt+he, top),
			pt(b.X-hu*cu, b.JoinUpper+hu*su),
			pt(b.X, b.JoinUpper)}),
	}
}

// bowlFrom returns the bowlSpec of the bowl whose design values are b, on
// the stem edge at x with its free extreme at left.  Its free side has the thin
// parts, the overshoot and the squareness of the o, and the thickness, height
// and extremes at top and bottom which b gives, see design.Bowl.
func bowlFrom(p design.Params, b design.Bowl, x, left float64) bowlSpec {
	return bowlSpec{
		X: x, Left: left, Half: p.BowlHalf(b),
		Bottom: -p.OvershootLower, Top: p.BowlTop(b),
		ThinBottom: p.ThinLower, ThinTop: p.ThinLower, Thick: p.BowlThick(b),
		NotchUpper: b.NotchUpper, JoinUpper: b.JoinUpper,
		NotchLower: b.NotchLower, JoinLower: b.JoinLower,
		JoinThick:      p.BowlJoinThick,
		SupernessOuter: p.SupernessOuter, SupernessInner: p.SupernessInner,
		ThickPos: b.ThickPos,
	}
}
