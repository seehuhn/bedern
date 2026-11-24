// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"fmt"
	"math"
)

// The placement of a ball by the rule of the spiral (BallSpec.Power).
//
// The outer edge of the curl leaves the end node of the stroke with the
// direction and the curvature k0 which the stroke gives it there and winds
// into the circle of the ball on a spiral whose curvature rises as the power
// M of the length from k0 to 1/R (the spiral of windSpiral in the roman).  A
// spiral of rising curvature has nested circles of curvature, so it exists
// only where the ball lies inside the circle of curvature of the edge at its
// start.  For a given M its shape depends only on k0 R and on the depth h of
// the centre of the ball below the tangent of the outer edge at the end
// node, on the side of the curl: the length follows from h, and with it how
// far ahead of the end node the centre lies; a higher power puts it further
// ahead.  The rule:
//
//   - The power is solved so that the spiral from the end node runs into the
//     circle, as the roman's windSpiral does, within the range of Power.
//   - Where the power would leave its range, it stays at the nearest bound
//     and the end node slides along its own direction until the spiral of
//     that power runs into the circle.
//   - Where h lies outside the range of the rule, from ballMinDepth radii to
//     the depth of the spiral which turns by ballMaxTurn, the end node first
//     turns (the lead-in): away from the ball where h is too small, towards
//     it where h is too large, until h lies on the nearest bound.  The
//     stroke before the end node then changes its direction on its way to
//     the end, and the spiral starts from the turned node.
//
// The values of the end node which the rules of the stroke supply are taken
// at the place where the letter puts it and kept while it slides and turns,
// so that the outer edge there follows from the node alone.  The centre of
// the ball lies on whole units of the final frame, so that the circle
// survives the rounding of the outline: its extremes are whole points and the
// handles of its arcs move by at most half a unit.

// ballMinDepth is the least depth of the centre of the ball below the tangent
// of the outer edge at the end node, in radii of the ball: below it the
// spiral turns by less than about 40 degrees and the wedge has no room.
const ballMinDepth = 1.15

// ballMaxTurn is the greatest turn of the spiral, in degrees: beyond it the
// curl runs round the ball and back over the stroke.
const ballMaxTurn = 240.0

// ballMinRun is the least length of the last piece of the centre line, from
// the node before the end node to the end node, after the slide.
const ballMinRun = 20.0

// powerSpiral returns the length L of the spiral which starts in the
// direction 0 and turns to the left with the curvature k0 + (1/r - k0)
// (s/L)^m, whose last centre of curvature lies h above the line of its
// start, and how far ahead of the start that centre lies.  Of the lengths
// for which the spiral turns by at most ballMaxTurn the first is taken; ok
// is false where none reaches h, and hmax is then the greatest depth there.
func powerSpiral(k0, r, m, h float64) (L, ahead, hmax float64, ok bool) {
	w := spiral{Sig: 1, K0: k0, K1: 1 / r, M: m}
	at := func(l float64) Point {
		v := w
		v.L = l
		return v.centreAtN(l, 40)
	}
	turn := func(l float64) float64 {
		return (k0*l + (1/r-k0)*l/(m+1)) * 180 / math.Pi
	}
	step := 0.05 * r
	prevL, prevH := 0.0, r
	for l := step; turn(l) <= ballMaxTurn; l += step {
		c := at(l)
		if c.Y < prevH {
			break // the depth falls again: the spiral runs back
		}
		if c.Y >= h {
			lo, hi := prevL, l
			for i := 0; i < 50; i++ {
				mid := (lo + hi) / 2
				if at(mid).Y < h {
					lo = mid
				} else {
					hi = mid
				}
			}
			L = (lo + hi) / 2
			return L, at(L).X, 0, true
		}
		prevL, prevH = l, c.Y
	}
	return 0, 0, prevH, false
}

// A ballFit is the spiral which the rule chooses for an end node and a ball:
// its power and length, the depth and the distance ahead of the centre of the
// ball, the turn of the end node which the lead-in asks for (degrees, towards
// the ball positive) and its slide along its direction.
type ballFit struct {
	M, L, H, A  float64
	Turn, Slide float64
}

