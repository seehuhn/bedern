// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// p draws the p: the entry and the stem down to the descender line, where
// it ends in the foot serif, and the bowl, a stroke which leaves the stem
// where the arch of n leaves it, runs over the top, down the free side and
// round the bottom, and runs back into the stem at PJoin.
func p(p Params) letter {
	e := p.SideStemLeft
	c := e + p.Stem/2
	free := p.PWidth - p.SideRoundRight
	stem := stroke.Stroke{Nodes: p.entry(e, p.XHeight), Start: round}
	stem.Nodes = append(stem.Nodes, p.footNodes(c)...)
	stem.End = hidden

	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir, dw float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: dw}
	}
	q := p
	q.ArchStartDir, q.Branch = p.PStartDir, p.PBranch
	start := q.archNodes(e, e+p.Stem+p.NCounter)[0]
	// The end: the counter, on the right of the stroke, meets the right edge
	// of the stem at PJoin; the centre line lies PJoinBack inside the stem.
	d := p.PJoinDir * math.Pi / 180
	t, rn := pt(math.Cos(d), math.Sin(d)), pt(math.Sin(d), -math.Cos(d))
	w := p.ArchStartW
	join := pt(e+p.Stem, p.PJoin)
	end := pt(join.X-rn.X*w/2+p.PJoinBack*t.X, join.Y-rn.Y*w/2+p.PJoinBack*t.Y)
	bowl := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			start,
			node(c+p.PTopAt*(free-c), p.XHeight+p.Overshoot-wh/2, 0, 0),
			node(free-wv/2-oSide[1], p.BSideY*p.XHeight, -90, auto),
			node(c+p.PBottomAt*(free-c), -p.Overshoot+wh/2, 180, 0),
			{P: end, Dir: p.PJoinDir, K: auto, W: w, DW: 0},
		},
		Start: hidden,
		End:   hidden,
	}
	return letter{strokes: []stroke.Stroke{stem, bowl, p.footSerif(e)}, advance: p.PWidth}
}

// footNodes returns the node where a stem whose centre line stands at c at
// ShearHeight ends inside the foot serif at the descender line: in the middle
// of the serif, so that the cut at right angles lies inside it and the stem
// meets the top of the serif in two corners.
func (p Params) footNodes(c float64) []stroke.Node {
	t := p.FootThick * p.Hair
	return []stroke.Node{p.stemNode(c, -p.Descender+t/2)}
}

// footSerif returns the slab of the foot serif under the stem whose left
// edge stands at e: a stroke along the descender line, FootThick hairlines
// thick, reaching FootJut beyond the right edge of the stem and FootObtuse
// as far beyond the left one, its tips cut square.
func (p Params) footSerif(e float64) stroke.Stroke {
	t := p.FootThick * p.Hair
	y := -p.Descender + t/2
	return stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(e-p.FootObtuse*p.FootJut, y), Dir: 0, W: t},
			{P: pt(e+p.Stem+p.FootJut, y), Dir: 0, W: t},
		},
		Start: round,
		End:   round,
	}
}
