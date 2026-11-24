// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The spine of s carries the curve of the upper half into that of the lower,
// so that each of its edges changes from convex to concave once.
func init() {
	why := "the spine of s turns from the upper half into the lower: each edge inflects once"
	intendedFindings["s"] = []intended{
		{Kind: fair.Inflection, X: 174, Y: 208, Dir: -22, Reason: why},
		{Kind: fair.Inflection, X: 194, Y: 250, Dir: 158, Reason: why},
	}
}

// s draws the s: one stroke of the pen from the end at the upper right up
// over the top, down the left side of the upper half, through the spine,
// down the right side of the lower half, round the bottom and up to the end
// at the lower left; both ends are cut square, as the ends of the exits.
// The lower half is SLower wide in the ink and the upper SUpper, both
// centred on the set width at ShearHeight.
func s(p Params) letter {
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	mid := p.SWidth / 2
	ll, lr := mid-p.SLower/2, mid+p.SLower/2
	ul, ur := mid-p.SUpper/2+p.SUpperShift, mid+p.SUpper/2+p.SUpperShift
	top, bot := p.XHeight+p.Overshoot, -p.Overshoot
	r := p.SUpper / p.SLower
	cy := bot + (top-bot)/(1+r)
	node := func(x, y, dir, dw float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: dw}
	}
	w := p.SEndW * p.Thin
	return letter{
		strokes: []stroke.Stroke{{
			Nib: p.nib(),
			Nodes: []stroke.Node{
				{P: pt(ur-w/2, top-p.SEndDrop), Dir: p.SEndDir, K: auto, W: w, DW: 0},
				node((ul+ur)/2+p.STopShift, top-wh/2, 180, 0),
				node(ul+wv/2, cy+p.SSideAt*(top-cy), 270, auto),
				{P: pt(mid, cy), Dir: -p.SSpineAngle, K: 0, W: auto, DW: auto},
				node(lr-wv/2, cy-p.SSideAt*(cy-bot), 270, auto),
				node((ll+lr)/2-p.STopShift, bot+wh/2, 180, 0),
				{P: pt(ll+w/2, bot+p.SEndDrop), Dir: p.SEndDir, K: auto, W: w, DW: 0},
			},
			Start: round,
			End:   round,
		}},
		advance: p.SWidth,
	}
}
