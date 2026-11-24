// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// The Greek letters of an open round: ϵ, and ς, ζ and ξ, whose round runs on
// below the baseline into a hook.

func init() {
	addGreek("epsilonLunate", 'ϵ', epsilonLunate)
	addGreek("sigmaFinal", 'ς', sigmaFinal)
	addGreek("zeta", 'ζ', zeta)
	addGreek("xi", 'ξ', xi)
	addGreek("rhoVar", 'ϱ', rhoVar)
}

// epsilonLunate draws ϵ: the round of c, opened on the right, which starts
// in a cut at the upper right, the end of the exit turned upside down, and
// ends in the end of the exit; and the bar, a straight stroke of the
// hairline from inside the left side to the right, shorter than the round.
func epsilonLunate(p Params) letter {
	l := p.SideRoundLeft
	cx := l + p.Gk.LunateWidth/2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	cy := (top + bottom) / 2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	tip := l + p.Gk.LunateTip
	round := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(tip-p.ExitEndW/2, top-p.Gk.LunateTop+p.ExitEndW/2), Dir: 180 - p.ExitAngle, K: auto, W: p.ExitEndW, DW: 0},
			node(cx+oTurn, top-wh/2, 180),
			node(l+wv/2+oSide[0], cy, 270),
			node(cx-oTurn, bottom+wh/2, 0),
			p.exitEnd(tip),
		},
		Start: stroke.End{Kind: stroke.Cut}, End: stroke.End{Kind: stroke.Cut},
	}
	y := p.Gk.LunateBar * p.XHeight
	bar := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(l+wv/2, y), Dir: 0, W: p.Hair},
			{P: pt(l+p.Gk.LunateBarEnd-p.ExitEndW/2, y), Dir: 0, W: p.Hair},
		},
		Start: hidden, End: stroke.End{Kind: stroke.Cut},
	}
	return letter{strokes: []stroke.Stroke{round, bar}, advance: tip + p.SideRoundRight}
}

// hookNodes returns the nodes of the hook below the baseline in which ς, ζ
// and ξ end, after the left side of their round about cx: the round turns
// into a straight run down to the right across the baseline, where the
// stroke turns the other way, and the hook turns down at p.Gk.HookSide right
// of cx and curls back to the left at the depth of the tails, where it ends
// in a cut.
func (p Params) hookNodes(cx float64) []stroke.Node {
	wh := p.nibWidth(0)
	a := pt(cx+p.Gk.HookRunX, p.Gk.HookRunY)
	b := pt(a.X+p.Gk.HookRun*cosd(p.Gk.HookRunDir), a.Y+p.Gk.HookRun*sind(p.Gk.HookRunDir))
	return []stroke.Node{
		{P: a, Dir: p.Gk.HookRunDir, K: 0, W: auto, DW: 0},
		{P: b, Dir: p.Gk.HookRunDir, K: 0, W: auto, DW: 0},
		{P: pt(cx+p.Gk.HookSide, p.Gk.HookSideY), Dir: -90, K: auto, W: auto, DW: 0},
		{P: pt(cx+p.Gk.HookLow, -p.TailDepth+wh/2), Dir: 180, K: auto, W: auto, DW: 0},
		{P: pt(cx+p.Gk.HookEnd, -p.TailDepth+p.Gk.HookEndUp), Dir: 180 - p.Gk.HookEndDir, K: auto, W: p.ExitEndW, DW: 0},
	}
}

// sigmaFinal draws ς: the top of c with its ball, the left side and the
// bottom of its round, which runs on into the hook.
func sigmaFinal(p Params) letter {
	l := p.SideRoundLeft
	cx := l + p.CBowlWidth/2
	top := p.XHeight + p.Overshoot
	wv := p.nibWidth(90)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	low := p.Gk.SigmaFinalLow
	s := stroke.Stroke{
		Nib: p.nib(),
		Nodes: append([]stroke.Node{
			{P: pt(cx+cCrest, top-p.BallStalk/2), Dir: 180, K: auto, W: p.BallStalk, DW: 0},
			node(l+wv/2+oSide[0], (top+low)/2, 270),
		}, p.hookNodes(cx)...),
		Start: p.ballAtDepth(p.CBallLow - p.Ball/2),
		End:   stroke.End{Kind: stroke.Cut},
	}
	return letter{strokes: []stroke.Stroke{s}, advance: p.CWidth}
}

