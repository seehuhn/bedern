// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// figEight draws the figure 8: two closed bowls, the upper one smaller, one above
// the other on one vertical, with a heavy spine between the counters which
// falls from the upper left to the lower right.
//
// The outer contour is the outline of two super-ellipses with the squareness of
// the outer edge of o, which cross on each side of the waist in a notch.  The
// lower one is the full width of the ink, from the round bottom to
// EightWaistThin above EightWaist; the upper one is EightUpperWidth wide, from
// EightWaistThin below EightWaist to the round top.
//
// The spine lies between two parallel lines EightGap apart, about the point at
// the height EightWaist in the middle of the ink, which fall to the right at
// EightSpine degrees from the horizontal.  The upper counter (spineCounter) has
// its top Figures.Thin below the round top and its sides EightUL and EightUR
// inside the outer edges of the upper bowl, and its lower edge leaves the upper
// line of the spine at a corner on the right, where the thin arc of the bowl
// meets the spine.  The lower counter has its bottom Figures.Thin above the
// round bottom and its sides EightLL and EightLR inside those of the lower bowl,
// and is the upper one's construction turned through half a turn about that
// point.  So the heavy stroke runs from the upper left of the upper bowl along
// the spine to the lower right of the lower bowl, and the upper right and the
// lower left of the bowls are thin arcs which close the counters and run into
// it.
//
// The outer contour runs from the bottom round the right side of the lower bowl
// to the right notch, round the upper bowl to the left notch and round the left
// side of the lower bowl back to the bottom; the counters are two contours,
// clockwise.
func figEight(p design.Params) (outline.Path, float64) {
	f := p.Figures
	x0 := design.FigureLeft(p, 8)
	x1 := x0 + f.Ink[8]
	top, bot := figTop(p), figBottom(p)
	ul := x0 + (f.Ink[8]-f.EightUpperWidth)/2
	ur := ul + f.EightUpperWidth
	th := degrees(f.EightSpine)
	mid := outline.Point{X: (x0 + x1) / 2, Y: f.EightWaist}
	n := outline.Point{X: math.Sin(th), Y: math.Cos(th)} // across the spine, up
	h := f.EightGap / 2
	s := outline.Point{X: mid.X + h*n.X, Y: mid.Y + h*n.Y}
	turn := func(q outline.Point) outline.Point { return outline.Point{X: 2*mid.X - q.X, Y: 2*mid.Y - q.Y} }
	turnC := func(c cubic) cubic { return cubic{turn(c.P0), turn(c.P1), turn(c.P2), turn(c.P3)} }

	lo := newSuperShape(x0, bot, x1, f.EightWaist+f.EightWaistThin, p.SupernessOuter)
	uo := newSuperShape(ul, f.EightWaist-f.EightWaistThin, ur, top, p.SupernessOuter)
	upper := spineCounter(ul+f.EightUL, ur-f.EightUR, top-f.Thin, s, th, uo.cubic(3, true), f.EightJoin)
	// The lower counter is drawn turned through half a turn about the middle.
	lower := spineCounter(2*mid.X-x1+f.EightLR, 2*mid.X-x0-f.EightLL, 2*mid.Y-bot-f.Thin, s, th,
		turnC(lo.cubic(1, true)), f.EightJoin).mapPoints(turn)

	// The notches, where the outer edges of the ovals cross.
	loR, uoR := lo.cubic(0, true), uo.cubic(3, true) // right to top, bottom to right
	loL, uoL := lo.cubic(1, true), uo.cubic(2, true) // top to left, left to bottom
	ra, rb := crossAt("eight", loR, uoR)
	la, lb := crossAt("eight", uoL, loL)
	toR, fromR := loR.sub(0, ra), uoR.sub(rb, 1)
	fromR.P0 = toR.P3
	toL, fromL := uoL.sub(0, la), loL.sub(lb, 1)
	fromL.P0 = toL.P3
	outer := join(
		chain{curveSeg(lo.cubic(3, true))},
		curveChain(toR), curveChain(fromR),
		chain{curveSeg(uo.cubic(0, true)), curveSeg(uo.cubic(1, true))},
		curveChain(toL), curveChain(fromL),
		chain{curveSeg(lo.cubic(2, true))},
	)
	return pathOfChains(outer, upper, lower), f.Advance
}

