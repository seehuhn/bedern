// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capADiagonals returns the left (thin) and the right (thick) diagonal of A.  The outer
// edges lean by the same amount and pass through the ends of the flat of the apex,
// ApexFlat wide and ApexRise above the cap height, the mirror image of the vertex of V
// in a horizontal line.
func capADiagonals(p design.Params) (left, right diagonal) {
	d := p.CapDiag
	l, cx, top := design.CapALean(p), design.CapACentre(p), p.CapHeight+d.ApexRise
	left = diagFromEdge(edgeLeft, cx-d.ApexFlat/2, top, l, d.Thin)
	right = diagFromEdge(edgeRight, cx+d.ApexFlat/2, top, -l, d.Thick)
	return left, right
}

// capA draws the capital A: a thin diagonal on the left and a thick one on the right
// which meet in an apex cut flat, ApexRise above the cap height, with a thin crossbar
// of the thickness of the bar of H, centred at ABar, between them, and foot serifs
// on both sides of both diagonals.  The crotch of the apex is the corner above the
// counter, whose angle is twice that of a diagonal.  The contour is one
// counter-clockwise outline and the triangle above the crossbar a clockwise counter:
//
//	the flat of the apex from right to left, down the outer edge of the left diagonal,
//	its outer foot, along the baseline to its inner foot, up its inner edge to the
//	underside of the crossbar, along it, down the inner edge of the right diagonal, its
//	inner foot, along the baseline to its outer foot, up the outer edge.
func capA(p design.Params) (outline.Path, float64) {
	d := p.CapDiag
	l, r := capADiagonals(p)
	top := p.CapHeight + d.ApexRise
	thick, thin := capDiagSerifs(p)
	outL := l.footServif(edgeLeft, 0, thin)
	inL := l.footServif(edgeRight, 0, thin).reversed()
	inR := r.footServif(edgeLeft, 0, thick)
	outR := r.footServif(edgeRight, 0, thick).reversed()
	apexL, apexR := l.edgeAt(edgeLeft, top), r.edgeAt(edgeRight, top)
	lower, upper := d.ABar-p.Caps.Bar/2, d.ABar+p.Caps.Bar/2
	p1, p2 := l.edgeAt(edgeRight, lower), r.edgeAt(edgeLeft, lower)
	c := join(
		lineChain(apexL, outL.start()),
		outL,
		lineChain(outL.end(), inL.start()),
		inL,
		lineChain(inL.end(), p1),
		lineChain(p1, p2),
		lineChain(p2, inR.start()),
		inR,
		lineChain(inR.end(), outR.start()),
		outR,
		lineChain(outR.end(), apexR),
		lineChain(apexR, apexL),
	)
	q1, q2 := l.edgeAt(edgeRight, upper), r.edgeAt(edgeLeft, upper)
	crotch := edgeMeet(l, edgeRight, r, edgeLeft)
	counter := join(lineChain(q1, crotch), lineChain(crotch, q2), lineChain(q2, q1))
	path := capContour(c)
	return append(path, contourPath(counter)...), design.CapAAdvance(p)
}
