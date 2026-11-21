// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The ampersand.
//
// The form has a small upper loop shaped like a teardrop, whose heavy left side
// runs on without a break into a thick diagonal down to the lower right, where
// it turns into a short tail that rises to the right and is cut; a large round
// bowl below, heavy on its left; and a thin stroke which comes down the right
// of the loop, crosses the diagonal at the waist, runs down the left of the
// bowl and round its bottom, rises up the right of the bowl across the diagonal
// a second time and goes on as the arm, which ends in a serif at about the
// x-height.  The weight lies in the left of the loop (AmpThick less a fifth),
// the diagonal (AmpThick) and the left of the bowl (AmpBowlThick); every other
// part is thin.
//
// It is one stroke (symstroke.go), drawn from the arm to the tail: down the
// arm (AmpArmThick at its foot), across the diagonal at the lower crossing, the
// third of the height up, round the bowl clockwise, up across the waist into
// the loop, over its top and down its left side, down the diagonal and into the
// tail.  The centre line of the bowl is a circle which touches the vertical
// through the middle of its left side and the horizontal through the middle of
// its bottom and has its right extreme just left of the lower crossing; the
// stations on its right and bottom carry its curvature, and the extremes of the
// bowl have vertical and horizontal tangents.  The loop is narrow and its right
// side is thin, with a teardrop counter about 160 wide; the left of the loop
// runs into the straight diagonal over a curve whose curvature falls to nothing
// where the diagonal begins.  The stroke crosses itself twice, and the outline
// is the union of it and of itself (symunion.go); the crotches of the
// crossings are open V's, near right angles at the waist.  The arm is straight
// at its end and carries the head serif of a diagonal of the capitals
// (diagonal.go), with the overhangs AmpSerifLeft and AmpSerifRight.
//
// Side kinds: open, AmpSide to the ink.

// ampFrame holds the points from which the ampersand is drawn.
type ampFrame struct {
	x0, w, top, bottom, h float64
	diagA, diagB          outline.Point // the centre line of the straight diagonal
}

func newAmpFrame(p design.Params) ampFrame {
	s := p.Sym
	f := ampFrame{x0: s.AmpSide, w: s.AmpInk, top: s.AmpTop, bottom: s.AmpBottom}
	f.h = f.top - f.bottom
	f.diagA = f.at(0.235, 0.63)
	f.diagB = f.at(0.68, 0.12)
	return f
}

// at returns the point at the fractions u of the ink width and v of the height.
func (f ampFrame) at(u, v float64) outline.Point {
	return outline.Point{X: f.x0 + u*f.w, Y: f.bottom + v*f.h}
}

// onDiag returns the station on the diagonal at the fraction u from its upper
// end, running down it.
func (f ampFrame) onDiag(u, w float64) station {
	q := lerp(f.diagA, f.diagB, u)
	d := math.Atan2(f.diagB.Y-f.diagA.Y, f.diagB.X-f.diagA.X) * 180 / math.Pi
	return station{X: q.X, Y: q.Y, Dir: d, W: w}
}

