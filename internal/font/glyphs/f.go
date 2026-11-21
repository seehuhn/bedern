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
// stem and curls down into the ball terminal (ball.go), and a crossbar at the
// x-height.  It is one counter-clockwise contour without a counter; the bar and
// the stem are one outline.  It runs
//
//	the stem from the baseline up the right-hand edge to the underside of the
//	crossbar (stemUp)
//	the right-hand part of the bar (barRight), and on up the right-hand edge to the
//	height at which the underside of the hook leaves it
//	the hook as one chain (fHookChain): the underside, the terminal and the outer
//	edge, which ends on the left-hand edge of the stem
//	down the left-hand edge to the top of the bar, the left-hand part of the bar
//	(barLeft), and down the left-hand edge and the left foot serif (stemDown)
//
// The hook.  The left-hand edge of the stem becomes the outer edge of the hook, and
// the right-hand edge its underside; both leave the stem vertically, FOuterDrop and
// FInnerDrop below the top, and are each one cubic curve to the section of the
// terminal, which they meet with the direction and the curvature of the stalk.
// The crest of the hook lies FOvershoot above the ascender, the rightmost ink of
// the ball FReach beyond the cell of the stem (placeUnderCrest), and the terminal
// is turned by fBallAngle.
//
// The crossbar has the thickness Hair, its top at the x-height, and extends FBarLeft
// to the left and FBarRight to the right of the stem; its ends are plain cuts.
//
// Side kinds: the left side is measured to the end of the bar or the foot serif
// (design.FInkLeft); the right side is FSideRight (5) inside the end of the hook, and
// the hook may reach beyond the advance by MaxOverhang if FSideRight is negative.

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

// fHookChain returns the hook of f with the ball terminal tb: from the right-hand
// edge of the stem along the underside, round the ball and back along the outer
// edge to the left-hand edge.  Both edges leave the stem vertically.
func fHookChain(p design.Params, tb placedBall) chain {
	h := fHook(p)
	return stalkChain("f",
		footAt(outline.Point{X: h.XR, Y: h.InnerJoin}, math.Pi/2),
		tb,
		footAt(outline.Point{X: h.XL, Y: h.OuterJoin}, -math.Pi/2))
}

// fTerminal returns the ball terminal of f: its rightmost ink FReach beyond the
// cell of the stem, and the crest of the hook FOvershoot above the ascender, see
// rTerminal.
func fTerminal(p design.Params) placedBall {
	h := fHook(p)
	return placeUnderCrest(newBallPart(p), fBallAngle, h.Cell+h.Reach, h.Apex, func(tb placedBall) chain { return fHookChain(p, tb) })
}

// f draws the lowercase letter f, see the comment above.
func f(p design.Params) (outline.Path, float64) {
	return fOutline(p), design.FAdvance(p)
}

// fOutline draws the letter f.
func fOutline(p design.Params) outline.Path {
	stem, bar, hook := fStem(p), fBar(p), fHook(p)
	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, bar.Top-bar.Thick)
	barRight(&path, bar)
	path.LineTo(stem.right(), hook.InnerJoin)
	fHookChain(p, fTerminal(p)).appendTo(&path)
	path.LineTo(stem.Left, bar.Top)
	barLeft(&path, bar)
	stemDown(&path, p, stem)
	return path
}
