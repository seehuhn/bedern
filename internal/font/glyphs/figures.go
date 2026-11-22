// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The figures 0 to 9.
//
// The figures are lining and tabular: all ten have the advance Figures.Advance
// and each stands centred in it, moved by its optical shift (design.FigureLeft);
// their sides are of the kind sideFigure.  The flat tops (1, 4, 7, the flag of
// 5) are at Figures.Height, the round tops OverTop above it, and the round
// bottoms OvershootUpper below the baseline, as for the capital O.
//
// The parts are those of the letters.  The rounds are super-ellipses with the
// squareness of the o (super.go); the stems of 1 and 4 carry the bracketed foot
// serif of the capitals (capstem.go); the end of the base of 2 and the left end
// of the bar of 7 are the vertical serifs of the arms of E, F and T (caparm.go),
// at the weight of the heavy strokes; and 2, 3, 5, 6 and 9 end in the ball
// terminal of the lowercase (ball.go), scaled to the size Figures.Ball with its
// terminal of the lowercase (ball.go), scaled to the size Figures.Ball with its
// stalk (figStalk).  The rounds which carry a ball reach beyond it on the open
// side.
//
// The ball is placed on a round stroke as the part would sit on c, but of the
// part the figures draw only the circle of the ball and the tip of the wedge:
// the outer edge winds from the extreme of the round into the circle of the ball
// as a spiral whose curvature rises to that of the ball (windSpiral), without
// the step to the curl of the part, and the inner edge runs to the tip of the
// wedge either as the stalk of the part (3, 5) or as one round arch from the top
// of the counter (2, 6; roundTerminal).  The construction is written once, in
// the frame of c, where the ball hangs at the upper right of a round whose
// stroke comes over the top from the left; a figure whose terminal stands
// elsewhere takes the two quarters of its round into that frame (mirrored in a
// vertical line for a ball at the upper left, turned through half a turn for
// one at the lower left), and takes the result back.

// figBall returns the ball terminal of the figures: the part of the lowercase,
// Figures.Ball across on the stalk figStalk.
func figBall(p design.Params) ballPart {
	b, err := makeBallPart(p.Figures.Ball/2, figStalk(p), ballCurl, ballSection, ballBlend)
	if err != nil {
		panic(fmt.Sprintf("glyphs: figures: %v; glyphs.Check reports such parameters", err))
	}
	return b
}

// figStalk returns the thickness of the stalk of the ball terminal of the
// figures: the stalk of the lowercase scaled with the ball.
func figStalk(p design.Params) float64 { return p.BallStalk * p.Figures.Ball / p.BallDiameter }

// figTop and figBottom return the height of the round tops and the depth of the
// round bottoms of the figures.
func figTop(p design.Params) float64    { return p.Figures.Height + p.Figures.OverTop }
func figBottom(p design.Params) float64 { return -p.OvershootUpper }

// figBallAngles are the turns of the part (radians) between which roundTerminal
// searches; the more the part is turned clockwise, the further right the ball
// stands.
var figBallAngles = [2]float64{-85 * math.Pi / 180, 30 * math.Pi / 180}

// roundTerminalAt returns the part turned by angle and placed so that the outer
// edge of the stalk at its section lies on the curve outer, with the direction of
// outer there, and the parameter of that point.
func roundTerminalAt(b ballPart, outer cubic, angle float64) (placedBall, float64) {
	tb := b.place(ballPlace{Angle: angle})
	want := math.Atan2(tb.Out.Dir.Y, tb.Out.Dir.X)
	head := func(t float64) float64 { q := outer.tangent(t); return math.Atan2(q.Y, q.X) }
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
	q := outer.at(t)
	return b.place(ballPlace{Centre: outline.Point{X: q.X - tb.Out.At.X, Y: q.Y - tb.Out.At.Y}, Angle: angle}), t
}

