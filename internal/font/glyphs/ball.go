// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"sort"
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
// section.  The part is made in its own frame, in which the centre O of the ball
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
// ball continues it.  The ball is a circle drawn as cubic arcs of at most a
// quarter turn, cut at its extremes; its centre lies on the integer grid where
// the letter allows it (snapped).
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

// chainInFrame returns the part as a chain in its own frame: from the section on
// the inner edge to the section on the outer edge.
func (b ballPart) chainInFrame() chain {
	in := arcPieces(b.C, b.Ri, b.PhiS, b.PhiW)
	ball := arcPieces(outline.Point{}, b.R, b.BallW, b.Alpha+2*math.Pi)
	return join(in, ball, chain{curveSeg(b.Blend)}, arcPieces(b.C, b.Ro, b.PhiT, b.PhiS))
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

// place puts the part into a letter.  The chain is cut at its extremes, which
// are on-curve points, and the piece of the ball which starts at the tip of the
// wedge is cut by splitPart.
func (b ballPart) place(pl ballPlace) placedBall {
	raw := pl.mapChain(b.chainInFrame())
	var ch chain
	for _, pc := range raw {
		if isTip(pc.P0, pl.point(b.W)) {
			ch = append(ch, splitPart(pc.cubic)...)
		} else {
			ch = append(ch, splitAtExtrema(pc.cubic)...)
		}
	}
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

// splitPart returns the curve cut at its extremes in x and y, as splitAtExtrema
// does, except at an extreme which lies less than splitMinPiece from an end of the
// curve and beyond that end by less than splitMaxOvershoot: that extreme stays
// between the nodes.
func splitPart(c cubic) chain {
	var ts []float64
	for axis := 0; axis < 2; axis++ {
		get := func(q outline.Point) float64 {
			if axis == 0 {
				return q.X
			}
			return q.Y
		}
		for _, t := range outline.CubicExtrema(get(c.P0), get(c.P1), get(c.P2), get(c.P3)) {
			if t <= 1e-6 || t >= 1-1e-6 {
				continue
			}
			q := c.at(t)
			d0, d1 := math.Hypot(q.X-c.P0.X, q.Y-c.P0.Y), math.Hypot(q.X-c.P3.X, q.Y-c.P3.Y)
			over := math.Min(math.Abs(get(q)-get(c.P0)), math.Abs(get(q)-get(c.P3)))
			if math.Min(d0, d1) < splitMinPiece && over < splitMaxOvershoot {
				continue
			}
			if d0 >= 1 && d1 >= 1 {
				ts = append(ts, t)
			}
		}
	}
	sort.Float64s(ts)
	var ch chain
	prev := 0.0
	for _, t := range append(ts, 1) {
		if t-prev < 1e-6 {
			continue
		}
		ch = append(ch, curveSeg(c.sub(prev, t)))
		prev = t
	}
	return ch
}

// isTip reports whether q is the tip of the wedge w.
func isTip(q, w outline.Point) bool { return math.Hypot(q.X-w.X, q.Y-w.Y) < 1e-6 }

// splitMinPiece and splitMaxOvershoot are the limits of splitPart.
const (
	splitMinPiece     = 8.0
	splitMaxOvershoot = 0.3
)
