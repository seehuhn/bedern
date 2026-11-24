// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"fmt"
	"math"
)

// endCap returns the right and the left edge with the end e applied at their
// ends, and the chain of the end from the end of the right edge to the end of
// the left edge.  nd is the last node and er, el the edges there.
func (r *Result) endCap(right, left Chain, e End, nd Node, er, el edgeData) (Chain, Chain, Chain, error) {
	line := func(a, b Point, k PieceKind) Chain {
		if dist(a, b) < 1e-9 {
			return nil
		}
		return Chain{{C: Cubic{a, lerp(a, b, 1.0/3), lerp(a, b, 2.0/3), b}, Kind: k}}
	}
	switch e.Kind {
	case Cut, Clip:
		kind := EndPiece
		if e.Sever {
			kind = JoinPiece
		}
		if e.Hidden {
			kind = HiddenPiece
		}
		if e.Kind == Cut && !e.Line {
			return right, left, line(right.end(), left.end(), kind), nil
		}
		q := nd.P
		if e.Kind == Clip {
			q, kind = e.Q, JoinPiece
		}
		u := dirOf(e.Angle)
		ahead := add(nd.P, mul(dirOf(nd.Dir), 1000))
		rr, err := clipChain(right, q, u, ahead)
		if err != nil {
			return nil, nil, nil, err
		}
		ll, err := clipChain(left, q, u, ahead)
		if err != nil {
			return nil, nil, nil, err
		}
		return rr, ll, line(rr.end(), ll.end(), kind), nil
	case Round:
		t := dirOf(nd.Dir)
		a0 := math.Atan2(t.Y, t.X) - math.Pi/2 // the right edge: the direction minus a quarter turn
		return right, left, circleArc(nd.P, nd.W/2, a0, a0+math.Pi, EndPiece), nil
	case Wedge:
		t := dirOf(nd.Dir)
		tip := add(add(nd.P, mul(t, e.Length)), mul(rot90(t), e.Offset))
		cr, _ := hermite(er.P, er.T, er.K, tip, unit(sub(tip, er.P)), 0)
		cl, _ := hermite(el.P, el.T, el.K, tip, unit(sub(tip, el.P)), 0)
		return right, left, Chain{{C: cr, Kind: EndPiece}, {C: cl.Reversed(), Kind: EndPiece}}, nil
	case Ball:
		b := e.Ball
		if b.InkSide != 0 && b.Power[1] == 0 {
			// The ball moves sideways until its ink reaches InkX.
			n0 := len(r.Warnings)
			for i := 0; i < 12; i++ {
				cp, err := r.ball(b, nd, er, el)
				r.Warnings, r.WarnAt = r.Warnings[:n0], r.WarnAt[:n0]
				if err != nil {
					return nil, nil, nil, err
				}
				x0, x1 := chainXRange(cp)
				x := x0
				if b.InkSide > 0 {
					x = x1
				}
				if math.Abs(b.InkX-x) < 1e-6 {
					break
				}
				b.Centre.X += b.InkX - x
			}
		}
		cp, err := r.ball(b, nd, er, el)
		return right, left, cp, err
	}
	return nil, nil, nil, fmt.Errorf("stroke: unknown end %d", e.Kind)
}

// chainXRange returns the least and the greatest abscissa of the chain.
func chainXRange(ch Chain) (float64, float64) {
	x0, x1 := math.Inf(1), math.Inf(-1)
	for _, pc := range ch {
		for _, q := range pieceSamples(pc.C) {
			x0, x1 = math.Min(x0, q.X), math.Max(x1, q.X)
		}
	}
	return x0, x1
}