// roundTerminal returns, in the frame of c, the end of a round stroke in the
// ball terminal.  The curve outer is the outer edge of the round from its right
// extreme up to its top, and inner the inner edge from its top to its right
// extreme, as the contour walks them.  The ball is placed as the part would be
// placed on c (figBallPlace): turned so that the rightmost ink of the part lies
// at right, with the centre of the ball on an integer abscissa.  Of the part the
// figures draw the circle of the ball and the tip of the wedge with the direction
// of the inner edge there; the outer edge is a spiral (windSpiral) from the top of
// the round, or from the point of the hook which how names, into the circle of
// the ball, and the inner edge runs as how says to the tip of the wedge.  The
// chain runs from the top of the inner edge to the tip, round the ball, and back
// along the outer edge to its top.
func roundTerminal(name string, b ballPart, outer, inner cubic, right float64, how figStroke) chain {
	tb := figBallPlace(name, b, outer, right)

	// The inner edge, from its top to the tip of the wedge.
	var in chain
	if how.Arch {
		in = archToTip(name, inner.P0, inner.tangent(0), tb)
	} else {
		in = stalkInner(name, inner, tb)
	}

	// The outer edge, wound into the ball.
	start := 1.0
	if how.Hook {
		start = tAtFall(outer, how.Fall)
	}
	t0 := outer.tangent(start)
	w, err := newWindSpiral(outer.at(start), math.Atan2(-t0.Y, -t0.X), outer.curvature(start), tb.Centre, b.R)
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: %v; glyphs.Check reports such parameters", name, err))
	}
	wind, err := w.cubics()
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: the spiral into the ball: %v; glyphs.Check reports such parameters", name, err))
	}

	// The ball, counter-clockwise from the tip of the wedge to the end of the
	// spiral.
	aW := math.Atan2(tb.Tip.Y-tb.Centre.Y, tb.Tip.X-tb.Centre.X)
	aA := w.heading(w.L) + math.Pi/2
	for aA <= aW {
		aA += 2 * math.Pi
	}
	for aA > aW+2*math.Pi {
		aA -= 2 * math.Pi
	}
	ball := circleArc(tb.Centre, b.R, aW, aA, 0)
	ball[len(ball)-1].P3 = wind.start()

	c := join(in, ball, wind)
	if how.Hook {
		c = join(c, chain{curveSeg(outer.sub(start, 1))})
	}
	return c
}

// A figStroke says how a figure leads its stroke into the ball (roundTerminal).
// The outer edge is a spiral (windSpiral) from the top of the round, or, if Hook
// is set, from the point where the outer edge, walked towards the ball, falls
// Fall degrees below the horizontal.  If Arch is set the inner edge runs from
// its top to the tip of the wedge in one arch (archToTip); otherwise it is the
// inner edge of the stalk of the part (stalkInner).
type figStroke struct {
	Arch, Hook bool
	Fall       float64
}

