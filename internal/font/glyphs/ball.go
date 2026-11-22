// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The ball terminal.
//
// A thin stalk curls round and ends in a round ball, with a wedge of white
// between the stalk and the ball.  The terminal is one part, the same on every
// letter which has one (a, c, f, j, r, y): the letters place it, turned and
// perhaps mirrored, and draw the stalk from their stems or strokes to its
// section.  The letters c, f and r, like the figures, draw only the circle of the
// ball and the tip of the wedge of the part, and wind the outer edge of the
// stalk into the ball on a spiral (windSpiral, windHook).  The part is made in its own frame, in which the centre O of the ball
// is the origin and the stalk comes in from the upper left, as on r: it runs
// over the top to the right and curls down into the ball, and the wedge lies at
// the upper left of the ball.
//
// The construction.  Near the ball the stalk is a band of the thickness BallStalk
// round a centre of curl C: its outer edge is the circle of the radius Ro about C,
// its inner edge the circle of the radius Ri = Ro - BallStalk.  The outer circle
// encloses the ball and touches it on the right (C lies left of O, at the distance
// Ro - r).  The inner circle cuts the circle of the ball at W, the tip of the
// wedge, which is a corner: the white between the inner edge and the ball opens
// from there at the angle between the two circles.  The part starts at the
// section S of the band, ballSection beyond W seen from C, where the stalk of the
// letter ends: the outer edge and the inner edge there have the direction and the
// curvature of the two circles.  From S the inner edge is the arc of the inner
// circle down to W; from W the outline is the circle of the ball, round its far
// side to A; from A the outer edge is one cubic curve (blend) which leaves the
// circle of the ball with its direction and its curvature and joins the outer
// circle, ballBlend beyond the point where that circle touches the ball, with its
// direction and its curvature; from there the outer edge is the arc of the
// outer circle back to S.  The outer edge passes from the stalk into the ball
// without a corner and without a jump of the curvature, and the far side of the
// ball continues it.  The ball is a circle drawn as cubic arcs with the handles
// of a circular arc, which have their nodes at the extremes of the circle in the
// glyph and where another edge joins it, except an extreme within arcMinPiece of
// such a join which the arc passes by less than arcMaxOvershoot; its centre lies
// on the integer grid where the letter allows it (snapped).
//
// The constants.  The ball has the diameter BallDiameter (98, 1.21 times the stem
// of 81) and the stalk the thickness BallStalk (34, 0.42 stem) at the section;
// both are design parameters.  The shape constants of the part are ballCurl (the
// radius of the outer edge of the curl, 2.2 times the radius of the ball),
// ballSection (10 degrees) and ballBlend (25 degrees).  They give the angle round
// the ball over which the stalk is attached (from A to W, seen from O: 99
// degrees; ballPart.attach) and the opening of the wedge at its tip, the angle
// between the two circles there (53 degrees).
//
// The contour.  In the frame of the part the contour runs along the inner edge
// towards the ball, from S to W, round the ball counter-clockwise from W to A,
// and back along the outer edge from A to S.  A letter whose stalk runs the
// other way round (a) takes the mirror image, which mapChain reverses; the
// contour keeps its direction.

// ballCurl is the radius of the outer edge of the stalk at its section, as a
// multiple of the radius of the ball.
const ballCurl = 2.2

// ballSection is the angle, seen from the centre of curl, from the tip of the
// wedge back along the stalk to the section where the part begins (radians).
const ballSection = 10 * math.Pi / 180

// A ballPart is the ball terminal in its own frame.
type ballPart struct {
	R, T   float64       // the radius of the ball and the thickness of the stalk
	Ro, Ri float64       // the radii of the outer and the inner edge of the stalk at the section
	C      outline.Point // the centre of curl
	W      outline.Point // the tip of the wedge
	PhiW   float64       // the direction of W from C
	PhiS   float64       // the direction of the section from C
	BallW  float64       // the direction of W from the centre of the ball
	PhiT   float64       // the direction from C of the point where the outer edge leaves its circle
	Alpha  float64       // the direction from the centre of the ball of A, where the outer edge runs into the ball
	Blend  cubic         // the outer edge from A to its circle
}

// newBallPart returns the ball terminal of the parameters.  It panics if the
// stalk does not fit the ball.
func newBallPart(p design.Params) ballPart {
	b, err := makeBallPart(p.BallDiameter/2, p.BallStalk, ballCurl, ballSection, ballBlend)
	if err != nil {
		panic(fmt.Sprintf("glyphs: %v; glyphs.Check reports such parameters", err))
	}
	return b
}

// ballPartKey holds every input of makeBallPart.
type ballPartKey struct{ r, t, curl, section, blend float64 }

// ballPartResult is what makeBallPart returns.
type ballPartResult struct {
	part ballPart
	err  error
}

// ballParts holds the results of makeBallPart, keyed by its arguments.
var ballParts sync.Map // ballPartKey -> ballPartResult

// makeBallPart returns the ball terminal for the radius r of the ball, the
// thickness t of the stalk and the shape constants curl and section (see ballCurl
// and ballSection).
func makeBallPart(r, t, curl, section, blend float64) (ballPart, error) {
	k := ballPartKey{r, t, curl, section, blend}
	if v, ok := ballParts.Load(k); ok {
		res := v.(ballPartResult)
		return res.part, res.err
	}
	part, err := computeBallPart(r, t, curl, section, blend)
	ballParts.Store(k, ballPartResult{part, err})
	return part, err
}

