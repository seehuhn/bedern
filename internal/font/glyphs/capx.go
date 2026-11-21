// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capXDiagonals returns the thick diagonal of X and the two halves of the thin one.
func capXDiagonals(p design.Params) (thick, thinUpper, thinLower diagonal) {
	d := p.CapDiag
	xc := design.CapXCentre(p)
	thick = diagonal{X: xc, Y: d.XCross, Lean: -design.CapXLeanThick(p), Thick: d.XThick}
	thinUpper = diagonal{X: xc, Y: d.XCross, Lean: design.CapXLeanThin(p), Thick: d.Thin}
	thinLower = diagonal{X: xc + d.XThinOffset, Y: d.XCross, Lean: design.CapXLeanThin(p), Thick: d.Thin}
	return thick, thinUpper, thinLower
}

// capX draws the capital X, the lowercase x at the weight of the capitals: a thick
// diagonal from the upper left to the lower right, XThick thick, and a thin one
// crossing it, drawn in two halves of which the lower is moved by XThinOffset to the
// left; each half ends on an edge of the thick diagonal in a sharp inner corner.  The strokes have
// serifs at both ends on both sides, flat on the cap height and the baseline.  The
// contour is that of x.
func capX(p design.Params) (outline.Path, float64) {
	thick, thinU, thinL := capXDiagonals(p)
	top := p.CapHeight
	sThick, sThin := capDiagSerifs(p)
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
	return capContour(c), design.CapXAdvance(p)
}
