// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The Greek letters of a bowl with a bar or a stem: θ, ϕ, ρ and α.

func init() {
	intendedFindings["alpha"] = []intended{
		{Kind: fair.Wiggle, X: 221, Y: 429, Dir: 14, Reason: "the counter of α narrows towards the crossing: its top tightens on the way into the crotch"},
	}
	addGreek("theta", 'θ', theta)
	addGreek("phi", 'ϕ', phiStroked)
	addGreek("rho", 'ρ', rho)
	addGreek("alpha", 'α', alpha)
}

// theta draws θ: the bowl of o drawn tall, from the baseline to the
// ascender with the overshoot, and a straight bar of the hairline across
// its middle, whose ends lie inside the sides.
func theta(p Params) letter {
	l := p.SideRoundLeft
	r := l + p.Gk.ThetaWidth
	top, bottom := p.Ascender+p.Overshoot, -p.Overshoot
	wv := p.nibWidth(90)
	y := p.Gk.ThetaBar * p.XHeight
	bar := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(l+wv/2, y), Dir: 0, W: p.Hair},
			{P: pt(r-wv/2, y), Dir: 0, W: p.Hair},
		},
		Start: hidden, End: hidden,
	}
	return letter{
		strokes: []stroke.Stroke{p.roundBowl(l, r, bottom, top, p.Gk.ThetaTurn), bar},
		advance: r + p.SideRoundRight,
	}
}

// phiStroked draws ϕ, the closed form: the bowl of o, wider, and a straight
// stem through its middle from the height of the ascender less p.Gk.PhiLow
// down to the tails, in flat cuts.
func phiStroked(p Params) letter {
	l := p.SideRoundLeft
	r := l + p.Gk.PhiWidth
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	c := (l + r) / 2
	stem := stroke.Stroke{
		Nodes: []stroke.Node{p.stemNode(c, p.Ascender-p.Gk.PhiLow), p.stemNode(c, -p.TailDepth)},
		Start: flatTop, End: foot,
	}
	return letter{
		strokes: []stroke.Stroke{p.roundBowl(l, r, bottom, top, oTurn), stem},
		advance: r + p.SideRoundRight,
	}
}

// rho draws ρ: one stroke of the pen from inside the left side at the
// bottom, round the bottom, up the right side, over the top and down the
// left side, which runs on straight below the baseline as the stem of the
// descender and ends in a flat cut, so that the bowl closes on the stem in
// a crotch at its lower left.
func rho(p Params) letter {
	l := p.SideRoundLeft + p.Gk.RhoLeft
	r := l + p.Gk.RhoWidth
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	c := l + p.across(p.Stem)/2 + p.Gk.RhoStemIn
	nodes := p.roundNodes(l, r, bottom, top, 0, oTurn)
	s := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(c, p.Gk.RhoClose), Dir: p.Gk.RhoCloseDir, K: auto, W: p.Thin, DW: 0},
			nodes[3], nodes[0], nodes[1],
			p.stemNode(c, p.Gk.RhoJoin),
			p.stemNode(c, -p.TailDepth),
		},
		Start: hidden, End: foot,
	}
	return letter{strokes: []stroke.Stroke{s}, advance: r + p.SideRoundRight}
}

// alpha draws α: one stroke of the pen which starts at the tip at the upper
// right, runs down to the left through the waist into the bottom of the
// bowl, round its left side and over its top, and down to the right through
// the waist again into the turn of an exit, so that the bowl closes on the
// right where the stroke crosses itself, and the tip and the foot stand
// right of the crossing.
func alpha(p Params) letter {
	l := p.SideRoundLeft
	x := l + p.Gk.AlphaBowl // the centre line at the crossing
	cx := (l + x) / 2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	foot := x + p.Gk.AlphaFoot
	s := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(x+p.Gk.AlphaTip, p.XHeight-p.ExitEndW/2), Dir: p.Gk.AlphaTipDir, K: auto, W: p.ExitEndW, DW: auto},
			node(cx+p.Gk.AlphaTurn, bottom+wh/2, 180),
			node(l+wv/2+oSide[0], (top+bottom)/2, 90),
			node(cx-p.Gk.AlphaTurn, top-wh/2, 0),
			{P: pt(x, p.Gk.AlphaCross), Dir: p.Gk.AlphaDown, K: 0, W: auto, DW: 0},
			{P: pt(x+p.Gk.AlphaRun*cosd(p.Gk.AlphaDown), p.Gk.AlphaCross+p.Gk.AlphaRun*sind(p.Gk.AlphaDown)), Dir: p.Gk.AlphaDown, K: 0, W: auto, DW: 0},
			{P: pt(foot, -p.Overshoot+p.ExitLowW/2), Dir: 0, K: auto, W: p.ExitLowW, DW: 0},
			p.exitEnd(foot + p.Gk.AlphaFlick),
		},
		Start: round, End: round,
	}
	return letter{strokes: []stroke.Stroke{s}, advance: x + math.Max(p.Gk.AlphaTip, p.Gk.AlphaFoot+p.Gk.AlphaFlick) + p.Gk.AlphaSide}
}

// The proportions of θ, ϕ, ρ and α, in the upright description.
const ()

// cosd and sind return the cosine and the sine of an angle in degrees.
func cosd(a float64) float64 { return math.Cos(a * math.Pi / 180) }
func sind(a float64) float64 { return math.Sin(a * math.Pi / 180) }
