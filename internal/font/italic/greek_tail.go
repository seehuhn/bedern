// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

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

// zeta draws ζ: one stroke of the pen which starts in a round end at the
// top, runs down into a long, low loop which it crosses at its left end,
// turns back at the right end of the loop, leaves it at the crossing down to
// the left in the back and runs round the bottom into the hook.  The centre
// line turns to the left from the start to the hook.
func zeta(p Params) letter { return p.zetaXi(false) }

// xi draws ξ: the start and the loop of ζ, then a bow down to the left and
// back to the right into a second loop at the waist, made as the first, and
// from it the back of ζ round the bottom into the hook.
func xi(p Params) letter { return p.zetaXi(true) }

// zetaXi draws ζ, or ξ if xi is set.
func (p Params) zetaXi(xi bool) letter {
	g := p.Gk
	l := p.SideRoundLeft
	cx := l + g.ZetaWidth/2
	k := p.slant()
	// node returns a node whose abscissa and direction are given after the
	// shear.
	node := func(x, y, dir, w float64, hairpin bool) stroke.Node {
		d := math.Atan2(sind(dir), cosd(dir)-k*sind(dir)) * 180 / math.Pi
		return stroke.Node{P: pt(x-k*y, y), Dir: d, K: auto, W: w, DW: auto, Hairpin: hairpin}
	}
	wh, wv := p.nibWidth(0), p.nibWidth(90)
	d := wh + g.ZetaEye // the distance of the strands of a loop
	loop := func(xc, yc, length, in, out float64) []stroke.Node {
		return p.loopNodes(pt(xc, yc), 0, length, g.ZetaEye, g.ZetaTurnIn, wh, g.ZetaLoopW, g.ZetaLoopTaper,
			g.ZetaTurnDir, g.ZetaCrossIn, g.ZetaCrossOut, in, out)
	}
	sx := l + wv/2 + oSide[0]
	xc := sx + k*g.ZetaSide*p.XHeight + g.ZetaCross
	side := g.ZetaSide
	if xi {
		xc += g.XiHead
		side = g.XiSide
	}
	start := node(xc+g.ZetaStart, p.Ascender-p.ExitEndW/2+g.ZetaStartRise, g.ZetaStartDir, wh*g.ZetaStartW, false)
	start.DW = 0
	nodes := append([]stroke.Node{start}, loop(xc, g.ZetaLoopY+d/2, g.ZetaLoop, g.ZetaCrossShift, 0)...)
	if xi {
		nodes = append(nodes, stroke.Node{P: pt(sx+g.XiBowX, g.XiBowY), Dir: g.XiBowDir, K: auto, W: wv * g.XiBowW, DW: auto})
		nodes = append(nodes, loop(xc+g.XiCross, g.XiWaist, g.XiLoop, g.XiCrossShift, g.XiCrossShift)...)
	}
	nodes = append(nodes, stroke.Node{P: pt(sx+g.ZetaSideX, side*p.XHeight), Dir: g.ZetaSideDir, K: auto, W: wv * g.ZetaSideW, DW: auto})
	hook := p.hookNodes(cx)
	hook[0].W, hook[1].W = wv*g.ZetaRunW, wv*g.ZetaRunW
	hook[1].DW, hook[2].DW = auto, auto
	hook[2].P = pt(hook[2].P.X+g.ZetaHookSide, hook[2].P.Y+g.ZetaHookSideY)
	hook[3].P.X += g.ZetaHookLow
	if g.ZetaHookSideW > 0 {
		hook[2].W = wv * g.ZetaHookSideW
	}
	s := stroke.Stroke{
		Nib:   p.nib(),
		Nodes: append(nodes, hook...),
		Start: stroke.End{Kind: stroke.Round},
		End:   stroke.End{Kind: stroke.Cut},
	}
	return letter{strokes: []stroke.Stroke{s}, advance: l + g.ZetaWidth + p.SideRoundRight + g.ZetaOver}
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
