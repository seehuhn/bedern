// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The capital U.
//
// U is two stems joined at the bottom by a round: the left stem is the capital
// stem (Stem), the right one thin (UThin), and
// both carry the head serifs of H on both sides, those of the thin stem with
// the smaller brackets of every thin stem (thinStemParams).  The round is drawn as the
// lower half of an O whose sides are the two stems: the outer edge leaves the
// outer edges of both stems vertically at UCurve and has its lowest point,
// OvershootUpper below the baseline, in the middle between them; the counter
// leaves the inner edges of the stems a little lower (uInnerDrop) and has its
// lowest point RoundThin above that of the outer edge, in the middle of the
// counter.  The stroke is thickest in the left stem, thins round the bottom
// to RoundThin and runs on into the thin right stem without a break.  Each of
// the four quarters is one cubic from the stem, which it leaves vertically, to
// the lowest point, which it reaches horizontally; the handle at the stem is
// long, and the curvature starts small after the straight stem and grows
// evenly (uShape).
//
// The contour runs from the left-hand edge of the left stem at UCurve round
// the outer edge to the right-hand edge of the right stem, up it and over the
// head serifs of the right stem, down its left-hand edge and round the counter
// to the right-hand edge of the left stem, up it, over the head serifs of the
// left stem and down to where it started.

// A uShapeConst holds the handles of the quarters of the round of U: at the
// stem as a fraction of the drop of the quarter, at the lowest point as a
// fraction of its run, for the outer edge and the counter.
type uShapeConst struct {
	OuterStem, OuterLow float64
	InnerStem, InnerLow float64
}

var uShape = uShapeConst{
	OuterStem: 0.75, OuterLow: 0.55,
	InnerStem: 0.75, InnerLow: 0.55,
}

// uInnerDrop is how far below UCurve the counter leaves the inner edges of the
// stems.
const uInnerDrop = 12

// stemToLow returns the quarter from the stem at a, which it leaves
// vertically, to the lowest point at l, which it reaches horizontally, with
// the handles fa (of the drop) and fl (of the run).
func stemToLow(a, l outline.Point, fa, fl float64) cubic {
	return cubic{
		P0: a,
		P1: outline.Point{X: a.X, Y: a.Y + fa*(l.Y-a.Y)},
		P2: outline.Point{X: l.X + fl*(a.X-l.X), Y: l.Y},
		P3: l,
	}
}

// capUStems returns the four edges of the stems of U, from the left.
func capUStems(p design.Params) (ll, lr, rl, rr float64) {
	c := p.Caps
	ll = c.SideStemLeft
	lr = ll + c.Stem
	rl = lr + design.CapUCounter(p)
	rr = rl + c.UThin
	return
}

// capU draws the capital U, see above.
func capU(p design.Params) (outline.Path, float64) {
	c := p.Caps
	j := c.SerifJut
	ll, lr, rl, rr := capUStems(p)
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	cap := p.CapHeight
	low := -p.OvershootUpper
	lowOut := pt((ll+rr)/2, low)
	lowIn := pt((lr+rl)/2, low+c.RoundThin)
	hIn := c.UCurve - uInnerDrop

	outer := chain{
		curveSeg(stemToLow(pt(ll, c.UCurve), lowOut, uShape.OuterStem, uShape.OuterLow)),
		curveSeg(stemToLow(pt(rr, c.UCurve), lowOut, uShape.OuterStem, uShape.OuterLow).reversed()),
	}
	inner := chain{
		curveSeg(stemToLow(pt(rl, hIn), lowIn, uShape.InnerStem, uShape.InnerLow)),
		curveSeg(stemToLow(pt(lr, hIn), lowIn, uShape.InnerStem, uShape.InnerLow).reversed()),
	}
	// The top of a stem whose edges are at x0 and x1, from the right-hand edge
	// at the height from to the left-hand edge at the height to.
	top := func(q design.Params, x0, x1, from, to float64) chain {
		hr := capHeadRight(q, x1, cap, j)
		hl := capHeadLeft(q, x0, cap, j)
		return join(
			lineChain(pt(x1, from), hr.start()), hr,
			lineChain(hr.end(), hl.start()), hl,
			lineChain(hl.end(), pt(x0, to)),
		)
	}
	ch := join(
		outer,
		top(thinStemParams(p, c.UThin), rl, rr, c.UCurve, hIn),
		inner,
		top(p, ll, lr, hIn, c.UCurve),
	)
	return capContour(ch), design.CapUAdvance(p)
}