// zeta draws ζ: one stroke of the pen which starts in a cut at the upper
// right at the height of the ascender, runs left over a short top, down
// the long left side and round the bottom, and runs on into the hook.
func zeta(p Params) letter {
	l := p.SideRoundLeft
	cx := l + p.Gk.ZetaWidth/2
	top := p.Ascender
	wv := p.nibWidth(90)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	s := stroke.Stroke{
		Nib: p.nib(),
		Nodes: append([]stroke.Node{
			{P: pt(l+p.Gk.ZetaTop, top-p.ExitEndW/2), Dir: 180 + p.Gk.ZetaTopDir, K: auto, W: p.ExitEndW, DW: 0},
			node(l+wv/2+oSide[0], p.Gk.ZetaSide*p.XHeight, 270),
		}, p.hookNodes(cx)...),
		Start: stroke.End{Kind: stroke.Cut},
		End:   stroke.End{Kind: stroke.Cut},
	}
	return letter{strokes: []stroke.Stroke{s}, advance: l + p.Gk.ZetaWidth + p.SideRoundRight}
}

// xi draws ξ: the upper stroke, which starts in a cut at the upper right at
// the height of the ascender, runs left round the small upper lobe and ends
// at the waist in a cut, and the lower stroke, which starts inside the end
// of the upper one, runs left round the larger lower lobe and on into the
// hook, as ε is made of two rounds.
func xi(p Params) letter {
	l := p.SideRoundLeft
	cx := l + p.Gk.XiWidth/2
	wv := p.nibWidth(90)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	waist := p.Gk.XiWaist
	low := p.Gk.SigmaFinalLow
	upper := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(l+p.Gk.XiTop, p.Ascender-p.ExitEndW/2), Dir: 180 + p.Gk.ZetaTopDir, K: auto, W: p.ExitEndW, DW: 0},
			node(l+p.Gk.XiUpperIn+wv/2, (p.Ascender+waist)/2, 270),
			{P: pt(cx+p.Gk.XiSpur, waist), Dir: p.Gk.XiWaistDir, K: auto, W: p.Gk.XiWaistW * p.Thin, DW: 0},
		},
		Start: stroke.End{Kind: stroke.Cut},
		End:   stroke.End{Kind: stroke.Cut},
	}
	lower := stroke.Stroke{
		Nib: p.nib(),
		Nodes: append([]stroke.Node{
			{P: pt(cx+p.Gk.XiSpur-p.Gk.XiJoin, waist), Dir: 180 + p.Gk.XiWaistDir, K: auto, W: p.Gk.HiddenStart * p.Gk.XiWaistW * p.Thin, DW: auto},
			node(l+wv/2+oSide[0], (waist+low)/2, 270),
		}, p.hookNodes(cx)...),
		Start: hidden,
		End:   stroke.End{Kind: stroke.Cut},
	}
	return letter{strokes: []stroke.Stroke{upper, lower}, advance: l + p.Gk.XiWidth + p.SideRoundRight}
}

// rhoVar draws ϱ: the ρ whose left side runs down below the baseline and
// curls to the right into a hook at the depth of the tails, instead of
// running straight into a stem.
func rhoVar(p Params) letter {
	l := p.SideRoundLeft
	r := l + p.Gk.RhoWidth
	cx := (l + r) / 2
	top, bottom := p.XHeight+p.Overshoot, -p.Overshoot
	c := l + p.across(p.Stem)/2 + p.Gk.RhoStemIn
	wh := p.nibWidth(0)
	nodes := p.roundNodes(l, r, bottom, top, 0, oTurn)
	s := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(c, p.Gk.RhoClose), Dir: p.Gk.RhoCloseDir, K: auto, W: p.Thin, DW: 0},
			nodes[3], nodes[0], nodes[1],
			{P: pt(l+p.nibWidth(90)/2+oSide[0], p.Gk.RhoVarSide), Dir: 270, K: auto, W: auto, DW: 0},
			{P: pt(cx-p.Gk.RhoVarLow, -p.TailDepth+wh/2), Dir: 0, K: auto, W: auto, DW: 0},
			{P: pt(cx+p.Gk.RhoVarEnd, -p.TailDepth+p.Gk.RhoVarEndUp), Dir: p.Gk.RhoVarEndDir, K: auto, W: p.ExitEndW, DW: 0},
		},
		Start: hidden, End: stroke.End{Kind: stroke.Cut},
	}
	return letter{strokes: []stroke.Stroke{s}, advance: r + p.SideRoundRight}
}

// The proportions of ϵ, ς and ζ, in the upright description.
const ()