// clipChain returns the chain up to the last point where it crosses the line
// through q in the direction u, on its way to the side of the point ahead.  If
// its end lies before the line, it is continued beyond the end to the line.
func clipChain(ch Chain, q, u, ahead Point) (Chain, error) {
	f := func(p Point) float64 { return cross(u, sub(p, q)) }
	before := (f(ch.end()) > 0) != (f(ahead) > 0)
	for i := len(ch) - 1; i >= 0 && !before; i-- {
		c := ch[i].C
		tmax := 1.0
		const steps = 400
		prevT, prevF := tmax, f(c.At(tmax))
		for k := steps - 1; k >= 0; k-- {
			t := tmax * float64(k) / steps
			ft := f(c.At(t))
			if (ft > 0) != (prevF > 0) || ft == 0 {
				lo, hi, flo := t, prevT, ft
				for j := 0; j < 80; j++ {
					m := (lo + hi) / 2
					if fm := f(c.At(m)); (fm > 0) == (flo > 0) {
						lo, flo = m, fm
					} else {
						hi = m
					}
				}
				tc := (lo + hi) / 2
				out := append(Chain(nil), ch[:i]...)
				return append(out, Piece{C: c.Sub(0, tc), Kind: ch[i].Kind}), nil
			}
			prevT, prevF = t, ft
		}
	}
	// If the edge does not cross the line, it continues beyond its end on its
	// circle of curvature, to the line; the short piece is merged later.
	e, tt := ch.end(), ch[len(ch)-1].C.Tangent(1)
	k := ch[len(ch)-1].C.Curvature(1)
	at := func(s float64) (Point, Point) {
		if math.Abs(k*s) < 1e-9 {
			return add(e, mul(tt, s)), tt
		}
		n := rot90(tt)
		a := k * s
		return add(e, add(mul(tt, math.Sin(a)/k), mul(n, (1-math.Cos(a))/k))), rotate(tt, a*180/math.Pi)
	}
	prevS, prevF := 0.0, f(e)
	for s := 1.0; s <= 300; s++ {
		x, _ := at(s)
		if fx := f(x); (fx > 0) != (prevF > 0) || fx == 0 {
			lo, hi, flo := prevS, s, prevF
			for j := 0; j < 80; j++ {
				m := (lo + hi) / 2
				xm, _ := at(m)
				if fm := f(xm); (fm > 0) == (flo > 0) {
					lo, flo = m, fm
				} else {
					hi = m
				}
			}
			x, tx := at((lo + hi) / 2)
			c, _ := hermite(e, tt, k, x, tx, k)
			out := append(Chain(nil), ch...)
			return append(out, Piece{C: c, Kind: ch[len(ch)-1].Kind}), nil
		}
		prevS, prevF = s, f(x)
	}
	return nil, fmt.Errorf("stroke: an edge does not reach the line through %.1f,%.1f", q.X, q.Y)
}

// arcMinPiece is the shortest arc of a circle which ends at an extreme: an
// extreme nearer to an end of the arc is left inside the curve.
const arcMinPiece = 2.0

// circleArc returns the arc of the circle about c of the radius r from the
// angle a0 counter-clockwise to a1 (radians), as cubic curves with the handles
// of a circular arc, with nodes at the extremes of the circle.
func circleArc(c Point, r, a0, a1 float64, kind PieceKind) Chain {
	for a1 <= a0 {
		a1 += 2 * math.Pi
	}
	cuts := []float64{a0}
	for q := math.Ceil(a0/(math.Pi/2)) * math.Pi / 2; q < a1; q += math.Pi / 2 {
		if (q-cuts[len(cuts)-1])*r >= arcMinPiece && (a1-q)*r >= arcMinPiece {
			cuts = append(cuts, q)
		}
	}
	cuts = append(cuts, a1)
	var ch Chain
	for i := 1; i < len(cuts); i++ {
		ch = append(ch, Piece{C: arcCubic(c, r, cuts[i-1], cuts[i]), Kind: kind})
	}
	return ch
}

// arcCubic returns the cubic arc of the circle from a0 to a1, counter-clockwise.
func arcCubic(c Point, r, a0, a1 float64) Cubic {
	h := 4.0 / 3 * math.Tan((a1-a0)/4) * r
	p0 := add(c, Point{X: r * math.Cos(a0), Y: r * math.Sin(a0)})
	p3 := add(c, Point{X: r * math.Cos(a1), Y: r * math.Sin(a1)})
	t0 := Point{X: -math.Sin(a0), Y: math.Cos(a0)}
	t3 := Point{X: -math.Sin(a1), Y: math.Cos(a1)}
	return Cubic{p0, add(p0, mul(t0, h)), sub(p3, mul(t3, h)), p3}
}

