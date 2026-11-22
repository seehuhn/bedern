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
// The form is a figure of eight drawn in one stroke, with an arm.  A round
// loop stands at the top, left of the middle.  Its left side, ampLoopLeftW
// times AmpThick thick, turns into the heavy diagonal, a straight band at
// ampDiagAngle degrees to the horizontal which thickens to AmpThick and runs
// out at the lower right into a tail: the stroke turns along the bottom, thins
// to AmpTail and ends in a cut at right angles to it, rising ampTailDir
// degrees, above the baseline at the right end of the ink.  The right side of
// the loop is thin; the thin stroke comes down from it to the left, crosses the
// diagonal and swings into the lower bowl, the widest part of the sign, with
// its weight, AmpBowlThick, at the extreme of its left side.  From the bottom
// of the bowl the stroke rises to the right, crosses the diagonal again at the
// lower crossing and runs on straight as the arm, AmpArmThick thick, to the head
// serif of a diagonal of the capitals (diagonal.go), whose top lies at AmpArm,
// whose right tip lies ampSerifIn inside the right end of the ink and which
// reaches AmpSerifLeft and AmpSerifRight beyond the edges of the arm.  The
// bracket of the serif meets both edges of the arm at the mean of the runs of
// the obtuse and the acute side, with the reach of the acute side.
//
// It is one stroke (symstroke.go), drawn from the arm to the tail: down the
// arm, across the diagonal, round the bottom of the bowl and up its left side,
// across the diagonal to the right side of the loop, over its top and down its
// left side, down the diagonal and round into the tail.  Each station has the
// direction of the stroke and its thickness; the edges between the stations
// are cubics with the curvatures which the neighbouring stations ask for, cut
// at their extremes.  The places of the stations follow from a few rules:
//
//   - The diagonal passes through the point (ampUpX, ampUpY) at ampDiagAngle;
//     the lower crossing is the point of its centre line at ampLowY, and the arm
//     is the straight line from there to the serif.
//   - The stroke leaves the lower crossing in the direction of the arm and
//     turns on a circular arc into the bottom of the bowl, which it reaches at
//     the overshoot; the extreme of the left side lies on the quarter circle
//     from the bottom.
//   - The right side of the loop lies where a circular arc from (ampUpX,
//     ampUpY), rising ampUpDir degrees, turns upright at ampLoopRightY; its
//     left side lies at ampLoopLeftY where a curve whose curvature falls evenly
//     to nothing over ampLoopTurn turns from upright into the diagonal, where
//     the straight part begins.
//   - The straight part of the diagonal ends at ampDiagEndY, where its centre
//     line takes the small curvature ampTailK; the tail is one curve from there
//     to the cut, with the curvature at the cut which gives its centre line the
//     handles of a circular arc at the diagonal.
//
// The stroke crosses itself twice, and the outline is the union of it and of
// itself (symunion.go).  The places are fractions of the ink width AmpInk
// (abscissae) and of the height from AmpBottom to AmpTop (heights), and the
// directions are in degrees from the positive x axis.
//
// Side kinds: open, AmpSideLeft and AmpSideRight to the ink.
const (
	// The point on the diagonal from which the right side of the loop is
	// placed, and the direction of the arc from it.
	ampUpX   = 0.35
	ampUpY   = 0.54
	ampUpDir = 20.0

	// The loop: the heights of the extremes of its right and its left side,
	// the thickness of its left side as a fraction of AmpThick, the length of
	// the turn from there into the diagonal (units), and the abscissa of its
	// top as a fraction of the way from its left to its right.
	ampLoopRightY = 0.80
	ampLoopLeftY  = 0.86
	ampLoopLeftW  = 0.8
	ampLoopTurn   = 240.0
	ampLoopTopAt  = 0.5

	// The diagonal: its angle below the horizontal, its thickness where its
	// straight part begins as a fraction of AmpThick, and the height where the
	// straight part ends.
	ampDiagAngle  = 55.0
	ampDiagStartW = 0.9
	ampDiagEndY   = 0.15

	// The height of the lower crossing.
	ampLowY = 0.19

	// The tail: the curvature of the centre line where it leaves the
	// diagonal, the height of the middle of the cut above the baseline (units),
	// and the direction in which it ends.
	ampTailK    = 0.003
	ampTailEndY = 34.0
	ampTailDir  = 22.0

	// How far the right tip of the serif of the arm lies inside the right end
	// of the ink.
	ampSerifIn = 14.0
)