func ampersand(p design.Params) (outline.Path, float64) {
	s := p.Sym
	f := newAmpFrame(p)
	t := p.Figures.Thin
	st := func(q outline.Point, dir, w float64) station { return station{X: q.X, Y: q.Y, Dir: dir, W: w} }

	// The bowl: its centre line is a circle, walked clockwise, whose right
	// extreme lies ampBowlIn left of the lower crossing and which touches the
	// vertical through the middle of its heavy left side and the horizontal
	// through the middle of its thin bottom.  The stations on it carry the
	// curvature of the circle.
	r := (f.at(ampLowCross[0], 0).X - ampBowlIn - f.x0 - s.AmpBowlThick/2) / 2
	bc := outline.Point{X: f.x0 + s.AmpBowlThick/2 + r, Y: f.bottom + t/2 + r}
	onBowl := func(deg, w float64) station {
		a := deg * math.Pi / 180
		return station{X: bc.X + r*math.Cos(a), Y: bc.Y + r*math.Sin(a), Dir: deg - 90, W: w, K: -1 / r}
	}

	// Where the bowl swells to its heavy side, the edges take the curvature
	// which joins their neighbours most evenly (symstroke.go), not that of the
	// circle.
	free := func(s station) station { s.K = 0; return s }

	armFoot := f.at(ampArmFoot[0], ampArmFoot[1])
	da, db := f.onDiag(ampDiagStart, s.AmpThick), f.onDiag(1, s.AmpThick)
	da.Flat = true
	loopRight := f.at(ampLoopRight, 0).X
	loopY := f.at(0, ampLoopSideY).Y
	sts := []station{
		st(armFoot, ampArmAngle+180, s.AmpArmThick),
		{X: f.at(ampLowCross[0], 0).X, Y: f.at(0, ampLowCross[1]).Y, Dir: ampLowCross[2] + 180, W: 0.5 * (s.AmpArmThick + t)},
		onBowl(0, t+2),
		onBowl(-45, t),
		onBowl(-90, t),
		free(onBowl(-180, s.AmpBowlThick)),
		st(f.at(ampUpCross[0], ampUpCross[1]), ampUpCross[2], t),
		{X: loopRight - t/2, Y: loopY, Dir: 90, W: t},
		{X: f.at(ampLoopTop, 0).X, Y: f.top - t/2, Dir: 180, W: t},
		{X: f.at(ampLoopLeft, 0).X + ampLoopLeftW*s.AmpThick/2, Y: loopY, Dir: -90, W: ampLoopLeftW * s.AmpThick},
		da,
		db,
		{X: f.at(0.84, 0).X, Y: f.bottom + (t+8)/2, Dir: 0, W: t + 8},
		{X: f.at(0.99, 0.27).X - 0.65*t, Y: f.at(0.99, 0.27).Y, Dir: 80, W: 1.3 * t},
	}
	right, left := strokeOpen(sts)
	a := ampArmAngle * math.Pi / 180
	arm := diagonal{X: armFoot.X, Y: armFoot.Y, Lean: math.Cos(a) / math.Sin(a), Thick: s.AmpArmThick}
	br := design.CapDiagBracket(p)
	outR := arm.headServif(edgeRight, s.AmpArm, diagSerif{Over: s.AmpSerifRight, Thick: p.Caps.SerifThick, Bracket: br}).reversed()
	outL := arm.headServif(edgeLeft, s.AmpArm, diagSerif{Over: s.AmpSerifLeft, Thick: p.Caps.SerifThick, Bracket: br})
	stroke := join(right, lineChain(right.end(), left.start()), left, lineChain(left.end(), outR.start()), outR,
		lineChain(outR.end(), outL.start()), outL, lineChain(outL.end(), right.start()))
	return chainsPath(union(stroke)...), f.x0 + f.w + s.AmpSide
}

// The shape constants of the ampersand, as fractions of the ink width
// (abscissae) and of the height (heights), and directions in degrees: the
// direction of the straight arm and its foot; how far the right extreme of the
// centre line of the bowl lies left of the lower crossing (units); the right of
// the loop, the abscissa of its top, the left of the loop and the height of its
// side extremes; the thickness of the left of the loop as a fraction of
// AmpThick; the start of the straight diagonal as a fraction of its length.
const (
	ampArmAngle  = 70.0
	ampBowlIn    = 3.0
	ampLoopRight = 0.537
	ampLoopTop   = 0.366
	ampLoopLeft  = 0.136
	ampLoopSideY = 0.807
	ampLoopLeftW = 0.8
	ampDiagStart = 0.15
)

// The places of the two crossings: where the stroke from the arm crosses the
// diagonal on its way down into the bowl, and where the stroke rising from the
// bowl crosses it at the waist, each with its direction.
var (
	ampArmFoot  = [2]float64{0.571, 0.52}
	ampLowCross = [3]float64{0.507, 0.322, 80}
	ampUpCross  = [3]float64{0.284, 0.582, 42}
)
