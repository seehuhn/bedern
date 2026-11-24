// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"

	"seehuhn.de/go/bedern/internal/stroke"
)

// The Greek letters of diagonal strokes: λ, ν, χ and ψ.

func init() {
	addGreek("lambda", 'λ', lambda)
	addGreek("nu", 'ν', nu)
	addGreek("chi", 'χ', chi)
	addGreek("psi", 'ψ', psi)
}

// lambda draws λ: the long stroke, which begins with the entry at the
// ascender, turns over and runs straight down to the right against the
// slant to a flat cut on the baseline, and the leg, a straight stroke of
// p.Gk.LambdaLegW stems which leaves the long stroke from inside and runs
// down to the left to a flat cut on the baseline.
func lambda(p Params) letter {
	e := p.SideStemLeft + p.Gk.LambdaLeft
	c := e + p.Stem/2
	top := p.XHeight + p.AscenderEntryRise
	d := p.Gk.LambdaDir * math.Pi / 180
	t := pt(math.Cos(d), math.Sin(d))
	y1, y2 := p.Gk.LambdaHigh, 0.0
	x1 := c + p.Gk.LambdaOut
	x2 := x1 + (y2-y1)*t.X/t.Y
	s := stroke.Stroke{Nib: p.nib(), Nodes: p.entry(e, top)[:2], Start: round, End: foot}
	s.Nodes = append(s.Nodes,
		stroke.Node{P: pt(x1, y1), Dir: p.Gk.LambdaDir, K: 0, W: p.across(p.Stem), DW: 0},
		stroke.Node{P: pt(x2, y2), Dir: p.Gk.LambdaDir, K: 0, W: p.across(p.Stem), DW: 0},
	)
	// The leg leaves the long stroke at p.Gk.LambdaLegY.
	ly := p.Gk.LambdaLegY
	lx := x1 + (ly-y1)*t.X/t.Y
	ld := p.Gk.LambdaLegDir * math.Pi / 180
	end := pt(lx-ly*math.Cos(ld)/math.Sin(ld), 0)
	lw := p.Gk.LambdaLegW * p.Stem
	leg := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(lx, ly), Dir: 180 + p.Gk.LambdaLegDir, W: lw},
			{P: end, Dir: 180 + p.Gk.LambdaLegDir, W: lw},
		},
		Start: hidden,
		End:   stroke.End{Kind: stroke.Cut, Line: true, Angle: 0},
	}
	return letter{strokes: []stroke.Stroke{s, leg}, advance: x2 + p.Stem/2 - p.shift() + p.Gk.LambdaSide}
}

// nu draws ν: the entry and the stem down to the baseline, cut along the
// direction of the rise so that the foot is a point, and the rise, a stroke
// of the hairline which leaves the foot from inside the stem and rises to
// the right, bending back to the left at the x-height, where it ends in a
// cut.
func nu(p Params) letter {
	e := p.SideStemLeft
	c := e + p.Stem/2
	stem := stroke.Stroke{Nodes: p.entry(e, p.XHeight), Start: round}
	stem.Nodes = append(stem.Nodes, p.stemNode(c, p.Gk.NuFoot))
	stem.End = stroke.End{Kind: stroke.Cut, Line: true, Angle: p.Gk.NuCut}
	r := c + p.Gk.NuWidth
	w := p.Gk.NuRiseW * p.Hair
	rise := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: pt(c, p.Gk.NuFoot+w), Dir: p.Gk.NuRiseDir, K: 0, W: w, DW: 0},
			{P: pt(r, p.Gk.NuSide*p.XHeight), Dir: 90, K: auto, W: p.Gk.NuSideW * p.Stem, DW: auto},
			{P: pt(r-p.Gk.NuBack, p.XHeight-p.ExitEndW/2), Dir: p.Gk.NuEndDir, K: auto, W: p.ExitEndW, DW: 0},
		},
		Start: hidden, End: round,
	}
	return letter{strokes: []stroke.Stroke{stem, rise}, advance: r + p.Stem/2 + p.Gk.NuSideRight}
}

// chi draws χ: the heavy stroke, which starts in a hook at the upper left,
// turns over and runs straight down to the right against the slant to the
// depth of the tails, and the light stroke, a straight stroke of the
// hairline from the x-height at the right down to the left, both cut flat.
func chi(p Params) letter {
	l := p.SideRoundLeft + p.Gk.ChiLeft
	d := p.Gk.ChiDir * math.Pi / 180
	t := pt(math.Cos(d), math.Sin(d))
	w := p.across(p.Stem)
	top := pt(l+p.Gk.ChiHook, p.XHeight+p.Overshoot-p.EntryTopW/2)
	y1, y2 := p.Gk.ChiHigh, -p.TailDepth+w/2
	x1 := top.X + p.Gk.ChiRun
	x2 := x1 + (y2-y1)*t.X/t.Y
	heavy := stroke.Stroke{
		Nib: p.nib(),
		Nodes: []stroke.Node{
			{P: pt(l+p.EndW/2, p.EntryLow), Dir: p.EntryStartDir, K: auto, W: p.EndW, DW: auto},
			{P: top, Dir: 0, K: auto, W: p.EntryTopW, DW: auto},
			{P: pt(x1, y1), Dir: p.Gk.ChiDir, K: 0, W: w, DW: 0},
			{P: pt(x2, y2), Dir: p.Gk.ChiDir, K: 0, W: w, DW: 0},
		},
		Start: round, End: round,
	}
	lw := p.Gk.ChiThin * p.Stem
	ld := p.Gk.ChiThinDir * math.Pi / 180
	my := p.Gk.ChiCross
	mx := x1 + (my-y1)*t.X/t.Y
	a := pt(mx+(p.XHeight-my)*math.Cos(ld)/math.Sin(ld), p.XHeight)
	b := pt(mx-(my+p.TailDepth)*math.Cos(ld)/math.Sin(ld), -p.TailDepth)
	light := stroke.Stroke{
		Nodes: []stroke.Node{
			{P: a, Dir: 180 + p.Gk.ChiThinDir, W: lw},
			{P: b, Dir: 180 + p.Gk.ChiThinDir, W: lw},
		},
		Start: stroke.End{Kind: stroke.Cut, Line: true, Angle: 0},
		End:   stroke.End{Kind: stroke.Cut, Line: true, Angle: 0},
	}
	k := p.slant()
	right := math.Max(a.X+lw/2+k*(a.Y-p.ShearHeight), x2+w/2+k*(y2-p.ShearHeight))
	return letter{strokes: []stroke.Stroke{heavy, light}, advance: right + p.Gk.ChiSide}
}

// psi draws ψ: the υ, wider, and a straight stem through the middle of its
// cup from p.Gk.PsiTop down to the depth of the tails, in flat cuts.
func psi(p Params) letter {
	l := p.upsilonOf(p.Gk.PsiWidth)
	c := p.SideRoundLeft + p.Gk.UpsilonLeft + p.Gk.PsiWidth/2 + p.Gk.PsiStemShift
	stem := stroke.Stroke{
		Nodes: []stroke.Node{p.stemNode(c, p.Gk.PsiTop), p.stemNode(c, -p.TailDepth)},
		Start: flatTop, End: foot,
	}
	l.strokes = append(l.strokes, stem)
	return l
}

// The proportions of the diagonal letters, in the upright description.
const ()
