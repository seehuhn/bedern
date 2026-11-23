// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter f.
//
// The f is the ascender stem of l with its foot serifs (the left one SerifJut
// long, the right one FFootRight), a round hook which grows out of the top of the
// stem and winds down into the ball terminal (ball.go), and a crossbar at the
// x-height.  It is one counter-clockwise contour without a counter; the bar and
// the stem are one outline.  It runs
//
//	the stem from the baseline up the right-hand edge to the underside of the
//	crossbar (stemUp)
//	the right-hand part of the bar (barRight), and on up the right-hand edge to the
//	point at which the underside of the hook leaves it
//	the hook as one chain (fHookChain): the underside, the terminal and the outer
//	edge, which ends on the left-hand edge of the stem
//	down the left-hand edge to the top of the bar, the left-hand part of the bar
//	(barLeft), and down the left-hand edge and the left foot serif (stemDown)
//
// The hook.  The left-hand edge of the stem becomes the outer edge of the hook, and
// the right-hand edge its underside (windHook).  The ball terminal is placed with
// the crest of the edges which stalkChain draws from the stem, FOuterDrop and
// FInnerDrop below the top, to the section of the part FOvershoot above the
// ascender, the rightmost ink of the ball FReach beyond the cell of the stem
// (placeUnderCrest), and the part turned by fBallAngle.  The underside runs in one
// cubic curve to the tip of the wedge; the outer edge follows the edge of
// stalkChain over the crest and winds into the ball on a spiral from the point
// where it rises an angle of fWindRises above the horizontal.  Both edges leave
// the stem without curvature: the last stemRamp degrees of their turn towards
// the stem are a transition curve whose foot lies a little below the drop.
//
// The crossbar has the thickness Hair, its top at the x-height, and extends FBarLeft
// to the left and FBarRight to the right of the stem; its ends are plain cuts.
//
// Side kinds: the left side is measured to the end of the bar or the foot serif
// (design.FInkLeft); the right side is FSideRight from the end of the hook, which
// overhangs the advance by -FSideRight (55), at most MaxOverhang.

// fBallAngle is the angle by which the ball terminal of f is turned from its own
// frame (radians).
var fBallAngle = 40 * math.Pi / 180

// fStem returns the stem of f.
func fStem(p design.Params) stemSpec {
	return stemSpec{
		Left: design.FStemLeft(p), Width: p.AscenderStem, Top: p.Ascender,
		FootLeft: p.SerifJut, FootRight: p.FFootRight,
	}
}

// fHook returns the hookSpec of f.
func fHook(p design.Params) hookSpec {
	s := fStem(p)
	apex := design.FApex(p)
	return hookSpec{
		XL: s.Left, XR: s.right(), Cell: design.FCellRight(p),
		Apex: apex, Reach: p.FReach,
		OuterJoin: apex - p.FOuterDrop, InnerJoin: apex - p.FInnerDrop,
	}
}

// fBar returns the barSpec of f.
func fBar(p design.Params) barSpec {
	s := fStem(p)
	return barSpec{
		Top: p.XHeight, Thick: p.Hair, End: p.Hair,
		LeftEdge: s.Left, RightEdge: s.right(),
		Left: p.FBarLeft, Right: p.FBarRight,
	}
}

// fHookFeet returns the feet of the two edges of the hook on the stem: on the
// right-hand edge, where the underside begins to curve, and on the left-hand
// edge, where the outer edge does.  Both leave the stem vertically.
func fHookFeet(p design.Params) (in, out outline.Point) {
	h := fHook(p)
	return outline.Point{X: h.XR, Y: h.InnerJoin}, outline.Point{X: h.XL, Y: h.OuterJoin}
}

// fHookChain returns the hook of f with the ball terminal tb: from the right-hand
// edge of the stem along the underside, round the ball and back along the outer
// edge to the left-hand edge (windHook).  Both edges leave the stem without
// curvature.
func fHookChain(p design.Params, tb placedBall) chain {
	in, out := fHookFeet(p)
	return must(windHook("f",
		stalkFoot{At: in, Dir: dirOf(math.Pi / 2)},
		tb,
		stalkFoot{At: out, Dir: dirOf(-math.Pi / 2)}, fWindRises))
}

// fWindRises is the range of the directions (degrees above the horizontal) in
// which the outer edge of the hook, walked towards the ball, may run where it
// begins to wind into the ball (windHook).
var fWindRises = [2]float64{20, 40}

// fTerminal returns the ball terminal of f: its rightmost ink FReach beyond the
// cell of the stem, and the crest of the hook which stalkChain draws FOvershoot
// above the ascender, see rTerminal.
func fTerminal(p design.Params) placedBall {
	h := fHook(p)
	in, out := fHookFeet(p)
	return must(placeUnderCrest(newBallPart(p), fBallAngle, h.Cell+h.Reach, h.Apex, func(tb placedBall) chain {
		return stalkChain("f", footAt(in, math.Pi/2), tb, footAt(out, -math.Pi/2))
	}))
}

// f draws the lowercase letter f, see the comment above.
func f(p design.Params) (outline.Path, float64) {
	return fOutline(p), design.FAdvance(p)
}

// fOutline draws the letter f.
func fOutline(p design.Params) outline.Path {
	stem, bar := fStem(p), fBar(p)
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, bar.Top-bar.Thick)
	barRight(&path, bar)
	h := newFHookParts(p).chain
	path.LineTo(h.start().X, h.start().Y)
	h.appendTo(&path)
	path.LineTo(stem.Left, bar.Top)
	barLeft(&path, bar)
	stemDown(&path, p, stem)
	return path
}
