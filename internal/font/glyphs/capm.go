// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capM draws the capital M: the outer stems are vertical, the left stem and
// the right diagonal thin, the left diagonal and the right stem thick.  The
// diagonals meet in a vertex whose flat, MFlat wide, lies MHeight above the
// baseline and not on it.  Their outer edges, the edges which bound the two
// wedges of white between the diagonals and the stems, have the same lean in
// absolute value and leave the inner edges of the stems MCrotchDepth below the
// cap height: each wedge comes to a point, a sharp crotch under solid ink,
// narrower than the slit angle and listed as such for the clearance check
// (capMCrotches), with the crotch of the vertex.  The left stem is thin (thinStem) and has the head serif on its
// left and the foot serifs on both sides, the right stem the head serif on its
// right and the foot serifs on both sides; the diagonals end flush with the cap
// height on the stems.  The contour is one counter-clockwise outline:
//
//	the left stem from the tip of its left foot serif, round the right foot
//	serif, up its right-hand edge to the crotch of the wedge, down the outer
//	edge of the thick diagonal to the vertex, along the flat, up the outer
//	edge of the thin diagonal, to the crotch of the other wedge, down the
//	left-hand edge of the right stem, round its foot serifs, up its right-hand
//	edge, its head serif, along the top to the thin diagonal, down its inner
//	edge to the crotch, up the inner edge of the thick diagonal and along the
//	top to the head serif of the left stem.
func capM(p design.Params) (outline.Path, float64) {
	d := p.CapDiag
	jut := p.Caps.SerifJut
	left := newThinStem(p, capStem{Left: design.CapMStemLeft(p), FootLeft: jut, FootRight: jut, HeadLeft: jut})
	right := capStem{Left: design.CapMStemRight(p) - d.MStem, FootLeft: jut, FootRight: jut, HeadRight: jut}
	rp := p
	rp.Caps.Stem = d.MStem

	x0 := right.Left
	vx := capMVertex(p)
	yA := p.CapHeight - d.MCrotchDepth
	wedgeL := vx.Left.edgeAt(edgeLeft, yA)   // the left crotch, on the stem
	wedgeR := vx.Right.edgeAt(edgeRight, yA) // the right crotch
	topL := vx.Left.edgeAt(edgeRight, p.CapHeight)
	topR := vx.Right.edgeAt(edgeLeft, p.CapHeight)

	up1 := left.up(left.start(), yA)
	down2 := right.down(rp, outline.Point{X: x0, Y: yA})
	up2 := right.up(rp, down2.end(), p.Caps.BracketHeight)
	top2 := right.top(rp, up2.end(), p.CapHeight)
	top1 := left.top(topL, left.bracketHeight())
	down1 := left.down(top1.end())
	c := join(
		up1,
		lineChain(up1.end(), wedgeL),
		lineChain(wedgeL, vx.outerLeft()),
		vx.flat(),
		lineChain(vx.outerRight(), wedgeR),
		lineChain(wedgeR, down2.start()),
		down2,
		up2,
		top2,
		lineChain(top2.end(), topR),
		lineChain(topR, vx.crotch()),
		lineChain(vx.crotch(), topL),
		top1,
		down1,
	)
	return capContour(c), design.CapMAdvance(p)
}

// capMVertex returns the diagonals of M and their vertex.
func capMVertex(p design.Params) vertex {
	d := p.CapDiag
	x1 := design.CapMStemLeft(p) + d.ThinStem
	x0 := design.CapMStemRight(p) - d.MStem
	return newVertex((x1+x0)/2, d.MFlat, -d.MHeight, design.CapMLean(p), d.MThick, d.Thin)
}

// capMCrotches returns the three sharp crotches of M: the points of the white wedges
// between the stems and the diagonals, MCrotchDepth below the cap height on the inner
// edges of the stems, and the crotch of the vertex.
func capMCrotches(p design.Params) []sharpCrotch {
	d := p.CapDiag
	vx := capMVertex(p)
	y := p.CapHeight - d.MCrotchDepth
	down := outline.Point{Y: -1}
	upL, upR := vx.Left.down(), vx.Right.down()
	upL.X, upL.Y, upR.X, upR.Y = -upL.X, -upL.Y, -upR.X, -upR.Y
	return []sharpCrotch{
		{Name: "left wedge", Tip: vx.Left.edgeAt(edgeLeft, y), Dirs: [2]outline.Point{down, vx.Left.down()}},
		{Name: "right wedge", Tip: vx.Right.edgeAt(edgeRight, y), Dirs: [2]outline.Point{down, vx.Right.down()}},
		{Name: "vertex", Tip: vx.crotch(), Dirs: [2]outline.Point{upL, upR}},
	}
}