// ball returns the ball terminal at the end node nd: the chain from the end of
// the right edge er round the ball to the end of the left edge el.  The ball
// lies on the side of the stroke towards which it curls; the edge on the other
// side is the outer edge, which winds into the circle on a spiral, and the
// other edge runs to the tip of the wedge in one curve.
func (r *Result) ball(b BallSpec, nd Node, er, el edgeData) (Chain, error) {
	t := dirOf(nd.Dir)
	sigma := 1.0
	outer, inner := er, el
	if cross(t, sub(b.Centre, nd.P)) < 0 {
		sigma, outer, inner = -1, el, er
	}
	var spiral Chain
	if b.Power[1] > 0 {
		// The end node is placed so that a spiral of a power in the range
		// exists (placeBalls).
		f, err := fitBall(outer, sigma, b.Centre, b.R, b.Power[0], b.Power[1])
		if err != nil {
			return nil, err
		}
		if f.Turn != 0 || math.Abs(f.Slide) > 0.01 {
			return nil, fmt.Errorf("stroke: the end node before the ball about %.1f,%.1f is not placed", b.Centre.X, b.Centre.Y)
		}
		sp, err := newSpiralFrom(outer.P, outer.T, sigma, math.Max(sigma*outer.K, 0), b.Centre, b.R, f.L, f.M)
		if err != nil {
			return nil, err
		}
		spiral = sp.chain()
	} else if sp, err := newSpiral(outer.P, outer.T, sigma, sigma*outer.K, b.Centre, b.R); err == nil {
		spiral = sp.chain()
	} else {
		c, err2 := windCurve(outer, sigma, b.Centre, b.R)
		if err2 != nil {
			return nil, err
		}
		r.warn("ball: no spiral runs into the ball; the outer edge winds in on one curve")
		spiral = Chain{{C: c, Kind: EdgePiece}}
	}
	aA := angleOf(sub(spiral.end(), b.Centre)) * math.Pi / 180

	if b.Neck > 0 {
		tip, err := neckTip(spiral, b, aA, sigma)
		if err != nil {
			return nil, err
		}
		b.Tip = tip
	}
	if b.Power[1] > 0 {
		// The tip moves onto an extreme of the circle which lies nearer to
		// it than MinPiece, so that no short piece of the circle is left
		// between them.
		for q := -360.0; q <= 360; q += 90 {
			if math.Abs(b.Tip-q)*math.Pi/180*b.R < MinPiece {
				b.Tip = q
			}
		}
	}
	w := add(b.Centre, mul(dirOf(b.Tip), b.R))
	tau := mul(rot90(unit(sub(w, b.Centre))), sigma)
	dw := rotate(tau, sigma*b.Open)
	kw := b.TipK
	if isAuto(kw) {
		kw = inner.K
	}
	// The inner edge runs to the tip in one or two curves (edgeCurves); where
	// none exists with the curvature at the tip, that curvature moves in steps
	// of 5 per cent, up to a factor of three either way, to the nearest value
	// for which one does.
	tipData := edgeData{P: w, T: dw, K: kw}
	var stalkCs []Cubic
	ok := false
	if b.Power[1] > 0 {
		// By the rule of the spiral the tip is a corner whose curvature is
		// free: the inner edge is one cubic which leaves the end node with
		// the direction and the curvature of the edge there and arrives at
		// the tip in the direction of the wedge, its handle at the tip that
		// of the circular arc between the two directions.
		if c, good := stalkToTip(inner, w, dw); good {
			stalkCs, ok = []Cubic{c}, true
		}
	}
	if !ok {
		stalkCs, ok = edgeCurves(inner, tipData)
	}
	for step := 1; !ok && step <= 40; step++ {
		for _, f := range []float64{1 + 0.05*float64(step), 1 - 0.05*float64(step)} {
			if f < 1.0/3 {
				continue
			}
			tipData.K = kw * f
			if cs, ok2 := edgeCurves(inner, tipData); ok2 {
				stalkCs, ok = cs, true
				break
			}
		}
	}
	if !ok && b.Power[1] > 0 {
		// The handle at the tip of the circular arc leaves no cubic with
		// the curvature of the edge at the end node: a handle nearer to the
		// tip does.
		if c, good := stalkToTipShort(inner, w, dw); good {
			stalkCs, ok = []Cubic{c}, true
		}
	}
	if !ok {
		r.warn("ball: the inner edge to the tip of the wedge has the curvatures only approximately")
	}
	var stalkCh Chain
	for _, c := range stalkCs {
		stalkCh = append(stalkCh, Piece{C: c, Kind: EdgePiece})
	}
	aW := b.Tip * math.Pi / 180
	// The circle from A, where the spiral runs into it, to the tip.
	var arc Chain
	if sigma > 0 {
		arc = circleArc(b.Centre, b.R, aA, aW, EndPiece)
	} else {
		arc = circleArc(b.Centre, b.R, aW, aA, EndPiece).reversed()
	}
	if b.Power[1] > 0 {
		spiral, arc = joinShortWind(spiral, arc)
	}
	var cp Chain
	if sigma > 0 {
		cp = append(cp, spiral...)
		cp = append(cp, arc...)
		cp = append(cp, stalkCh.reversed()...)
	} else {
		cp = append(cp, stalkCh...)
		cp = append(cp, arc.reversed()...)
		cp = append(cp, spiral.reversed()...)
	}
	return cp, nil
}

