// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// y draws the y: the u whose right stem runs on below the baseline into the
// tail with its ball (tail).  The entry, the left stem, the turn and the
// rise are those of u between stems YCounter apart, the rise meeting the
// right stem in the direction YRiseDir and the left stem straight down to
// XHeight - YShoulder; the right stem starts
// at the x-height in a flat cut.
func y(p Params) letter {
	q := p
	q.NCounter, q.URiseDir, q.Shoulder = p.YCounter, p.YRiseDir, p.YShoulder
	turn := u(q).strokes[0]
	e2 := p.SideStemLeft + p.Stem + p.YCounter
	right := stroke.Stroke{Nodes: []stroke.Node{p.stemNode(e2+p.Stem/2, p.XHeight)}, Start: flatTop}
	nodes, end := p.tail(e2)
	right.Nodes, right.End = append(right.Nodes, nodes...), end
	return letter{
		strokes: []stroke.Stroke{turn, right},
		advance: e2 + p.Stem + p.YSideRight,
	}
}
