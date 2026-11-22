// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"sort"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The asterisk: six arms at sixty degrees, one pointing up, each a tapered
// stroke which ends in a round cap, a circle of AsteriskOuter across whose
// outermost point lies at the distance AsteriskSize/2 from the centre.  Halfway
// from the centre to the centre of the cap the arm has the width of a stroke
// which widens evenly from AsteriskInner at the centre to AsteriskOuter at the
// end.  The diameter is AsteriskSize and the top
// AsteriskTop.  The arms are one contour, whose crotches are where the straight
// edges of neighbouring arms meet.  Each straight edge runs into the cap on a
// spiral whose curvature rises linearly with the length run from zero to that
// of the cap, over the turn asteriskSpiralTurn; the edge is the line through
// its point halfway out on which that spiral ends on the circle of the cap.

// asteriskSpiralTurn is the angle (radians) by which the edge of an arm turns
// on the spiral into its cap.
const asteriskSpiralTurn = 30 * math.Pi / 180

// asteriskSpiralPower is the power of the length run with which the curvature
// of that spiral rises.
const asteriskSpiralPower = 2.0

// asteriskCap is the edge of one arm from the end of the straight on its right
// edge round the cap to the end of the straight on its left edge, in the frame
// of the arm: the arm points along the positive x-axis from the origin, the
// centre of the cap is (Reach, 0), and the length s runs along the edge.
type asteriskCap struct {
	R, Reach float64
	// Alpha is the direction of the straight of the right edge, L the length
	// of each spiral and Arc that of the circle between them.
	Alpha, L, Arc float64
	P0            outline.Point
}

// newAsteriskCap returns the cap of the radius r about (reach, 0) for an arm
// whose right edge passes through (xp, -yp).
func newAsteriskCap(r, reach, xp, yp float64) asteriskCap {
	c := asteriskCap{R: r, Reach: reach, L: (asteriskSpiralPower + 1) * r * asteriskSpiralTurn}
	// The spiral from the origin in the direction 0 ends with its centre of
	// curvature at (xc, yc).
	e := c.spiralPoint(outline.Point{}, 0, c.L)
	xc, yc := e.X-r*math.Sin(asteriskSpiralTurn), e.Y+r*math.Cos(asteriskSpiralTurn)
	// The straight through (xp, -yp) passes the centre of the cap at the
	// distance yc on its left.
	a, b := yp, reach-xp
	c.Alpha = math.Acos(yc/math.Hypot(a, b)) - math.Atan2(b, a)
	u := dirOf(c.Alpha)
	t := b*u.X + a*u.Y - xc
	c.P0 = outline.Point{X: xp + t*u.X, Y: -yp + t*u.Y}
	c.Arc = r * (math.Pi - 2*c.Alpha - 2*asteriskSpiralTurn)
	return c
}

// spiralPoint returns the point after the length s on the spiral from p0 in the
// direction phi0 whose curvature rises linearly to 1/R at the length L, by
// Simpson's rule.
func (c asteriskCap) spiralPoint(p0 outline.Point, phi0, s float64) outline.Point {
	const n = 400
	h := s / n
	q := p0
	for i := 0; i <= n; i++ {
		w := 2.0
		if i == 0 || i == n {
			w = 1
		} else if i%2 == 1 {
			w = 4
		}
		t := float64(i) * h
		phi := phi0 + c.turn(t)
		q.X += w * h / 3 * math.Cos(phi)
		q.Y += w * h / 3 * math.Sin(phi)
	}
	return q
}

// turn returns the angle by which the spiral has turned after the length s.
func (c asteriskCap) turn(s float64) float64 {
	return c.L / (c.R * (asteriskSpiralPower + 1)) * math.Pow(s/c.L, asteriskSpiralPower+1)
}

// length returns the length of the edge.
func (c asteriskCap) length() float64 { return 2*c.L + c.Arc }

// at returns the point, the direction and the curvature of the edge after the
// length s.
func (c asteriskCap) at(s float64) (outline.Point, float64, float64) {
	total := c.length()
	switch {
	case s <= c.L:
		return c.spiralPoint(c.P0, c.Alpha, s), c.Alpha + c.turn(s), math.Pow(s/c.L, asteriskSpiralPower) / c.R
	case s >= total-c.L:
		q, h, k := c.at(total - s)
		return outline.Point{X: q.X, Y: -q.Y}, math.Pi - h, k
	default:
		h := c.Alpha + asteriskSpiralTurn + (s-c.L)/c.R
		return outline.Point{X: c.Reach + c.R*math.Sin(h), Y: -c.R * math.Cos(h)}, h, 1 / c.R
	}
}

// heading returns the direction of the edge after the length s.
func (c asteriskCap) heading(s float64) float64 {
	_, h, _ := c.at(s)
	return h
}

