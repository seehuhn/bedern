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
// wedge as the stalk of the part (the upper terminal of 3), as a spiral whose
// curvature rises from the bottom of the counter (the lower terminals of 3 and
// 5; stalkRising) or as one round arch from the top of the counter (2, 6;
// roundTerminal).  The construction is written once, in
// the frame of c, where the ball hangs at the upper right of a round whose
// stroke comes over the top from the left; a figure whose terminal stands
// elsewhere takes the two quarters of its round into that frame (mirrored in a
// vertical line for a ball at the upper left, turned through half a turn for
// one at the lower left), and takes the result back.

// figBall returns the ball terminal of the figures: the part of the lowercase,
// Figures.Ball across on the stalk figStalk.
func figBall(p design.Params) ballPart { return must(figBallPart(p)) }

// figBallPart is figBall which returns an error where the stalk does not fit the
// ball.
func figBallPart(p design.Params) (ballPart, error) {
	b, err := makeBallPart(p.Figures.Ball/2, figStalk(p), ballCurl, ballSection, ballBlend)
	if err != nil {
		return ballPart{}, fmt.Errorf("figures: %v", err)
	}
	return b, nil
}

// figStalk returns the thickness of the stalk of the ball terminal of the
// figures: the stalk of the lowercase scaled with the ball.
func figStalk(p design.Params) float64 { return p.BallStalk * p.Figures.Ball / p.BallDiameter }

// curlBall returns the ball terminal of the curled end of 6 and 9: the
// part of the lowercase, Figures.Ball across on the stalk Figures.CurlStalk.
func curlBall(p design.Params) ballPart { return must(curlBallPart(p)) }

// curlBallPart is curlBall which returns an error where the stalk does not fit the
// ball.
func curlBallPart(p design.Params) (ballPart, error) {
	b, err := makeBallPart(p.Figures.Ball/2, p.Figures.CurlStalk, ballCurl, ballSection, ballBlend)
	if err != nil {
		return ballPart{}, fmt.Errorf("figures: %v", err)
	}
	return b, nil
}

// curlTerminal returns, in the frame of c, the curled end of a round stroke in
// the ball terminal of 6 and 9.  The curves outer and inner are as for
// roundTerminal.  The ball has its centre at centre.  The outer edge is the
// spiral (windSpiral) from the top of the round, or from the point of the hook
// which how names, into the circle of the ball, and the part is turned so that
// the tip of its wedge lies Figures.CurlNeck from the outer edge (curlTip).  The
// inner edge runs as how says to the tip.  The chain runs as that of roundTerminal.
func curlTerminal(name string, p design.Params, outer, inner cubic, centre outline.Point, how figStroke) chain {
	return must(curlTerminalTry(name, p, outer, inner, centre, how))
}

// curlTerminalTry is curlTerminal which returns an error where a part of the
// construction fails for the parameters.
func curlTerminalTry(name string, p design.Params, outer, inner cubic, centre outline.Point, how figStroke) (c chain, err error) {
	defer recovered(&err)
	b, err := curlBallPart(p)
	if err != nil {
		return nil, err
	}
	start := 1.0
	if how.Hook {
		start = tAtFall(outer, how.Fall)
	}
	t0 := outer.tangent(start)
	q0, phi0, k0 := outer.at(start), math.Atan2(-t0.Y, -t0.X), outer.curvature(start)
	w, err := newWindSpiral(q0, phi0, k0, centre, b.R)
	if err != nil {
		return nil, fmt.Errorf("glyph %s: %v", name, err)
	}
	tip := curlTip(name, w, centre, b.R, p.Figures.CurlNeck)
	tb := b.place(ballPlace{Centre: centre, Angle: tip - b.BallW})

	var in chain
	switch {
	case how.Arch:
		in, err = archToTip(name, inner.P0, inner.tangent(0), tb)
	case how.Tip:
		in, err = stalkToTip(name, inner, tb)
	default:
		in, err = stalkInner(name, inner, tb)
	}
	if err != nil {
		return nil, err
	}
	wound, err := windBall(name, tb, q0, phi0, k0)
	if err != nil {
		return nil, err
	}
	c = join(in, wound)
	if how.Hook {
		c = join(c, chain{curveSeg(outer.sub(start, 1))})
	}
	return c, nil
}

