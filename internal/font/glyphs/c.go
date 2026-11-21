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
// top as the round of the o, runs down the right-hand side and curls into the ball,
// which lies inside the end of the stroke.  The contour runs
//
//	from the top extreme of the outer edge, over the top and down the left side
//	round the bottom to the lower terminal (outerU, outerD, lowerTerminal), the
//	upper edge of the terminal back to the lowest point of the inner edge
//	the inner edge round the left side and over the top, as far as the point where
//	the inner edge of the stalk leaves it (innerD, innerU)
//	the inner edge of the stalk to the section of the terminal (cInner), the
//	terminal, and from its section the outer edge of the o up to the top extreme,
//	where the contour closes
//
// The stalk is the stroke of the o.  The terminal is placed with the outer edge
// of its section on the outer edge of the o, with the direction of the o there
// (cTerminalAt): the outer edge of the o runs on into the curl without a corner,
// and the thin top of the o, 32 to 35 units thick, is the stalk.  The turn of the
// part puts the rightmost ink of the ball CBallInset inside the rightmost ink of
// the letter, which is the tip of the lower terminal (cTerminal); the height of
// the ball follows from it.  The inner edge of the stalk is one cubic curve from
// the inner edge of the o, which it leaves with its direction and its curvature,
// to the section, at the cut with the smallest rate of change of curvature.
//
// Side kinds: a round left side, and an open right side, which is measured from
// the right extreme of the bowl, as for e: SideOpenRight.  The tip of the lower
// terminal lies a little inside that extreme; c has the same advance as e.

// cBallAngles are the angles (radians) between which cTerminal searches for the
// turn of the ball terminal of c.
var cBallAngles = [2]float64{-85 * math.Pi / 180, 30 * math.Pi / 180}

// cCuts returns the fractions of the parameter of the upper right quarter of the o,
// counted from the top, at which an edge of the stalk may leave the o.  The
// choice among them is made in cInner.
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

// cInner returns the inner edge of the upper stroke of c to the ball terminal tb:
// it leaves the inner edge of the o with its direction and its curvature, at the
// cut (cCuts) where its curvature changes least abruptly (curvatureRate), or with its
// direction only if no cut allows the curvature.  It returns the edge and the
// parameter of the cut on the upper right quarter of the inner edge of the o,
// counted from the top.
func cInner(f cFrame, tb placedBall) (cubic, float64) {
	quarterI := f.innerU.cubic(3, false) // top to the right extreme
	best, in, tI := math.Inf(1), cubic{}, 0.0
	for _, free := range []bool{false, true} {
		for _, u := range cCuts() {
			fi := stalkFoot{At: quarterI.at(u), Dir: quarterI.tangent(u), Kappa: quarterI.curvature(u)}
			if free {
				fi.Kappa = math.NaN()
			}
			if c, err := edgeIn(fi, tb.In); err == nil {
				if e := curvatureRate(c); e < best {
					best, in, tI = e, c, u
				}
			}
		}
		if !math.IsInf(best, 1) {
			return in, tI
		}
	}
	panic("glyphs: glyph c: no inner edge of the upper stroke leaves the o and reaches the ball terminal; glyphs.Check reports such parameters")
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

// cTerminalAt returns the ball terminal of c turned by angle and placed so that the
// outer edge of the stalk at its section touches the outer edge of the o, with the
// direction of the o there, and the parameter of that point on the upper right
// quarter of the outer edge (from the right extreme to the top).
func cTerminalAt(p design.Params, f cFrame, angle float64) (placedBall, float64) {
	quarterO := f.outerU.cubic(0, true)
	b := newBallPart(p)
	tb := b.place(ballPlace{Angle: angle})
	want := math.Atan2(tb.Out.Dir.Y, tb.Out.Dir.X)
	head := func(t float64) float64 { q := quarterO.tangent(t); return math.Atan2(q.Y, q.X) }
	lo, hi := 0.0, 1.0
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		if head(mid) < want {
			lo = mid
		} else {
			hi = mid
		}
	}
	t := (lo + hi) / 2
	q := quarterO.at(t)
	return b.place(ballPlace{Centre: outline.Point{X: q.X - tb.Out.At.X, Y: q.Y - tb.Out.At.Y}, Angle: angle}), t
}

// cTerminal returns the ball terminal of c: its rightmost ink CBallInset inside the
// rightmost ink of the letter, and its stalk the continuation of the outer edge of
// the o, which runs into the section of the part with its direction
// (cTerminalAt).  The turn of the part is found by bisection between cBallAngles.
func cTerminal(p design.Params, f cFrame) (placedBall, float64) {
	right := func(angle float64) float64 {
		tb, _ := cTerminalAt(p, f, angle)
		_, _, x1, _ := chainBBox(tb.Chain)
		return x1
	}
	solve := func(target float64) float64 {
		lo, hi := cBallAngles[0], cBallAngles[1]
		for i := 0; i < 60; i++ {
			mid := (lo + hi) / 2
			if right(mid) > target {
				lo = mid
			} else {
				hi = mid
			}
		}
		return (lo + hi) / 2
	}
	// The centre of the ball is moved to the nearest integer abscissa.
	target := f.tipX - p.CBallInset
	tb, _ := cTerminalAt(p, f, solve(target))
	target += math.Round(tb.Centre.X) - tb.Centre.X
	return cTerminalAt(p, f, solve(target))
}

// c draws the lowercase letter c, see the comment above.
func c(p design.Params) (outline.Path, float64) {
	return cOutline(p), design.CAdvance(p)
}

// cOutline draws the letter c.
func cOutline(p design.Params) outline.Path {
	f := newCFrame(p)
	top := p.XHeight + p.OvershootLower
	quarterO := f.outerU.cubic(0, true)  // right extreme to the top
	quarterI := f.innerU.cubic(3, false) // top to the right extreme
	terminal := lowerTerminal(f.outerD.cubic(3, true), p.CTipHeight, p.ETerminalTip,
		outline.Point{X: f.cx, Y: f.innerBottomY}, eShape)

	var path outline.Path
	path.MoveTo(f.cx, top)
	f.outerU.segment(&path, 1, true) // top to the left extreme
	f.outerD.segment(&path, 2, true) // left extreme to the bottom
	terminal.appendTo(&path)
	f.innerD.segment(&path, 1, false) // bottom to the left extreme
	f.innerU.segment(&path, 2, false) // left extreme to the top
	tb, tO := cTerminal(p, f)
	in, tI := cInner(f, tb)
	path.CurveTo(quarterI.sub(0, tI).args())
	join(splitAtExtrema(in), tb.Chain).appendTo(&path)
	path.CurveTo(quarterO.sub(tO, 1).args())
	return path
}
