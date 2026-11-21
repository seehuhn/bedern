// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capNStems returns the left and the right stem of N, both thin: the left stem has the
// head serif on its left and the foot serifs on both sides, the right stem the head
// serifs on both sides and no foot, since the diagonal ends there in a vertex.
func capNStems(p design.Params) (l, r thinStem) {
	jut := p.Caps.SerifJut
	l = newThinStem(p, capStem{Left: design.CapNStemLeft(p), FootLeft: jut, FootRight: jut, HeadLeft: jut})
	r = newThinStem(p, capStem{Left: design.CapNStemRight(p) - p.CapDiag.ThinStem, HeadLeft: jut, HeadRight: jut})
	return l, r
}

// capN draws the capital N in the classical pattern: a thin left stem, a thick
// diagonal (NThick) from the upper left to the lower right and a thin right stem.  The
// diagonal starts at the top of the left stem, its upper edge flush with the cap height
// and its lower edge leaving the stem NJoin below it, and its lower end makes a
// pointed vertex with the right-hand edge of the right stem, cut flat NFlat wide
// and NDepth below the baseline; it overshoots the baseline slightly like
// the vertex of V.  The inner edge of the diagonal meets the
// left-hand edge of the right stem in a sharp inner corner above the baseline.  The
// contour is one counter-clockwise outline:
//
//	the left stem from the tip of its left foot serif, along the baseline and round
//	the right foot serif, up the right-hand edge to the wedge, down the
//	lower edge of the diagonal to the vertex, along its flat, up the right-hand edge of
//	the right stem, its head serifs, down its left-hand edge to the inner corner, up
//	the upper edge of the diagonal to the cap height, along the top to the head serif
//	of the left stem and down its left-hand edge.
func capN(p design.Params) (outline.Path, float64) {
	d := p.CapDiag
	l, r := capNStems(p)
	xr1 := r.rightEdge()
	diag, crotch := capNDiagonal(p)
	topRight := diag.edgeAt(edgeRight, p.CapHeight)
	vl := outline.Point{X: xr1 - d.NFlat, Y: -d.NDepth}
	vr := outline.Point{X: xr1, Y: -d.NDepth}

	up1 := l.up(l.start(), p.CapHeight-d.NJoin)
	up2 := r.top(vr, crotch.Y)
	top1 := l.top(topRight, l.bracketHeight())
	down1 := l.down(top1.end())
	c := join(
		up1,
		lineChain(up1.end(), vl),
		lineChain(vl, vr),
		up2,
		lineChain(up2.end(), topRight),
		top1,
		down1,
	)
	return capContour(c), design.CapNAdvance(p)
}

// capNDiagonal returns the diagonal of N and its crotch with the right stem, where
// its right-hand edge meets the left-hand edge of the stem.
func capNDiagonal(p design.Params) (diag diagonal, crotch outline.Point) {
	d := p.CapDiag
	l, r := capNStems(p)
	diag = diagFromEdge(edgeLeft, l.rightEdge(), p.CapHeight-d.NJoin, -design.CapNLean(p), d.NThick)
	stemD := diagonal{X: r.Left + d.ThinStem/2, Lean: 0, Thick: d.ThinStem}
	return diag, edgeMeet(diag, edgeRight, stemD, edgeLeft)
}

// capNCrotches returns the two crotches of N: the point of the white wedge between
// the left stem and the diagonal, NJoin below the cap height, and the inner corner
// between the diagonal and the right stem above the baseline.
func capNCrotches(p design.Params) []sharpCrotch {
	diag, crotch := capNDiagonal(p)
	l, _ := capNStems(p)
	down := diag.down()
	return []sharpCrotch{
		{Name: "upper wedge", Tip: outline.Point{X: l.rightEdge(), Y: p.CapHeight - p.CapDiag.NJoin}, Dirs: [2]outline.Point{{Y: -1}, down}},
		{Name: "lower wedge", Tip: crotch, Dirs: [2]outline.Point{{Y: 1}, {X: -down.X, Y: -down.Y}}},
	}
}
