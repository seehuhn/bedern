// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capE draws the capital E: the capital stem with serifs on the left at the
// head and the foot, and three arms on the right.  The top arm (ETopArm) has its
// serif hanging TopDrop below it, the middle arm (EMiddleArm, centred at
// EMiddleHeight) a serif MiddleReach above and below, and the bottom arm
// (EBottomArm), the longest, a serif rising BottomRise above it.  The contour
// runs from the tip of the left foot serif along the baseline to the end of the
// bottom arm, round its serif and back to the stem, up the right-hand edge to
// the middle arm and round it, up to the top arm and round it, along the top to
// the head serif and down the left-hand edge.
func capE(p design.Params) (outline.Path, float64) {
	c := p.Caps
	s := capStem{Left: design.CapStemLeft(p), FootLeft: c.SerifJut, HeadLeft: c.SerifJut}
	r := s.right(p)

	bottom := capArmSpec(p, r, r+c.EBottomArm, c.Bar)
	bottom.Rise = c.BottomRise
	middle := capArmSpec(p, r, r+c.EMiddleArm, c.EMiddleHeight+c.Bar/2)
	middle.Rise, middle.Drop = c.MiddleReach, c.MiddleReach
	top := capArmSpec(p, r, r+c.ETopArm, p.CapHeight)
	top.Drop = c.TopDrop

	b := capArmBottom(bottom)
	m := capArmMiddle(middle)
	t := capArmTop(top)
	head := capHeadLeft(p, s.Left, p.CapHeight, c.SerifJut)
	ch := join(
		lineChain(s.start(), b.start()), b,
		lineChain(b.end(), m.start()), m,
		lineChain(m.end(), t.start()), t,
		lineChain(t.end(), head.start()), head,
	)
	return capContour(join(ch, s.down(p, ch.end()))), design.CapEAdvance(p)
}