// fitBall applies the rule to the outer edge outer of an end node whose ball,
// of the radius r about centre, lies on the side sigma of the stroke (+1
// left), with the powers between pmin and pmax.
func fitBall(outer edgeData, sigma float64, centre Point, r, pmin, pmax float64) (ballFit, error) {
	d := sub(centre, outer.P)
	h := sigma * cross(outer.T, d)
	a := dot(outer.T, d)
	k0 := math.Max(sigma*outer.K, 0)
	f := ballFit{H: h, A: a}
	// The lead-in: the turn which brings h to target.
	turnTo := func(target float64) (ballFit, error) {
		rho := math.Hypot(h, a)
		if target >= rho {
			return f, fmt.Errorf("stroke: the ball about %.0f,%.0f lies too near the end of the stroke for a spiral", centre.X, centre.Y)
		}
		beta := math.Atan2(h, a)
		beta1 := math.Asin(target / rho)
		if beta > math.Pi/2 {
			beta1 = math.Pi - beta1
		}
		f.Turn = (beta - beta1) * 180 / math.Pi
		return f, nil
	}
	if h < ballMinDepth*r {
		return turnTo(ballMinDepth * r)
	}
	type sol struct{ m, l, a float64 }
	at := func(m float64) (sol, float64, bool) {
		l, ah, hmax, ok := powerSpiral(k0, r, m, h)
		return sol{m, l, ah}, hmax, ok
	}
	lo, hmaxLo, okLo := at(pmin)
	hi, hmaxHi, okHi := at(pmax)
	switch {
	case !okLo && !okHi:
		return turnTo(math.Max(hmaxLo, hmaxHi) - 0.01*r)
	case !okHi:
		// The highest power which reaches h.
		good, bad := pmin, pmax
		for i := 0; i < 40; i++ {
			mid := (good + bad) / 2
			if _, _, ok := at(mid); ok {
				good = mid
			} else {
				bad = mid
			}
		}
		hi, _, _ = at(good)
	case !okLo:
		good, bad := pmax, pmin
		for i := 0; i < 40; i++ {
			mid := (good + bad) / 2
			if _, _, ok := at(mid); ok {
				good = mid
			} else {
				bad = mid
			}
		}
		lo, _, _ = at(good)
	}
	var s sol
	switch {
	case a <= lo.a:
		s = lo
	case a >= hi.a:
		s = hi
	default:
		// The power for which the centre lies a ahead, by bisection: the
		// distance ahead grows with the power.
		for i := 0; i < 50; i++ {
			mid, _, ok := at((lo.m + hi.m) / 2)
			if !ok || mid.a > a {
				hi = mid
			} else {
				lo = mid
			}
		}
		s = sol{(lo.m + hi.m) / 2, (lo.l + hi.l) / 2, a}
	}
	f.M, f.L, f.Slide = s.m, s.l, a-s.a
	return f, nil
}

// placeBalls returns the stroke with its ends of the kind Ball placed by the
// rule of the spiral where they ask for it (Power).
func (s Stroke) placeBalls() (Stroke, error) {
	if s.Closed {
		return s, nil
	}
	var err error
	if s.End.Kind == Ball && s.End.Ball.Power[1] > 0 {
		if s, err = s.placeEndBall(); err != nil {
			return s, err
		}
	}
	if s.Start.Kind == Ball && s.Start.Ball.Power[1] > 0 {
		rs, err := s.reversedStroke().placeEndBall()
		if err != nil {
			return s, err
		}
		s = rs.reversedStroke()
	}
	return s, nil
}

// reversedStroke returns the stroke walked from its end to its start.
func (s Stroke) reversedStroke() Stroke {
	out := s
	n := len(s.Nodes)
	out.Nodes = make([]Node, n)
	for i, nd := range s.Nodes {
		r := nd
		r.Dir = nd.Dir + 180
		r.K, r.DW = -nd.K, -nd.DW
		if nd.Edge != nil {
			r.Edge = &EdgeK{Right: -nd.Edge.Left, Left: -nd.Edge.Right}
		}
		out.Nodes[n-1-i] = r
	}
	out.Start, out.End = s.End, s.Start
	return out
}

// ballSide returns the side of the stroke at the end node nd on which the
// ball about centre lies (+1 left) and the outer edge there.
func ballSide(nd Node, centre Point) (float64, edgeData, error) {
	sigma := 1.0
	if cross(dirOf(nd.Dir), sub(centre, nd.P)) < 0 {
		sigma = -1
	}
	outer, err := edgeAt(nd, -sigma)
	return sigma, outer, err
}

