// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The capital sharp s ẞ.
//
// A capital stem on the left with the foot serif of the capitals on the left
// rises into a rounded shoulder and a flat top, which runs to the right to a
// corner.  From the corner a straight diagonal runs down to the left, at
// capEszettAngle degrees from the horizontal and capEszettDiag stems thick at
// right angles to it, into a lower bowl.  The lower bowl is the lower bowl of
// ß with the weights of the capitals: a pair of super-ellipses with the
// squareness of O, whose outer edge leaves the right-hand edge of the
// diagonal in a crotch; its inner edge leaves the lower end of the diagonal
// in a point and rises in one curve to the top of the counter; it turns at the bottom to the left and ends in the ball terminal near
// the foot of the stem.  The white between the stem and the diagonal runs
// down past the point into the lower counter and out between the stem and the
// ball, so that nothing joins the stem below the top.
//
// The contour runs from the tip of the foot serif along the baseline, up the
// right-hand edge of the stem, round the inner corner under the top, along the
// underside of the top to the diagonal, down its left-hand edge to the point,
// round the inner edge of the lower bowl to the ball, round it, back along the
// outer edge of the lower bowl to the crotch, up the right-hand edge of the
// diagonal to the corner, along the top, round the shoulder and down the
// left-hand edge of the stem to the foot serif.

// Shape constants of ẞ.
const (
	capEszettInk      = 0.88  // the ink width, in ink widths of H
	capEszettCorner   = 0.95  // the corner of the top, in ink widths from the left of the ink
	capEszettAngle    = 58.0  // the diagonal, degrees from the horizontal
	capEszettDiag     = 0.86  // the diagonal at right angles to it, in capital stems
	capEszettTopBar   = 1.1   // the top at its right end, in bars of the capitals
	capEszettBowl     = 1.08  // the lower bowl at its right extreme, in capital stems
	capEszettCounter  = 0.50  // the top of the lower counter, in cap heights
	capEszettPoint    = 20.0  // the point where the diagonal ends, below the top of the lower counter
	capEszettWhite    = 0.4   // the least white between the stem and the point, in capital stems
	capEszettShoulder = 280.0 // the outer edge of the shoulder: its run on the top and on the stem
	capEszettInner    = 180.0 // the inner corner under the top: its run on the underside and on the stem
)

// newCapEszettFrame returns the frame of ẞ: the stem, the serif, the heights
// and the weights of the capitals.  The fields of eszettFrame which hold the
// upper bowl and the waist of ß are not used; waist holds the top of the
// lower counter.  The ink is capEszettInk of that of H wide, or wider where
// the point at the lower end of the diagonal would otherwise come closer than
// capEszettWhite stems to the stem.
func newCapEszettFrame(p design.Params) eszettFrame {
	c := p.Caps
	s := design.CapStemLeft(p)
	x0 := s - c.SerifJut
	w := capEszettInk * design.CapHInk(p)
	a := capEszettAngle * math.Pi / 180
	top, yp := p.CapHeight+p.OvershootUpper, capEszettCounter*p.CapHeight-capEszettPoint
	// The point lies capEszettCorner*w right of x0 less the run of the
	// diagonal from the corner down to it.
	least := s + c.Stem + capEszettWhite*c.Stem + capEszettDiag*c.Stem/math.Sin(a) + (top-yp)/math.Tan(a)
	w = math.Max(w, math.Ceil((least-x0)/capEszettCorner))
	f := eszettFrame{
		name: "uni1E9E", stem: s, width: c.Stem, jut: c.SerifJut,
		bottom: -p.OvershootUpper, top: top,
		waist: capEszettCounter * p.CapHeight,
		right: math.Round(x0 + w), upperRight: math.Round(x0 + capEszettCorner*w),
		thin: c.RoundThin, middle: capEszettTopBar * c.Bar, bowl: capEszettBowl * c.Stem,
		open: capEszettOpen,
		ball: newBallPart(p),
	}
	f.ballLeft = s + c.Stem + eszettBallGap
	f.lowerLeft = f.ballLeft - eszettLowerLeft
	f.advance = math.Round(f.right + c.SideRoundRight)
	return f
}