// ballArcMin is the length under which the first piece of the circle is
// joined with the last piece of the spiral (joinShortWind), where MinPiece is
// that for the last piece of the spiral: the handles of a shorter arc are
// under 8 units long and the rounding of the outline turns them by about 2
// degrees.
const ballArcMin = 25.0

// joinShortWind draws the last piece of the spiral and the first piece of the
// circle after it, where the last piece of the spiral is shorter than MinPiece or the
// first of the circle than ballArcMin, without a
// short piece.  Where the other is at least twice MinPiece long, the short
// one and the part of the other which it needs to reach MinPiece become one
// curve, which leaves the spiral with its direction and curvature and
// arrives with those of the circle; the rest of the long piece stays as it
// is.  Otherwise the two are one curve, as the roman's windBall draws them,
// whose curvature swings by more than the spiral allows over a long turn.
// The spiral runs from the end node to A, the arc from A on.
func joinShortWind(spiral, arc Chain) (Chain, Chain) {
	if len(spiral) < 2 || len(arc) < 1 {
		return spiral, arc
	}
	last := spiral[len(spiral)-1].C
	a := arc[0].C
	if last.Length() >= MinPiece && a.Length() >= ballArcMin {
		return spiral, arc
	}
	k1 := a.Curvature(1)
	switch {
	case a.Length() < ballArcMin && last.Length() >= 2*MinPiece:
		// The end of the last piece, as long as the first arc needs to
		// reach MinPiece with a margin, and the first arc.
		l := last.Length()
		tail := math.Min(math.Max(MinPiece+2-a.Length(), 0.15*l), 0.5*l)
		first, second := last.Split(1 - tail/l)
		c, ok := hermite(second[0], second.Tangent(0), second.Curvature(0), a[3], a.Tangent(1), k1)
		if !ok {
			break
		}
		out := append(spiral[:len(spiral)-1:len(spiral)-1], Piece{C: first, Kind: EdgePiece}, Piece{C: c, Kind: EdgePiece})
		return out, arc[1:]
	case last.Length() < MinPiece && a.Length() >= 2*MinPiece:
		// The last piece and half of the first arc, which stays half a
		// circle piece.
		m := a.At(0.5)
		_, rest := a.Split(0.5)
		c, ok := hermite(last[0], last.Tangent(0), last.Curvature(0), m, a.Tangent(0.5), a.Curvature(0.5))
		if !ok {
			break
		}
		out := append(spiral[:len(spiral)-1:len(spiral)-1], Piece{C: c, Kind: EdgePiece})
		return out, append(Chain{{C: rest, Kind: arc[0].Kind}}, arc[1:]...)
	}
	c, ok := hermite(last[0], last.Tangent(0), last.Curvature(0), a[3], a.Tangent(1), k1)
	if !ok {
		return spiral, arc
	}
	spiral = append(spiral[:len(spiral)-1:len(spiral)-1], Piece{C: c, Kind: EdgePiece})
	return spiral, arc[1:]
}

