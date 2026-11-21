// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capDBowl returns the bowl of D: one bowl from the baseline to the cap height,
// DWidth wide, whose strokes run straight for the fraction DFlat of it.
func capDBowl(p design.Params) capBowl {
	s := capBowlStem(p, false)
	return capBowlFrom(p, s.right(p), p.Caps.DWidth, p.Caps.DFlat, 0, p.CapHeight)
}

// capD draws the capital D: the capital stem with serifs on the left at head and
// foot, and one tall bowl on its right from the baseline to the cap height
// (capBowl).  The contour runs from the tip of the left foot serif along the
// baseline, round the bowl, along the cap height, over the head serif and down
// the left-hand edge; the counter is a second contour.
func capD(p design.Params) (outline.Path, float64) {
	s := capBowlStem(p, false)
	b := capDBowl(p)
	start := lineChain(s.start(), outline.Point{X: b.Edge})
	return capBowlLetter(p, s, join(start, b.outer()), b.counter()), design.CapDAdvance(p)
}