// computeBallPart is makeBallPart without the memory.
func computeBallPart(r, t, curl, section, blend float64) (ballPart, error) {
	b := ballPart{R: r, T: t, Ro: curl * r}
	b.Ri = b.Ro - t
	d := b.Ro - r // the distance of C from O
	if !(t > 0 && b.Ri > 0 && t < r && d > 0) {
		return b, fmt.Errorf("ball terminal: a stalk of %v does not fit a ball of the radius %v", t, r)
	}
	b.C = outline.Point{X: -d}
	// W lies on both circles: |W - C| = Ri and |W| = r.
	x := (b.Ri*b.Ri + d*d - r*r) / (2 * d) // the abscissa of W relative to C, towards O
	y := b.Ri*b.Ri - x*x
	if !(y > 0) {
		return b, fmt.Errorf("ball terminal: the inner edge of the stalk does not meet the ball")
	}
	b.W = outline.Point{X: b.C.X + x, Y: math.Sqrt(y)}
	b.PhiW = math.Atan2(b.W.Y-b.C.Y, b.W.X-b.C.X)
	b.PhiS = b.PhiW + section
	b.BallW = math.Atan2(b.W.Y, b.W.X)
	return b.blend(blend)
}

// ballBlend is the angle, seen from the centre of curl, over which the outer edge
// of the stalk leaves its circle before it reaches the ball (radians).
var ballBlend = 25 * math.Pi / 180

// blend fills in the curve by which the outer edge passes from its circle into the
// circle of the ball: it leaves the outer circle at the angle blend after the
// point where that circle touches the ball, seen from C, and runs into the ball
// at A with the direction and the curvature of the ball.  A is the point, from
// 60 degrees before the point of touch to it in steps of a quarter of a degree,
// for which the curve turns one way, its curvature rises all the way, and it is
// fairest (bending).
func (b ballPart) blend(blend float64) (ballPart, error) {
	b.PhiT = blend
	q := outline.Point{X: b.C.X + b.Ro*math.Cos(blend), Y: b.C.Y + b.Ro*math.Sin(blend)}
	tq := outline.Point{X: math.Sin(blend), Y: -math.Cos(blend)} // walked from the stalk to the ball
	best := math.Inf(1)
	for deg := 0.0; deg >= -60; deg -= 0.25 {
		al := deg * math.Pi / 180
		a := outline.Point{X: b.R * math.Cos(al), Y: b.R * math.Sin(al)}
		ta := outline.Point{X: math.Sin(al), Y: -math.Cos(al)}
		c, err := g2Cubic(q, tq, -1/b.Ro, a, ta, -1/b.R)
		if err != nil || !oneWay(c, -1) || !curvatureRises(c) {
			continue
		}
		if e := bending(c); e < best {
			best, b.Alpha = e, al
			b.Blend = cubic{c.P3, c.P2, c.P1, c.P0} // from A to the circle, along the contour
		}
	}
	if math.IsInf(best, 1) {
		return b, fmt.Errorf("ball terminal: the outer edge of the stalk cannot run into the ball")
	}
	return b, nil
}

// A windSpiral is the outer edge by which the figures and c, f and r lead a
// stroke into the ball, in place of the curl and the blend of the part: a spiral which leaves
// P0 in the direction Phi0 (radians), turns clockwise, and whose curvature rises
// from K0 to K1, the curvature of the ball, with the power M of the length run,
//
//	k(s) = K0 + (K1 - K0) (s/L)^M,
//
// and ends on the circle of the ball with its direction and its curvature.
type windSpiral struct {
	P0     outline.Point
	Phi0   float64
	K0, K1 float64
	L, M   float64
}

// newWindSpiral returns the spiral from p0, leaving in the direction phi0 with
// the curvature k0 (positive, turning clockwise), which ends on the circle about
// centre of the radius r with its direction and its curvature: the length L and
// the power M are the solution of the two equations which put the centre of
// curvature at the end of the spiral at centre, found by Newton's method.  It
// reports an error if there is none with M at least 1.
func newWindSpiral(p0 outline.Point, phi0, k0 float64, centre outline.Point, r float64) (windSpiral, error) {
	w := windSpiral{P0: p0, Phi0: phi0, K0: k0, K1: 1 / r}
	miss := func(L, m float64) (float64, float64) {
		v := w
		v.L, v.M = L, m
		c := v.centreAt(L)
		return c.X - centre.X, c.Y - centre.Y
	}
	L, m := 1.5*math.Hypot(centre.X-p0.X, centre.Y-p0.Y), 2.0
	for i := 0; i < 200; i++ {
		fx, fy := miss(L, m)
		if math.Hypot(fx, fy) < 1e-7 {
			w.L, w.M = L, m
			if !(m >= 0.5) {
				return w, fmt.Errorf("the spiral into the ball needs the power %.2f", m)
			}
			return w, nil
		}
		const h = 1e-5
		ax, ay := miss(L+h, m)
		bx, by := miss(L, m+h)
		j11, j21, j12, j22 := (ax-fx)/h, (ay-fy)/h, (bx-fx)/h, (by-fy)/h
		det := j11*j22 - j12*j21
		if det == 0 || math.IsNaN(det) {
			break
		}
		dL, dm := (fx*j22-fy*j12)/det, (j11*fy-j21*fx)/det
		for math.Abs(dL) > 0.2*L || math.Abs(dm) > 0.5 {
			dL, dm = dL/2, dm/2
		}
		L, m = L-dL, math.Max(m-dm, 0.1)
	}
	return w, fmt.Errorf("no spiral from %v runs into the ball about %v", p0, centre)
}

// heading returns the direction of the spiral after the length s.
func (w windSpiral) heading(s float64) float64 {
	return w.Phi0 - w.K0*s - (w.K1-w.K0)*w.L/(w.M+1)*math.Pow(s/w.L, w.M+1)
}

// curvature returns the curvature of the spiral after the length s (positive).
func (w windSpiral) curvature(s float64) float64 {
	return w.K0 + (w.K1-w.K0)*math.Pow(s/w.L, w.M)
}

