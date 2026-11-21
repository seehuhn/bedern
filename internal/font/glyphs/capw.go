// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capWVertices returns the two vertices of W, each of a thick diagonal on the left
// and a thin one on the right.
func capWVertices(p design.Params) (v1, v2 vertex) {
	d := p.CapDiag
	lean := design.CapWLean(p)
	c1 := design.CapWCentre(p)
	c2 := c1 + design.CapWPeriod(p)
	v1 = newVertex(c1, d.VertexFlat, d.VertexDepth, lean, d.WThick, d.Thin)
	v2 = newVertex(c2, d.VertexFlat, d.VertexDepth, lean, d.WThick, d.Thin)
	return v1, v2
}

// capW draws the capital W: two narrowed V's side by side, thick, thin, thick
// and thin from the left, which meet at two vertices below the baseline and at
// an apex in the middle.  The thick diagonals are thinned a little (WThick).
// The apex is pointed in the way of the lowercase w: it has no serif and is cut
// flat, WApexFlat wide, WApexHeight, a few units under the cap height, where
// the outer edge of the second diagonal meets that of the third.  The contour
// is that of w.
func capW(p design.Params) (outline.Path, float64) {
	d := p.CapDiag
	v1, v2 := capWVertices(p)
	d1, d2, d3, d4 := v1.Left, v1.Right, v2.Left, v2.Right
	top := p.CapHeight
	thick, thin := capDiagSerifs(p)
	outL := d1.headServif(edgeLeft, top, thick)
	outR := d4.headServif(edgeRight, top, thin)
	inR := d4.headServif(edgeLeft, top, thin)
	inL := d1.headServif(edgeRight, top, thick)
	notch := edgeMeet(d2, edgeRight, d3, edgeLeft)
	apexL, apexR := d2.edgeAt(edgeLeft, d.WApexHeight), d3.edgeAt(edgeRight, d.WApexHeight)
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
	return capContour(c), design.CapWAdvance(p)
}
