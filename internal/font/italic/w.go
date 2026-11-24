// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// w draws the w: the entry and the left stem, which turns at the bottom as the
// turn of u does and runs up to the right into the foot of the second stem,
// WCounter from the first; the second stem starts at the x-height in a flat
// cut and turns at its foot into the free side and the end of v.  The first
// turn is a circular arc from its lowest node to its end inside the second
// stem, which it reaches travelling at WJoinDir, so that the white above it
// meets the stem in an open corner.
func w(p Params) letter {
	e1 := p.SideStemLeft
	e2 := e1 + p.Stem + p.WCounter
	c1, c2 := e1+p.Stem/2, e2+p.Stem/2
	yb := -p.Overshoot + p.nibWidth(0)/2
	xb := c1 + p.WBottom*(c2-c1)
	join := pt(c2, yb+(c2-xb)*math.Tan(p.WJoinDir/2*math.Pi/180))
	first := stroke.Stroke{Nib: p.nib(), Nodes: p.entry(e1, p.XHeight), Start: round, End: hidden}
	first.Nodes = append(first.Nodes,
		p.stemNode(c1, p.VTurn),
		p.vertNode(c1, p.VTurn, p.VVert, p.ShoulderW),
		stroke.Node{P: pt(xb, yb), Dir: 0, K: auto, W: auto, DW: 0},
		stroke.Node{P: join, Dir: p.WJoinDir, K: auto, W: p.URiseW, DW: 0},
	)
	adv := c2 + p.WFree + p.VSideX
	nodes, end := p.vTurn(c2, adv)
	second := stroke.Stroke{Nib: p.nib(), Nodes: append([]stroke.Node{p.stemNode(c2, p.XHeight)}, nodes...), Start: flatTop, End: end}
	return letter{strokes: []stroke.Stroke{first, second}, advance: adv}
}
