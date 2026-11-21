// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The currency signs dollar, sterling and euro.  All three stand at the figure
// height on the common advance of the figures (Figures.Advance).

// dollarParams returns p with the values of the S of the dollar in place of
// those of the capital S, at the figure height, for the curve code of S
// (capss.go).
func dollarParams(p design.Params) design.Params {
	s := p.Sym
	q := p
	q.CapHeight, q.OvershootUpper = p.Figures.Height, p.OvershootLower
	g := &q.Caps.CGSJQ
	g.SLowerWidth, g.SUpperWidth, g.SUpperShift = s.DollarLowerWidth, s.DollarUpperWidth, s.DollarUpperShift
	g.SSpineThick, g.SSpineAngle = s.DollarSpine, s.DollarSpineAngle
	g.SArcThick, g.SUpperArcThick, g.SSideThick = s.DollarArc, s.DollarUpperArc, s.DollarSideThick
	g.SUpperDrop, g.SLowerDrop = s.DollarUpperDrop, s.DollarLowerDrop
	g.SSideLeft = math.Round((p.Figures.Advance - s.DollarLowerWidth) / 2)
	g.SSideRight = p.Figures.Advance - s.DollarLowerWidth - g.SSideLeft
	return q
}

// dollar draws the dollar: the capital S drawn at the figure height with the
// values of the dollar (narrower and lighter than S, with its beaks), and a
// vertical stroke DollarStroke thick through the middle of its ink, from
// DollarBottom to DollarTop, which runs through the S and shows in both
// counters.  The outline is the union of the two (symunion.go).
//
// Side kinds: open; the S is centred on the advance of the figures.
func dollar(p design.Params) (outline.Path, float64) {
	s := p.Sym
	q := dollarParams(p)
	f := newCapSFrame(q)
	var path outline.Path
	path.MoveTo(f.C.X+f.LowerE.Inner[0].P0.X, f.C.Y+f.LowerE.Inner[0].P0.Y)
	f.circuit(&path, f.LowerE, 1)
	f.circuit(&path, f.UpperE, -1)
	cx := math.Round(q.Caps.CGSJQ.SSideLeft + s.DollarLowerWidth/2)
	bar := rectChain(cx-s.DollarStroke/2, s.DollarBottom, cx+s.DollarStroke/2, s.DollarTop)
	return chainsPath(union(append(pathChains(path), bar)...)...), p.Figures.Advance
}

// sterling draws the pound sterling: a stem SterStem thick which rises from a
// flat base, leaning a little to the right, into a hook over the top which is
// thin at the top (Thin) and ends in the ball terminal of the lowercase at the
// upper right (ball.go), not mirrored, as on r; a bar SterBar thick across the
// stem at SterBarY; and a base across the whole ink, a stroke which is SterBase
// thick at its thickest, thinner at its upturned ends, with a gentle wave.  The
// outline is the union of the stroke, the bar and the base.
//
// Side kinds: open; the ink is centred on the advance of the figures.
func sterling(p design.Params) (outline.Path, float64) {
	s := p.Sym
	w := s.SterInk
	x0 := math.Round((p.Figures.Advance - w) / 2)
	x1 := x0 + w
	top := p.Figures.Height + p.OvershootLower
	sts := []station{
		{X: x0 + 0.60*w, Y: top - p.Figures.Thin/2, Dir: 180, W: p.Figures.Thin},
		{X: x0 + 0.33*w, Y: 0.68 * top, Dir: -90, W: s.SterStem},
		{X: x0 + 0.30*w, Y: s.SterBarY, Dir: -95, W: s.SterStem},
		{X: x0 + 0.24*w, Y: s.SterBase / 2, Dir: -102, W: 0.92 * s.SterStem},
	}
	right, left := strokeOpen(sts)
	r := p.BallDiameter / 2
	ball := ballStart("sterling", p, right, left, x1-0.10*w-r, span(0.70*top, 0.82*top, 0.01*top), span(-40, 60, 2), false)
	stroke := join(right, lineChain(right.end(), left.start()), left, ball)
	bx := x0 + 0.04*w
	bar := rectChain(bx, s.SterBarY-s.SterBar/2, bx+s.SterBarLen, s.SterBarY+s.SterBar/2)
	// The base is a stroke with a gentle wave: it leaves the foot of the
	// stem as a thin upturned end, dips to its full thickness SterBase, which
	// touches the baseline, a third of the way along, and rises again to the right
	// to a thin end which turns up a little.  Both edges are curved.
	bt := s.SterBase
	bsts := []station{
		{X: x0 + 0.40*bt, Y: 0.95 * bt, Dir: -32, W: 0.70 * bt},
		{X: x0 + 0.30*w, Y: bt / 2, Dir: 0, W: bt},
		{X: x0 + 0.68*w, Y: 0.72 * bt, Dir: 10, W: 0.80 * bt},
		{X: x1 - 0.45*bt, Y: 1.25 * bt, Dir: 36, W: 0.62 * bt},
	}
	base := strokeChain(bsts)
	return chainsPath(union(stroke, bar, base)...), p.Figures.Advance
}

// euro draws the euro: the capital C (capc.go) drawn at the figure height,
// EuroRoundW wide, with the round side EuroThick and the thin parts of the
// capitals, its upper beak cut vertically EuroBeakX from the left extreme and
// its lower terminal ending on the same vertical, as on C; and two horizontal
// bars EuroBar thick, EuroBarGap apart between their centres about the middle
// of the figure height, which reach EuroReach beyond the left of the round and
// end in the counter.  The outline is the union of the round and the bars.
//
// Side kinds: open; the ink is centred on the advance of the figures.
func euro(p design.Params) (outline.Path, float64) {
	s := p.Sym
	x0 := math.Round((p.Figures.Advance - s.EuroInk) / 2)
	q := p
	q.CapHeight, q.OvershootUpper = p.Figures.Height, p.OvershootLower
	c := &q.Caps
	c.SideRoundLeft, c.OWidth, c.Round = x0+s.EuroReach, s.EuroRoundW, s.EuroThick
	c.CGSJQ.BeakX, c.CGSJQ.BeakEnd = s.EuroBeakX, s.EuroBeakEnd
	c.CGSJQ.CTailHeight *= p.Figures.Height / p.CapHeight
	f := capRound(q)
	round := join(capRoundLeft(f), capCTail(q, f), capRoundLeftInner(f), capUpperBeak(q, f))
	bar := func(y float64) chain { return rectChain(x0, y-s.EuroBar/2, x0+s.EuroBarEnd, y+s.EuroBar/2) }
	return chainsPath(union(round, bar(p.Figures.Height/2-s.EuroBarGap/2), bar(p.Figures.Height/2+s.EuroBarGap/2))...), p.Figures.Advance
}
