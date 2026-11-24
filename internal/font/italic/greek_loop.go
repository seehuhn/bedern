// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The Greek letters of rounds with necks, loops and waists: δ, ε, φ, ω, β
// and γ.

func init() {
	tip := "the start of φ turns in at the upper left, so that each edge of the left side inflects once below it"
	intendedFindings["phiLoop"] = []intended{
		{Kind: fair.Inflection, X: 95, Y: 197, Dir: 75, Reason: tip},
		{Kind: fair.Inflection, X: 106, Y: 279, Dir: 86, Reason: tip},
	}
	tips := "the tips of ω turn in at the x-height as those of υ do, so that each edge of both sides inflects once below them"
	intendedFindings["omega"] = []intended{
		{Kind: fair.Inflection, X: 84, Y: 372, Dir: -134, Reason: tips},
		{Kind: fair.Inflection, X: 114, Y: 276, Dir: 76, Reason: tips},
		{Kind: fair.Inflection, X: 590, Y: 363, Dir: -113, Reason: tips},
		{Kind: fair.Inflection, X: 559, Y: 265, Dir: -105, Reason: tips},
	}
	addGreek("delta", 'δ', delta)
	addGreek("epsilon", 'ε', epsilon)
	addGreek("phiLoop", 'φ', phiLoop)
	addGreek("omega", 'ω', omega)
	addGreek("beta", 'β', beta)
	addGreek("gamma", 'γ', gamma)
}

// delta draws δ: the bowl of o, smaller, and the neck, a stroke of the pen
// which leaves the top of the bowl from inside, rises to the left, turns up
// at its left extreme and over to the right into the crest at the height of
// the ascender, from which the ball hangs on the right, by the rule of the
// spiral, as on c.
func delta(p Params) letter {
	l := p.SideRoundLeft
	r := l + p.Gk.DeltaWidth
	cx := (l + r) / 2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	neck := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(cx+p.Gk.DeltaStart, top-p.nibWidth(0)/2), Dir: p.Gk.DeltaStartDir, K: auto, W: p.Gk.HiddenStart * p.Thin, DW: auto},
			{P: pt(cx-p.Gk.DeltaNeck, p.Gk.DeltaNeckY), Dir: 90, K: auto, W: auto, DW: 0},
			{P: pt(cx+p.Gk.DeltaCrest, p.Ascender+p.Overshoot-p.BallStalk/2), Dir: 0, K: auto, W: p.BallStalk, DW: 0},
		},
		Start: hidden,
		End:   p.ballAtDepth(p.Gk.DeltaBall),
	}
	return letter{
		strokes: []stroke.Stroke{p.roundBowl(l, r, bottom, top, oTurn), neck},
		advance: r + p.SideRoundRight,
	}
}

// epsilon draws ε, the open form: the upper stroke, which starts in the ball
// at the upper right as c does, runs over the top and down the left side
// of the smaller upper round and ends at the waist in a cut, and the lower
// stroke, which starts inside the end of the upper one, runs left round the
// larger lower round and out into the end of the exit, as c does.
func epsilon(p Params) letter {
	l := p.SideRoundLeft
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	waist := p.Gk.EpsWaist * p.XHeight
	ux := l + p.Gk.EpsUpperIn + p.Gk.EpsUpper/2
	lx := l + p.Gk.EpsLower/2
	r := p.Ball / 2
	upper := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(ux+cCrest, top-p.BallStalk/2), Dir: 180, K: auto, W: p.BallStalk, DW: 0},
			node(l+p.Gk.EpsUpperIn+wv/2, (top+waist)/2, 270),
			{P: pt(lx+p.Gk.EpsSpur, waist), Dir: p.Gk.EpsWaistDir, K: auto, W: p.Gk.EpsWaistW * p.Thin, DW: 0},
		},
		Start: p.ballAtDepth(p.Gk.EpsBallLow - r),
		End:   round,
	}
	lower := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(lx+p.Gk.EpsSpur-p.Gk.EpsJoin, waist), Dir: 180 + p.Gk.EpsWaistDir, K: auto, W: p.Gk.HiddenStart * p.Gk.EpsWaistW * p.Thin, DW: auto},
			node(l+wv/2+oSide[0], (waist+bottom)/2, 270),
			node(lx-oTurn, bottom+wh/2, 0),
			p.exitEnd(l + p.Gk.EpsTip),
		},
		Start: hidden,
		End:   round,
	}
	return letter{strokes: []stroke.Stroke{upper, lower}, advance: l + p.Gk.EpsTip + p.SideRoundRight}
}

// phiLoop draws φ, the open form: one stroke of the pen which starts at the
// upper left, runs down the left side and round the bottom of a round, up
// its right side and over its top to the left, and turns down into a
// straight stem through the middle to the depth of the tails, where it ends
// in a flat cut.
func phiLoop(p Params) letter {
	l := p.SideRoundLeft
	r := l + p.Gk.PhiLoopWidth
	cx := (l + r) / 2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	c := cx + p.Gk.PhiLoopStem
	s := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(l+p.Gk.PhiLoopStart, p.XHeight-p.ExitEndW/2), Dir: p.Gk.PhiLoopStartDir, K: auto, W: p.ExitEndW, DW: auto},
			node(l+wv/2+oSide[0], p.Gk.PhiLoopSide*p.XHeight, 270),
			node(cx-oTurn, bottom+wh/2, 0),
			node(r-wv/2-oSide[1], (top+bottom)/2, 90),
			node(c+p.Gk.PhiLoopTop, top-wh/2, 180),
			p.stemNode(c, p.Gk.PhiLoopJoin),
			p.stemNode(c, -p.TailDepth),
		},
		Start: round, End: foot,
	}
	return letter{strokes: []stroke.Stroke{s}, advance: r + p.SideRoundRight}
}