// tAtFall returns the parameter of the point where the curve c, walked from its
// end to its start, falls fall degrees below the horizontal; c is the outer
// edge of a hook as roundTerminal takes it, which turns clockwise from the
// horizontal at its end.
func tAtFall(c cubic, fall float64) float64 {
	head := func(t float64) float64 { q := c.tangent(t); return math.Atan2(-q.Y, -q.X) * 180 / math.Pi }
	lo, hi := 0.0, 1.0
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		if head(mid) < -fall {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// archToTip returns the inner edge from its top q0, where it leaves in the
// direction t0, to the tip of the wedge of tb, where it arrives in the direction
// of the inner edge of the part: one cubic curve which leaves q0 with the
// curvature of the circular arc which turns from the one direction to the other
// over the chord, and arrives at the tip with the handle of that arc
// (arcHandle).  The figure gives the inner edge on the other side of q0 the
// same curvature at q0.
func archToTip(name string, q0, t0 outline.Point, tb placedBall) chain {
	var tW outline.Point
	for _, pc := range tb.Chain {
		if isTip(pc.P3, tb.Tip) {
			tW = pc.tangent(1)
		}
	}
	turn := math.Atan2(t0.X*tW.Y-t0.Y*tW.X, t0.X*tW.X+t0.Y*tW.Y) // negative: clockwise
	chord := math.Hypot(tb.Tip.X-q0.X, tb.Tip.Y-q0.Y)
	k := 2 * math.Sin(turn/2) / chord
	c, err := g2Start(q0, t0, k, tb.Tip, tW, arcHandle(q0, t0, tb.Tip, tW))
	if err != nil || !oneWay(c, -1) {
		panic(fmt.Sprintf("glyphs: glyph %s: no arch from the top of the counter to the tip of the wedge; glyphs.Check reports such parameters", name))
	}
	return splitAtExtrema(c)
}

// stalkInner returns the inner edge from the top of inner to the tip of the
// wedge of tb: the inner edge of the round to a cut, one cubic curve from there
// to the section of the part, which leaves the round with its direction and its
// curvature, at the cut where its curvature changes least abruptly
// (curvatureRate), and the inner edge of the part from its section to the tip.
func stalkInner(name string, inner cubic, tb placedBall) chain {
	best, in, tI := math.Inf(1), cubic{}, 0.0
	for _, free := range []bool{false, true} {
		for _, u := range cCuts() {
			fi := stalkFoot{At: inner.at(u), Dir: inner.tangent(u), Kappa: inner.curvature(u)}
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
			break
		}
	}
	if math.IsInf(best, 1) {
		panic(fmt.Sprintf("glyphs: glyph %s: no inner edge of the stalk reaches the ball terminal; glyphs.Check reports such parameters", name))
	}
	c := join(chain{curveSeg(inner.sub(0, tI))}, splitAtExtrema(in))
	for _, pc := range tb.Chain {
		c = append(c, pc)
		if isTip(pc.P3, tb.Tip) {
			break
		}
	}
	return c
}

// figBallPlace returns the part placed as on c (roundTerminalAt), turned so that
// the rightmost ink of the part lies at right with the centre of the ball on an
// integer abscissa.
func figBallPlace(name string, b ballPart, outer cubic, right float64) placedBall {
	x1 := func(angle float64) float64 {
		tb, _ := roundTerminalAt(b, outer, angle)
		_, _, x1, _ := chainBBox(tb.Chain)
		return x1
	}
	solve := func(target float64) float64 {
		lo, hi := figBallAngles[0], figBallAngles[1]
		if !(x1(lo) >= target && x1(hi) <= target) {
			panic(fmt.Sprintf("glyphs: glyph %s: the ball cannot be placed with its ink at %.1f (%.1f to %.1f); glyphs.Check reports such parameters", name, target, x1(hi), x1(lo)))
		}
		for i := 0; i < 60; i++ {
			mid := (lo + hi) / 2
			if x1(mid) > target {
				lo = mid
			} else {
				hi = mid
			}
		}
		return (lo + hi) / 2
	}
	tb, _ := roundTerminalAt(b, outer, solve(right))
	target := right + math.Round(tb.Centre.X) - tb.Centre.X
	tb, _ = roundTerminalAt(b, outer, solve(target))
	return tb
}

// mirrorCubic returns the curve reflected in the vertical line x = m and walked
// the other way, as mirrorX does for a chain.
func mirrorCubic(c cubic, m float64) cubic {
	f := func(q outline.Point) outline.Point { return outline.Point{X: 2*m - q.X, Y: q.Y} }
	return cubic{f(c.P3), f(c.P2), f(c.P1), f(c.P0)}
}

// turnCubic returns the curve turned through half a turn about the origin.
func turnCubic(c cubic) cubic {
	f := func(q outline.Point) outline.Point { return outline.Point{X: -q.X, Y: -q.Y} }
	return cubic{f(c.P0), f(c.P1), f(c.P2), f(c.P3)}
}

// upperLeftBall returns the ball terminal at the upper left of a round: outer is
// the upper left quarter of the outer edge (from the top to the left extreme,
// counter-clockwise) and inner that of the inner edge (from the left extreme to
// the top, clockwise), and left is the leftmost ink of the ball.  The chain runs
// from the top of the outer edge round the ball to the top of the inner edge.
func upperLeftBall(name string, b ballPart, outer, inner cubic, left float64, how figStroke) chain {
	return roundTerminal(name, b, mirrorCubic(outer, 0), mirrorCubic(inner, 0), -left, how).mirrorX(0)
}

// lowerLeftBall returns the ball terminal at the lower left of a round: outer is
// the lower left quarter of the outer edge (from the left extreme to the bottom,
// counter-clockwise) and inner that of the inner edge (from the bottom to the left
// extreme, clockwise), and left is the leftmost ink of the ball.  The chain runs
// from the bottom of the inner edge round the ball to the bottom of the outer edge.
func lowerLeftBall(name string, b ballPart, outer, inner cubic, left float64, how figStroke) chain {
	return roundTerminal(name, b, turnCubic(outer), turnCubic(inner), -left, how).halfTurn(0, 0)
}

// curveChain returns the curve as a chain, cut at its extremes.
func curveChain(c cubic) chain { return splitAtExtrema(c) }

// figCurve returns the curve from p0, leaving in the direction t0 with the
// curvature k0, to p3, arriving in the direction t3 with the curvature k3, for the
// glyph name.  Where no such curve exists the curvature at p0 is left free.
func figCurve(name string, p0, t0 outline.Point, k0 float64, p3, t3 outline.Point, k3 float64) cubic {
	if c, err := g2Cubic(p0, t0, k0, p3, t3, k3); err == nil && oneWay(c, math.Copysign(1, k3+k0)) {
		return c
	}
	c, err := fairEnd(p0, t0, p3, t3, k3)
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: %v", name, err))
	}
	return c
}

// crossAt returns the parameters at which the curves a and b cross, for the glyph
// name; it panics if they do not.
func crossAt(name string, a, b cubic) (float64, float64) {
	ta, tb, ok := crossing(a, b)
	if !ok {
		panic(fmt.Sprintf("glyphs: glyph %s: two edges do not meet", name))
	}
	return ta, tb
}

// ellipseChain returns the closed super-ellipse in the box as a chain, starting
// at its right extreme, counter-clockwise if ccw is set and clockwise otherwise.
func ellipseChain(x0, y0, x1, y1, superness float64, ccw bool) chain {
	q := newSuperShape(x0, y0, x1, y1, superness)
	var c chain
	for i := 0; i < 4; i++ {
		c = append(c, curveSeg(q.cubic(i, ccw)))
	}
	return c
}

// pathOfChains returns the closed chains as the contours of one path.
func pathOfChains(cs ...chain) outline.Path {
	var path outline.Path
	for _, c := range cs {
		c.addContour(&path)
	}
	return path
}

// mergedPath returns the closed chains as the contours of one path, each with its
// collinear straight pieces merged.
func mergedPath(cs ...chain) outline.Path {
	var path outline.Path
	for _, c := range cs {
		mergeLines(c).addContour(&path)
	}
	return path
}

// degrees returns the angle a given in degrees in radians.
func degrees(a float64) float64 { return a * math.Pi / 180 }
