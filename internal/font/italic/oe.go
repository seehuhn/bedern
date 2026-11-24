// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// oe draws the œ: the left half of o, open on the right, and the e moved to
// the right so that its left side stands where the right side of the o
// would; the top and the bottom of the o run into the left side of the e
// and end inside it, OeJoin degrees above and below its middle on its
// centre line, in the direction of the o there.  The left side of the e is
// the stroke which the two letters share.
func oe(p Params) letter {
	l := p.SideRoundLeft
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	cy := (top + bottom) / 2
	b := (top - bottom - wh) / 2
	ao := (p.OBowlWidth - wv) / 2
	cxo := l + p.OBowlWidth/2
	dx := math.Round(p.OBowlWidth - wv + p.OeGap)
	ae := (p.EBowlWidth - wv) / 2
	cxe := l + dx + p.EBowlWidth/2
	join := func(deg float64) stroke.Node {
		t := deg * math.Pi / 180
		dir := math.Atan2(b*math.Cos(t), -ao*math.Sin(t)) * 180 / math.Pi
		return stroke.Node{P: pt(cxe-ae*math.Cos(t)+p.OeIn, cy+b*math.Sin(t)), Dir: dir, K: auto, W: auto, DW: auto}
	}
	// The o travels anticlockwise from the upper join over the top, down
	// the left side and round the bottom to the lower join.
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: auto}
	}
	half := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			join(p.OeJoin),
			node(cxo+oTurn+p.OeTurn, top-wh/2, 180),
			node(l+wv/2+oSide[0], cy, 270),
			node(cxo-oTurn+p.OeTurn, bottom+wh/2, 0),
			join(-p.OeJoin),
		},
		Start: hidden,
		End:   hidden,
	}
	le := e(p)
	out := letter{strokes: []stroke.Stroke{half}, advance: dx + le.advance}
	for _, s := range le.strokes {
		out.strokes = append(out.strokes, moved(s, dx))
	}
	return out
}

// ae draws the æ: the bowl of a, and the e whose left side is the stem of
// a.  The e is moved to the right so that the outer left extreme of its
// round, were it drawn, would stand AeGap right of the left edge of the
// stem.  One stroke runs from the flat cut of a down the stem to AeStemEnd,
// turns there as the exit does and runs on as the bottom and the exit of
// the e; the other is the bar, the right side and the top of the e, which
// runs into the stem at AeTopJoin as the bowl of a does, ending inside it.
// The bowl of a runs into the stem at AeLowJoin in the direction AeLowDir.
func ae(p Params) letter {
	edge := p.bowlStemEdge()
	c := edge + p.Stem/2
	// The bar of the e runs straight over AeBarRun of its round.
	pe := p
	pe.EBarRun = p.AeBarRun
	le := e(pe)
	dx := edge + p.AeGap - p.SideRoundLeft
	en := moved(le.strokes[0], dx).Nodes
	st := stroke.Stroke{
		Nib: p.nib(),
		Nodes: append([]stroke.Node{
			p.stemNode(c, p.BowlStemTop),
			p.stemNode(c, p.AeStemEnd),
			p.vertNode(c, p.AeStemEnd, p.AeStemEnd-p.AeTopIn, p.across(p.Stem))},
			en[5:]...),
		Start: flatTop,
		End:   round,
	}
	d := p.AeTopDir * math.Pi / 180
	es := stroke.Stroke{
		Nib: p.nib(),
		Nodes: append(en[:4:4], stroke.Node{
			P:   pt(c+p.BowlDepth*math.Cos(d), p.AeTopJoin+p.BowlDepth*math.Sin(d)),
			Dir: p.AeTopDir, K: auto, W: p.BowlJoinW, DW: auto}),
		Start: hidden,
		End:   hidden,
	}
	q := p
	q.BowlLowJoin, q.BowlLowDir = p.AeLowJoin, p.AeLowDir
	return letter{strokes: []stroke.Stroke{q.bowl(edge), st, es}, advance: dx + le.advance}
}
