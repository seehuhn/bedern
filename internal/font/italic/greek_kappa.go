// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// The variant kappa ϰ.

func init() {
	addGreek("kappaVar", 'ϰ', kappaVar)
}

// kappaVar draws ϰ: the arm and the leg of κ on a curved stroke in place of
// its stem, which begins with the entry at the x-height, bows to the left
// and curls at the baseline to the left into a cut.
func kappaVar(p Params) letter {
	q := p
	q.AscenderEntryRise = 0
	q.KArmX = 150 // the arm of ϰ keeps its reach: a longer arm leaves the union with the curved stroke open
	l := k(q)
	e := p.SideStemLeft - p.HLeftLess
	c := e + p.Stem/2
	s := stroke.Stroke{Nib: p.nib(), Nodes: p.entry(e, p.XHeight)[:3], Start: round, End: stroke.End{Kind: stroke.Cut}}
	s.Nodes = append(s.Nodes,
		stroke.Node{P: pt(c-p.Gk.KappaVarBow, p.Gk.KappaVarBowY), Dir: -90, K: auto, W: auto, DW: 0},
		stroke.Node{P: pt(c-p.Gk.KappaVarEnd, p.ExitEndW/2+p.Gk.KappaVarEndUp), Dir: 180 + p.Gk.KappaVarEndDir, K: auto, W: p.ExitEndW + p.Gk.KappaVarEndW, DW: 0},
	)
	l.strokes[0] = s
	return l
}

// The proportions of ϰ, in the upright description.