// stalkToTip returns the inner edge of a curl from the edge a to the tip w,
// where it arrives in the direction t: a cubic whose handle at w is that of
// the circular arc which turns from the direction of a to t over the chord,
// and whose handle at a gives it the curvature of a there.  It reports false
// where no such cubic exists or where its curvature changes sign.
func stalkToTip(a edgeData, w, t Point) (Cubic, bool) {
	chord := dist(a.P, w)
	theta := math.Abs(math.Atan2(cross(a.T, t), dot(a.T, t)))
	hb := chord / 3
	if theta > 1e-6 {
		hb = chord * 2 / 3 * math.Tan(theta/4) / math.Sin(theta/2)
	}
	p2 := sub(w, mul(t, hb))
	// k(0) = 2 cross(t0, p2 - p0) / (3 ha^2).
	cr := cross(a.T, sub(p2, a.P))
	var ha float64
	switch {
	case math.Abs(a.K) < 1e-6:
		ha = chord / 3
	case cr/a.K > 0:
		ha = math.Sqrt(2 * cr / (3 * a.K))
	default:
		return Cubic{}, false
	}
	if ha < minHandle || ha > chord {
		return Cubic{}, false
	}
	c := Cubic{a.P, add(a.P, mul(a.T, ha)), p2, w}
	k0 := c.Curvature(0)
	for i := 1; i <= 32; i++ {
		if c.Curvature(float64(i)/32)*k0 < 0 {
			return Cubic{}, false
		}
	}
	return c, true
}

// tipHandleMin is the shortest handle of the inner edge of a curl whose chord
// is short: half of minHandle.
const tipHandleMin = minHandle / 2

// stalkToTipShort is stalkToTip with the handle at the tip shortened from
// that of the circular arc, by the largest step of 2 per cent of it for which
// the cubic with the curvature of a at its start exists and its curvature
// keeps its sign, down to a handle of tipHandleMin.
func stalkToTipShort(a edgeData, w, t Point) (Cubic, bool) {
	chord := dist(a.P, w)
	theta := math.Abs(math.Atan2(cross(a.T, t), dot(a.T, t)))
	hb0 := chord / 3
	if theta > 1e-6 {
		hb0 = chord * 2 / 3 * math.Tan(theta/4) / math.Sin(theta/2)
	}
	for f := 0.98; f*hb0 >= tipHandleMin; f -= 0.02 {
		p2 := sub(w, mul(t, f*hb0))
		cr := cross(a.T, sub(p2, a.P))
		var ha float64
		switch {
		case math.Abs(a.K) < 1e-6:
			ha = chord / 3
		case cr/a.K > 0:
			ha = math.Sqrt(2 * cr / (3 * a.K))
		default:
			continue
		}
		if ha < tipHandleMin || ha > chord {
			continue
		}
		c := Cubic{a.P, add(a.P, mul(a.T, ha)), p2, w}
		k0 := c.Curvature(0)
		good := true
		for i := 1; i <= 32; i++ {
			if c.Curvature(float64(i)/32)*k0 < 0 {
				good = false
				break
			}
		}
		if good {
			return c, true
		}
	}
	return Cubic{}, false
}

// neckMaxTurn is the greatest angle, in degrees, from where the outer edge
// runs into the ball back to the tip of the wedge which neckTip considers.
const neckMaxTurn = 260.0

// neckTip returns the direction in degrees from the centre of the ball of the
// tip of the wedge: of the points of the circle from A, where the outer edge
// runs into it, back against the turn of the curl towards the stalk, the first
// whose distance from the outer edge is the neck, between 10 and 200 degrees
// from A.  This is the rule of curlTip in the figures of the font.
func neckTip(outer Chain, b BallSpec, aA, sigma float64) (float64, error) {
	var pts []Point
	for _, pc := range outer {
		pts = append(pts, pieceSamples(pc.C)...)
	}
	gap := func(a float64) float64 {
		q := add(b.Centre, Point{X: b.R * math.Cos(a), Y: b.R * math.Sin(a)})
		d := math.Inf(1)
		for _, pt := range pts {
			d = math.Min(d, dist(pt, q))
		}
		return d
	}
	lo, hi := aA-sigma*10*math.Pi/180, aA-sigma*neckMaxTurn*math.Pi/180
	if !(gap(lo) < b.Neck && gap(hi) > b.Neck) {
		return 0, fmt.Errorf("stroke: no tip of the wedge lies %.1f from the outer edge of the ball", b.Neck)
	}
	for i := 0; i < 50; i++ {
		m := (lo + hi) / 2
		if gap(m) < b.Neck {
			lo = m
		} else {
			hi = m
		}
	}
	return (lo + hi) / 2 * 180 / math.Pi, nil
}

