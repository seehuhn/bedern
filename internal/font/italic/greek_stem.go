// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"seehuhn.de/go/bedern/internal/fair"
	"seehuhn.de/go/bedern/internal/stroke"
)

// The Greek letters of stems and arches: ι, τ, π, η, μ, κ and υ, made of the
// stem, the entry, the exit, the arch and the turn of the italic.

func init() {
	intendedFindings["mu"] = []intended{
		{Kind: fair.Inflection, X: 136, Y: 72, Dir: 91, Reason: "the turn of μ leaves its stem downwards and bends the other way into its bottom: its outer edge inflects once below the crotch"},
	}
	tips := "the tips of υ turn in at the x-height; each edge of both sides inflects once below them"
	upsilonTips := []intended{
		{Kind: fair.Inflection, X: 384, Y: 318, Dir: -110, Reason: tips},
		{Kind: fair.Inflection, X: 350, Y: 214, Dir: -103, Reason: tips},
		{Kind: fair.Inflection, X: 105, Y: 196, Dir: 77, Reason: tips},
		{Kind: fair.Inflection, X: 113, Y: 298, Dir: 89, Reason: tips},
	}
	intendedFindings["upsilon"] = upsilonTips
	intendedFindings["psi"] = upsilonTips
	why := "the leg of κ, that of k, bows out from the stem and turns the other way into the exit: each edge inflects once"
	intendedFindings["kappa"] = []intended{
		{Kind: fair.Inflection, X: 259, Y: 118, Dir: -90, Reason: why},
		{Kind: fair.Inflection, X: 327, Y: 82, Dir: 81, Reason: why},
	}
	addGreek("iota", 'ι', iota)
	addGreek("tau", 'τ', tau)
	addGreek("pi", 'π', pi)
	addGreek("eta", 'η', eta)
	addGreek("mu", 'μ', mu)
	addGreek("kappa", 'κ', kappa)
	addGreek("upsilon", 'υ', upsilon)
}

// iota draws ι: the stem from a flat cut at the x-height down into the
// exit, without the entry and the dot of i.
func iota(p Params) letter {
	e := p.SideStemLeft - p.Gk.IotaLeftLess
	s := stroke.Stroke{Nodes: []stroke.Node{p.stemNode(e+p.Stem/2, p.XHeight)}, Start: flatTop, End: round}
	s.Nodes = append(s.Nodes, p.exit(e)...)
	return letter{strokes: []stroke.Stroke{s}, advance: e + p.Stem + p.SideStemRight}
}

// greekBar returns the bar of τ and π from l to r at the x-height: a stroke
// of the pen which starts low at the left, rises to the right into the top,
// and runs level to its end, where it is cut square; its top edge stands at
// the x-height.
func (p Params) greekBar(l, r float64) stroke.Stroke {
	w := p.Gk.BarW * p.Thin
	y := p.XHeight - w/2
	return stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(l, y-p.Gk.BarDrop), Dir: p.Gk.BarStartDir, K: auto, W: p.ExitEndW, DW: auto},
			{P: pt(l+p.Gk.BarRise, y), Dir: 0, K: 0, W: w, DW: 0},
			{P: pt(r-p.ExitEndW/2, y), Dir: 0, K: 0, W: w, DW: 0},
		},
		Start: round, End: round,
	}
}

// tau draws τ: the bar, and the stem, which starts inside the bar and runs
// down into the exit.
func tau(p Params) letter {
	l := p.SideRoundLeft
	e := l + p.Gk.TauStem
	st := stroke.Stroke{Nodes: []stroke.Node{p.stemNode(e+p.Stem/2, p.XHeight-p.Gk.BarW*p.Thin/2)}, Start: hidden, End: round}
	st.Nodes = append(st.Nodes, p.exit(e)...)
	r := l + p.Gk.TauWidth
	return letter{strokes: []stroke.Stroke{p.greekBar(l, r), st}, advance: r + p.SideRoundRight}
}

