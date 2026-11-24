// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// b draws the b: one stroke from the entry at the ascender down the stem,
// which turns at the bottom as the turn of u does into the bowl, runs round
// the free side and over the top and back into the stem where the arch of n
// leaves it, so that the bowl closes at the top as an arch.  Its left side is
// BLeftLess less than that of n.
func b(p Params) letter {
	e := p.SideStemLeft - p.BLeftLess
	free := p.BWidth - p.SideRoundRight
	s := stroke.Stroke{Nib: p.nib(), Nodes: p.entry(e, p.XHeight+p.AscenderEntryRise), Start: round, End: hidden}
	s.Nodes = append(s.Nodes, p.bpBowlNodes(e, free)...)
	q := p
	q.ArchStartDir = p.BBackDir
	back := q.archNodes(e, e+p.Stem+p.NCounter)[0]
	back.Dir += 180
	s.Nodes = append(s.Nodes, back)
	return letter{strokes: []stroke.Stroke{s}, advance: p.BWidth}
}

// bpBowlNodes returns the nodes of the bowl of b and p right of the stem whose
// left edge stands at e, from the stem down round the bottom, up the free
// side, whose outer edge stands at free at ShearHeight, and over the top: the
// stem straight down to where the turn of u leaves it, the node where the
// turn runs vertically, the bottom, the free side and the top, the thickness
// from the pen.
func (p Params) bpBowlNodes(e, free float64) []stroke.Node {
	c := e + p.Stem/2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	low := p.XHeight - p.Shoulder
	node := func(x, y, dir, dw float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: dw}
	}
	return []stroke.Node{
		p.stemNode(c, low),
		p.vertNode(c, low, p.BTurnVert*p.XHeight, p.BTurnW*p.Curve),
		node(c+p.BBottomAt*(free-c), -p.Overshoot+wh/2, 0, 0),
		node(free-wv/2-oSide[1], p.BSideY*p.XHeight, 90, auto),
		node(c+p.BTopAt*(free-c), p.XHeight+p.Overshoot-wh/2, 180, 0),
	}
}
