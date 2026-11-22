// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The diagonal strokes of the capitals A, V, W, X, Y, Z, K, M and N.
//
// The strokes are those of the lowercase (diagonal.go) at the weight of the
// capitals: a thick diagonal of about 0.97 of the capital stem and a thin one of
// 42, the serifs with the capital serif thickness and the bracket of the capital
// stem adapted to the inclined edge (design.CapDiagBracket), the same overhang on
// both sides of a stroke (design.CapDiag.OverThick and OverThin), vertices cut flat
// below the baseline, an apex of A cut flat above the cap height.  The capitals
// add the serifs with the capital values and the thin stem of M and N.
//
// The letters are drawn as one counter-clockwise contour from chains, as v, w, x
// and k are: the serifs and the straight edges between them, the flats at the
// vertices and the stems with the parts of capstem.go.

// capDiagSerifs returns the head and foot serifs of the diagonals of the capitals:
// the one on the thick strokes and the one on the thin strokes, with the serif
// thickness of the capitals at the tip.
func capDiagSerifs(p design.Params) (thick, thin diagSerif) {
	d, b := p.CapDiag, design.CapDiagBracket(p)
	return diagSerif{Over: d.OverThick, Thick: p.Caps.SerifThick, Bracket: b, Fair: true}, diagSerif{Over: d.OverThin, Thick: p.Caps.SerifThick, Bracket: b, Fair: true}
}

// A thinStem is a thin vertical stem of the capitals, the stem of M and N on the
// side where the weight is that of the thin diagonal: a capStem of the width
// design.CapDiag.ThinStem with the serifs of the capital stem and the brackets of a
// thin stem (thinStemParams).  The methods are those of capStem with the width
// and the brackets changed.
type thinStem struct {
	capStem
	q design.Params // the parameters with the stem width and the brackets changed
}

// thinStemBracket is the size of the brackets of the serifs of every thin stem of the
// capitals (the thin stems of M, N and U), as a fraction of those of the capital stem:
// the serifs have the length, the tip and the flat of all capital serifs.
const thinStemBracket = 0.7

// thinStemParams returns p with the capital stem of the width width and the brackets
// of the capital serifs made smaller by thinStemBracket, for a thin stem.
func thinStemParams(p design.Params, width float64) design.Params {
	q := p
	q.Caps.Stem = width
	q.Caps.BracketHeight *= thinStemBracket
	q.Caps.BracketReach *= thinStemBracket
	return q
}

// newThinStem returns the thin stem with the serifs of s.
func newThinStem(p design.Params, s capStem) thinStem {
	return thinStem{capStem: s, q: thinStemParams(p, p.CapDiag.ThinStem)}
}

// rightEdge returns the position of the right-hand edge.
func (t thinStem) rightEdge() float64 { return t.capStem.right(t.q) }

// bracketHeight returns the height above the baseline (below the cap height) at
// which the brackets of the serifs leave the stem.
func (t thinStem) bracketHeight() float64 { return t.q.Caps.BracketHeight }

// up is capStem.up for the thin stem.
func (t thinStem) up(from outline.Point, to float64) chain { return t.capStem.up(t.q, from, to) }

// top is capStem.top for the thin stem.
func (t thinStem) top(from outline.Point, to float64) chain { return t.capStem.top(t.q, from, to) }

// down is capStem.down for the thin stem.
func (t thinStem) down(from outline.Point) chain { return t.capStem.down(t.q, from) }

// capSideSpace returns the side space of the kinds of the diagonal capitals, and
// whether k is one of them.
func capSideSpace(k sideKind, p design.Params, left bool) (float64, bool) {
	var s design.CapSides
	switch k {
	case sideCapA:
		s = p.CapDiag.SideA
	case sideCapV:
		s = p.CapDiag.SideV
	case sideCapW:
		s = p.CapDiag.SideW
	case sideCapX:
		s = p.CapDiag.SideX
	case sideCapY:
		s = p.CapDiag.SideY
	case sideCapZ:
		s = p.CapDiag.SideZ
	case sideCapK:
		s = p.CapDiag.SideK
	case sideCapM:
		s = p.CapDiag.SideM
	case sideCapN:
		s = p.CapDiag.SideN
	default:
		return 0, false
	}
	if left {
		return s.Left, true
	}
	return s.Right, true
}
