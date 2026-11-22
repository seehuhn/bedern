// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The capital R.
//
// R is the stem of P with its serifs, the bowl of P with the values of R (RBar,
// RWidth, RBowlRound) and a leg.  The leg is a straight diagonal, RLeg thick at
// right angles, which leaves the underside of the bowl near the stem and runs
// down to the right, where it ends on the baseline in the foot serif of the
// diagonals (diagonal.footServif), the serif of the leg of K.
// Its left edge leaves the underside of the
// bowl RLegTop from the stem and reaches the baseline RLegFoot from the stem;
// the angle follows from the two.  Both edges of the leg end on the outer edge
// of the bowl in plain corners, as the leg of K ends on its arm: the left edge
// on the straight underside of the lower stroke, the right edge on the
// underside or on the lower quarter of the round, wherever its line meets it.
//
// The contour runs from the tip of the left foot serif along the baseline,
// round the right foot serif and up the right-hand edge of the stem to the
// underside of the bowl, along it to the leg, down the left edge of the leg,
// round the serifs of the leg and up its right edge to the bowl, round the bowl
// to the top of the stem, over the head serif and down the left-hand edge; the
// counter of the bowl is a second contour.

// capRBowl returns the bowl of R.
func capRBowl(p design.Params) capBowl {
	c := p.Caps
	s := capBowlStem(p, true)
	b := capBowlFrom(p, s.right(p), c.RWidth, c.BowlFlat, c.RBar-c.Bar/2, p.CapHeight)
	b.Thick = c.RBowlRound
	return b
}

// capRLeg returns the leg of R as a diagonal.
func capRLeg(p design.Params) diagonal {
	c := p.Caps
	edge := capBowlStem(p, true).right(p)
	lean := design.CapRLegLean(p)
	w := design.DiagWidth(c.RLeg, lean)
	return diagonal{X: edge + c.RLegFoot + w/2, Y: 0, Lean: lean, Thick: c.RLeg}
}

// capRJoin returns the point q where the right edge of the leg meets the outer
// edge of the bowl, and the outer edge of the bowl from q up to the extreme.
func capRJoin(b capBowl, leg diagonal) (q outline.Point, rest chain) {
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	bl := b.outerLower()
	if x := leg.edgeAt(edgeRight, b.Bottom).X; x < b.Flat-0.5 {
		q = pt(x, b.Bottom)
		return q, join(lineChain(q, pt(b.Flat, b.Bottom)), chain{curveSeg(bl)})
	}
	r0 := leg.edgeAt(edgeRight, 0)
	d := leg.down()
	far := pt(r0.X-1600*d.X, r0.Y-1600*d.Y)
	_, tb, ok := crossing(cubic{r0, r0, far, far}, bl)
	if !ok {
		panic("glyphs: R: the right edge of the leg does not meet the bowl; glyphs.Check reports such parameters")
	}
	q = bl.at(tb)
	return q, chain{curveSeg(bl.sub(tb, 1))}
}

// capREaseBracket returns the right foot serif of the leg, walked up the leg,
// with the end of its bracket, over the last capRRamp degrees of its turn, a
// curve (footCubic) which meets the right edge of the leg, on the way to q, at
// the same point without curvature; the bracket is long, and its curvature
// would otherwise jump to the straight edge.
func capREaseBracket(out chain, q outline.Point) chain {
	n := len(out)
	last := out[n-1]
	if last.line {
		return out
	}
	c := last.cubic
	up := unit(c.P3, q)
	u := tTurned(c, 1, capRRamp)
	ramp := straightFoot(c.at(u), c.tangent(u), c.curvature(u), c.P3, up)
	res := append(chain(nil), out[:n-1]...)
	return append(res, curveSeg(c.sub(0, u)), curveSeg(ramp))
}

// capRRamp is the turn in degrees of the end of the bracket of the right foot
// serif of the leg which capREaseBracket redraws.
const capRRamp = 15.0

// capR draws the capital R, see above.
func capR(p design.Params) (outline.Path, float64) {
	c := p.Caps
	s := capBowlStem(p, true)
	b := capRBowl(p)
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }
	leg := capRLeg(p)
	a := leg.edgeAt(edgeLeft, b.Bottom)

	thick := c.SerifThick
	in := leg.footServif(edgeLeft, 0, diagSerif{Over: c.RLegSerifLeft, Thick: thick, Bracket: design.CapDiagBracket(p)})
	out := leg.footServif(edgeRight, 0, diagSerif{Over: c.RLegSerifRight, Thick: thick, Bracket: design.CapDiagBracket(p)}).reversed()
	q, rest := capRJoin(b, leg)
	out = capREaseBracket(out, q)

	ch := join(
		s.up(p, s.start(), b.Bottom),
		lineChain(pt(b.Edge, b.Bottom), a),
		lineChain(a, in.start()), in,
		lineChain(in.end(), out.start()), out,
		lineChain(out.end(), q), rest,
		chain{curveSeg(b.outerUpper())},
		lineChain(pt(b.Flat, b.Top), pt(b.Edge, b.Top)),
	)
	return capBowlLetter(p, s, ch, b.counter()), design.CapRAdvance(p)
}
