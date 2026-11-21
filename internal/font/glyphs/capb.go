// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capBBowls returns the two bowls of B: the lower one from the baseline to the
// upper edge of the waist, BLower wide, and the upper one from the lower edge
// of the waist to the cap height, BUpper wide.  The waist is the stroke which
// the two bowls share at the stem: the upper stroke of the lower bowl and the
// lower stroke of the upper bowl are one stroke, Bar thick and centred at
// BWaist.
func capBBowls(p design.Params) (lower, upper capBowl) {
	c := p.Caps
	x := capBowlStem(p, false).right(p)
	lower = capBowlFrom(p, x, c.BLower, c.BowlFlat, 0, c.BWaist+c.Bar/2)
	upper = capBowlFrom(p, x, c.BUpper, c.BowlFlat, c.BWaist-c.Bar/2, p.CapHeight)
	lower.Thick, upper.Thick = c.BLowerRound, c.BUpperRound
	return lower, upper
}

// capB draws the capital B: the capital stem with serifs on the left at head
// and foot and two bowls on its right (capBowl), the lower one wider and taller.
// The waist leaves the stem as one thin stroke, and the outer edges of the two
// bowls run on from it until they cross, which gives the outer contour a notch
// where the upper bowl comes down into the lower one.  The contour runs from the
// tip of the left foot serif along the baseline, round the lower bowl to the
// notch, round the upper bowl to the cap height, over the head serif and down the
// left-hand edge; the two counters are contours of their own.
func capB(p design.Params) (outline.Path, float64) {
	s := capBowlStem(p, false)
	lower, upper := capBBowls(p)
	lu, ul := lower.outerUpper(), upper.outerLower()
	ta, tb, ok := crossing(lu, ul)
	if !ok {
		panic("B: the outer edges of the two bowls do not cross")
	}
	c := join(
		lineChain(s.start(), outline.Point{X: lower.Flat}),
		chain{curveSeg(lower.outerLower()), curveSeg(lu.sub(0, ta)), curveSeg(ul.sub(tb, 1)), curveSeg(upper.outerUpper())},
		lineChain(outline.Point{X: upper.Flat, Y: upper.Top}, outline.Point{X: upper.Edge, Y: upper.Top}),
	)
	return capBowlLetter(p, s, c, upper.counter(), lower.counter()), design.CapBAdvance(p)
}
