// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// g draws the g: the bowl of a, GBowl of its width, joined to the plain stem
// as on q (plainBowl), and the stem from a flat cut at the x-height down
// into the open tail with its ball (tail), which j and y share.  The right
// side is that of q.
func g(p Params) letter {
	b := p
	b.BowlReach = math.Round(p.GBowl * p.BowlReach)
	e := b.bowlStemEdge()
	st := stroke.Stroke{Nodes: []stroke.Node{p.stemNode(e+p.Stem/2, p.BowlStemTop)}, Start: flatTop}
	t := p
	t.TailWider, t.TailTurn = p.GTailWider, p.GTailTurn
	nodes, end := t.tail(e)
	st.Nodes, st.End = append(st.Nodes, nodes...), end
	return letter{
		strokes: []stroke.Stroke{b.plainBowl(e), st},
		advance: e + p.QWidth - p.bowlStemEdge(),
	}
}