// point returns the point of the spiral after the length s, by Simpson's rule.
func (w windSpiral) point(s float64) outline.Point {
	const n = 400
	h := s / n
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

// centreAt returns the centre of curvature of the spiral after the length s.
func (w windSpiral) centreAt(s float64) outline.Point {
	q, phi, k := w.point(s), w.heading(s), w.curvature(s)
	return outline.Point{X: q.X + math.Sin(phi)/k, Y: q.Y - math.Cos(phi)/k}
}

// lengthAt returns the length after which the spiral has the direction phi,
// which must lie between its directions at its ends.
func (w windSpiral) lengthAt(phi float64) float64 {
	lo, hi := 0.0, w.L
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		if w.heading(mid) > phi {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// cubics returns the spiral as cubic curves walked from its end on the ball back
// to P0, as a contour which runs round the ball counter-clockwise and then out
// along the spiral takes it.  The nodes are the ends and the extremes, where the
// direction is a multiple of a quarter turn, except an extreme which lies within
// windMinPiece of P0; each curve has the points, the directions and the
// curvatures of the spiral at its ends (g2Cubic).
func (w windSpiral) cubics() (chain, error) {
	cuts := []float64{0}
	end := w.heading(w.L)
	for q := math.Floor(w.Phi0/(math.Pi/2)) * math.Pi / 2; q > end; q -= math.Pi / 2 {
		if q >= w.Phi0 {
			continue
		}
		if s := w.lengthAt(q); s > windMinPiece && w.L-s > 1 {
			cuts = append(cuts, s)
		}
	}
	cuts = append(cuts, w.L)
	var ch chain
	for i := len(cuts) - 1; i > 0; i-- {
		a, b := cuts[i], cuts[i-1]
		pa, pb := w.point(a), w.point(b)
		// Walked from a back to b the spiral turns counter-clockwise.
		ta := dirOf(w.heading(a) + math.Pi)
		tb := dirOf(w.heading(b) + math.Pi)
		c, err := g2Cubic(pa, ta, w.curvature(a), pb, tb, w.curvature(b))
		if err != nil {
			return nil, err
		}
		ch = append(ch, curveSeg(c))
	}
	return ch, nil
}

// windBall returns the circle of the ball of tb, counter-clockwise from the tip
// of the wedge, and the outer edge which winds into it, walked back out to q0:
// the spiral (windSpiral) which leaves q0 in the direction phi0 (radians) with
// the curvature k0 and ends on the circle of the ball with its direction and its
// curvature.
func windBall(name string, tb placedBall, q0 outline.Point, phi0, k0 float64) chain {
	w, err := newWindSpiral(q0, phi0, k0, tb.Centre, tb.Part.R)
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: %v; glyphs.Check reports such parameters", name, err))
	}
	wind, err := w.cubics()
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: the spiral into the ball: %v; glyphs.Check reports such parameters", name, err))
	}
	aW := math.Atan2(tb.Tip.Y-tb.Centre.Y, tb.Tip.X-tb.Centre.X)
	aA := w.heading(w.L) + math.Pi/2
	for aA <= aW {
		aA += 2 * math.Pi
	}
	for aA > aW+2*math.Pi {
		aA -= 2 * math.Pi
	}
	ball := circleArc(tb.Centre, tb.Part.R, aW, aA, 0)
	ball[len(ball)-1].P3 = wind.start()
	return join(ball, wind)
}

// windMinPiece is the shortest arc of the spiral from P0 which is drawn as a
// curve of its own: an extreme nearer to P0 than this is left between the nodes.
const windMinPiece = 20.0

// curvatureRises reports whether the absolute curvature of the curve grows from
// its start to its end and stays between its values there, both within 2 per cent
// of the difference of the two.
func curvatureRises(c cubic) bool {
	k0, k1 := math.Abs(c.curvature(0)), math.Abs(c.curvature(1))
	return curvatureRisesWithin(c, 0.02*(k1-k0))
}

// curvatureRisesWithin is curvatureRises with the overshoot of the curvature which
// is allowed given as slack.
func curvatureRisesWithin(c cubic, slack float64) bool {
	k0, k1 := math.Abs(c.curvature(0)), math.Abs(c.curvature(1))
	low, high := k0, k0
	for i := 1; i <= 200; i++ {
		k := math.Abs(c.curvature(float64(i) / 200))
		if k < high-slack || k > k1+slack {
			return false
		}
		low, high = math.Min(low, k), math.Max(high, k)
	}
	return low > k0-slack
}

// attach returns the angle round the ball, seen from its centre, over which the
// stalk is attached: from A, where the outer edge runs into the circle, to the
// tip of the wedge.
func (b ballPart) attach() float64 { return b.BallW - b.Alpha }

// wedge returns the opening of the wedge at its tip, in degrees: the angle
// between the inner circle of the stalk and the circle of the ball there.
func (b ballPart) wedge() float64 {
	in := math.Atan2(b.W.Y-b.C.Y, b.W.X-b.C.X) // the direction of the inner circle's radius at W
	ball := math.Atan2(b.W.Y, b.W.X)           // that of the ball's
	return 180 - (math.Pi-math.Abs(in-ball))*180/math.Pi
}

// arcPieces returns the arc of the circle about c of the radius r from the
// direction a0 to the direction a1 (radians; counter-clockwise if a1 > a0) as
// cubic curves of at most a quarter turn each.
func arcPieces(c outline.Point, r, a0, a1 float64) chain {
	n := int(math.Ceil(math.Abs(a1-a0)/(math.Pi/2) - 1e-9))
	if n < 1 {
		n = 1
	}
	on := func(a float64) outline.Point { return outline.Point{X: c.X + r*math.Cos(a), Y: c.Y + r*math.Sin(a)} }
	var ch chain
	step := (a1 - a0) / float64(n)
	k := 4.0 / 3 * math.Tan(step/4) * r
	for i := 0; i < n; i++ {
		a, e := a0+float64(i)*step, a0+float64(i+1)*step
		pa, pe := on(a), on(e)
		ch = append(ch, curveSeg(cubic{pa,
			outline.Point{X: pa.X - k*math.Sin(a), Y: pa.Y + k*math.Cos(a)},
			outline.Point{X: pe.X + k*math.Sin(e), Y: pe.Y - k*math.Cos(e)},
			pe}))
	}
	return ch
}

// partsInFrame returns the part in its own frame in three chains: the inner
// edge from the section to the tip of the wedge, the circle of the ball, cut at
// the directions phase plus a multiple of a quarter turn (circleArc), and the
// outer edge back to the section.
func (b ballPart) partsInFrame(phase float64) (in, ball, rest chain) {
	in = arcPieces(b.C, b.Ri, b.PhiS, b.PhiW)
	ball = circleArc(outline.Point{}, b.R, b.BallW, b.Alpha+2*math.Pi, phase)
	rest = join(chain{curveSeg(b.Blend)}, arcPieces(b.C, b.Ro, b.PhiT, b.PhiS))
	return in, ball, rest
}

// circleArc returns the arc of the circle about c of the radius r from the
// direction a0 counter-clockwise to the direction a1 > a0 (radians), cut at the
// directions phase + k/4 turn for integers k.  A cut whose arc to an end of the
// arc is shorter than arcMinPiece and which lies less than arcMaxOvershoot
// beyond that end is left out.  Each piece has the handles of a circular arc.
func circleArc(c outline.Point, r, a0, a1, phase float64) chain {
	keep := func(a, end float64) bool {
		d := math.Abs(a - end)
		return r*d >= arcMinPiece || r*(1-math.Cos(d)) >= arcMaxOvershoot
	}
	cuts := []float64{a0}
	for k := math.Floor((a0-phase)/(math.Pi/2)) + 1; ; k++ {
		a := phase + k*math.Pi/2
		if a >= a1-1e-9 {
			break
		}
		if a > a0+1e-9 && keep(a, a0) && keep(a, a1) {
			cuts = append(cuts, a)
		}
	}
	cuts = append(cuts, a1)
	var ch chain
	for i := 1; i < len(cuts); i++ {
		ch = append(ch, curveSeg(arcCubic(c, r, cuts[i-1], cuts[i])))
	}
	return ch
}

// arcCubic returns the arc of the circle about c of the radius r from the
// direction a0 to the direction a1 (radians) as one cubic curve.
func arcCubic(c outline.Point, r, a0, a1 float64) cubic {
	on := func(a float64) outline.Point { return outline.Point{X: c.X + r*math.Cos(a), Y: c.Y + r*math.Sin(a)} }
	k := 4.0 / 3 * math.Tan((a1-a0)/4) * r
	pa, pe := on(a0), on(a1)
	return cubic{pa,
		outline.Point{X: pa.X - k*math.Sin(a0), Y: pa.Y + k*math.Cos(a0)},
		outline.Point{X: pe.X + k*math.Sin(a1), Y: pe.Y - k*math.Cos(a1)},
		pe}
}

// A ballPlace says where a letter puts the part: the centre of the ball, the angle
// by which the part is turned (radians, counter-clockwise), and whether it is
// mirrored in the vertical through the centre before it is turned.
type ballPlace struct {
	Centre outline.Point
	Angle  float64
	Mirror bool
}

// point maps a point of the frame of the part to the glyph.
func (pl ballPlace) point(q outline.Point) outline.Point {
	if pl.Mirror {
		q.X = -q.X
	}
	c, s := math.Cos(pl.Angle), math.Sin(pl.Angle)
	return outline.Point{X: pl.Centre.X + c*q.X - s*q.Y, Y: pl.Centre.Y + s*q.X + c*q.Y}
}

// mapChain maps a chain of the frame of the part to the glyph, reversed if the
// place mirrors it.
func (pl ballPlace) mapChain(c chain) chain {
	m := c.mapPoints(pl.point)
	if pl.Mirror {
		m = m.reversed()
	}
	return m
}

// A stalkEnd is the place where a letter's stalk meets the part: a point of the
// contour, the unit vector in which the contour runs there and its signed
// curvature (positive where the contour turns to the left).
type stalkEnd struct {
	At, Dir outline.Point
	Kappa   float64
}

// A placedBall is the part placed in a letter: its chain in the glyph, from the
// point where the contour enters it to the point where it leaves, and the two
// ends, which the letter's edges must meet.
type placedBall struct {
	Part    ballPart
	Place   ballPlace
	Chain   chain
	In, Out stalkEnd
	Centre  outline.Point
	Tip     outline.Point // the tip of the wedge
	Tangent outline.Point // the point A, where the outer edge runs into the circle
}

// place puts the part into a letter.  The edges of the stalk are cut at their
// extremes, which are on-curve points, and the circle of the ball is cut at the
// extremes of the glyph (circleArc).
func (b ballPart) place(pl ballPlace) placedBall {
	phase := -pl.Angle
	if pl.Mirror {
		phase = pl.Angle
	}
	in, ball, rest := b.partsInFrame(phase)
	split := func(c chain) chain {
		var out chain
		for _, pc := range pl.mapChain(c) {
			out = append(out, splitAtExtrema(pc.cubic)...)
		}
		return out
	}
	parts := []chain{split(in), pl.mapChain(ball), split(rest)}
	if pl.Mirror {
		parts[0], parts[2] = parts[2], parts[0]
	}
	ch := join(parts...)
	raw := pl.mapChain(join(in, ball, rest))
	first, last := raw[0].cubic, raw[len(raw)-1].cubic
	return placedBall{
		Part: b, Place: pl, Chain: ch,
		In:      stalkEnd{At: first.P0, Dir: first.tangent(0), Kappa: first.curvature(0)},
		Out:     stalkEnd{At: last.P3, Dir: last.tangent(1), Kappa: last.curvature(1)},
		Centre:  pl.Centre,
		Tip:     pl.point(b.W),
		Tangent: pl.point(outline.Point{X: b.R * math.Cos(b.Alpha), Y: b.R * math.Sin(b.Alpha)}),
	}
}

// g2Cubic returns the cubic curve from p0, leaving in the direction of the unit
// vector t0 with the signed curvature k0, to p3, arriving in the direction t3
// with the curvature k3.  The handles a (at p0) and b (at p3) solve
//
//	3/2 k0 a^2 = t0 x D - b t0 x t3,   3/2 k3 b^2 = D x t3 - a t0 x t3,
//
// where D = p3 - p0 and x is the cross product, which Newton's method finds from
// the handles of a circular arc.  It reports an error if there is no solution with
// positive handles shorter than the chord.
func g2Cubic(p0, t0 outline.Point, k0 float64, p3, t3 outline.Point, k3 float64) (cubic, error) {
	cr := func(u, v outline.Point) float64 { return u.X*v.Y - u.Y*v.X }
	d := outline.Point{X: p3.X - p0.X, Y: p3.Y - p0.Y}
	chord := math.Hypot(d.X, d.Y)
	s, c1, c2 := cr(t0, t3), cr(t0, d), cr(d, t3)
	// For a handle a at p0 the second equation gives b; the first is then a function
	// of a alone, whose root nearest the handle of a circular arc is taken.
	bOf := func(a float64) float64 {
		v := 2.0 / 3 * (c2 - a*s) / k3
		if !(v > 0) {
			return math.NaN()
		}
		return math.Sqrt(v)
	}
	f := func(a float64) float64 { return 1.5*k0*a*a - c1 + bOf(a)*s }
	h := arcHandle(p0, t0, p3, t3)
	best, bestDist := math.NaN(), math.Inf(1)
	const n = 400
	for i := 0; i < n; i++ {
		lo, hi := chord*float64(i)/n, chord*float64(i+1)/n
		flo, fhi := f(lo), f(hi)
		if math.IsNaN(flo) || math.IsNaN(fhi) || (flo > 0) == (fhi > 0) {
			continue
		}
		for k := 0; k < 60; k++ {
			mid := (lo + hi) / 2
			if fm := f(mid); (fm > 0) == (flo > 0) {
				lo, flo = mid, fm
			} else {
				hi = mid
			}
		}
		if r := (lo + hi) / 2; math.Abs(r-h) < bestDist {
			best, bestDist = r, math.Abs(r-h)
		}
	}
	a := best
	bb := bOf(a)
	if !(a > 0 && bb > 0 && bb < chord) {
		return cubic{}, fmt.Errorf("no cubic curve from %v to %v has the curvatures %v and %v at its ends", p0, p3, k0, k3)
	}
	return cubic{p0,
		outline.Point{X: p0.X + a*t0.X, Y: p0.Y + a*t0.Y},
		outline.Point{X: p3.X - bb*t3.X, Y: p3.Y - bb*t3.Y},
		p3}, nil
}

// g2End returns the cubic curve from p0, leaving in the direction t0 with the
// handle a, to p3, arriving in the direction t3 with the signed curvature k3.  It
// reports an error if there is no such curve with a positive handle at p3 shorter
// than the chord.
func g2End(p0, t0 outline.Point, a float64, p3, t3 outline.Point, k3 float64) (cubic, error) {
	cr := func(u, v outline.Point) float64 { return u.X*v.Y - u.Y*v.X }
	d := outline.Point{X: p3.X - p0.X, Y: p3.Y - p0.Y}
	num := 2.0 / 3 * (cr(d, t3) - a*cr(t0, t3)) / k3
	if !(num > 0) || !(a > 0) {
		return cubic{}, fmt.Errorf("no cubic curve from %v with the handle %v arrives at %v with the curvature %v", p0, a, p3, k3)
	}
	b := math.Sqrt(num)
	if b >= math.Hypot(d.X, d.Y) {
		return cubic{}, fmt.Errorf("the curve from %v to %v needs a handle of %v at its end", p0, p3, b)
	}
	return cubic{p0,
		outline.Point{X: p0.X + a*t0.X, Y: p0.Y + a*t0.Y},
		outline.Point{X: p3.X - b*t3.X, Y: p3.Y - b*t3.Y},
		p3}, nil
}

// g2Start is g2End walked the other way: the curve from p0, leaving in the
// direction t0 with the curvature k0, to p3, arriving in the direction t3 with
// the handle b.
func g2Start(p0, t0 outline.Point, k0 float64, p3, t3 outline.Point, b float64) (cubic, error) {
	c, err := g2End(p3, outline.Point{X: -t3.X, Y: -t3.Y}, b, p0, outline.Point{X: -t0.X, Y: -t0.Y}, -k0)
	if err != nil {
		return c, err
	}
	return cubic{c.P3, c.P2, c.P1, c.P0}, nil
}

// arcHandle returns the length of the handles of a circular arc from p0, leaving
// in the direction t0, to p3, arriving in the direction t3.
func arcHandle(p0, t0, p3, t3 outline.Point) float64 {
	chord := math.Hypot(p3.X-p0.X, p3.Y-p0.Y)
	theta := math.Abs(math.Atan2(t0.X*t3.Y-t0.Y*t3.X, t0.X*t3.X+t0.Y*t3.Y))
	if theta < 1e-6 {
		return chord / 3
	}
	return 4.0 / 3 * math.Tan(theta/4) * chord / (2 * math.Sin(theta/2))
}

// dirOf returns the unit vector of the direction a (radians).
func dirOf(a float64) outline.Point { return outline.Point{X: math.Cos(a), Y: math.Sin(a)} }

// neg returns the vector q reversed.
func neg(q outline.Point) outline.Point { return outline.Point{X: -q.X, Y: -q.Y} }

// bending returns the fairness measure of a curve: the integral of the square of
// the rate of change of its curvature along the arc length, from 200 samples.  A
// circular arc has zero, and a curve whose curvature changes evenly has less than
// one which has the same change in a short stretch.
func bending(c cubic) float64 {
	const n = 200
	var e float64
	prevK := c.curvature(0)
	prev := c.at(0)
	for i := 1; i <= n; i++ {
		t := float64(i) / n
		q, k := c.at(t), c.curvature(t)
		ds := math.Hypot(q.X-prev.X, q.Y-prev.Y)
		if ds > 0 {
			e += (k - prevK) * (k - prevK) / ds
		}
		prev, prevK = q, k
	}
	return e
}

// fairEnd returns the curve of g2End whose handle at p0 makes it fairest
// (bending), among the handles from 0.05 to 0.9 times the chord.
func fairEnd(p0, t0, p3, t3 outline.Point, k3 float64) (cubic, error) {
	return fairEndRising(p0, t0, p3, t3, k3, false)
}

// fairEndRising is fairEnd; if rising is set, only curves whose curvature grows from
// the start to the end without a hump (curvatureRises) are considered.  If there
// is none, it is fairEnd.
func fairEndRising(p0, t0, p3, t3 outline.Point, k3 float64, rising bool) (cubic, error) {
	if rising {
		if c, err := fairEndWith(p0, t0, p3, t3, k3, true); err == nil {
			return c, nil
		}
	}
	return fairEndWith(p0, t0, p3, t3, k3, false)
}

func fairEndWith(p0, t0, p3, t3 outline.Point, k3 float64, rising bool) (cubic, error) {
	chord := math.Hypot(p3.X-p0.X, p3.Y-p0.Y)
	best, bestE := cubic{}, math.Inf(1)
	var lastErr error
	for i := 0; i <= 170; i++ {
		a := (0.05 + 0.005*float64(i)) * chord
		c, err := g2End(p0, t0, a, p3, t3, k3)
		if err == nil && !oneWay(c, math.Copysign(1, k3)) {
			err = fmt.Errorf("the curve from %v to %v with the handle %v turns both ways", p0, p3, a)
		}
		if err == nil && rising && !curvatureRisesWithin(c, 0.06*math.Abs(k3)) {
			err = fmt.Errorf("the curve from %v to %v with the handle %v has a hump of the curvature", p0, p3, a)
		}
		if err != nil {
			lastErr = err
			continue
		}
		if e := bending(c); e < bestE {
			best, bestE = c, e
		}
	}
	if math.IsInf(bestE, 1) {
		return best, lastErr
	}
	return best, nil
}

// fairStart is fairEnd walked the other way: the curve from p0, leaving in the
// direction t0 with the curvature k0, to p3, arriving in the direction t3.
func fairStart(p0, t0 outline.Point, k0 float64, p3, t3 outline.Point) (cubic, error) {
	c, err := fairEnd(p3, neg(t3), p0, neg(t0), -k0)
	return cubic{c.P3, c.P2, c.P1, c.P0}, err
}

// A stalkFoot is where a letter's edge of a stalk starts or ends away from the
// ball: a point, the unit vector in which the contour runs there, and its signed
// curvature, or NaN where the edge leaves a stem or a stroke at a corner or the
// curvature is free.
type stalkFoot struct {
	At, Dir outline.Point
	Kappa   float64
	Rising  bool // the curvature of the edge, if free, grows without a hump up to the part
}

// footRising returns the foot at q in the direction a (radians) with a free
// curvature, which grows without a hump up to the part (j and y).
func footRising(q outline.Point, a float64) stalkFoot {
	return stalkFoot{At: q, Dir: dirOf(a), Kappa: math.NaN(), Rising: true}
}

// footAt returns the foot at q in the direction a (radians) with a free curvature.
func footAt(q outline.Point, a float64) stalkFoot {
	return stalkFoot{At: q, Dir: dirOf(a), Kappa: math.NaN()}
}

// edgeIn returns the edge of the stalk from the foot f to the point where the
// contour enters the part: one cubic curve which arrives with the direction and
// the curvature of the part, and leaves the foot with its curvature, or, if that
// is free, as fair as it can be (fairEnd).
func edgeIn(f stalkFoot, e stalkEnd) (cubic, error) {
	if f.Rising {
		// The edge of a stem or a diagonal begins straight; if no such curve
		// reaches the part, the fairest curve whose curvature rises
		// (fairEndRising).
		if c, err := g2Cubic(f.At, f.Dir, 0, e.At, e.Dir, e.Kappa); err == nil && oneWay(c, math.Copysign(1, e.Kappa)) && curvatureRisesWithin(c, 0.06*math.Abs(e.Kappa)) {
			return c, nil
		}
	}
	if math.IsNaN(f.Kappa) {
		return fairEndRising(f.At, f.Dir, e.At, e.Dir, e.Kappa, f.Rising)
	}
	c, err := g2Cubic(f.At, f.Dir, f.Kappa, e.At, e.Dir, e.Kappa)
	if err == nil && !oneWay(c, math.Copysign(1, e.Kappa)) {
		err = fmt.Errorf("the edge from %v to %v turns both ways", f.At, e.At)
	}
	return c, err
}

// edgeOut returns the edge of the stalk from the point where the contour leaves
// the part to the foot f, see edgeIn.
func edgeOut(e stalkEnd, f stalkFoot) (cubic, error) {
	if math.IsNaN(f.Kappa) {
		return fairStart(e.At, e.Dir, e.Kappa, f.At, f.Dir)
	}
	c, err := g2Cubic(e.At, e.Dir, e.Kappa, f.At, f.Dir, f.Kappa)
	if err == nil && !oneWay(c, math.Copysign(1, e.Kappa)) {
		err = fmt.Errorf("the edge from %v to %v turns both ways", e.At, f.At)
	}
	return c, err
}

// stalkChain returns the stalk and the ball of a letter as one chain: the edge from
// the foot in to the part, the part, and the edge from the part to the foot out.
// It panics with the letter's name if an edge cannot be drawn; Check reports it.
func stalkChain(letter string, in stalkFoot, tb placedBall, out stalkFoot) chain {
	a, err := edgeIn(in, tb.In)
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: the edge into the ball terminal: %v", letter, err))
	}
	b, err := edgeOut(tb.Out, out)
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: the edge out of the ball terminal: %v", letter, err))
	}
	return join(splitAtExtrema(a), tb.Chain, splitAtExtrema(b))
}

