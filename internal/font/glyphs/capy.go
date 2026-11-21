// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capY draws the capital Y: the upper part of V on a capital stem.  The two diagonals,
// thick on the left and thin on the right, have outer edges which meet the edges of
// the stem at one height (design.CapYJoint), as the outer edges of V pass through the
// ends of a flat, here as wide as the stem; their inner edges meet in the crotch at
// the height YJunction.  The arms have head serifs on both sides and the stem has
// foot serifs on both sides.  The contour is one counter-clockwise outline:
//
//	the outer serif of the thick diagonal, down its outer edge to the stem, down the
//	left-hand edge of the stem round its left foot serif, along the baseline, round the
//	right foot serif, up the right-hand edge of the stem, up the outer edge of the thin
//	diagonal, its outer serif, along the top to its inner serif, down the inner edge
//	to the crotch, up the inner edge of the thick diagonal, its inner serif and along
//	the top.
func capY(p design.Params) (outline.Path, float64) {
	d := p.CapDiag
	lean, joint, cx := design.CapYLean(p), design.CapYJoint(p), design.CapYCentre(p)
	half := p.Caps.Stem / 2
	l := diagFromEdge(edgeLeft, cx-half, joint, -lean, d.Thick)
	r := diagFromEdge(edgeRight, cx+half, joint, lean, d.Thin)
	top := p.CapHeight
	thick, thin := capDiagSerifs(p)
	outL := l.headServif(edgeLeft, top, thick)
	outR := r.headServif(edgeRight, top, thin)
	inR := r.headServif(edgeLeft, top, thin)
	inL := l.headServif(edgeRight, top, thick)
	stem := capStem{Left: cx - half, FootLeft: p.Caps.SerifJut, FootRight: p.Caps.SerifJut}
	jl := outline.Point{X: cx - half, Y: joint}
	jr := outline.Point{X: cx + half, Y: joint}
	down := stem.down(p, jl)
	up := stem.up(p, down.end(), joint)
	crotch := edgeMeet(l, edgeRight, r, edgeLeft)
	c := join(
		outL,
		lineChain(outL.end(), jl),
		down,
		up,
		lineChain(jr, outR.end()),
		outR.reversed(),
		lineChain(outR.start(), inR.start()),
		inR,
		lineChain(inR.end(), crotch),
		lineChain(crotch, inL.end()),
		inL.reversed(),
		lineChain(inL.start(), outL.start()),
	)
	return capContour(c), design.CapYAdvance(p)
}
