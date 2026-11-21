// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capV draws the capital V: two diagonals which meet in a vertex cut flat below the
// baseline, a thick one on the left (Thick) and a thin one on the right (Thin), with
// head serifs on both sides of both.  It is the lowercase v at the weight of the
// capitals.  The outer edges lean by the same amount, which follows from the width of
// the ink (design.CapVLean).  The contour is the one of v:
//
//	the outer serif of the left diagonal, down the outer edge to the flat of the
//	vertex, along the flat, up the outer edge of the right diagonal, its outer
//	serif, along the top to its inner serif, down the inner edge to the crotch, up
//	the inner edge of the left diagonal, its inner serif and along the top.
func capV(p design.Params) (outline.Path, float64) {
	d := p.CapDiag
	vx := newVertex(design.CapVCentre(p), d.VertexFlat, d.VertexDepth, design.CapVLean(p), d.Thick, d.Thin)
	l, r := vx.Left, vx.Right
	top := p.CapHeight
	thick, thin := capDiagSerifs(p)
	outL := l.headServif(edgeLeft, top, thick)
	outR := r.headServif(edgeRight, top, thin)
	inR := r.headServif(edgeLeft, top, thin)
	inL := l.headServif(edgeRight, top, thick)
	c := join(
		outL,
		lineChain(outL.end(), vx.outerLeft()),
		vx.flat(),
		lineChain(vx.outerRight(), outR.end()),
		outR.reversed(),
		lineChain(outR.start(), inR.start()),
		inR,
		lineChain(inR.end(), vx.crotch()),
		lineChain(vx.crotch(), inL.end()),
		inL.reversed(),
		lineChain(inL.start(), outL.start()),
	)
	return capContour(c), design.CapVAdvance(p)
}
