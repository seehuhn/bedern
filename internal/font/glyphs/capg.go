// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capG draws the capital G: the round of O with the upper beak of C, and at the
// lower right a short stem which rises from the round to GStemTop and carries a
// capital head serif on each side, GSerifLeft into the counter and GSerifRight
// beyond the stem: the bar of G.
//
// The right-hand edge of the stem is the cut of the beak, BeakX from the left
// extreme.  It is straight from the bar down to GJoin above the baseline, where
// the lower arc of the round runs into it: the outer edge of that arc is one
// curve from the bottom extreme, which it leaves with the direction and the
// curvature of O, to the edge of the stem, which it meets running straight up
// with a fraction (capGFootCurve) of that curvature, and it ends on the vertical
// without a corner, a notch or a point (capGFoot).  The stem is Stem wide: its
// left-hand edge runs straight down from the foot of the bracket of the left
// serif until just above the counter and then, in one short curve which meets the
// counter with its direction and its curvature, into the counter (capGRise,
// capGDrop).
//
// The contour runs from the top extreme of the outer edge over the top, down the
// left side and round the bottom, up the lower right of the round to the spur,
// up the stem and round its serifs, down the counter, round the bottom and the
// left side of the counter, over its top and round the beak.
//
// Side kinds: the round side of O on the left, and on the right the tip of the
// right serif, GSideRight from the edge.
func capG(p design.Params) (outline.Path, float64) {
	return capContour(capGChain(p)), design.CapGAdvance(p)
}

// capGFootCurve is the curvature with which the outer edge of the lower arc of
// G meets the stem, as a fraction of its curvature at the bottom extreme.
const capGFootCurve = 1.0

// The left-hand edge of the stem of G runs straight down from the foot of the
// bracket to capGRise above the point where it would cross the counter, and
// then in one short curve into the counter, which it meets capGDrop below that
// point.
const (
	capGRise = 20
	capGDrop = 30
)

// capGChain returns the contour of G as a closed chain.
func capGChain(p design.Params) chain {
	c, g := p.Caps, p.Caps.CGSJQ
	f := capRound(p)
	xR := f.beakX(p)
	xL := xR - c.Stem
	top := g.GStemTop

	foot, join0 := capGFoot(p, f)

	// The left-hand edge of the stem: from the foot of the bracket of the left
	// serif straight down, then in one short curve into the counter, which it
	// meets with its direction and its curvature.
	qi := f.inner.cubic(0, false) // right extreme to the bottom
	bracket := outline.Point{X: xL, Y: top - c.BracketHeight}
	cross := qi.at(qi.tAtX(xL)).Y
	low := outline.Point{X: xL, Y: math.Min(cross+capGRise, bracket.Y)}
	tL := qi.tAtY(cross - capGDrop)
	blend, err := fairEnd(low, outline.Point{Y: -1}, qi.at(tL), qi.tangent(tL), qi.curvature(tL))
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph G: the stem does not run into the counter: %v; glyphs.Check reports such parameters", err))
	}

	hr := capHeadRight(p, xR, top, g.GSerifRight)
	hl := capHeadLeft(p, xL, top, g.GSerifLeft)
	return join(
		capRoundLeft(f),
		chain{curveSeg(foot)},
		lineChain(join0, hr.start()), hr,
		lineChain(hr.end(), hl.start()), hl,
		lineChain(bracket, low),
		splitAtExtrema(blend),
		chain{curveSeg(qi.sub(tL, 1))},
		capRoundLeftInner(f),
		capUpperBeak(p, f),
	)
}

// capGFoot returns the outer edge of the lower arc of G, from the bottom extreme
// of the round to the right-hand edge of the stem, and the point where it meets
// the stem, see the comment above.
func capGFoot(p design.Params, f capRoundFrame) (cubic, outline.Point) {
	g := p.Caps.CGSJQ
	xR := f.beakX(p)
	bottom := f.outer.cubic(3, true) // bottom to the right extreme
	j := outline.Point{X: xR, Y: g.GJoin}
	c, err := fairEnd(bottom.at(0), bottom.tangent(0), j, outline.Point{Y: 1}, capGFootCurve*bottom.curvature(0))
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph G: the lower arc does not run into the stem: %v; glyphs.Check reports such parameters", err))
	}
	return c, j
}