// ampFrame holds the frame of the ampersand.
type ampFrame struct {
	x0, w, top, bottom, h float64
}

func newAmpFrame(p design.Params) ampFrame {
	s := p.Sym
	f := ampFrame{x0: s.AmpSideLeft, w: s.AmpInk, top: s.AmpTop, bottom: s.AmpBottom}
	f.h = f.top - f.bottom
	return f
}

// x and y return the abscissa at the fraction u of the ink width and the height
// at the fraction v of the height.
func (f ampFrame) x(u float64) float64 { return f.x0 + u*f.w }
func (f ampFrame) y(v float64) float64 { return f.bottom + v*f.h }

// diagX returns the abscissa of the centre line of the diagonal at the height
// y.
func (f ampFrame) diagX(y float64) float64 {
	return f.x(ampUpX) + (f.y(ampUpY)-y)/math.Tan(ampDiagAngle*math.Pi/180)
}

// ampArm returns the arm as a diagonal: a straight stroke through the point low
// whose right edge lies AmpSerifRight left of the right tip of the serif at the
// underside of the serif.
func ampArm(p design.Params, f ampFrame, low outline.Point) diagonal {
	s := p.Sym
	tip := f.x0 + f.w - ampSerifIn
	under := s.AmpArm - p.Caps.SerifThick
	d := diagonal{X: low.X, Y: low.Y, Thick: s.AmpArmThick}
	for i := 0; i < 8; i++ {
		d.Lean = (tip - s.AmpSerifRight - d.width()/2 - low.X) / (under - low.Y)
	}
	return d
}

// ampTurnRadius returns the radius of the circular arc which turns from the
// direction dir (degrees, a descent) to the horizontal over the drop d.
func ampTurnRadius(dir, d float64) float64 {
	return d / (1 - math.Cos(dir*math.Pi/180))
}

// ampSpiral returns the displacement along a curve of the length l which
// leaves in the direction from (degrees) and turns through turn degrees with a
// curvature which falls evenly to nothing at its end; reversed, it rises from
// nothing.
func ampSpiral(from, turn, l float64) outline.Point {
	const n = 200
	k0 := 2 * turn * math.Pi / 180 / l
	var d outline.Point
	for i := 0; i < n; i++ {
		u := (float64(i) + 0.5) / n * l
		a := from*math.Pi/180 + k0*(u-u*u/(2*l))
		d.X += math.Cos(a) * l / n
		d.Y += math.Sin(a) * l / n
	}
	return d
}

// ampEndCurvature returns the curvature at q3, reached in the direction d3
// (degrees), of the cubic from q0, which it leaves in the direction d0 with
// the curvature k0 and the handle of the circular arc between the two
// directions.
func ampEndCurvature(q0 outline.Point, d0, k0 float64, q3 outline.Point, d3 float64) float64 {
	cr := func(u, v outline.Point) float64 { return u.X*v.Y - u.Y*v.X }
	t0, t3 := dirOf(d0*math.Pi/180), dirOf(d3*math.Pi/180)
	d := outline.Point{X: q3.X - q0.X, Y: q3.Y - q0.Y}
	s, c1, c2 := cr(t0, t3), cr(t0, d), cr(d, t3)
	a := arcHandle(q0, t0, q3, t3)
	b := (c1 - 1.5*k0*a*a) / s
	return (c2 - a*s) / (1.5 * b * b)
}

