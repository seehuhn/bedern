// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capH draws the capital H: two capital stems with all four serifs each, the
// inner ones as long as the outer ones, and the crossbar between them, Bar thick
// with its centre at HBarHeight.  It is one contour, walked from the tip of the
// left foot serif of the left stem:
//
//	the right-hand side of the left stem up to the bar, the underside of the bar,
//	down the left-hand edge of the right stem and round its inner foot serif,
//	the right-hand side of the right stem with its two outer serifs, the top of
//	the right stem down to the bar, the top of the bar, the top of the left
//	stem, and down its left-hand edge.
func capH(p design.Params) (outline.Path, float64) {
	c := p.Caps
	j := c.SerifJut
	left := capStem{Left: design.CapStemLeft(p), FootLeft: j, FootRight: j, HeadLeft: j, HeadRight: j}
	right := capStem{Left: left.right(p) + c.HCounter, FootLeft: j, FootRight: j, HeadLeft: j, HeadRight: j}
	lower := c.HBarHeight - c.Bar/2
	upper := c.HBarHeight + c.Bar/2

	up1 := left.up(p, left.start(), lower)
	bar1 := capCrossbar(left.right(p), right.Left, lower)
	down2 := right.down(p, bar1.end())
	up2 := right.up(p, down2.end(), upper)
	top2 := right.top(p, up2.end(), upper)
	bar2 := capCrossbar(right.Left, left.right(p), upper)
	top1 := left.top(p, bar2.end(), lower)
	down1 := left.down(p, top1.end())
	return capContour(join(up1, bar1, down2, up2, top2, bar2, top1, down1)), design.CapHAdvance(p)
}
