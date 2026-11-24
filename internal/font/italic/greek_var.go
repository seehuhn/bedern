// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// The variant forms ϖ and ϑ.

func init() {
	// The ω of ϖ, moved to the right by PiVarOver, has the tips of ω.
	for _, e := range intendedFindings["omega"] {
		e.X += defaultGreek().PiVarOver
		intendedFindings["piVar"] = append(intendedFindings["piVar"], e)
	}
	addGreek("piVar", 'ϖ', piVar)
	addGreek("thetaVar", 'ϑ', thetaVar)
}

// piVar draws ϖ: the ω, a little lower, under the bar of π, which runs from
// left of its left side to right of its right side; the tips of the ω end
// inside the bar.
func piVar(p Params) letter {
	l := omega(p)
	x0 := p.SideRoundLeft - p.Gk.PiVarOver
	x1 := p.SideRoundLeft + p.Gk.OmegaWidth + p.Gk.PiVarOver
	for i := range l.strokes {
		l.strokes[i] = moved(l.strokes[i], p.Gk.PiVarOver)
	}
	l.strokes = append(l.strokes, moved(p.greekBar(x0, x1), p.Gk.PiVarOver))
	l.advance = x1 + p.Gk.PiVarOver + p.SideRoundRight
	return l
}

// thetaVar draws ϑ, the cursive theta: one stroke of the pen which begins
// with the entry at the x-height, runs down the left side and round the
// bottom as the turn of u, up the tall right side to the height of the
// ascender, over the top of the loop and down its left side, and back to the
// right across the right side into a level end at p.Gk.ThetaVarEndY.
func thetaVar(p Params) letter {
	e := p.SideStemLeft - p.Gk.ThetaVarLeft
	c := e + p.Stem/2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	r := c + p.Gk.ThetaVarWidth
	lx := r - p.Gk.ThetaVarLoop
	s := stroke.Stroke{Nib: p.nib(), Nodes: p.entry(e, p.XHeight)[:2], Start: round, End: stroke.End{Kind: stroke.Cut}}
	// The next node lies low on the left side; the top takes the curvature
	// of the hook.
	s.Nodes[1].K = p.Gk.ThetaVarTopK
	s.Nodes = append(s.Nodes,
		node(c-p.Gk.ThetaVarIn, p.Gk.ThetaVarLowSide, 270),
		node((c+r)/2, -p.Overshoot+wh/2, 0),
		node(r-wv/2, p.Gk.ThetaVarSide, 90),
		node((lx+r)/2, p.Ascender+p.Overshoot-wh/2, 180),
		node(lx+wv/2, p.Gk.ThetaVarLoopSide, 270),
		stroke.Node{P: pt(r+p.Gk.ThetaVarEnd, p.Gk.ThetaVarEndY), Dir: p.Gk.ThetaVarEndDir, K: auto, W: p.ExitEndW + p.Gk.ThetaVarEndW, DW: 0},
	)
	return letter{strokes: []stroke.Stroke{s}, advance: r + p.Gk.ThetaVarEnd + p.SideRoundRight}
}

// The proportions of ϖ and ϑ, in the upright description.
const ()