// curlTip returns the direction (radians) from the centre of the ball in which
// the tip of the wedge lies: of the points of the circle of the radius r about
// centre counter-clockwise from the point where the spiral w runs into it, the
// first whose distance from the spiral is neck, found by bisection between 10
// and 200 degrees from that point.
func curlTip(name string, w windSpiral, centre outline.Point, r, neck float64) float64 {
	n := int(math.Ceil(w.L))
	pts := make([]outline.Point, n+1)
	for i := range pts {
		pts[i] = w.point(w.L * float64(i) / float64(n))
	}
	aA := w.heading(w.L) + math.Pi/2
	gap := func(a float64) float64 {
		q := outline.Point{X: centre.X + r*math.Cos(a), Y: centre.Y + r*math.Sin(a)}
		d := math.Inf(1)
		for _, pt := range pts {
			d = math.Min(d, math.Hypot(pt.X-q.X, pt.Y-q.Y))
		}
		return d
	}
	lo, hi := aA+degrees(10), aA+degrees(200)
	if !(gap(lo) < neck && gap(hi) > neck) {
		panic(fmt.Sprintf("glyphs: glyph %s: no tip of the wedge lies %v from the outer edge; glyphs.Check reports such parameters", name, neck))
	}
	for i := 0; i < 50; i++ {
		mid := (lo + hi) / 2
		if gap(mid) < neck {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

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
	return must(roundTerminalTry(name, b, outer, inner, right, how))
}

// roundTerminalTry is roundTerminal which returns an error where a part of the
// construction fails for the parameters.
func roundTerminalTry(name string, b ballPart, outer, inner cubic, right float64, how figStroke) (c chain, err error) {
	defer recovered(&err)
	tb, err := figBallPlace(name, b, outer, right)
	if err != nil {
		return nil, err
	}

	// The inner edge, from its top to the tip of the wedge.
	var in chain
	switch {
	case how.Arch:
		in, err = archToTip(name, inner.P0, inner.tangent(0), tb)
	case how.Tip:
		in, err = stalkToTip(name, inner, tb)
	case how.Rise:
		in, err = stalkRising(name, how.Before, tb)
	default:
		in, err = stalkInner(name, inner, tb)
	}
	if err != nil {
		return nil, err
	}

	// The outer edge, wound into the ball.
	start := 1.0
	if how.Hook {
		start = tAtFall(outer, how.Fall)
	}
	t0 := outer.tangent(start)
	wound, err := windBall(name, tb, outer.at(start), math.Atan2(-t0.Y, -t0.X), outer.curvature(start))
	if err != nil {
		return nil, err
	}

	c = join(in, wound)
	if how.Hook {
		c = join(c, chain{curveSeg(outer.sub(start, 1))})
	}
	return c, nil
}

// A figStroke says how a figure leads its stroke into the ball (roundTerminal).
// The outer edge is a spiral (windSpiral) from the top of the round, or, if Hook
// is set, from the point where the outer edge, walked towards the ball, falls
// Fall degrees below the horizontal.  If Arch is set the inner edge runs from
// its top to the tip of the wedge in one arch (archToTip); if Tip is set it
// leaves the inner edge of the round and runs to the tip in one curve
// (stalkToTip); if Rise is set it runs from the start of Before, the quarter of
// the inner edge which ends at the top of inner, to the tip on a spiral whose
// curvature rises all the way (stalkRising), and the chain begins at the start
// of Before; otherwise it is the inner edge of the stalk of the part
// (stalkInner).
type figStroke struct {
	Arch, Hook, Tip, Rise bool
	Fall                  float64
	Before                cubic
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
// (arcHandle).  Where no such curve turns one way only, the curve arrives in the
// direction nearest to that of the part, turned towards that of the circular arc
// from q0 to the tip, in which the curve so made turns one way only and leaves
// q0 with a handle longer than archMinHandle.  The figure gives the inner edge
// on the other side of q0 the same curvature at q0.
func archToTip(name string, q0, t0 outline.Point, tb placedBall) (chain, error) {
	var tW outline.Point
	for _, pc := range tb.Chain {
		if isTip(pc.P3, tb.Tip) {
			tW = pc.tangent(1)
		}
	}
	c, ok := archCurve(q0, t0, tb.Tip, tW, 0)
	if !ok {
		aW := math.Atan2(tW.Y, tW.X)
		aC := math.Atan2(t0.Y, t0.X) + 2*(math.Atan2(tb.Tip.Y-q0.Y, tb.Tip.X-q0.X)-math.Atan2(t0.Y, t0.X))
		lo, hi := 0.0, 1.0
		cHi, okHi := archCurve(q0, t0, tb.Tip, dirOf(aC), archMinHandle)
		if !okHi {
			return nil, fmt.Errorf("glyph %s: no arch from the top of the counter to the tip of the wedge", name)
		}
		for i := 0; i < 50; i++ {
			mid := (lo + hi) / 2
			if cm, ok := archCurve(q0, t0, tb.Tip, dirOf(aW+mid*(aC-aW)), archMinHandle); ok {
				hi, cHi = mid, cm
			} else {
				lo = mid
			}
		}
		c = cHi
	}
	return splitAtExtrema(c), nil
}

// archCurve returns the curve of archToTip from q0, leaving in the direction t0,
// to tip, arriving in the direction tW, and whether it turns one way only and
// leaves q0 with a handle longer than minHandle.
func archCurve(q0, t0, tip, tW outline.Point, minHandle float64) (cubic, bool) {
	turn := math.Atan2(t0.X*tW.Y-t0.Y*tW.X, t0.X*tW.X+t0.Y*tW.Y) // negative: clockwise
	chord := math.Hypot(tip.X-q0.X, tip.Y-q0.Y)
	k := 2 * math.Sin(turn/2) / chord
	c, err := g2Start(q0, t0, k, tip, tW, arcHandle(q0, t0, tip, tW))
	if err != nil || !oneWay(c, -1) || !(math.Hypot(c.P1.X-c.P0.X, c.P1.Y-c.P0.Y) > minHandle) {
		return c, false
	}
	return c, true
}

// archMinHandle is the shortest handle with which the arch of archToTip leaves
// the top of the counter where it arrives at the tip of the wedge in another
// direction than the part.
const archMinHandle = 2.0

// stalkInner returns the inner edge from the top of inner to the tip of the
// wedge of tb: the inner edge of the round to a cut, one cubic curve from there
// to the section of the part, which leaves the round with its direction and its
// curvature, at the cut where its curvature changes least abruptly
// (curvatureRate), and the inner edge of the part from its section to the tip.
func stalkInner(name string, inner cubic, tb placedBall) (chain, error) {
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
		return nil, fmt.Errorf("glyph %s: no inner edge of the stalk reaches the ball terminal", name)
	}
	c := join(chain{curveSeg(inner.sub(0, tI))}, splitAtExtrema(in))
	for _, pc := range tb.Chain {
		c = append(c, pc)
		if isTip(pc.P3, tb.Tip) {
			break
		}
	}
	return c, nil
}

// stalkToTip returns the inner edge from the top of inner to the tip of the
// wedge of tb: the inner edge of the round to a cut, and one cubic curve from
// there to the tip, which leaves the round with its direction and its curvature
// and arrives with the direction and the curvature of the inner edge of the part
// (tipEnd), at the cut where its curvature changes least abruptly
// (curvatureRate).
func stalkToTip(name string, inner cubic, tb placedBall) (chain, error) {
	tip := tipEnd(tb)
	best, in, tI := math.Inf(1), cubic{}, 0.0
	for _, free := range []bool{false, true} {
		for _, u := range cCuts() {
			fi := stalkFoot{At: inner.at(u), Dir: inner.tangent(u), Kappa: inner.curvature(u)}
			if free {
				fi.Kappa = math.NaN()
			}
			if c, err := edgeIn(fi, tip); err == nil {
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
		return nil, fmt.Errorf("glyph %s: no inner edge of the stalk reaches the tip of the wedge", name)
	}
	return join(chain{curveSeg(inner.sub(0, tI))}, splitAtExtrema(in)), nil
}

// stalkRising returns the inner edge from the start of before, the quarter of
// the inner edge of the round which ends where inner begins, to the tip of the
// wedge of tb: the round to a cut, and a spiral (newRiseSpiral) which leaves the
// round there with its direction and its curvature and arrives at the tip in
// the direction of the inner edge of the part, its curvature rising with the
// power M of the length run.  The cut is the start of inner if the spiral from
// there has M at least 1; otherwise it is the point of before nearest to its
// end where M is 1, so that the curvature rises evenly, found by bisection to
// riseTolerance.
func stalkRising(name string, before cubic, tb placedBall) (chain, error) {
	tip := tipEnd(tb)
	spiral := func(u float64) (windSpiral, bool) {
		t := before.tangent(u)
		w, err := newRiseSpiral(before.at(u), math.Atan2(t.Y, t.X), -before.curvature(u), tip.At, math.Atan2(tip.Dir.Y, tip.Dir.X))
		return w, err == nil
	}
	u := 1.0
	if w, ok := spiral(1); !ok || w.M < 1 {
		hi, lo := 1.0, math.NaN()
		for v := 1 - riseStep; v > 0; v -= riseStep {
			if w, ok := spiral(v); ok && w.M >= 1 {
				lo = v
				break
			} else if ok {
				hi = v
			}
		}
		if math.IsNaN(lo) {
			return nil, fmt.Errorf("glyph %s: no inner edge with rising curvature reaches the tip of the wedge", name)
		}
		for hi-lo > riseTolerance {
			mid := (lo + hi) / 2
			if w, ok := spiral(mid); ok && w.M >= 1 {
				lo = mid
			} else {
				hi = mid
			}
		}
		u = lo
	}
	w, _ := spiral(u)
	in, err := w.riseCubics()
	if err != nil {
		return nil, fmt.Errorf("glyph %s: the inner edge to the tip of the wedge: %v", name, err)
	}
	in[len(in)-1].P3 = tip.At
	var c chain
	if u < 1 {
		c = chain{curveSeg(before.sub(0, u))}
	} else {
		c = chain{curveSeg(before)}
	}
	return join(c, in), nil
}

// riseStep is the step of the parameter of the round by which stalkRising
// scans for the cut, and riseTolerance the length of the interval of the
// parameter to which the bisection narrows it.
const (
	riseStep      = 0.01
	riseTolerance = 1e-9
)

// searchRiseSpiral returns the spiral (windSpiral) from p0, leaving in the
// direction phi0 (radians) with the curvature k0 (positive turning clockwise),
// which ends at p3 in the direction phi3: the length L, the power M and the
// curvature K1 at the end are the solution of the three equations for the end
// point and the direction there, found by Newton's method with M at least
// riseMinPower.  It reports an error if there is none, or if the method stays
// at riseMinPower for riseFloorSteps steps in a row or makes too little
// progress (riseStall).
func searchRiseSpiral(p0 outline.Point, phi0, k0 float64, p3 outline.Point, phi3 float64) (windSpiral, error) {
	w := windSpiral{P0: p0, Phi0: phi0, K0: k0}
	for phi3-phi0 > math.Pi {
		phi3 -= 2 * math.Pi
	}
	for phi3-phi0 < -math.Pi {
		phi3 += 2 * math.Pi
	}
	// For a length L and a power m the direction at the end gives K1.
	set := func(L, m float64) windSpiral {
		v := w
		v.L, v.M = L, m
		v.K1 = k0 + (m+1)*((phi0-phi3)-k0*L)/L
		return v
	}
	miss := func(L, m float64) (float64, float64) {
		v := set(L, m)
		q := v.point(L)
		return q.X - p3.X, q.Y - p3.Y
	}
	L, m := 1.02*math.Hypot(p3.X-p0.X, p3.Y-p0.Y), 2.0
	floor, first := 0, 0.0
	for i := 0; i < 200; i++ {
		fx, fy := miss(L, m)
		r := math.Hypot(fx, fy)
		if r < 1e-7 {
			return set(L, m), nil
		}
		if i == 0 {
			first = r
		} else if i == riseStallSteps && r > riseStall*first {
			break
		}
		const h = 1e-6
		ax, ay := miss(L+h, m)
		bx, by := miss(L, m+h)
		j11, j21, j12, j22 := (ax-fx)/h, (ay-fy)/h, (bx-fx)/h, (by-fy)/h
		det := j11*j22 - j12*j21
		if det == 0 || math.IsNaN(det) {
			break
		}
		dL, dm := (fx*j22-fy*j12)/det, (j11*fy-j21*fx)/det
		for math.Abs(dL) > 0.05*L || math.Abs(dm) > 0.5 {
			dL, dm = dL/2, dm/2
		}
		L, m = L-dL, math.Max(m-dm, riseMinPower)
		if m == riseMinPower {
			floor++
			if floor == riseFloorSteps {
				break
			}
		} else {
			floor = 0
		}
	}
	return w, fmt.Errorf("no spiral from %v runs to %v", p0, p3)
}

// riseMinPower is the least power M of a spiral of newRiseSpiral and
// riseFloorSteps the number of steps in a row at that power after which the
// search ends without a spiral; the search also ends without one if after
// riseStallSteps steps the miss of the end point is more than riseStall times
// the first.
const (
	riseMinPower   = 0.1
	riseFloorSteps = 4
	riseStallSteps = 20
	riseStall      = 0.8
)

// riseCubics returns the spiral as cubic curves walked from P0 to its end, cut
// at its extremes and halved where a piece turns by more than riseMaxTurn; each
// curve has the points, the directions and the curvatures of the spiral at its
// ends (g2Cubic).
func (w windSpiral) riseCubics() (chain, error) {
	cuts := []float64{0}
	end := w.heading(w.L)
	lo, hi := math.Min(w.Phi0, end), math.Max(w.Phi0, end)
	for q := math.Ceil(lo/(math.Pi/2)) * math.Pi / 2; q < hi; q += math.Pi / 2 {
		if q <= lo {
			continue
		}
		if s := w.lengthAt(q); s > 1 && w.L-s > 1 {
			cuts = append(cuts, s)
		}
	}
	cuts = append(cuts, w.L)
	var ch chain
	for i := 1; i < len(cuts); i++ {
		a, b := cuts[i-1], cuts[i]
		n := int(math.Ceil(math.Abs(w.heading(b)-w.heading(a))/riseMaxTurn - 1e-9))
		if n < 1 {
			n = 1
		}
		for k := 0; k < n; k++ {
			sa, sb := a+(b-a)*float64(k)/float64(n), a+(b-a)*float64(k+1)/float64(n)
			c, err := g2Cubic(w.point(sa), dirOf(w.heading(sa)), -w.curvature(sa), w.point(sb), dirOf(w.heading(sb)), -w.curvature(sb))
			if err != nil {
				return nil, err
			}
			ch = append(ch, curveSeg(c))
		}
	}
	return ch, nil
}

// riseMaxTurn is the largest turn (radians) of one cubic curve of riseCubics.
const riseMaxTurn = math.Pi / 4

// figBallPlace returns the part placed as on c (roundTerminalAt), turned so that
// the rightmost ink of the part lies at right with the centre of the ball on an
// integer abscissa, or an error where no turn of the part puts its ink there.
func figBallPlace(name string, b ballPart, outer cubic, right float64) (tb placedBall, err error) {
	defer recovered(&err)
	x1 := func(angle float64) float64 {
		tb, _ := roundTerminalAt(b, outer, angle)
		_, _, x1, _ := chainBBox(tb.Chain)
		return x1
	}
	solve := func(target float64) (float64, error) {
		lo, hi := figBallAngles[0], figBallAngles[1]
		if !(x1(lo) >= target && x1(hi) <= target) {
			return 0, fmt.Errorf("glyph %s: the ball cannot be placed with its ink at %.1f (%.1f to %.1f)", name, target, x1(hi), x1(lo))
		}
		for i := 0; i < 60; i++ {
			mid := (lo + hi) / 2
			if x1(mid) > target {
				lo = mid
			} else {
				hi = mid
			}
		}
		return (lo + hi) / 2, nil
	}
	angle, err := solve(right)
	if err != nil {
		return placedBall{}, err
	}
	tb, _ = roundTerminalAt(b, outer, angle)
	target := right + math.Round(tb.Centre.X) - tb.Centre.X
	if angle, err = solve(target); err != nil {
		return placedBall{}, err
	}
	tb, _ = roundTerminalAt(b, outer, angle)
	return tb, nil
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
	how.Before = mirrorCubic(how.Before, 0)
	return roundTerminal(name, b, mirrorCubic(outer, 0), mirrorCubic(inner, 0), -left, how).mirrorX(0)
}

// lowerLeftBall returns the ball terminal at the lower left of a round: outer is
// the lower left quarter of the outer edge (from the left extreme to the bottom,
// counter-clockwise) and inner that of the inner edge (from the bottom to the left
// extreme, clockwise), and left is the leftmost ink of the ball.  The chain runs
// from the bottom of the inner edge round the ball to the bottom of the outer edge.
func lowerLeftBall(name string, b ballPart, outer, inner cubic, left float64, how figStroke) chain {
	how.Before = turnCubic(how.Before)
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
