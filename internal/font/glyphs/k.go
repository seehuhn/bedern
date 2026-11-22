// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter k.
//
// The k is the stem of h (the ascender stem with its head serif on the left and
// both feet) with an arm and a leg on its right.  The arm is a thin diagonal of the
// hairline thickness which rises from the stem to the upper right at the angle
// KArmAngle to the vertical and ends in the head serif of the diagonal letters at
// the x-height, overhanging both edges by KOverArm.  The leg is a thick diagonal,
// KLeg thick, which leaves the arm and descends to the lower right at KLegAngle,
// ending in a foot serif on the baseline which overhangs both edges by KOverLeg.
// Both are measured at right angles to the stroke.
//
// The junction.  The centre lines of the arm and the leg meet at the point
// (design.KJunctionX, KJunction), which stands off the stem; the leg leaves the
// arm and not the stem.  The arm is one
// straight stroke from the stem to its serif and the leg is a straight stroke
// which ends on the arm: the edges of the arm continue through the junction, and
// the edges of the leg end on the arm.  So, from below, the outline goes up the
// stem to the point P1 where the lower edge of the arm (its right-hand edge) leaves
// it, and follows that edge to P2, where the left-hand edge of the leg meets it;
// there it turns, in a sharp inner corner, onto the left-hand edge of the leg and
// runs down it to the foot.  On the way up the right-hand edge of the leg ends on the
// lower edge of the arm in Q, a second inner corner, above P2, and the lower edge of
// the arm goes on to the serif.  Along the top the outline comes down the upper edge
// of the arm to the stem at P0, the third inner corner, and goes on up the stem.  The
// stretch of the lower edge of the arm between P2 and Q is inside the leg.  Between
// the stem, the lower edge of the arm and the left-hand edge of the leg lies a white
// triangle which is open at the bottom: it has the corners P1 and P2 and ends at the
// baseline, and its width at the baseline is the distance between the right foot of
// the stem and the foot of the leg.
//
// The width.  KReach is the distance from the right-hand edge of the cell of the
// stem (a stem of the width Stem, see design.KStemRight) to the end of the foot of
// the leg, which is the rightmost ink, and the junction follows from it
// (design.KJunctionX).  The junction is 53 from the stem.
// The contour is one counter-clockwise outline, drawn from the parts of h: the start
// of the stem, stemUp to P1, the chain of arm and leg, stemCap from P0 and stemDown.
//
// Side kinds: the left side is that of h, the right side the tip of the foot of the
// leg, KSideRight from the edge of the advance.

// kDiagonals returns the arm and the leg.
func kDiagonals(p design.Params) (arm, leg diagonal) {
	xj, yj := design.KJunctionX(p), p.KJunction
	arm = diagonal{X: xj, Y: yj, Lean: design.KArmLean(p), Thick: p.Hair}
	leg = diagonal{X: xj, Y: yj, Lean: -design.KLegLean(p), Thick: p.KLeg}
	return arm, leg
}

// kStemSpec returns the stem of k, which is the left stem of h.
func kStemSpec(p design.Params) stemSpec {
	return stemSpec{
		Left: p.SideStemLeft + design.AscenderInset(p), Width: p.AscenderStem, Top: p.Ascender,
		FootLeft: p.SerifJut, FootRight: p.SerifJutInner,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
}

// k draws the lowercase letter k, see the comment above.
func k(p design.Params) (outline.Path, float64) {
	stem := kStemSpec(p)
	arm, leg := kDiagonals(p)
	right := stem.right()
	// Where the edges of the arm reach the right-hand edge of the stem.
	atStem := func(e edge) outline.Point {
		x := arm.edgeAt(e, arm.Y).X
		return outline.Point{X: right, Y: arm.Y + (right-x)/arm.Lean}
	}
	p1, p0 := atStem(edgeRight), atStem(edgeLeft)
	p2 := edgeMeet(arm, edgeRight, leg, edgeLeft)
	q := edgeMeet(arm, edgeRight, leg, edgeRight)
	sArm := diagSerif{Over: p.KOverArm, Thick: p.SerifThick, Bracket: design.LowerDiagBracket(p), Fair: true}
	sLeg := diagSerif{Over: p.KOverLeg, Thick: p.SerifThick, Bracket: design.LowerDiagBracket(p), Fair: true}
	armOut := arm.headServif(edgeRight, p.XHeight, sArm)
	armIn := arm.headServif(edgeLeft, p.XHeight, sArm)
	legIn := leg.footServif(edgeLeft, 0, sLeg)
	legOut := leg.footServif(edgeRight, 0, sLeg).reversed()
	c := join(
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
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, p1.Y)
	c.appendTo(&path)
	stemCap(&path, p, stem)
	stemDown(&path, p, stem)
	return path, design.KAdvance(p)
}