func ampersand(p design.Params) (outline.Path, float64) {
	s := p.Sym
	f := newAmpFrame(p)
	t := p.Caps.RoundThin
	st := func(q outline.Point, dir, w float64) station { return station{X: q.X, Y: q.Y, Dir: dir, W: w} }
	pt := func(x, y float64) outline.Point { return outline.Point{X: x, Y: y} }

	up := pt(f.x(ampUpX), f.y(ampUpY))
	low := pt(f.diagX(f.y(ampLowY)), f.y(ampLowY))

	arm := ampArm(p, f, low)
	armDir := math.Atan2(1, arm.Lean) * 180 / math.Pi
	br := design.CapDiagBracket(p)
	br.Obtuse = (br.Obtuse + br.Acute) / 2
	br.Acute, br.Reach = br.Obtuse, br.ReachAcute

	// The bowl.
	botY := f.bottom + t/2
	rb := ampTurnRadius(armDir, low.Y-botY)
	bot := pt(low.X-rb*math.Sin(armDir*math.Pi/180), botY)
	leftX := f.x0 + s.AmpBowlThick/2
	side := pt(leftX, botY+(bot.X-leftX))

	// The loop and the start of the straight diagonal.
	ry := f.y(ampLoopRightY)
	rx := up.X + math.Tan((90-ampUpDir)/2*math.Pi/180)*(ry-up.Y)
	ly := f.y(ampLoopLeftY)
	dl := ampSpiral(-90, 90-ampDiagAngle, ampLoopTurn)
	qy := ly + dl.Y
	lx := f.diagX(qy) - dl.X
	lw := ampLoopLeftW * s.AmpThick
	topX := lx - lw/2 + ampLoopTopAt*(rx+t/2-(lx-lw/2))
	da := st(pt(f.diagX(qy), qy), -ampDiagAngle, ampDiagStartW*s.AmpThick)
	da.Flat = true

	// The end of the straight diagonal and the tail.
	dy := f.y(ampDiagEndY)
	db := st(pt(f.diagX(dy), dy), -ampDiagAngle, s.AmpThick)
	db.K = ampTailK
	ta := ampTailDir * math.Pi / 180
	end := pt(f.x0+f.w-s.AmpTail/2*math.Sin(ta), ampTailEndY)
	kEnd := ampEndCurvature(pt(db.X, db.Y), -ampDiagAngle, ampTailK, end, ampTailDir)

	kBowl := -1 / (bot.X - leftX)

	// The stroke starts where the brackets of the serif end, with a
	// horizontal nib, and is straight down to the lower crossing.
	brY := s.AmpArm - p.Caps.SerifThick - br.Obtuse
	sts := []station{
		{X: arm.centreAt(brY), Y: brY, Dir: armDir + 180, W: arm.width(), Nib: 360, Flat: true},
		st(low, armDir+180, s.AmpArmThick),
		{X: bot.X, Y: bot.Y, Dir: 180, W: t, K: kBowl},
		{X: side.X, Y: side.Y, Dir: 90, W: s.AmpBowlThick},
		st(pt(rx, ry), 90, t),
		st(pt(topX, f.top-t/2), 180, t),
		st(pt(lx, ly), -90, lw),
		da,
		db,
		{X: end.X, Y: end.Y, Dir: ampTailDir, W: s.AmpTail, K: kEnd},
	}
	right, left := strokeOpen(sts)
	outR := arm.headServif(edgeRight, s.AmpArm, diagSerif{Over: s.AmpSerifRight, Thick: p.Caps.SerifThick, Bracket: br}).reversed()
	outL := arm.headServif(edgeLeft, s.AmpArm, diagSerif{Over: s.AmpSerifLeft, Thick: p.Caps.SerifThick, Bracket: br})
	// The top of the serif is one straight piece from tip to tip.
	outR, outL = outR[:len(outR)-1], outL[1:]
	// The edges of the stroke begin where the brackets end.
	stroke := join(right, lineChain(right.end(), left.start()), left, outR,
		lineChain(outR.end(), outL.start()), outL)
	path := chainsPath(union(stroke)...)
	_, _, x1, _ := path.BBox()
	return path, x1 + s.AmpSideRight
}