// windCurve returns the outer edge of a curl which runs into the circle about
// centre of the radius r as one cubic, where no spiral does: from the end of
// the edge with its direction and curvature to the point of the circle where
// the curve arrives with the direction and the curvature of the circle.  Of the
// points of the circle in steps of 2 degrees, in the order of the turn of the
// edge from the end of the stroke to the point, from 30 to 300 degrees, the
// first is taken for which the cubic exists and turns only to the side of the
// curl.
func windCurve(e edgeData, sigma float64, centre Point, r float64) (Cubic, error) {
	phi0 := math.Atan2(e.T.Y, e.T.X)
	for turn := 30.0; turn <= 300; turn += 2 {
		heading := phi0 + sigma*turn*math.Pi/180
		t := Point{X: math.Cos(heading), Y: math.Sin(heading)}
		// The point of the circle where the circle runs in the direction t,
		// turning to the side sigma.
		a := add(centre, mul(rot90(t), -sigma*r))
		c, ok := hermite(e.P, e.T, e.K, a, t, sigma/r)
		if !ok {
			continue
		}
		rising := true
		for i := 1; i <= 16; i++ {
			if c.Curvature(float64(i)/16)*sigma < -1e-4 {
				rising = false
				break
			}
		}
		if rising {
			return c, nil
		}
	}
	return Cubic{}, fmt.Errorf("stroke: no curve runs into the ball about %.1f,%.1f", centre.X, centre.Y)
}

// rotate returns v turned by a degrees.
func rotate(v Point, a float64) Point {
	c, s := math.Cos(a*math.Pi/180), math.Sin(a*math.Pi/180)
	return Point{X: c*v.X - s*v.Y, Y: s*v.X + c*v.Y}
}

// A spiral leaves P0 in the direction Phi0 (radians) and turns to the side
// Sigma (+1 left) with the curvature
//
//	k(s) = K0 + (K1 - K0) (s/L)^M,
//
// which rises to K1, the curvature of the circle it ends on.  It is the curve
// of windSpiral in the glyphs of the font, written again here.
type spiral struct {
	P0        Point
	Phi0, Sig float64
	K0, K1    float64
	L, M      float64
}

// newSpiral returns the spiral from p0 in the direction t0 with the curvature
// k0 (a magnitude, towards the side sigma) which ends on the circle about
// centre of the radius r with its direction and its curvature: L and M put the
// centre of curvature at the end on the centre, by Newton's method from a
// fixed list of starting values.
func newSpiral(p0, t0 Point, sigma, k0 float64, centre Point, r float64) (spiral, error) {
	// Newton's method from a list of starting values, the first which
	// converges.
	for _, start := range [][2]float64{{1.5, 2}, {1, 2}, {2.5, 2}, {1.5, 1}, {2.5, 3}, {3.5, 2}} {
		v, err := newSpiralFrom(p0, t0, sigma, k0, centre, r, start[0]*dist(centre, p0), start[1])
		if err == nil || v.L > 0 {
			return v, err
		}
	}
	return spiral{}, fmt.Errorf("stroke: no spiral from %.1f,%.1f runs into the ball about %.1f,%.1f", p0.X, p0.Y, centre.X, centre.Y)
}

