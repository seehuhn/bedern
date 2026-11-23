// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter j.
//
// The j is the stem, the head serif and the dot of i, without foot serifs, with the
// stem continued below the baseline and curved to the left into a tail which curls
// up into the ball terminal (ball.go), the hook of f turned through half a turn.
// It is one counter-clockwise contour for the stem and a second for the dot
// (DotDiameter, DotCentre, DotShift and DotSuperness, those of i).  The stem
// outline runs
//
//	the top of the stem with the head serif (stemCap), from the top right corner to
//	the left-hand edge
//	down the left-hand edge to the height at which it begins to curve, JInnerDrop above
//	the lowest point, below the baseline
//	the tail (jTailChain): the inner edge, the terminal and the outer edge, which
//	ends on the right-hand edge of the stem, JOuterDrop above the lowest point
//	up the right-hand edge to the top right corner
//
// Each edge of the tail is one cubic curve from the stem, which it leaves
// vertically, to the section of the terminal, with the handle at the stem which
// makes it fairest; the stretch of the outer edge next to the stem, over which it
// turns by jRamp, is a transition curve which meets the stem without curvature.
// The ball is placed with the edges as stalkChain draws them.  The terminal is turned by jBallAngle, that of f and half a
// turn; the lowest ink lies JDepth below the baseline and the leftmost ink JReach to
// the left of the left-hand edge of the stem (placeAtExtreme).  The left edge of the
// stem stands at JStemLeft, which gives the tail room.
//
// Side kinds: the left side is an overhang (the tail may reach beyond the advance by
// MaxOverhang, and the left side space is negative), the right side that of the
// narrow letters.

// jBallAngle is the angle by which the ball terminal of j is turned from its own
// frame (radians): half a turn and that of f.
var jBallAngle = math.Pi + 15*math.Pi/180

// jStem returns the stem of j: the stem of i with its head serif, without the foot
// serifs.
func jStem(p design.Params) stemSpec {
	return stemSpec{
		Left: p.JStemLeft, Width: p.Stem, Bottom: 0, Top: p.XHeight,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
}

// jHook returns the hook of j in the frame in which it stands on the stem, before
// it is turned round: the crest is at the height JDepth, and the edges of the stem
// are those of the stem of j.
func jHook(p design.Params) hookSpec {
	s := jStem(p)
	return hookSpec{
		XL: s.Left, XR: s.right(), Cell: s.right(),
		Apex: p.JDepth, Reach: p.JReach,
		OuterJoin: p.JDepth - p.JOuterDrop, InnerJoin: p.JDepth - p.JInnerDrop,
	}
}

// jTailChain returns the tail of j with the ball terminal tb: from the left-hand
// edge of the stem, where it begins to curve, along the inner edge, round the ball
// and back along the outer edge to the right-hand edge.  The ball is placed with
// these edges (jTerminal).
func jTailChain(p design.Params, tb placedBall) chain {
	in, out := jTailFeet(p)
	return stalkChain("j", in, tb, out)
}

// jTailFeet returns the feet of the edges of the tail of j on the stem.
func jTailFeet(p design.Params) (in, out stalkFoot) {
	s, h := jStem(p), jHook(p)
	return footRising(outline.Point{X: s.Left, Y: -h.InnerJoin}, -math.Pi/2),
		footAt(outline.Point{X: s.right(), Y: -h.OuterJoin}, math.Pi/2)
}

// jRamp is the turn in degrees over which the outer edge of the tail of j runs
// into the stem without curvature (stalkChainStraight).
const jRamp = 15.0

// jTails holds the tails of j by the parameters, so that j, ŋ and the
// ligature of f and j of one set of parameters draw the tail once.
var jTails sync.Map // design.Params -> func() chain

// jTail returns jTailDrawn with the terminal of j.  The chain is shared and
// must not be changed.
func jTail(p design.Params) chain {
	f, _ := jTails.LoadOrStore(p, sync.OnceValue(func() chain { return jTailDrawn(p, jTerminal(p)) }))
	return f.(func() chain)()
}

// jTailDrawn returns the tail of j as it is drawn: jTailChain with an outer
// edge which meets the straight right-hand edge of the stem without curvature
// (stalkChainStraight).
func jTailDrawn(p design.Params, tb placedBall) chain {
	in, out := jTailFeet(p)
	return stalkChainStraight("j", in, tb, out, jRamp)
}

// jTerminal returns the ball terminal of j: its leftmost ink JReach left of the
// stem, and the lowest point of the tail JDepth below the baseline.
func jTerminal(p design.Params) placedBall {
	return placeAtExtreme(newBallPart(p), ballPlace{Angle: jBallAngle},
		func(x0, _ float64) float64 { return design.JInkLeft(p) - x0 }, -p.JDepth, false,
		func(tb placedBall) chain { return jTailChain(p, tb) })
}

// j draws the lowercase letter j, see the comment above.
func j(p design.Params) (outline.Path, float64) {
	return jOutline(p), design.JAdvance(p)
}

// jOutline draws the letter j.
func jOutline(p design.Params) outline.Path {
	stem := jStem(p)
	var path outline.Path
	path.MoveTo(stem.right(), stem.Top)
	stemCap(&path, p, stem)
	tail := jTail(p)
	path.LineTo(tail.start().X, tail.start().Y)
	tail.appendTo(&path)
	path.LineTo(stem.right(), stem.Top)
	cx := stem.Left + p.Stem/2 + design.DotOffset(p)
	r := p.DotDiameter / 2
	superEllipse(&path, cx-r, p.DotCentre-r, cx+r, p.DotCentre+r, p.DotSuperness, true)
	return path
}
