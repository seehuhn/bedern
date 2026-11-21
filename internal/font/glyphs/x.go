// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter x.
//
// The x is two diagonals which cross: a thick one from the upper left to the lower
// right, XThick thick at right angles, and a thin one from the lower left to the
// upper right, of the hairline thickness.  The widths of the ink are the
// parameters, XInk at the baseline from tip to tip of the outer serifs, and the leans
// follow (design.XLeanSum): the thin diagonal is steeper than the thick one by
// XLeanDiff.  The ink at the x-height is narrower than at the baseline: the
// crossing, XCross, lies above the middle of the x-height.
//
// The thin stroke is drawn in two halves: the lower half is moved by XThinOffset
// to the right of the line of the upper half (the offset is negative, so the half
// lies to the left).  The halves are parallel and each ends on the edge of the thick
// diagonal, the upper half on its right-hand edge and the lower half on its left-hand
// edge, where the outline has an inner corner.  The thick diagonal is one straight
// stroke.
//
// There are four serifs: the head serifs of the diagonal letters (diagonal.go) at the
// x-height with a flat top, overhanging both edges of the thick diagonal by XOverThick
// and of the thin one by XOverThin, and at the baseline their feet, the same serifs
// standing flat on the baseline (footServif).  The contour is one counter-clockwise outline,
// without a counter:
//
//	the outer serif of the thick diagonal at the top, down its left-hand edge to the
//	notch on the left, down the left-hand edge of the lower half of the thin stroke,
//	its outer foot, along the baseline to its inner foot, up its right-hand edge to
//	the notch at the bottom, down the left-hand edge of the thick diagonal, its inner
//	foot, along the baseline to its outer foot, up its right-hand edge to the notch
//	on the right, up the right-hand edge of the upper half of the thin stroke, its
//	outer serif, along the top to its inner serif, down its left-hand edge to the
//	notch at the top, up the right-hand edge of the thick diagonal, its inner serif
//	and along the top.
//
// Side kinds: both sides are those of the diagonal letters, measured to the tips of
// the serifs, which stand further out at the baseline than at the top.

// xDiagonals returns the thick diagonal and the two halves of the thin one.
func xDiagonals(p design.Params) (thick, thinUpper, thinLower diagonal) {
	xc := design.XCentre(p)
	thick = diagonal{X: xc, Y: p.XCross, Lean: -design.XLeanThick(p), Thick: p.XThick}
	thinUpper = diagonal{X: xc, Y: p.XCross, Lean: design.XLeanThin(p), Thick: p.Hair}
	thinLower = diagonal{X: xc + p.XThinOffset, Y: p.XCross, Lean: design.XLeanThin(p), Thick: p.Hair}
	return thick, thinUpper, thinLower
}

// x draws the lowercase letter x, see the comment above.
func x(p design.Params) (outline.Path, float64) {
	thick, thinU, thinL := xDiagonals(p)
	top := p.XHeight
	sThick := diagSerif{Over: p.XOverThick, Thick: p.SerifThick, Bracket: design.LowerDiagBracket(p)}
	sThin := diagSerif{Over: p.XOverThin, Thick: p.SerifThick, Bracket: design.LowerDiagBracket(p)}
	outL := thick.headServif(edgeLeft, top, sThick)
	inL := thick.headServif(edgeRight, top, sThick)
	inR := thinU.headServif(edgeLeft, top, sThin)
	outR := thinU.headServif(edgeRight, top, sThin)
	blOut := thinL.footServif(edgeLeft, 0, sThin)
	blIn := thinL.footServif(edgeRight, 0, sThin).reversed()
	brIn := thick.footServif(edgeLeft, 0, sThick)
	brOut := thick.footServif(edgeRight, 0, sThick).reversed()
	notchTop := edgeMeet(thick, edgeRight, thinU, edgeLeft)
	notchRight := edgeMeet(thick, edgeRight, thinU, edgeRight)
	notchBottom := edgeMeet(thick, edgeLeft, thinL, edgeRight)
	notchLeft := edgeMeet(thick, edgeLeft, thinL, edgeLeft)
	c := join(
		outL,
		lineChain(outL.end(), notchLeft),
		lineChain(notchLeft, blOut.start()),
		blOut,
		lineChain(blOut.end(), blIn.start()),
		blIn,
		lineChain(blIn.end(), notchBottom),
		lineChain(notchBottom, brIn.start()),
		brIn,
		lineChain(brIn.end(), brOut.start()),
		brOut,
		lineChain(brOut.end(), notchRight),
		lineChain(notchRight, outR.end()),
		outR.reversed(),
		lineChain(outR.start(), inR.start()),
		inR,
		lineChain(inR.end(), notchTop),
		lineChain(notchTop, inL.end()),
		inL.reversed(),
		lineChain(inL.start(), outL.start()),
	)
	return contourPath(c), design.XAdvance(p)
}
