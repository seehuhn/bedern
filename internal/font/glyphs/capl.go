// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capL draws the capital L: the capital stem with head serifs on both sides and
// a foot serif on the left, and the bottom arm of E, LArm long, with its serif
// rising BottomRise above it.
func capL(p design.Params) (outline.Path, float64) {
	c := p.Caps
	s := capStem{Left: design.CapStemLeft(p), FootLeft: c.SerifJut, HeadLeft: c.SerifJut, HeadRight: c.SerifJut}
	r := s.right(p)

	bottom := capArmSpec(p, r, r+c.LArm, c.Bar)
	bottom.Rise = c.BottomRise
	b := capArmBottom(bottom)
	top := s.top(p, b.end(), p.CapHeight/2)
	ch := join(lineChain(s.start(), b.start()), b, top)
	return capContour(join(ch, s.down(p, ch.end()))), design.CapLAdvance(p)
}
