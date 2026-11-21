// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter w.
//
// The w is two narrowed v's side by side: four diagonals, thick, thin, thick
// and thin from the left, which meet at two vertices on the baseline and at an
// apex in the middle.  The thick diagonals are WStemCut thinner than the stem
// and the thin ones have the hairline thickness.  All four have the lean WLean;
// the thick ones are parallel to each other and so are the thin ones.  The
// vertices are those of v (a flat of DiagVertexFlat, DiagOvershoot below the
// baseline), the head serifs are those of v on the two outer diagonals, and the
// apex has no serif but is cut flat, WApexFlat wide, at WApexHeight, where the
// outer edge of the second diagonal meets the outer edge of the third.  The
// lean follows from the width of the ink, WInk, and the distance between the
// vertices from the apex (design.WLean, design.WPeriod).  The contour is one
// counter-clockwise outline:
//
//	the outer serif of the first diagonal, down its outer edge to the first vertex,
//	along its flat, up the inner edge of the second diagonal to the notch under the
//	apex, down the inner edge of the third to the second vertex, along its flat, up
//	the outer edge of the fourth, its outer serif, along the top to its inner serif,
//	down the inner edge to the crotch of the second vertex, up the outer edge of the
//	third diagonal to the apex, along the flat of the apex, down the outer edge of
//	the second diagonal to the crotch of the first vertex, up the inner edge of the
//	first diagonal, its inner serif and along the top.
//
// Side kinds: both sides are those of the diagonal letters.

// wVertices returns the two vertices of w.
func wVertices(p design.Params) (v1, v2 vertex) {
	lean := design.WLean(p)
	x0 := p.DiagSideLeft
	c1 := x0 + design.DiagJutOuter(p, p.DiagOverThick, lean) + lean*design.DiagHeight(p) + p.DiagVertexFlat/2
	c2 := c1 + design.WPeriod(p)
	v1 = newVertex(c1, p.DiagVertexFlat, p.DiagOvershoot, lean, design.WThick(p), p.Hair)
	v2 = newVertex(c2, p.DiagVertexFlat, p.DiagOvershoot, lean, design.WThick(p), p.Hair)
	return v1, v2
}

// wApex returns the two ends of the flat of the apex: the left end on the outer
// edge of the second diagonal and the right end on the outer edge of the third.
func wApex(p design.Params) (left, right outline.Point) {
	v1, v2 := wVertices(p)
	return v1.Right.edgeAt(edgeLeft, p.WApexHeight), v2.Left.edgeAt(edgeRight, p.WApexHeight)
}

// w draws the lowercase letter w, see the comment above.
func w(p design.Params) (outline.Path, float64) {
	v1, v2 := wVertices(p)
	d1, d2, d3, d4 := v1.Left, v1.Right, v2.Left, v2.Right
	top := p.XHeight
	thick, thin := diagSerifs(p)
	outL := d1.headServif(edgeLeft, top, thick)
	outR := d4.headServif(edgeRight, top, thin)
	inR := d4.headServif(edgeLeft, top, thin)
	inL := d1.headServif(edgeRight, top, thick)
	notch := edgeMeet(d2, edgeRight, d3, edgeLeft)
	apexL, apexR := wApex(p)
	c := join(
		outL,
		lineChain(outL.end(), v1.outerLeft()),
		v1.flat(),
		lineChain(v1.outerRight(), notch),
		lineChain(notch, v2.outerLeft()),
		v2.flat(),
		lineChain(v2.outerRight(), outR.end()),
		outR.reversed(),
		lineChain(outR.start(), inR.start()),
		inR,
		lineChain(inR.end(), v2.crotch()),
		lineChain(v2.crotch(), apexR),
		lineChain(apexR, apexL),
		lineChain(apexL, v1.crotch()),
		lineChain(v1.crotch(), inL.end()),
		inL.reversed(),
		lineChain(inL.start(), outL.start()),
	)
	return contourPath(c), design.WAdvance(p)
}