// placeEndBall places the ball at the end of the stroke by the rule of the
// spiral: the centre on whole units (by Depth, or with InkSide so that the
// ink of the terminal reaches InkX), the end node turned where the depth
// lies outside the range and slid where the power would.
func (s Stroke) placeEndBall() (Stroke, error) {
	b := s.End.Ball
	s.Nodes = append([]Node(nil), s.Nodes...)
	n := len(s.Nodes)
	if n < 2 {
		return s, fmt.Errorf("stroke: a stroke needs two nodes")
	}
	if nd := s.Nodes[n-1]; isAuto(nd.K) || isAuto(nd.DW) || isAuto(nd.W) || nd.Edge != nil {
		ns, err := s.resolved()
		if err != nil {
			return s, err
		}
		s.Nodes[n-1] = ns[n-1]
		s.Nodes[n-1].Edge = nil
	}
	centre := b.Centre
	if b.Depth != 0 {
		// The ball by its depth: the centre lies where the spiral of the
		// middle power from the end node puts it.
		nd := s.Nodes[n-1]
		sigma := math.Copysign(1, -b.Depth)
		outer, err := edgeAt(nd, -sigma)
		if err != nil {
			return s, err
		}
		h := math.Abs(b.Depth)
		m := (b.Power[0] + b.Power[1]) / 2
		_, ahead, _, ok := powerSpiral(math.Max(sigma*outer.K, 0), b.R, m, h)
		if !ok {
			return s, fmt.Errorf("stroke: no spiral of the power %.2f reaches the depth %.1f below the end at %.0f,%.0f", m, h, nd.P.X, nd.P.Y)
		}
		centre = add(add(outer.P, mul(outer.T, ahead)), mul(rot90(outer.T), sigma*h))
	}
	centre = Point{X: math.Round(centre.X), Y: math.Round(centre.Y)}
	for pass := 0; pass < 8; pass++ {
		if err := s.fitEnd(centre, b); err != nil {
			return s, err
		}
		if b.InkSide == 0 {
			break
		}
		// The ink of the terminal as it is now, and the move of the centre
		// which brings it to InkX.
		ns, sides, err := s.edges()
		if err != nil {
			return s, err
		}
		var r Result
		bb := b
		bb.Centre = centre
		cp, err := r.ball(bb, ns[n-1], sides[0][n-1], sides[1][n-1])
		if err != nil {
			return s, err
		}
		x0, x1 := chainXRange(cp)
		x := x0
		if b.InkSide > 0 {
			x = x1
		}
		dx := math.Round(b.InkX - x)
		if dx == 0 {
			break
		}
		centre.X += dx
	}
	s.End.Ball.Centre = centre
	s.End.Ball.InkSide = 0
	s.End.Ball.Depth = 0
	return s, nil
}

// fitEnd turns (the lead-in) and slides the end node so that a spiral of a
// power in the range of b from the outer edge at the end node runs into the
// circle about centre.
func (s Stroke) fitEnd(centre Point, b BallSpec) error {
	n := len(s.Nodes)
	for iter := 0; iter < 60; iter++ {
		nd := &s.Nodes[n-1]
		sigma, outer, err := ballSide(*nd, centre)
		if err != nil {
			return err
		}
		f, err := fitBall(outer, sigma, centre, b.R, b.Power[0], b.Power[1])
		if err != nil {
			return err
		}
		if f.Turn != 0 {
			nd.Dir += sigma * f.Turn
			continue
		}
		if math.Abs(f.Slide) < 1e-4 {
			prev := s.Nodes[n-2]
			if dot(sub(nd.P, prev.P), dirOf(nd.Dir)) < ballMinRun {
				return fmt.Errorf("stroke: the ball about %.0f,%.0f lies too near the node at %.0f,%.0f: the end node would slide behind it", centre.X, centre.Y, prev.P.X, prev.P.Y)
			}
			return nil
		}
		nd.P = add(nd.P, mul(dirOf(nd.Dir), f.Slide))
	}
	return fmt.Errorf("stroke: the end node before the ball about %.0f,%.0f finds no place", centre.X, centre.Y)
}
