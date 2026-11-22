// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The letter c.
//
// The c is the curve of the o, open on the right: one counter-clockwise contour
// without a counter.  The left side, the top and the bottom are the curves of e
// (newEFrameWidth): the outer edge has the thickest part at the left extreme,
// the thin parts at the top and the bottom, the overshoots and the squareness of
// the o, and the inner edge is that of the o.  The lower end is the terminal of
// e (lowerTerminal, eShape): the outer edge is cut CTipHeight above the baseline
// and the upper edge of the terminal runs back to the lowest point of the inner
// edge.  The upper end is the ball terminal (ball.go): the stroke comes over the
// top as the round of the o, runs down the right-hand side and winds into the
// ball, which lies inside the end of the stroke.  The contour runs
//
//	from the top extreme of the outer edge, over the top and down the left side
//	round the bottom to the lower terminal (outerU, outerD, lowerTerminal), the
//	upper edge of the terminal back to the lowest point of the inner edge
//	the inner edge round the left side and over the top, as far as the point where
//	the inner edge of the stalk leaves it (innerD, innerU)
//	the inner edge of the stalk to the tip of the wedge, round the ball, and the
//	outer edge back to the top extreme, where the contour closes (cStalk)
//
// The end of the stroke is that of the figures (roundTerminal with Tip).  The
// part is placed with the outer edge of its section on the outer edge of the o,
// with the direction of the o there (figBallPlace), turned so that the rightmost
// ink of the ball lies CBallInset inside the rightmost ink of the letter, which
// is the tip of the lower terminal (cTerminal).  Of the part the letter draws the
// circle of the ball and the tip of the wedge.  The outer edge winds from the
// top extreme of the o into the ball on a spiral (windSpiral).  The inner edge of
// the stalk is one cubic curve from the inner edge of the o, which it leaves with
// its direction and its curvature, to the tip of the wedge, where it has the
// direction and the curvature of the inner edge of the part, at the cut with the
// smallest rate of change of curvature (stalkToTip).
//
// Side kinds: a round left side, CSideLeft to the left extreme, and an open right
// side, CSideRight from the rightmost ink, the tip of the lower terminal.

// cCuts returns the fractions of the parameter of the upper right quarter of the o,
// counted from the top, at which the inner edge of the stalk may leave the o.  The
// choice among them is made in stalkToTip.
func cCuts() []float64 {
	var us []float64
	for u := 0.02; u <= 0.7+1e-9; u += 0.01 {
		us = append(us, u)
	}
	return us
}

// cFrame holds the curves from which c is drawn.
type cFrame struct {
	eFrame
	tipX float64 // the rightmost ink of the letter, the tip of the lower terminal
}

// newCFrame returns the curves of c.  The width of the bowl is found by iteration such
// that the tip of the lower terminal is CBowlWidth from the left extreme.
func newCFrame(p design.Params) cFrame {
	var f cFrame
	width := p.CBowlWidth
	for i := 0; i < 50; i++ {
		f.eFrame = newEFrameAt(p, design.CBowlLeft(p), width)
		outer := f.outerD.cubic(3, true)
		f.tipX = outer.at(outer.tAtY(p.CTipHeight)).X
		step := f.left + p.CBowlWidth - f.tipX
		if math.Abs(step) < 1e-9 {
			break
		}
		width += step
	}
	return f
}

// curvatureRate returns the largest rate of change of the curvature of c along
// the arc length, from 400 samples.
func curvatureRate(c cubic) float64 {
	const n = 400
	var m float64
	prevK, prev := c.curvature(0), c.at(0)
	for i := 1; i <= n; i++ {
		t := float64(i) / n
		q, k := c.at(t), c.curvature(t)
		if ds := math.Hypot(q.X-prev.X, q.Y-prev.Y); ds > 0 {
			m = math.Max(m, math.Abs(k-prevK)/ds)
		}
		prev, prevK = q, k
	}
	return m
}

// cTerminal returns the ball terminal of c as it is placed on the outer edge of
// the o (figBallPlace): its rightmost ink CBallInset inside the tip of the lower
// terminal.
func cTerminal(p design.Params, f cFrame) placedBall {
	return figBallPlace("c", newBallPart(p), f.outerU.cubic(0, true), f.tipX-p.CBallInset)
}

// cStalk returns the end of the upper stroke of c in the ball terminal, from the
// top of the inner edge of the o round the ball to the top of the outer edge
// (roundTerminal).
func cStalk(p design.Params, f cFrame) chain {
	return roundTerminal("c", newBallPart(p), f.outerU.cubic(0, true), f.innerU.cubic(3, false), f.tipX-p.CBallInset, figStroke{Tip: true})
}

// c draws the lowercase letter c, see the comment above.
func c(p design.Params) (outline.Path, float64) {
	return cOutline(p), design.CAdvance(p)
}

// cOutline draws the letter c.
func cOutline(p design.Params) outline.Path {
	f := newCFrame(p)
	top := p.XHeight + p.OvershootLower
	terminal := lowerTerminal(f.outerD.cubic(3, true), p.CTipHeight, p.ETerminalTip,
		outline.Point{X: f.cx, Y: f.innerBottomY}, eShape)

	var path outline.Path
	path.MoveTo(f.cx, top)
	f.outerU.segment(&path, 1, true) // top to the left extreme
	f.outerD.segment(&path, 2, true) // left extreme to the bottom
	terminal.appendTo(&path)
	f.innerD.segment(&path, 1, false) // bottom to the left extreme
	f.innerU.segment(&path, 2, false) // left extreme to the top
	cStalk(p, f).appendTo(&path)
	return path
}
