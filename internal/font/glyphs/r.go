// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter r.
//
// The r is the left stem of n, with its head serif and both feet, and an arm
// which leaves the stem as the arch of n does, runs over the top as a thin, even
// arc and curls down into the ball terminal (ball.go) at the upper right.  It is
// one counter-clockwise contour, which runs
//
//	the stem from the baseline up to the join of the arm (stemUp), the right
//	foot being RFootRight long
//	the arm, as one chain (stalkChain): the underside from the stem to the
//	section of the terminal, the terminal, and the outer edge back over the top
//	to the notch on the stem
//	the top of the stem with the head serif, and its left-hand edge and foot
//	(stemCap, stemDown)
//
// The outer edge leaves the stem at the notch at rArmOuterAngle (66 degrees; the
// edge of the arch of n leaves at 56), the underside at the join at
// rArmInnerAngle.  Each is one cubic curve from the stem to the section of the
// terminal, where it takes the direction and the curvature of the stalk, with
// the handle at the stem which makes it fairest (fairEnd).  The arm passes into
// the curl of the terminal without a change of curvature, and the curvature
// changes evenly along it.
//
// The terminal is turned by rBallAngle from its own frame.  Its rightmost ink lies
// RReach right of the stem, and the crest of the arm, its highest point, at the
// height of the peak of the arch of n, XHeight + ArchOvershoot (placeUnderCrest).
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

// rArmChain returns the arm of r with the ball terminal tb: from the join
// on the stem along the underside, round the ball and back along the outer edge to
// the notch.
func rArmChain(p design.Params, tb placedBall) chain {
	x1 := rStem(p).right()
	return stalkChain("r",
		footAt(outline.Point{X: x1, Y: p.RJoin}, rArmInnerAngle*math.Pi/180),
		tb,
		footAt(outline.Point{X: x1, Y: p.RNotch}, rArmOuterAngle*math.Pi/180+math.Pi))
}

// rTerminal returns the ball terminal of r: its rightmost ink RReach right of the
// stem, and the highest point of the outer edge of the arm at the peak of the arch
// of n.
func rTerminal(p design.Params) placedBall {
	return placeUnderCrest(newBallPart(p), rBallAngle, rStem(p).right()+p.RReach, p.XHeight+p.ArchOvershoot,
		func(tb placedBall) chain { return rArmChain(p, tb) })
}

// rOutline draws the letter r.
func rOutline(p design.Params) outline.Path {
	stem := rStem(p)
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, p.RJoin)
	rArmChain(p, rTerminal(p)).appendTo(&path)
	stemCap(&path, p, stem)
	stemDown(&path, p, stem)
	return path
}
