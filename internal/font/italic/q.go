// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// q draws the q: the bowl and the flat top of a, with the stem carried down
// to the descender line into the foot serif of p (footNodes, footSerif),
// without an exit.  The bowl runs into the stem lower than on a
// (plainBowl).
func q(p Params) letter {
	e := p.bowlStemEdge()
	st := stroke.Stroke{Nodes: []stroke.Node{p.stemNode(e+p.Stem/2, p.BowlStemTop)}, Start: flatTop, End: hidden}
	st.Nodes = append(st.Nodes, p.footNodes(e+p.Stem/2)...)
	return letter{
		strokes: []stroke.Stroke{p.plainBowl(e), st, p.footSerif(e)},
		advance: p.QWidth,
	}
}
