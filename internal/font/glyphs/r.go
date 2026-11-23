// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter r.
//
// The r is the left stem of n, with its head serif and both feet, and an arm
// which leaves the stem as the arch of n does, runs over the top as a thin, even
// arc and winds down into the ball terminal (ball.go) at the upper right.  It is
// one counter-clockwise contour, which runs
//
//	the stem from the baseline up to the join of the arm (stemUp), the right
//	foot being RFootRight long
//	the arm, as one chain (windHook): the underside from the stem to the tip of
//	the wedge, the ball, and the outer edge back over the top to the notch on
//	the stem
//	the top of the stem with the head serif, and its left-hand edge and foot
//	(stemCap, stemDown)
//
// The outer edge leaves the stem at the notch at rArmOuterAngle (66 degrees; the
// edge of the arch of n leaves at 56), the underside at the join at
// rArmInnerAngle.  The edges which stalkChain draws from there to the section of
// the part, each one cubic curve with the direction and the curvature of the
// stalk at the section and the handle at the stem which makes it fairest
// (fairEnd), place the terminal: it is turned by rBallAngle from its own frame,
// its rightmost ink lies RReach right of the stem, and the crest of the outer
// edge, its highest point, at the height of the peak of the arch of n, XHeight +
// ArchOvershoot (placeUnderCrest).  The outer edge follows that edge from the
// notch over the top and winds into the ball on a spiral from the point where
// it rises an angle of rWindRises above the horizontal; the underside runs in
// one cubic curve to the tip of the wedge (windHook).
//
// Side kinds: a stem on the left, with RSideLeft; the right side is open, the
// advance being made from RReach and RSideRight (design.RAdvance).

// rArmOuterAngle and rArmInnerAngle are the directions (degrees) in which the outer
// edge and the underside of the arm leave the stem.
const (
	rArmOuterAngle = 66.0
	rArmInnerAngle = 62.0
)

// rStem returns the stem of r.
func rStem(p design.Params) stemSpec {
	return stemSpec{
		Left: design.RStemLeft(p), Width: p.Stem, Top: p.XHeight,
		FootLeft: p.SerifJut, FootRight: p.RFootRight,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
}

// r draws the lowercase letter r, see the comment above.
func r(p design.Params) (outline.Path, float64) {
	return rOutline(p), design.RAdvance(p)
}

// rBallAngle is the angle by which the ball terminal of r is turned from its own
// frame (radians), clockwise.
var rBallAngle = 40 * math.Pi / 180

// rArmFeet returns the feet of the two edges of the arm on the stem: the join,
// where the underside leaves, and the notch, where the outer edge ends.
func rArmFeet(p design.Params) (join, notch stalkFoot) {
	x1 := rStem(p).right()
	return footAt(outline.Point{X: x1, Y: p.RJoin}, rArmInnerAngle*math.Pi/180),
		footAt(outline.Point{X: x1, Y: p.RNotch}, rArmOuterAngle*math.Pi/180+math.Pi)
}

// rArms holds the arms of r by the parameters, so that r and the tests of one
// set of parameters draw the arm once.
var rArms sync.Map // design.Params -> func() chain

// rArm returns rArmChain with the terminal of r.  The chain is shared and must
// not be changed.
func rArm(p design.Params) chain {
	f, _ := rArms.LoadOrStore(p, sync.OnceValue(func() chain { return rArmChain(p, rTerminal(p)) }))
	return f.(func() chain)()
}

// rArmChain returns the arm of r with the ball terminal tb: from the join on the
// stem along the underside, round the ball and back along the outer edge to the
// notch (windHook).
func rArmChain(p design.Params, tb placedBall) chain {
	join, notch := rArmFeet(p)
	return windHook("r", join, tb, notch, rWindRises)
}

// rWindRises is the range of the directions (degrees above the horizontal) in
// which the outer edge of the arm, walked towards the ball, may run where it
// begins to wind into the ball (windHook).
var rWindRises = [2]float64{25, 45}

// rTerminal returns the ball terminal of r: its rightmost ink RReach right of the
// stem, and the highest point of the outer edge of the arm which stalkChain draws
// at the peak of the arch of n.
func rTerminal(p design.Params) placedBall {
	join, notch := rArmFeet(p)
	return placeUnderCrest(newBallPart(p), rBallAngle, rStem(p).right()+p.RReach, p.XHeight+p.ArchOvershoot,
		func(tb placedBall) chain { return stalkChain("r", join, tb, notch) })
}

// rOutline draws the letter r.
func rOutline(p design.Params) outline.Path {
	stem := rStem(p)
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, p.RJoin)
	rArm(p).appendTo(&path)
	stemCap(&path, p, stem)
	stemDown(&path, p, stem)
	return path
}
