// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// a draws the one-storey a: the bowl (bowl), and the stem from a flat cut at
// the x-height down into the exit.
func a(p Params) letter {
	e := p.bowlStemEdge()
	c := e + p.Stem/2
	st := stroke.Stroke{Nodes: []stroke.Node{p.stemNode(c, p.BowlStemTop)}, Start: flatTop, End: round}
	st.Nodes = append(st.Nodes, p.exit(e)...)
	return letter{
		strokes: []stroke.Stroke{p.bowl(e), st},
		advance: e + p.Stem + p.SideStemRight,
	}
}
