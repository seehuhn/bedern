// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capF draws the capital F: the stem of E with a foot serif on both sides, and
// the top and the middle arm of E with their serifs, the top arm FTopArm and the
// middle arm FMiddleArm long, the middle arm centred at FMiddleHeight.
func capF(p design.Params) (outline.Path, float64) {
	c := p.Caps
	s := capStem{Left: design.CapStemLeft(p), FootLeft: c.SerifJut, FootRight: c.SerifJut, HeadLeft: c.SerifJut}
	r := s.right(p)

	middle := capArmSpec(p, r, r+c.FMiddleArm, c.FMiddleHeight+c.Bar/2)
	middle.Rise, middle.Drop = c.MiddleReach, c.MiddleReach
	top := capArmSpec(p, r, r+c.FTopArm, p.CapHeight)
	top.Drop = c.TopDrop

	up := s.up(p, s.start(), middle.Lower)
	m := capArmMiddle(middle)
	t := capArmTop(top)
	head := capHeadLeft(p, s.Left, p.CapHeight, c.SerifJut)
	ch := join(
		up, m,
		lineChain(m.end(), t.start()), t,
		lineChain(t.end(), head.start()), head,
	)
	return capContour(join(ch, s.down(p, ch.end()))), design.CapFAdvance(p)
}