// newSpiralFrom is newSpiral from the one starting value L, m.  Where
// Newton's method does not converge the spiral has the length zero.
func newSpiralFrom(p0, t0 Point, sigma, k0 float64, centre Point, r, L, m float64) (spiral, error) {
	w := spiral{P0: p0, Phi0: math.Atan2(t0.Y, t0.X), Sig: sigma, K0: k0, K1: 1 / r}
	miss := func(L, m float64) Point {
		v := w
		v.L, v.M = L, m
		return sub(v.centreAt(L), centre)
	}
	{
		for i := 0; i < 200; i++ {
			f := miss(L, m)
			if norm(f) < 1e-7 {
				w.L, w.M = L, m
				if m < 0.5 {
					return w, fmt.Errorf("stroke: the spiral into the ball needs the power %.2f", m)
				}
				return w, nil
			}
			const h = 1e-5
			a, b := miss(L+h, m), miss(L, m+h)
			j11, j21, j12, j22 := (a.X-f.X)/h, (a.Y-f.Y)/h, (b.X-f.X)/h, (b.Y-f.Y)/h
			det := j11*j22 - j12*j21
			if det == 0 || math.IsNaN(det) {
				break
			}
			dL, dm := (f.X*j22-f.Y*j12)/det, (j11*f.Y-j21*f.X)/det
			for math.Abs(dL) > 0.2*L || math.Abs(dm) > 0.5 {
				dL, dm = dL/2, dm/2
			}
			L, m = L-dL, math.Max(m-dm, 0.1)
		}
	}
	return spiral{}, fmt.Errorf("stroke: no spiral from %.1f,%.1f runs into the ball about %.1f,%.1f", p0.X, p0.Y, centre.X, centre.Y)
}

func (w spiral) heading(s float64) float64 {
	return w.Phi0 + w.Sig*(w.K0*s+(w.K1-w.K0)*w.L/(w.M+1)*math.Pow(s/w.L, w.M+1))
}

func (w spiral) curvature(s float64) float64 { return w.K0 + (w.K1-w.K0)*math.Pow(s/w.L, w.M) }

func (w spiral) point(s float64) Point { return w.pointN(s, 400) }

// pointN is point by Simpson's rule with n steps (n even).
func (w spiral) pointN(s float64, n int) Point {
	h := s / float64(n)
	q := w.P0
	for i := 0; i <= n; i++ {
		c := 2.0
		if i == 0 || i == n {
			c = 1
		} else if i%2 == 1 {
			c = 4
		}
		phi := w.heading(float64(i) * h)
		q.X += c * h / 3 * math.Cos(phi)
		q.Y += c * h / 3 * math.Sin(phi)
	}
	return q
}

func (w spiral) centreAt(s float64) Point { return w.centreAtN(s, 400) }

// centreAtN is centreAt with the point by pointN.
func (w spiral) centreAtN(s float64, steps int) Point {
	phi := w.heading(s)
	n := Point{X: -math.Sin(phi), Y: math.Cos(phi)}
	return add(w.pointN(s, steps), mul(n, w.Sig/w.curvature(s)))
}

// chain returns the spiral from P0 to its end as cubic curves with nodes at
// the extremes, each with the points, directions and curvatures of the spiral
// at its ends.  An extreme within 20 units of P0 stays inside a curve.
func (w spiral) chain() Chain {
	end := w.heading(w.L)
	cuts := []float64{0}
	lo, hi := math.Min(w.Phi0, end), math.Max(w.Phi0, end)
	var qs []float64
	for q := math.Ceil(lo/(math.Pi/2)) * math.Pi / 2; q < hi; q += math.Pi / 2 {
		qs = append(qs, q)
	}
	if w.Sig < 0 {
		for i, j := 0, len(qs)-1; i < j; i, j = i+1, j-1 {
			qs[i], qs[j] = qs[j], qs[i]
		}
	}
	for _, q := range qs {
		s := w.lengthAt(q)
		if s > 20 && w.L-s > 1 && s-cuts[len(cuts)-1] > MinPiece {
			cuts = append(cuts, s)
		}
	}
	cuts = append(cuts, w.L)
	var ch Chain
	for i := 1; i < len(cuts); i++ {
		a, b := cuts[i-1], cuts[i]
		ta, tb := w.heading(a), w.heading(b)
		c, _ := hermite(w.point(a), Point{X: math.Cos(ta), Y: math.Sin(ta)}, w.Sig*w.curvature(a),
			w.point(b), Point{X: math.Cos(tb), Y: math.Sin(tb)}, w.Sig*w.curvature(b))
		ch = append(ch, Piece{C: c, Kind: EdgePiece})
	}
	return ch
}

// lengthAt returns the length after which the spiral has the heading phi.
func (w spiral) lengthAt(phi float64) float64 {
	lo, hi := 0.0, w.L
	for i := 0; i < 60; i++ {
		m := (lo + hi) / 2
		if (w.heading(m)-phi)*w.Sig < 0 {
			lo = m
		} else {
			hi = m
		}
	}
	return (lo + hi) / 2
}
