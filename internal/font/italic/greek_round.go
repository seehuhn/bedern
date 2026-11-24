// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// The round Greek letters: ο, σ, θ, ϕ and ρ, made of the bowl of o (one
// closed stroke of the pen through the four extremes of its centre line)
// and of straight strokes.

func init() {
	addGreek("omicron", 'ο', omicron)
	addGreek("sigma", 'σ', sigma)
}

// roundNodes returns the four nodes of the bowl of o between the outer
// extremes l and r and between the heights bottom and top, from its right
// side anticlockwise.
func (p Params) roundNodes(l, r, bottom, top, dw, turn float64) []stroke.Node {
	cx, cy := (l+r)/2, (top+bottom)/2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: dw}
	}
	return []stroke.Node{
		node(r-wv/2-oSide[1], cy, 90),
		node(cx+turn, top-wh/2, 180),
		node(l+wv/2+oSide[0], cy, 270),
		node(cx-turn, bottom+wh/2, 0),
	}
}

// round returns the closed bowl between the outer extremes l and r and the
// heights bottom and top.
func (p Params) roundBowl(l, r, bottom, top, turn float64) stroke.Stroke {
	return stroke.Stroke{Closed: true, Nib: p.nib(), Nodes: p.roundNodes(l, r, bottom, top, 0, turn)}
}

// omicron draws ο, which is the o: no reader tells them apart, and
// mathematics does not use it.
func omicron(p Params) letter { return o(p) }

// sigma draws σ: one stroke of the pen which starts at the right end of the
// bar at the x-height, runs left along the top of the bowl of o, down its
// left side, round its bottom and up its right side, and ends inside the
// bar, so that the bowl closes under the bar in a crotch.
func sigma(p Params) letter {
	l := p.SideRoundLeft
	r := l + p.OBowlWidth
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	nodes := p.roundNodes(l, r, bottom, top, 0, oTurn)
	wh := p.nibWidth(0)
	barEnd := l + p.Gk.SigmaBar
	start := stroke.Node{P: pt(barEnd-p.ExitEndW/2, top-wh/2-p.Gk.SigmaFall), Dir: 180 - p.Gk.SigmaFallDir, K: auto, W: wh, DW: 0}
	end := stroke.Node{P: pt(r-p.Gk.SigmaClose, top-wh/2-p.Gk.SigmaDip), Dir: p.Gk.SigmaCloseDir, K: auto, W: wh, DW: 0}
	return letter{
		strokes: []stroke.Stroke{{
			Nib:   p.nib(),
			Nodes: []stroke.Node{start, nodes[1], nodes[2], nodes[3], nodes[0], end},
			Start: stroke.End{Kind: stroke.Cut},
			End:   hidden,
		}},
		advance: barEnd + p.SideRoundRight,
	}
}

// The bar of σ: its ink reaches p.Gk.SigmaBar right of the left of the
// bowl; the stroke of the bowl ends p.Gk.SigmaClose right of the middle and
// p.Gk.SigmaDip below the centre line of the bar.
