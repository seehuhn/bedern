// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The ties.
//
// A tie is an arc over two letters: a crescent between the upper halves of
// two ellipses, thick at the top and thin at its two ends, which are cut
// horizontally, as the breve turned over.  It is as thick as the breve at the
// top (Acc.BreveThick) and as thin as the breve at its ends (Acc.BreveThin).  The tie of the lowercase stands with its ends tieLift
// above the height of the lowercase accents (Acc.Gap above the x-height),
// the tie of the capitals as far above that of the capital accents
// (Acc.GapCap above the cap height).
//
// A tie is a spacing glyph which is set as an accent over the first of two
// letters: its advance is half that of n, or of H for the capitals, and its
// ink starts tieStart to the left of the middle of the advance and reaches
// beyond the advance on the right.  The rounder forms, with the suffix .new,
// are narrower arcs of the same height and weights, which stand in the
// middle of the advance of the figures.

// Shape constants of the ties.
const (
	tieWidth       = 380.0 // the ink width of the tie of the lowercase
	tieWidthCap    = 460.0 // that of the tie of the capitals
	tieNewWidth    = 320.0 // that of the rounder tie of the lowercase
	tieNewWidthCap = 400.0 // that of the rounder tie of the capitals
	tieHeight      = 125.0 // the ink height of the ties
	tieLift        = 48.0  // the ends above the height of the accents
	tieStart       = 18.0  // the left end of the ink to the left of the middle of the advance

	// newSuffix is the suffix of the name of the rounder form of a tie.
	newSuffix = ".new"
)

// tieList lists the ties and the two forms of the o with a rogate, in glyph
// order.
var tieList = []entry{
	{"tieaccentlowercase", privateUseRune("tieaccentlowercase"), sideOpen, sideTie, tieLower},
	{"tieaccentcapital", privateUseRune("tieaccentcapital"), sideOpen, sideTie, tieCapital},
	{"tieaccentlowercase" + newSuffix, 0, sideOpen, sideOpen, tieLowerNew},
	{"tieaccentcapital" + newSuffix, 0, sideOpen, sideOpen, tieCapitalNew},
	{"Orogate", privateUseRune("Orogate"), sideCapRound, sideCapRound, capORogate},
	{"orogate", privateUseRune("orogate"), sideRound, sideRound, oRogate},
}

// tieChain returns the tie w wide with its left end at the abscissa 0 and
// its ends on the height 0.
func tieChain(p design.Params, w float64) chain {
	a := p.Acc
	thick := a.BreveThick
	outer := newSuperShape(0, -tieHeight, w, tieHeight, degSuperness)
	inner := newSuperShape(a.BreveThin, -(tieHeight - thick), w-a.BreveThin, tieHeight-thick, degSuperness)
	// Counter-clockwise: from the right end over the top to the left end
	// along the outer edge, across the left end, back along the inner edge
	// and across the right end.
	c := chain{curveSeg(outer.cubic(0, true)), curveSeg(outer.cubic(1, true))}
	in := chain{curveSeg(inner.cubic(2, false)), curveSeg(inner.cubic(3, false))}
	return join(c, lineChain(c.end(), in.start()), in, lineChain(in.end(), c.start()))
}

// tieBottom returns the height of the ends of the tie of the lowercase or,
// if capital is set, of the capitals.
func tieBottom(p design.Params, capital bool) float64 {
	if capital {
		return capitalMarkHeight(p) + tieLift
	}
	return math.Round(p.XHeight+p.Acc.Gap) + tieLift
}

// tie draws a tie w wide in the advance adv, with its ink starting tieStart
// to the left of the middle of the advance.
func tie(p design.Params, w, adv float64, capital bool) (outline.Path, float64) {
	adv = math.Round(adv)
	path := chainsPath(tieChain(p, w))
	return shiftInk(path, math.Round(adv/2-tieStart), tieBottom(p, capital)), adv
}

// newTie draws the rounder form of a tie, w wide, in the middle of the
// advance of the figures.
func newTie(p design.Params, w float64, capital bool) (outline.Path, float64) {
	path := shiftInk(chainsPath(tieChain(p, w)), 0, tieBottom(p, capital))
	return placeCentred(path, p.Figures.Advance)
}

func tieLower(p design.Params) (outline.Path, float64) {
	return tie(p, tieWidth, design.NAdvance(p)/2, false)
}

func tieCapital(p design.Params) (outline.Path, float64) {
	return tie(p, tieWidthCap, design.CapHAdvance(p)/2, true)
}

func tieLowerNew(p design.Params) (outline.Path, float64) { return newTie(p, tieNewWidth, false) }

func tieCapitalNew(p design.Params) (outline.Path, float64) { return newTie(p, tieNewWidthCap, true) }
