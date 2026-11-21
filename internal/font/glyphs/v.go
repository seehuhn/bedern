// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter v.
//
// The v is two diagonals which meet in a vertex at the baseline: a thick one on
// the left, which has the thickness of the stem, and a thin one on the right,
// which has the hairline thickness Hair, both measured at right angles
// (diagonal.go).  The outer edges lean by the same amount, VLean, and the
// outer serifs stand VInk apart at the x-height, and DiagVertexFlat apart at the
// flat of the vertex, which lies DiagOvershoot below the baseline.  Each diagonal
// has a head serif on both sides: a flat slab whose top lies on the x-height,
// SerifThick thick at the tip, which overhangs the edge by DiagOverThick (thick
// diagonal) or DiagOverThin (thin one) on each side, measured at the level of the
// underside.  The contour is one counter-clockwise outline:
//
//	the outer serif of the left diagonal, down the outer edge to the flat of the
//	vertex, along the flat, up the outer edge of the right diagonal, its outer
//	serif, along the top to its inner serif, down the inner edge to the crotch,
//	up the inner edge of the left diagonal, its inner serif and along the top.
//
// Side kinds: both sides are those of the diagonal letters, measured to the tips
// of the serifs.

// diagSerifs returns the head serifs of the diagonal letters: the outer one and
// the inner one.
func diagSerifs(p design.Params) (thick, thin diagSerif) {
	thick = diagSerif{Over: p.DiagOverThick, Thick: p.SerifThick, Bracket: design.LowerDiagBracket(p)}
	thin = diagSerif{Over: p.DiagOverThin, Thick: p.SerifThick, Bracket: design.LowerDiagBracket(p)}
	return thick, thin
}

// vVertex returns the vertex of v, in the middle of the ink.
func vVertex(p design.Params) vertex {
	cx := design.VCentre(p)
	return newVertex(cx, p.DiagVertexFlat, p.DiagOvershoot, design.VLean(p), p.Stem, p.Hair)
}

// contourPath returns the chain as a path of one closed contour.
func contourPath(c chain) outline.Path {
	var path outline.Path
	path.MoveTo(c.start().X, c.start().Y)
	c.appendTo(&path)
	return path
}

// v draws the lowercase letter v, see the comment above.
func v(p design.Params) (outline.Path, float64) {
	vx := vVertex(p)
	l, r := vx.Left, vx.Right
	top := p.XHeight
	thick, thin := diagSerifs(p)
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
	return contourPath(c), design.VAdvance(p)
}