// pi draws π: the bar, the left leg, a stem from inside the bar down to a
// plain foot, and the right leg, a stem from inside the bar into the exit.
func pi(p Params) letter {
	l := p.SideRoundLeft
	e1 := l + p.Gk.PiLeft
	e2 := e1 + p.Stem + p.Gk.PiCounter
	top := p.XHeight - p.Gk.BarW*p.Thin/2
	left := stroke.Stroke{Nodes: []stroke.Node{p.stemNode(e1+p.Stem/2, top), p.stemNode(e1+p.Stem/2, 0)}, Start: hidden, End: foot}
	right := stroke.Stroke{Nodes: []stroke.Node{p.stemNode(e2+p.Stem/2, top)}, Start: hidden, End: round}
	right.Nodes = append(right.Nodes, p.exit(e2)...)
	r := e2 + p.Stem + p.Gk.PiOver
	return letter{strokes: []stroke.Stroke{p.greekBar(l, r), left, right}, advance: r + p.SideRoundRight}
}

// eta draws η: the n whose right stem runs on below the baseline to the
// depth of the tails and ends in a flat cut.
func eta(p Params) letter {
	e1 := p.SideStemLeft
	e2 := e1 + p.Stem + p.NCounter
	a := stroke.Stroke{Nib: p.nib(), Nodes: p.archNodes(e1, e2), Start: hidden, End: foot}
	a.Nodes = append(a.Nodes, p.stemNode(e2+p.Stem/2, -p.TailDepth))
	return letter{
		strokes: []stroke.Stroke{p.stem(e1, p.XHeight, false), a},
		advance: e2 + p.Stem + p.SideStemRight,
	}
}

// mu draws μ: the u whose left stem runs on straight below the baseline to
// the depth of the tails and ends in a flat cut; the turn leaves it from
// inside the stem at the node where it runs vertically in the italic.
func mu(p Params) letter {
	l := u(p)
	turn := l.strokes[0]
	k := len(p.entry(0, p.XHeight))
	c1 := p.SideStemLeft + p.Stem/2
	stem := stroke.Stroke{Nodes: append([]stroke.Node(nil), turn.Nodes[:k]...), Start: turn.Start, End: foot}
	stem.Nodes = append(stem.Nodes, p.stemNode(c1, -p.TailDepth))
	turn.Nodes = append([]stroke.Node{{P: pt(c1, p.Gk.MuStart), Dir: p.Gk.MuDir, K: auto, W: p.across(p.Stem) - p.Gk.MuInside, DW: auto}}, turn.Nodes[k+1:]...)
	turn.Start = hidden
	return letter{strokes: []stroke.Stroke{stem, turn, l.strokes[1]}, advance: l.advance}
}

// kappa draws κ: the k with its entry at the x-height instead of the
// ascender.
func kappa(p Params) letter {
	q := p
	q.AscenderEntryRise = 0
	return k(q)
}

// upsilon draws υ: one stroke of the pen from a tip at the upper left down
// the round left side, round the bottom and up the right side to the
// x-height, where it ends in a cut bent a little back to the left; it has
// neither the entry and the straight stem of u and v nor the ball of v.
func upsilon(p Params) letter { return p.upsilonOf(p.Gk.UpsilonWidth) }

// upsilonOf returns υ with the given width between its outer extremes.
func (p Params) upsilonOf(width float64) letter {
	l := p.SideRoundLeft + p.Gk.UpsilonLeft
	r := l + width
	cx := (l + r) / 2
	wv, wh := p.nibWidth(90), p.nibWidth(0)
	node := func(x, y, dir float64) stroke.Node {
		return stroke.Node{P: pt(x, y), Dir: dir, K: auto, W: auto, DW: 0}
	}
	s := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(l+p.Gk.UpsilonTip, p.XHeight-p.ExitEndW/2), Dir: 180 + p.Gk.UpsilonTipDir, K: auto, W: p.ExitEndW, DW: auto},
			node(l+wv/2+oSide[0], p.Gk.UpsilonSide*p.XHeight, 270),
			node(cx, -p.Overshoot+wh/2, 0),
			node(r-wv/2-oSide[1], p.Gk.UpsilonSide*p.XHeight, 90),
			{P: pt(r-p.Gk.UpsilonBack, p.XHeight-p.ExitEndW/2+p.Gk.UpsilonRaise), Dir: p.Gk.UpsilonEndDir, K: auto, W: p.ExitEndW, DW: auto},
		},
		Start: round, End: round,
	}
	return letter{strokes: []stroke.Stroke{s}, advance: r + p.SideRoundRight}
}

// The proportions of the stem letters, in the upright description.
const ()
