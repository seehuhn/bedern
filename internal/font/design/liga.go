// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package design

// LigaturePitch returns the distance from the left edge of the cell of one
// stem of a ligature of f to that of the next: the stem and the gap between
// the stems of two adjacent letters n.
func LigaturePitch(p Params) float64 { return p.Stem + p.SideStemRight + p.SideStemLeft }

// LigatureCell returns the left edge of the cell of the stem with the index k
// (from 0) of a ligature of f.
func LigatureCell(p Params, k int) float64 {
	return NarrowStemLeft(p) + float64(k)*LigaturePitch(p)
}

// LigatureAdvance returns the advance of a ligature of n letters whose last
// letter is last: the advance of last, moved by the pitch for each letter
// before it.  The stem of j stands in the cell of the stem of i, and the
// ligature ends with the right side of i.
func LigatureAdvance(p Params, n int, last string) float64 {
	shift := LigatureCell(p, n-1) - NarrowStemLeft(p)
	if last == "f" {
		return shift + FAdvance(p)
	}
	return shift + NarrowAdvance(p)
}

// LigatureCaret returns the caret between the letters k and k+1 of a ligature
// of f: midway between the cell of stem k and that of stem k+1.
func LigatureCaret(p Params, k int) float64 {
	return LigatureCell(p, k) + p.Stem + (p.SideStemRight+p.SideStemLeft)/2
}