// windHook returns the stalk and the ball of a letter whose stroke leaves a stem
// (f, r) as one chain, from the foot in round the ball to the foot out.  The
// inner edge is one cubic curve from the foot in to the tip of the wedge, which
// arrives with the direction and the curvature of the inner edge of the part
// (tipEnd); where the inner circle of the part has an extreme within
// windMinPiece of the tip, the curve ends at the extreme and the circle runs on
// to the tip.  The outer edge is the curve which stalkChain draws from the part
// to the foot out, from the foot as far as the point where, walked towards the
// ball, it rises by an angle within rises (degrees) above the horizontal; from
// there it winds into the ball (windBall).  The angle is the largest in the range
// for which the spiral meets the circle of the ball at one of its extremes, a
// node which the circle has anyway.  A foot whose curvature is zero, on a
// straight stem, is reached by a transition curve (straightFoot) over the last
// stemRamp degrees of the turn of the edge.
func windHook(letter string, in stalkFoot, tb placedBall, out stalkFoot, rises [2]float64) chain {
	tip, arc := tipEnd(tb), chain(nil)
	cg := tb.Place.point(tb.Part.C)
	aW := math.Atan2(tb.Tip.Y-cg.Y, tb.Tip.X-cg.X)
	if e := math.Ceil(aW/(math.Pi/2)) * math.Pi / 2; (e-aW)*tb.Part.Ri < windMinPiece {
		// The inner circle of the part has an extreme just before the tip: the
		// edge runs to the extreme and on along the circle.
		arc = chain{curveSeg(arcCubic(cg, tb.Part.Ri, e, aW))}
		tip = stalkEnd{At: arc.start(), Dir: dirOf(e - math.Pi/2), Kappa: -1 / tb.Part.Ri}
	}
	free := func(f stalkFoot) stalkFoot { return stalkFoot{At: f.At, Dir: f.Dir, Kappa: math.NaN()} }
	a, err := edgeIn(free(in), tip)
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: the edge into the ball terminal: %v", letter, err))
	}
	inner := splitAtExtrema(a)
	if in.Kappa == 0 {
		u := tTurned(a, 0, stemRamp)
		ramp := straightFoot(a.at(u), neg(a.tangent(u)), -a.curvature(u), in.At, neg(in.Dir)).reversed()
		inner = join(chain{curveSeg(ramp)}, splitAtExtrema(a.sub(u, 1)))
	}
	inner = join(inner, arc)
	outer, err := edgeOut(tb.Out, free(out))
	if err != nil {
		panic(fmt.Sprintf("glyphs: glyph %s: the edge out of the ball terminal: %v", letter, err))
	}
	spiral := func(rise float64) (windSpiral, float64, error) {
		t := tAtFall(outer, -rise)
		q := outer.tangent(t)
		w, err := newWindSpiral(outer.at(t), math.Atan2(-q.Y, -q.X), outer.curvature(t), tb.Centre, tb.Part.R)
		return w, t, err
	}
	// The direction from the centre of the ball in which the spiral meets it,
	// less the nearest multiple of a quarter turn, in degrees.
	miss := func(rise float64) float64 {
		w, _, err := spiral(rise)
		if err != nil {
			return math.NaN()
		}
		d := (w.heading(w.L) + math.Pi/2) * 180 / math.Pi
		return d - 90*math.Round(d/90)
	}
	lo, hi, found := 0.0, 0.0, false
	b := miss(rises[1])
	for r := rises[1]; r > rises[0]; r-- {
		a := miss(r - 1)
		if math.Abs(a) < 30 && math.Abs(b) < 30 && (a <= 0) != (b <= 0) {
			lo, hi, found = r-1, r, true
			break
		}
		b = a
	}
	if !found {
		panic(fmt.Sprintf("glyphs: glyph %s: the spiral into the ball does not meet it at an extreme; glyphs.Check reports such parameters", letter))
	}
	mlo := miss(lo)
	for i := 0; i < 30; i++ {
		m := (lo + hi) / 2
		if mm := miss(m); (mm <= 0) == (mlo <= 0) {
			lo, mlo = m, mm
		} else {
			hi = m
		}
	}
	_, t, _ := spiral((lo + hi) / 2)
	q := outer.tangent(t)
	wound := windBall(letter, tb, outer.at(t), math.Atan2(-q.Y, -q.X), outer.curvature(t))
	rest := outer.sub(t, 1)
	tail := splitAtExtrema(rest)
	if out.Kappa == 0 {
		u := tTurned(rest, 1, stemRamp)
		ramp := straightFoot(rest.at(u), rest.tangent(u), rest.curvature(u), out.At, out.Dir)
		tail = join(splitAtExtrema(rest.sub(0, u)), chain{curveSeg(ramp)})
	}
	return join(inner, wound, tail)
}

