// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// capKDiagonals returns the arm and the leg of K, which meet at the junction of their
// centre lines: the leg leaves the arm and not the stem, as in the lowercase k.
func capKDiagonals(p design.Params) (arm, leg diagonal) {
	d := p.CapDiag
	xj, yj := design.CapKJunctionX(p), d.KJunction
	arm = diagonal{X: xj, Y: yj, Lean: design.CapKArmLean(p), Thick: d.Thin}
	leg = diagonal{X: xj, Y: yj, Lean: -design.CapKLegLean(p), Thick: d.KLeg}
	return arm, leg
}

// capK draws the capital K: the capital stem with serifs on both sides at the head and
// at the foot, a thin arm which rises from the stem to the upper right and ends in the
// head serif of the diagonals, and a thick leg which leaves the arm and descends to
// the lower right, ending in a foot serif.  The arm is one straight stroke from the
// stem to its serif, and its underside runs from the stem to the leg as one
// straight line; the leg is a straight stroke which ends on the arm.  The contour is
// that of k: from below, up the stem to the point where the underside of the arm
// leaves it, along that edge to the left-hand edge of the leg, down it to the foot, the
// foot serif, up the right-hand edge of the leg to the underside of the arm, along it
// to the serif of the arm, along the top, down the upper edge of the arm to the stem and
// up the stem to its head.
func capK(p design.Params) (outline.Path, float64) {
	d := p.CapDiag
	jut := p.Caps.SerifJut
	stem := capStem{Left: design.CapKStemLeft(p), FootLeft: jut, FootRight: jut, HeadLeft: jut, HeadRight: jut}
	arm, leg := capKDiagonals(p)
	right := stem.right(p)
	// Where the edges of the arm reach the right-hand edge of the stem.
	atStem := func(e edge) outline.Point {
		x := arm.edgeAt(e, arm.Y).X
		return outline.Point{X: right, Y: arm.Y + (right-x)/arm.Lean}
	}
	p1, p0 := atStem(edgeRight), atStem(edgeLeft)
	p2 := edgeMeet(arm, edgeRight, leg, edgeLeft)
	q := edgeMeet(arm, edgeRight, leg, edgeRight)
	sArm := diagSerif{Over: d.KOverArm, Thick: p.Caps.SerifThick, Bracket: design.CapDiagBracket(p)}
	sLeg := diagSerif{Over: d.KOverLeg, Thick: p.Caps.SerifThick, Bracket: design.CapDiagBracket(p)}
	armOut := arm.headServif(edgeRight, p.CapHeight, sArm)
	armIn := arm.headServif(edgeLeft, p.CapHeight, sArm)
	legIn := leg.footServif(edgeLeft, 0, sLeg)
	legOut := leg.footServif(edgeRight, 0, sLeg).reversed()
	up := stem.up(p, stem.start(), p1.Y)
	mid := join(
		lineChain(p1, p2),
		lineChain(p2, legIn.start()),
		legIn,
		lineChain(legIn.end(), legOut.start()),
		legOut,
		lineChain(legOut.end(), q),
		lineChain(q, armOut.end()),
		armOut.reversed(),
		lineChain(armOut.start(), armIn.start()),
		armIn,
		lineChain(armIn.end(), p0),
	)
	top := stem.top(p, p0, p.Caps.BracketHeight)
	down := stem.down(p, top.end())
	return capContour(join(up, mid, top, down)), design.CapKAdvance(p)
}
