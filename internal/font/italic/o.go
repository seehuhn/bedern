// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// o draws the o: one closed stroke of the pen through four nodes at the
// extremes of its centre line, upright, the outer edge OBowlWidth wide at
// ShearHeight and overshooting the x-height and the baseline.
func o(p Params) letter {
	l := p.SideRoundLeft
	r := l + p.OBowlWidth
	cx := (l + r) / 2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	cy := (top + bottom) / 2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: auto}
	}
	return letter{
		strokes: []stroke.Stroke{{
			Closed: true,
			Nib:    p.nib(),
			Nodes: []stroke.Node{
				node(r-wv/2-oSide[1], cy, 90),
				node(cx+oTurn, top-wh/2, 180),
				node(l+wv/2+oSide[0], cy, 270),
				node(cx-oTurn, bottom+wh/2, 0),
			},
		}},
		advance: r + p.SideRoundRight,
	}
}

// oSide moves the nodes at the sides of o inwards, left and right, so that
// the outer edge, whose extreme the pen puts a little outside the node, has
// its extremes at the sides of o at ShearHeight.
var oSide = [2]float64{1.4, 1.6}

// oTurn moves the node at the top of o to the right and the one at the
// bottom to the left, so that the upper right and the lower left of the
// bowl, which the shear draws out, are as full as the other two quarters.
const oTurn = 4.0
