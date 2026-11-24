// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// tail returns the tail of g, j and y below the stem whose left edge stands
// at edge: the nodes which follow the stem's last node above the baseline,
// and the end of the stroke.  The stem runs straight down to TailTurn below
// the baseline and turns to the left into the lowest node, TailLow left of
// the centre of the stem, where the stroke runs horizontally, BallStalk
// thick, its outer edge TailDepth below the baseline; from there the outer
// edge winds into the ball by the rule of the spiral, the centre of the ball
// TailBallDepth above the tangent of the outer edge there.  A letter calls it
// as
//
//	s := stroke.Stroke{Nodes: stemNodes, Start: start}
//	nodes, end := p.tail(edge)
//	s.Nodes, s.End = append(s.Nodes, nodes...), end
//
// with the stem's nodes from its top (the entry of j, the flat cut of g and
// y) down to a node above the baseline.
func (p Params) tail(edge float64) ([]stroke.Node, stroke.End) {
	c := edge + p.Stem/2
	return []stroke.Node{
		p.stemNode(c, -p.TailTurn),
		{P: pt(c-p.TailLow-p.TailWider, -p.TailDepth+p.BallStalk/2), Dir: 180, K: auto, W: p.BallStalk, DW: 0},
	}, p.ballAtDepth(p.TailBallDepth)
}