// capEszett draws ẞ.
func capEszett(p design.Params) (outline.Path, float64) {
	f := newCapEszettFrame(p)
	return capEszettOutline(p, f), f.advance
}

// capEszettOutline draws ẞ with the frame f, see the comment above.
func capEszettOutline(p design.Params, f eszettFrame) outline.Path {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	sr := f.stem + f.width
	under := f.top - f.middle // the underside of the top
	// The lower bowl: the counter's top at f.waist, the outer edge f.thin
	// above it.
	lo := newSuperShape(f.lowerLeft, f.bottom, f.right, f.waist+f.thin, p.SupernessOuter)
	li := newSuperShape(f.lowerLeft+f.open, f.bottom+f.thin, f.right-f.bowl, f.waist, p.SupernessInner)
	lower := lowerLeftBall(f.name, f.ball, lo.cubic(2, true), li.cubic(1, false), f.ballLeft,
		figStroke{Rise: true, Before: li.cubic(0, false)})
	// The diagonal: its right-hand edge runs from the corner of the top at
	// capEszettAngle, its left-hand edge parallel to it, the thickness to its
	// left.
	corner := pt(f.upperRight, f.top)
	yp := f.waist - capEszettPoint
	a := capEszettAngle * math.Pi / 180
	d := outline.Point{X: math.Cos(a), Y: math.Sin(a)} // up to the right
	shift := capEszettDiag * f.width / math.Sin(a)
	long := 2 * f.top
	rightEdge := lineSeg(pt(corner.X-long*d.X, corner.Y-long*d.Y), corner).cubic
	lp := pt(corner.X-shift, corner.Y)
	crotch := meetLines(lp, d, pt(sr, under), pt(1, 0))
	// The point: the left-hand edge of the diagonal ends capEszettPoint
	// below the top of the lower counter, where the inner edge of the bowl
	// leaves it in the direction of a circular arc to the top of the counter
	// and with its handle, in one curve to the top of the counter, which
	// it meets with the curvature of the super-ellipse there.
	point := pt(lp.X-(lp.Y-yp)/math.Tan(a), yp)
	top := li.cubic(3, false) // the top of the counter to its right extreme
	leave := 2 * math.Atan2(top.P0.Y-point.Y, top.P0.X-point.X)
	rise, err := g2End(point, dirOf(leave), arcHandle(point, dirOf(leave), top.P0, pt(1, 0)), top.P0, pt(1, 0), top.curvature(0))
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: no curve from the point to the top of the counter", f.name))
	}
	fromPoint := join(curveChain(rise), chain{curveSeg(top)})
	var toCrotch chain
	outer := []cubic{lo.cubic(0, true), lo.cubic(1, true)}
	for i, c := range outer {
		if tc, _, ok := crossing(c, rightEdge); ok {
			for _, c := range outer[:i] {
				toCrotch = append(toCrotch, curveSeg(c))
			}
			toCrotch = append(toCrotch, curveChain(c.sub(0, tc))...)
			break
		}
	}
	if toCrotch == nil {
		panic(fmt.Sprintf("glyphs: glyph %s: the diagonal misses the lower bowl", f.name))
	}
	// The shoulder and the inner corner: each one curve between the two
	// straights, with the curvature zero at both ends.
	sh := capEszettShoulder
	shoulder := cubic{pt(f.stem+sh, f.top), pt(f.stem, f.top), pt(f.stem, f.top), pt(f.stem, f.top-sh)}
	in := capEszettInner
	innerCorner := cubic{pt(sr, under-in), pt(sr, under), pt(sr, under), pt(sr+in, under)}
	foot := capFootLeft(p, f.stem, f.jut)
	c := join(
		lineChain(foot.end(), pt(sr, 0)),
		lineChain(pt(sr, 0), innerCorner.P0),
		curveChain(innerCorner),
		lineChain(innerCorner.P3, crotch),
		lineChain(crotch, fromPoint.start()),
		fromPoint,
		lower,
		chain{curveSeg(lo.cubic(3, true))},
		toCrotch,
		lineChain(toCrotch.end(), corner),
		lineChain(corner, shoulder.P0),
		curveChain(shoulder),
		lineChain(shoulder.P3, foot.start()),
		foot,
	)
	return contourPath(c)
}
