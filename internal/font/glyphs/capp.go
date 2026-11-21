// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capPBowl returns the bowl of P: from the thin stroke centred at PBar to the
// cap height, PWidth wide.
func capPBowl(p design.Params) capBowl {
	c := p.Caps
	s := capBowlStem(p, true)
	return capBowlFrom(p, s.right(p), c.PWidth, c.BowlFlat, c.PBar-c.Bar/2, p.CapHeight)
}

// capP draws the capital P: the capital stem with serifs on both sides at the
// foot and on the left at the head, and the bowl in the upper part (capBowl).
// The contour runs from the tip of the left foot serif along the baseline and
// round the right foot serif, up the right-hand edge of the stem to the
// underside of the bowl, round the bowl to the top of the stem, over the head
// serif and down the left-hand edge; the counter of the bowl is a second
// contour.
func capP(p design.Params) (outline.Path, float64) {
	s := capBowlStem(p, true)
	b := capPBowl(p)
	up := s.up(p, s.start(), b.Bottom)
	return capBowlLetter(p, s, join(up, b.outer()), b.counter()), design.CapPAdvance(p)
}