// omega draws ω: two strokes of the pen.  The left one starts at a tip at
// the upper left, runs down the left side and round the first bottom and
// rises to the waist in the middle; the right one starts inside the end of
// the left one, runs down round the second bottom, up the right side and
// ends at a tip at the upper right, bent back to the left.
func omega(p Params) letter {
	l := p.SideRoundLeft
	r := l + p.Gk.OmegaWidth
	m := (l + r) / 2
	bottom := -p.Overshoot
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	waist := p.Gk.OmegaWaist * p.XHeight
	// The right stroke starts on the centre line of the left one, OmegaJoin
	// back from its end along the direction of the end.
	a := p.Gk.OmegaWaistDir * math.Pi / 180
	join := pt(m-p.Gk.OmegaJoin*math.Cos(a), waist-p.Gk.OmegaJoin*math.Sin(a))
	left := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(l+p.Gk.OmegaTip, p.XHeight-p.ExitEndW/2), Dir: 180 + p.Gk.OmegaTipDir, K: 0, W: p.ExitEndW, DW: auto},
			node(l+wv/2+oSide[0], p.Gk.OmegaSide*p.XHeight, 270),
			node((l+m)/2, bottom+wh/2, 0),
			{P: pt(m, waist), Dir: p.Gk.OmegaWaistDir, K: auto, W: auto, DW: 0},
		},
		Start: round, End: round,
	}
	right := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: join, Dir: -p.Gk.OmegaWaistDir, K: auto, W: p.Gk.HiddenStart * wv, DW: auto},
			node((m+r)/2, bottom+wh/2, 0),
			node(r-wv/2-oSide[1], p.Gk.OmegaSide*p.XHeight, 90),
			{P: pt(r-p.Gk.OmegaTip, p.XHeight-p.ExitEndW/2), Dir: p.Gk.OmegaTipDir, K: auto, W: p.ExitEndW, DW: auto},
		},
		Start: hidden, End: round,
	}
	return letter{strokes: []stroke.Stroke{left, right}, advance: r + p.SideRoundRight}
}

// beta draws β: the stem, from the top of the upper bowl down to the depth
// of the tails in a flat cut, which the stroke of the upper bowl begins:
// from the waist inside the stroke of the lower bowl it runs right and up
// round the upper bowl, over its top and down into the stem; and the lower
// bowl, which leaves the waist, runs right and down round the larger lower
// round, along the bottom and back into the stem.
func beta(p Params) letter {
	e := p.SideStemLeft + p.Gk.BetaLeft
	c := e + p.Stem/2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	top := p.Ascender + p.Overshoot
	waist := p.Gk.BetaWaist
	ur := e + p.Stem + p.Gk.BetaUpper
	lr := e + p.Stem + p.Gk.BetaLower
	upper := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(c, waist), Dir: p.Gk.BetaWaistDir, K: auto, W: p.Gk.BetaWaistW * p.Thin, DW: auto},
			node(ur-wv/2-oSide[1], (top+waist)/2, 90),
			node((c+ur)/2+p.Gk.BetaTopShift, top-wh/2, 180),
			p.stemNode(c, p.Gk.BetaJoin),
			p.stemNode(c, -p.TailDepth),
		},
		Start: hidden, End: foot,
	}
	lower := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(c, waist-p.Gk.BetaJoinBack), Dir: -p.Gk.BetaWaistDir, K: auto, W: p.Gk.BetaWaistW * p.Thin, DW: auto},
			node(lr-wv/2-oSide[1], p.Gk.BetaLowSide*waist, -90),
			node((c+lr)/2, -p.Overshoot+wh/2, 180),
			{P: pt(c, p.Gk.BetaClose), Dir: p.Gk.BetaCloseDir, K: auto, W: p.Thin, DW: 0},
		},
		Start: hidden, End: hidden,
	}
	return letter{strokes: []stroke.Stroke{upper, lower}, advance: lr + p.SideRoundRight}
}

// gamma draws γ: one stroke of the pen which begins with the entry at the
// x-height, runs down to the right across the baseline into the loop, round
// the loop to the left and up, and up to the right as a straight stroke
// through the crossing to the x-height, where it ends in a cut.
func gamma(p Params) letter {
	e := p.SideStemLeft
	m := e + p.Gk.GammaCross
	wh := p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	s := stroke.Stroke{Nib: p.nib(), Nodes: p.entry(e, p.XHeight)[:2], Start: round, End: round}
	loop := -p.TailDepth
	s.Nodes = append(s.Nodes,
		stroke.Node{P: pt(m-p.Gk.GammaArm, p.Gk.GammaArmY), Dir: p.Gk.GammaArmDir, K: auto, W: auto, DW: 0},
		node(m+p.Gk.GammaLoopR, p.Gk.GammaLoopSide*loop, -90),
		node(m, loop+wh/2, 180),
		node(m-p.Gk.GammaLoopL, p.Gk.GammaLoopSide*loop, 90),
		stroke.Node{P: pt(m+p.Gk.GammaRise, p.Gk.GammaRiseY), Dir: p.Gk.GammaRiseDir, K: 0, W: p.Gk.GammaRiseW * p.Stem, DW: 0},
		stroke.Node{P: pt(e+p.Gk.GammaWidth, p.XHeight-p.ExitEndW/2), Dir: p.Gk.GammaRiseDir, K: auto, W: p.ExitEndW, DW: 0},
	)
	return letter{strokes: []stroke.Stroke{s}, advance: e + p.Gk.GammaWidth + p.Gk.GammaSide}
}

// The proportions of δ, ε, φ, ω, β and γ, in the upright description.
const ()
