// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capT draws the capital T: the capital stem with a foot serif on both sides,
// and the bar at the cap height, TBar thick, whose two arms reach TArm beyond the
// edges of the stem and end in serifs hanging TDrop below the bar.  The contour
// runs from the tip of the left foot serif, round the right foot, up the
// right-hand edge, along the underside of the right arm and round its serif,
// along the top to the end of the left arm, round its serif, along its
// underside to the stem and down the left-hand edge.
func capT(p design.Params) (outline.Path, float64) {
	c := p.Caps
	s := capStem{Left: design.CapTStemLeft(p), FootLeft: c.SerifJut, FootRight: c.SerifJut}
	r := s.right(p)
	lower := p.CapHeight - c.TBar

	right := capArm{Edge: r, End: r + c.TArm, Lower: lower, Upper: p.CapHeight, Drop: c.TDrop, SerifThick: c.ArmSerifThick}
	left := capArm{Edge: s.Left, End: s.Left - c.TArm, Lower: lower, Upper: p.CapHeight, Drop: c.TDrop, SerifThick: c.ArmSerifThick}

	up := s.up(p, s.start(), lower)
	ra := capArmTop(right)
	la := capArmTopLeft(left)
	ch := join(up, ra, lineChain(ra.end(), la.start()), la)
	return capContour(join(ch, s.down(p, ch.end()))), design.CapTAdvance(p)
}
