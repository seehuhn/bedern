// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// bowl returns the bowl of a, d, g and q left of the stem whose left edge
// stands at edge: one stroke of the pen which leaves the stem BowlTopJoin
// below the x-height in the direction BowlTopDir, runs over the top, down the
// left side, whose outer extreme lies BowlReach left of the stem, and round
// the bottom, and runs back into the stem at BowlLowJoin in the direction
// BowlLowDir.  Both ends lie BowlDepth inside the stem, so that the union
// with the stem makes the corners where the bowl meets it.  The extremes of
// the centre line are the nodes, with the thickness of the pen not changing
// there, as on o, except at the top, where the thickness is still growing
// from BowlTopW at the join.
func (p Params) bowl(edge float64) stroke.Stroke {
	left := edge - p.BowlReach
	cx := (left + edge) / 2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	cy := (top + bottom) / 2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: auto}
	}
	inside := func(y, dir, depth float64) Point {
		d := dir * math.Pi / 180
		return pt(edge+depth*math.Cos(d), y+depth*math.Sin(d))
	}
	enter := stroke.Node{P: inside(p.XHeight-p.BowlTopJoin, p.BowlTopDir, -p.BowlDepth), Dir: p.BowlTopDir, K: auto, W: p.BowlTopW, DW: auto}
	// The top stroke grows from BowlTopW to the thickness of the pen at the
	// top at an even rate, the fraction BowlTopRise of that of the chord, so that it
	// arrives still growing and goes on into the left side without a plateau.  With
	// BowlTopRise zero the pen's rule gives the rate.
	crown := node(cx+p.BowlTop, top-wh/2, 180)
	if p.BowlTopRise != 0 {
		crown.DW = p.BowlTopRise * (wh - enter.W) / math.Hypot(crown.P.X-enter.P.X, crown.P.Y-enter.P.Y)
	}
	return stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			enter,
			crown,
			node(left+wv/2+oSide[0], cy, 270),
			node(cx+p.BowlBottom, bottom+wh/2, 360+p.BowlBottomDir),
			{P: inside(p.BowlLowJoin, p.BowlLowDir, p.BowlDepth), Dir: p.BowlLowDir, K: auto, W: p.BowlJoinW, DW: auto},
		},
		Start: hidden,
		End:   hidden,
	}
}

// plainBowl returns the bowl of q and g, whose stem runs straight past the
// lower join: it runs into the stem at BowlPlainJoin in the direction
// BowlPlainDir, lower and flatter than on a and d, where the turn of the
// exit takes the corner and the join stands above it.
func (p Params) plainBowl(edge float64) stroke.Stroke {
	q := p
	q.BowlLowJoin, q.BowlLowDir = p.BowlPlainJoin, p.BowlPlainDir
	return q.bowl(edge)
}

// bowlStemEdge returns the left edge of the stem of a, d, g and q, right of
// the bowl.
func (p Params) bowlStemEdge() float64 { return p.SideRoundLeft + p.BowlReach }
