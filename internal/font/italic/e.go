// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// e draws the e: one stroke of the pen which starts inside the left side at
// the left end of the bar, runs up to the right along the bar, which rises
// at EBarRise, turns up the right side of the eye, over the top, down the
// left side and round the bottom into the end of the exit.  The bar passes
// through the left side, and the union of the stroke with itself leaves no
// stub of it inside.  The bar runs straight over EBarRun of the width of
// the round, so that the eye and the turn into its right side keep their
// proportions as the round narrows; EBar is the height of the centre line of
// the bar at the middle of the round.
func e(p Params) letter {
	l := p.SideRoundLeft
	rt := l + p.EBowlWidth
	cx := (l + rt) / 2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	cy := (top + bottom) / 2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	tan := math.Tan(p.EBarRise * math.Pi / 180)
	bar := func(x float64) stroke.Node {
		return stroke.Node{P: pt(x, p.EBar+(x-cx)*tan), Dir: p.EBarRise, K: 0, W: p.EBarW, DW: 0}
	}
	x0 := l + p.EBarIn
	return letter{
		strokes: []stroke.Stroke{{
			Nib: p.nib(),
			Nodes: []stroke.Node{
				bar(x0),
				bar(x0 + p.EBarRun*p.EBowlWidth),
				node(rt-wv/2, p.EEye, 90),
				node(cx-p.ETurn, top-wh/2, 180),
				node(l+wv/2+oSide[0], cy, 270),
				node(cx-oTurn+p.ETurn, bottom+wh/2, 0),
				p.exitEnd(l + p.ETipInk),
			},
			Start: hidden,
			End:   round,
		}},
		advance: p.EWidth,
	}
}