// cuts returns the lengths at which the edge, turned through the angle th, has
// its nodes: its ends, its extremes, where the direction is a multiple of a
// quarter turn, and the ends of the spirals, except those within windMinPiece
// of an extreme.
func (c asteriskCap) cuts(th float64) []float64 {
	total := c.length()
	var ext []float64
	h0, h1 := th+c.Alpha, th+math.Pi-c.Alpha
	for q := math.Ceil(h0/(math.Pi/2)) * math.Pi / 2; q < h1; q += math.Pi / 2 {
		lo, hi := 0.0, total
		for i := 0; i < 60; i++ {
			mid := (lo + hi) / 2
			if th+c.heading(mid) < q {
				lo = mid
			} else {
				hi = mid
			}
		}
		if s := (lo + hi) / 2; s > 1 && total-s > 1 {
			ext = append(ext, s)
		}
	}
	cuts := append([]float64{0, total}, ext...)
	for _, s := range []float64{c.L, total - c.L} {
		near := false
		for _, e := range ext {
			if math.Abs(e-s) < windMinPiece {
				near = true
			}
		}
		if !near {
			cuts = append(cuts, s)
		}
	}
	sort.Float64s(cuts)
	return cuts
}

// asteriskPath returns the asterisk with its centre on the abscissa 0.
func asteriskPath(p design.Params) outline.Path {
	q := p.Punct
	r := q.AsteriskSize / 2
	cy := q.AsteriskTop - r
	capR := q.AsteriskOuter / 2
	reach := r - capR
	// The edges pass through the points of the arm, halfway to the cap, at which
	// a stroke which widens evenly from AsteriskInner at the centre to
	// AsteriskOuter at the distance r has its edges.
	w := q.AsteriskInner + (q.AsteriskOuter-q.AsteriskInner)*reach/2/r
	cp := newAsteriskCap(capR, reach, reach/2, w/2)
	const arms = 6
	type arm struct {
		th   float64
		u, n outline.Point
	}
	as := make([]arm, arms)
	for k := range as {
		th := math.Pi/2 + float64(k)*2*math.Pi/arms
		as[k] = arm{th, dirOf(th), outline.Point{X: -math.Sin(th), Y: math.Cos(th)}}
	}
	toGlobal := func(a arm, v outline.Point) outline.Point {
		return outline.Point{X: v.X*a.u.X + v.Y*a.n.X, Y: cy + v.X*a.u.Y + v.Y*a.n.Y}
	}
	// The straights of an arm, as their outer end and their direction outward.
	right := func(a arm) (outline.Point, outline.Point) {
		return toGlobal(a, cp.P0), dirOf(a.th + cp.Alpha)
	}
	left := func(a arm) (outline.Point, outline.Point) {
		return toGlobal(a, outline.Point{X: cp.P0.X, Y: -cp.P0.Y}), dirOf(a.th - cp.Alpha)
	}
	crotches := make([]outline.Point, arms)
	for k, a := range as {
		lp, ld := left(as[(k+arms-1)%arms])
		rp, rd := right(a)
		crotches[k] = meetLines(lp, ld, rp, rd)
	}
	var ch chain
	for k, a := range as {
		cuts := cp.cuts(a.th)
		n := len(cuts) - 1
		for i := 1; i <= n; i++ {
			pa, ha, ka := cp.at(cuts[i-1])
			pb, hb, kb := cp.at(cuts[i])
			ga, gb := toGlobal(a, pa), toGlobal(a, pb)
			ta, tb := dirOf(a.th+ha), dirOf(a.th+hb)
			var c cubic
			var err error
			if kb != 0 {
				c, err = g2Cubic(ga, ta, ka, gb, tb, kb)
			} else {
				// g2Cubic needs a curvature at the end; this curve is drawn the
				// other way.
				c, err = g2Cubic(gb, neg(tb), -kb, ga, neg(ta), -ka)
				c = cubic{c.P3, c.P2, c.P1, c.P0}
			}
			// The first and the last curve start and end at the crotches, with
			// the handle there lengthened by the straight they replace.
			if err == nil && i == 1 {
				h := math.Hypot(c.P1.X-ga.X, c.P1.Y-ga.Y)
				cr := crotches[k]
				c, err = g2End(cr, ta, h+math.Hypot(ga.X-cr.X, ga.Y-cr.Y), gb, tb, kb)
			}
			if err == nil && i == n {
				h := math.Hypot(c.P2.X-gb.X, c.P2.Y-gb.Y)
				cr := crotches[(k+1)%arms]
				c, err = g2Start(c.P0, ta, ka, cr, tb, h+math.Hypot(gb.X-cr.X, gb.Y-cr.Y))
			}
			if err != nil {
				panic(fmt.Sprintf("glyphs: the asterisk: %v", err))
			}
			ch = append(ch, curveSeg(c))
		}
	}
	var path outline.Path
	ch.addContour(&path)
	return path
}

// asterisk draws the asterisk.
func asterisk(p design.Params) (outline.Path, float64) {
	return placeInk(asteriskPath(p), p.Punct.Asterisk)
}