// stemRamp is the turn (degrees) over which the edge of a hook which leaves a
// straight stem without curvature (windHook) is a transition curve
// (straightFoot).
const stemRamp = 20.0

// tTurned returns the parameter of the point of c whose direction differs by
// turn degrees from the direction at the parameter end (0 or 1); c turns one way.
func tTurned(c cubic, end, turn float64) float64 {
	t0 := c.tangent(end)
	diff := func(t float64) float64 {
		q := c.tangent(t)
		return math.Abs(math.Atan2(t0.X*q.Y-t0.Y*q.X, t0.X*q.X+t0.Y*q.Y)) * 180 / math.Pi
	}
	lo, hi := 0.0, 1.0
	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		if (diff(mid) < turn) == (end == 0) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// straightFoot returns the cubic curve from p0, leaving in the direction t0
// with the signed curvature k0, to a point on the line through p3 in the
// direction t3, a straight stem, where it arrives in the direction t3 without
// curvature, and has handles of equal length.  The point is p3 + h t3, with h
// chosen so that the two handles are equal.
func straightFoot(p0, t0 outline.Point, k0 float64, p3, t3 outline.Point) cubic {
	cr := func(u, v outline.Point) float64 { return u.X*v.Y - u.Y*v.X }
	s := cr(t0, t3)
	// The curvature 0 at the end fixes the handle at the start, a = (d x t3)/s,
	// which does not depend on h, and the handle at the end is
	// b = (t0 x d - 3/2 k0 a^2)/s.
	d := outline.Point{X: p3.X - p0.X, Y: p3.Y - p0.Y}
	a := cr(d, t3) / s
	h := (a*s + 1.5*k0*a*a - cr(t0, d)) / s
	q := outline.Point{X: p3.X + h*t3.X, Y: p3.Y + h*t3.Y}
	return cubic{p0,
		outline.Point{X: p0.X + a*t0.X, Y: p0.Y + a*t0.Y},
		outline.Point{X: q.X - a*t3.X, Y: q.Y - a*t3.Y},
		q}
}

// chainBBox returns the bounding box of the chain, from its curves.
func chainBBox(c chain) (x0, y0, x1, y1 float64) {
	var path outline.Path
	path.MoveTo(c.start().X, c.start().Y)
	c.appendTo(&path)
	return path.BBox()
}

// placeUnderCrest places the part turned by angle so that its rightmost ink is at
// right and the highest point of the stroke which arm draws to it is at apex, see
// placeAtExtreme.
func placeUnderCrest(b ballPart, angle, right, apex float64, arm func(placedBall) chain) placedBall {
	return placeAtExtreme(b, ballPlace{Angle: angle}, func(_, x1 float64) float64 { return right - x1 }, apex, true, arm)
}

// placeAtExtreme places the part, turned and mirrored as pl says: dx returns the
// abscissa of the centre from the extent in x of the part placed at the origin,
// and the ball is moved up or down until the highest point (top set) or the lowest
// point of the stroke which arm draws to it is at the height y, which may lie in
// the edge the letter draws.
func placeAtExtreme(b ballPart, pl ballPlace, dx func(x0, x1 float64) float64, y float64, top bool, arm func(placedBall) chain) placedBall {
	pl.Centre = outline.Point{}
	x0, y0, x1, y1 := chainBBox(b.place(pl).Chain)
	pl.Centre = outline.Point{X: dx(x0, x1), Y: y - y0}
	if top {
		pl.Centre.Y = y - y1
	}
	tb := b.place(pl)
	for i := 0; i < 30; i++ {
		_, y0, _, y1 := chainBBox(arm(tb))
		ext := y0
		if top {
			ext = y1
		}
		if math.Abs(ext-y) < 1e-7 {
			break
		}
		pl.Centre.Y += y - ext
		tb = b.place(pl)
	}
	return tb.snapped()
}

// snapped returns the terminal with the centre of the ball moved to the nearest
// point with integer coordinates.  The radius of the ball is an integer, and the
// extremes of the ball are then on the grid.
func (tb placedBall) snapped() placedBall {
	pl := tb.Place
	pl.Centre = outline.Point{X: math.Round(pl.Centre.X), Y: math.Round(pl.Centre.Y)}
	return tb.Part.place(pl)
}

// oneWay reports whether the curve turns one way only, by less than half a turn,
// and its curvature has the sign sign where it is not almost straight.
func oneWay(c cubic, sign float64) bool {
	const n = 100
	var turn float64
	prev := c.tangent(0)
	for i := 0; i <= n; i++ {
		t := float64(i) / n
		if k := c.curvature(t); k*sign < -1e-5 {
			return false
		}
		if i > 0 {
			q := c.tangent(t)
			turn += math.Atan2(prev.X*q.Y-prev.Y*q.X, prev.X*q.X+prev.Y*q.Y)
			prev = q
		}
	}
	return math.Abs(turn) < math.Pi
}

// tipEnd returns the end of the inner edge of the part at the tip of the wedge
// of tb: the tip, and the direction and the curvature of the inner edge there.
func tipEnd(tb placedBall) stalkEnd {
	for _, pc := range tb.Chain {
		if isTip(pc.P3, tb.Tip) {
			return stalkEnd{At: tb.Tip, Dir: pc.tangent(1), Kappa: pc.curvature(1)}
		}
	}
	panic("glyphs: the ball terminal has no tip of the wedge")
}

// isTip reports whether q is the tip of the wedge w.
func isTip(q, w outline.Point) bool { return math.Hypot(q.X-w.X, q.Y-w.Y) < 1e-6 }

// arcMinPiece and arcMaxOvershoot are the limits of circleArc.
const (
	arcMinPiece     = 8.0
	arcMaxOvershoot = 0.3
)