// spineCounter returns the clockwise upper counter of 8.  The line of the
// spine passes through s and falls to the right at the angle th; the corner c
// of the counter is the point of the line, right of s, which lies join inside
// out, the outer edge of the bowl from its bottom to its right extreme
// (eightCorner).  The inner edge of the thin arc is out offset inwards by a
// thickness which falls in a smooth step from that of the side at the right extreme of
// out to join at eightThinBy of the way to c and keeps that.  The counter runs
// from c along the line, without curvature there, to the left extreme at xl,
// at the middle height between s and the top, over the top at yt above the
// middle of xl and xr, to the inner edge of the arc at the height of the right
// extreme of out, and down to c, which it reaches with the direction and the
// curvature of the inner edge of the arc.  Between these stations it is one
// cubic curve each, with the mean curvature of the circles through the
// neighbouring stations at the stations which do not give one (strokeEdge, on
// stations of no width).
func spineCounter(xl, xr, yt float64, s outline.Point, th float64, out cubic, join float64) chain {
	along := outline.Point{X: -math.Cos(th), Y: math.Sin(th)} // up the spine
	c, tc := eightCorner(s, th, out, join)
	cy := (s.Y + yt) / 2
	deg := func(v outline.Point) float64 { return math.Atan2(v.Y, v.X) * 180 / math.Pi }
	// The inner edge of the thin arc, from the right extreme to c: the outer
	// edge offset inwards by a thickness which falls from the side's at the
	// extreme to join at eightThinBy of the way to c and keeps that.
	side := out.P3.X - xr
	inner := func(u float64) outline.Point {
		t := 1 - u*(1-tc)
		v := math.Min(1, u/eightThinBy)
		w := side - (side-join)*v*v*(3-2*v)
		q, g := out.at(t), out.tangent(t)
		l := math.Hypot(g.X, g.Y)
		return outline.Point{X: q.X - g.Y/l*w, Y: q.Y + g.X/l*w}
	}
	const e = 1e-3
	r0, r1 := inner(0), inner(e)
	a, b, z := inner(1-2*e), inner(1-e), inner(1)
	// The curvature of the offset curve at c, from three of its points.
	d1, d2 := outline.Point{X: b.X - a.X, Y: b.Y - a.Y}, outline.Point{X: z.X - b.X, Y: z.Y - b.Y}
	kc := 2 * (d1.X*d2.Y - d1.Y*d2.X) / (dist(a, b) * dist(b, z) * dist(a, z))
	return strokeEdge([]station{
		{X: c.X, Y: c.Y, Dir: deg(along), K: -1e-9}, // no curvature (a K of 0 is not given)
		{X: xl, Y: cy, Dir: 90},
		{X: (xl + xr) / 2, Y: yt, Dir: 0},
		{X: r0.X, Y: r0.Y, Dir: deg(unit(r0, r1))},
		{X: c.X, Y: c.Y, Dir: deg(unit(a, c)), K: kc},
	}, 1)
}

// eightThinBy is the share of the way from the right extreme to the corner at
// which the thin arc of 8 reaches its least thickness.
const eightThinBy = 0.8

// eightCorner returns the corner c of the upper counter of 8 (spineCounter)
// and the parameter of the point of out nearest to it.
func eightCorner(s outline.Point, th float64, out cubic, join float64) (outline.Point, float64) {
	along := outline.Point{X: -math.Cos(th), Y: math.Sin(th)}
	// The distance of a point from the outer edge and the parameter of its
	// nearest point.
	near := func(q outline.Point) (float64, float64) {
		best, bt := math.Inf(1), 0.0
		for i := 0; i <= 400; i++ {
			t := float64(i) / 400
			if d := dist(out.at(t), q); d < best {
				best, bt = d, t
			}
		}
		for step := 1.0 / 800; step > 1e-9; step /= 2 {
			for _, t := range []float64{bt - step, bt + step} {
				if t >= 0 && t <= 1 {
					if d := dist(out.at(t), q); d < best {
						best, bt = d, t
					}
				}
			}
		}
		return best, bt
	}
	// c by bisection along the line to the right of s.
	pos := func(u float64) outline.Point { return outline.Point{X: s.X - u*along.X, Y: s.Y - u*along.Y} }
	lo, hi := 0.0, 0.0
	for d, _ := near(pos(hi)); d > join; d, _ = near(pos(hi)) {
		lo, hi = hi, hi+10
		if hi > 1000 {
			panic("glyphs: glyph eight: the spine does not reach the thin arc; glyphs.Check reports such parameters")
		}
	}
	for k := 0; k < 60; k++ {
		if d, _ := near(pos((lo + hi) / 2)); d > join {
			lo = (lo + hi) / 2
		} else {
			hi = (lo + hi) / 2
		}
	}
	c := pos((lo + hi) / 2)
	_, tc := near(c)
	return c, tc
}

// figG2 returns the cubic curve from p0, leaving in the direction t0 with the
// curvature k0, to p3, arriving in the direction t3 with the curvature k3
// (g2Cubic), drawn from p3 if the curve ends without curvature.  It panics
// with the name of the glyph if there is no such curve.
func figG2(name string, p0, t0 outline.Point, k0 float64, p3, t3 outline.Point, k3 float64) cubic {
	if k3 == 0 {
		c, err := g2Cubic(p3, neg(t3), 0, p0, neg(t0), -k0)
		if err != nil {
			panic("glyphs: glyph " + name + ": " + err.Error() + "; glyphs.Check reports such parameters")
		}
		return c.reversed()
	}
	c, err := g2Cubic(p0, t0, k0, p3, t3, k3)
	if err != nil {
		panic("glyphs: glyph " + name + ": " + err.Error() + "; glyphs.Check reports such parameters")
	}
	return c
}
