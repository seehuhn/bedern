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

// The round of C, G and Q and the beaks of C and G.
//
// C, G and Q are cut from the round of O: capRound returns the two
// super-ellipses of capO, the outer one and the counter, and the left side, the
// top and the bottom of the three letters are the curve of O itself, with its
// thickest part, its thin parts, its overshoot and its squareness.
//
// A beak of C and G is a serif on the thin end of an arc, of the family of the
// serifs on the arms of E, F and T (caparm.go, zBeak); capBeak draws it at the
// top of the round.  The arc comes over the top of the round and ends in a corner
// on the vertical cut, BeakX from the left extreme and Drop below the top
// extreme.  Its outer edge is one curve from the corner to the top extreme, which
// it meets running west with the curvature of the round there; of the curves
// which come down to the corner at capBeakRoundArriveMin to capBeakRoundArriveMax
// degrees below the horizontal, the fairest is taken (bending), which is nearly
// an arc of a circle; the last stretch of it is replaced by a short hollow to a
// tip a little higher (lift).  The arc ends where the round is still thin and
// bends at less than half a right angle.  The serif hangs from the corner: the
// cut runs down to End, the flat end is as wide as on the serifs of the arms, and
// the inner face leaves it at design.ZBeakTilt from the vertical and is
// ArmSerifThick from the cut halfway along the cut.  The face runs straight up to
// where the bracket of an arm leaves it: (reach - Bar)/2 below the point V in
// which it meets the underside of an arc Arc thick (the line at right angles Arc
// in from the outer edge), reach being how far the serif reaches beyond V, as the
// bracket of an arm leaves the face (Drop - Bar)/2 below its bar.  The inner edge
// of the arc is one curve from the counter of the round near its top, which it
// leaves with the direction and the curvature of the counter, to the top of the
// face, which it meets in the direction of the face; of the points at which it
// may leave the counter, the one whose curve changes its curvature least abruptly
// is taken (curvatureRate).  The underside of the arc runs thin from the top and
// turns, bending more and more and without a knee, into the face: the bracket.
// With a ramp (capBeakSpec.Ramp), the last stretch of the bracket and the face
// are one curve which arrives at the foot of the face without curvature
// (capBeakParts.ease).
//
// capBeakParts.chain returns the beak as the chain from the counter of the
// round, at the top extreme, along the inner edge and down the face, across the
// flat end, up the cut and along the outer edge to the top extreme of the outer
// contour: the right-hand end of a contour which comes over the top from the
// left on the outside and goes back over the top to the right on the inside.

// The outer edge of the arc of a beak of C and G comes down to the corner on the
// cut at capBeakRoundArriveMin to capBeakRoundArriveMax degrees below the horizontal.
const (
	capBeakRoundArriveMin = 30
	capBeakRoundArriveMax = 65
)

// A capRoundFrame holds the curves of the round of O.
type capRoundFrame struct {
	outer, inner superShape
	left, right  float64 // the extremes of the outer contour
	cx           float64 // the abscissa of the top and the bottom extremes
	top, bottom  float64 // the extremes of the outer contour
	innerTop     float64 // the top extreme of the counter
	innerBottom  float64 // the bottom extreme of the counter
}

// capRound returns the round of O, with its left extreme at the round side.
func capRound(p design.Params) capRoundFrame {
	c := p.Caps
	var f capRoundFrame
	f.left = c.SideRoundLeft
	f.right = f.left + c.OWidth
	f.cx = (f.left + f.right) / 2
	f.bottom = -p.OvershootUpper
	f.top = p.CapHeight + p.OvershootUpper
	f.innerBottom = f.bottom + c.RoundThin
	f.innerTop = f.top - c.RoundThin
	f.outer = newSuperShape(f.left, f.bottom, f.right, f.top, p.SupernessOuter)
	f.inner = newSuperShape(f.left+c.Round, f.innerBottom, f.right-c.Round, f.innerTop, p.SupernessInner)
	return f
}

// beakX returns the abscissa of the cut of the beaks.
func (f capRoundFrame) beakX(p design.Params) float64 { return f.left + p.Caps.CGSJQ.BeakX }

// capCuts returns the parameters at which the curves of the beak may meet a
// quarter of the round.
func capCuts(lo, hi float64) []float64 {
	var us []float64
	for u := lo; u <= hi+1e-9; u += 0.01 {
		us = append(us, u)
	}
	return us
}

// A capBeakParts holds the curves and the points of a beak of C and G.
type capBeakParts struct {
	Counter       cubic         // the upper right quarter of the counter, top to right extreme
	TI            float64       // the parameter on Counter at which Inner leaves it
	Inner         cubic         // the inner edge of the arc, from the counter to Face (to Ease, if there is one)
	Ease          cubic         // the end of the bracket and the face, from Inner to TipIn; empty without Ramp
	Face          outline.Point // the top of the straight inner face of the serif; TipIn where the bracket eases into the face (Ease)
	TipIn, TipOut outline.Point // the ends of the flat end, inside and on the cut
	Corner        outline.Point // the top of the cut, where the outer edge meets it
	Outer         cubic         // the outer edge of the round, from Hollow (or from Corner) to the top extreme
	Hollow        cubic         // the last stretch of the outer edge, from Corner to the start of Outer; empty without a lift
}

// A capBeakSpec gives the size of a beak at the top of the round: Drop is the
// depth below the top extreme at which the round meets the cut, Arc the thickness
// of the arc which the serif is placed against, and End the height of the end.
// The tip of the serif, the corner, stands Lift higher, where the outer edge
// lifts off the round Run (of arc length) before the cut, see capBeakParts.lift.
// Ramp, if positive, is the turn in degrees of the end of the bracket which
// eases into the face, see capBeakParts.ease.
type capBeakSpec struct{ Drop, Arc, End, Lift, Run, Ramp float64 }

// capUpperBeakSpec returns the size of the upper beak of C and G.
func capUpperBeakSpec(p design.Params) capBeakSpec {
	g := p.Caps.CGSJQ
	return capBeakSpec{Drop: g.BeakDrop, Arc: g.BeakArc, End: g.BeakEnd, Lift: g.BeakLift, Run: g.BeakRun}
}

// beakRamp is the turn in degrees of the end of the bracket of the beaks of s,
// S, C and G which eases into the face of the serif (capBeakParts.ease).
const beakRamp = 15.0

// capBeak returns the beak of the size s at the top of the round, see the
// comment above.
func capBeak(p design.Params, f capRoundFrame, s capBeakSpec) capBeakParts {
	qo := f.outer.cubic(0, true) // right extreme to the top
	qi := f.inner.cubic(3, false)
	a := beakArc{
		Top:  outline.Point{X: f.cx, Y: f.top},
		KTop: qo.curvature(1),
		Leave: func(u float64) (outline.Point, outline.Point, float64) {
			return qi.at(u), qi.tangent(u), qi.curvature(u)
		},
		Us: capCuts(0.02, 0.8),
		X:  f.beakX(p),
	}
	b, err := a.beak(p, s)
	if err != nil {
		panic(fmt.Sprintf("glyphs: a beak of C or G: %v; glyphs.Check reports such parameters", err))
	}
	b.Counter = qi
	return b
}

// A beakArc describes the arc which carries a beak, in the frame in which the
// beak is at the top on the right: the top extreme of the outer edge and its
// curvature there, the points at which the inner edge may leave the counter
// (Leave returns the point, the direction of the counter walked towards the
// beak and its curvature for each of the parameters Us), and the abscissa of
// the cut.
type beakArc struct {
	Top   outline.Point
	KTop  float64
	Leave func(u float64) (outline.Point, outline.Point, float64)
	Us    []float64
	X     float64
}

// beak returns the beak of the size s on the arc a, see the comment above; the
// Counter of the result is left empty.
func (a beakArc) beak(p design.Params, s capBeakSpec) (capBeakParts, error) {
	return a.beakSized(s, p.Caps.ArmSerifThick, p.Caps.Bar)
}

// beakSized returns the beak of the size s on the arc a with a serif which is
// thick from the cut halfway along it and a bracket which leaves the face as
// that of an arm under a bar of the thickness bar; beak gives it the values of
// the capitals and the lowercase s (s.go) its own.
func (a beakArc) beakSized(s capBeakSpec, thick, bar float64) (capBeakParts, error) {
	var b capBeakParts
	x := a.X
	b.Corner = outline.Point{X: x, Y: a.Top.Y - s.Drop}

	// The outer edge, from the corner to the top extreme.
	out, ok := capBeakOuter(b.Corner, a.Top, a.KTop)
	if !ok {
		return b, fmt.Errorf("the outer edge cannot reach the corner at %v", b.Corner)
	}
	b.Outer = out

	// The cut, the flat end and the face.
	cut := b.Corner.Y - s.End
	tilt := design.ZBeakTilt
	b.TipOut = outline.Point{X: x, Y: s.End}
	b.TipIn = outline.Point{X: x - (thick - cut/2*math.Tan(tilt)), Y: s.End}
	faceX := func(y float64) float64 { return b.TipIn.X - (y-s.End)*math.Tan(tilt) }

	// V, where the face meets the line Arc in from the outer edge.  The outer
	// edge runs west and the ink lies on its left.
	inside := func(t float64) outline.Point {
		q, tg := out.at(t), out.tangent(t)
		return outline.Point{X: q.X - s.Arc*tg.Y, Y: q.Y + s.Arc*tg.X}
	}
	atX := func(x float64) outline.Point {
		lo, hi := 0.0, 1.0
		for i := 0; i < 60; i++ {
			m := (lo + hi) / 2
			if inside(m).X > x {
				lo = m
			} else {
				hi = m
			}
		}
		return inside((lo + hi) / 2)
	}
	vy := s.End + cut/2
	for i := 0; i < 40; i++ {
		vy = atX(faceX(vy)).Y
	}
	rise := (vy - s.End - bar) / 2
	b.Face = outline.Point{X: faceX(vy - rise), Y: vy - rise}

	// The inner edge, from the counter to the top of the face.
	down := outline.Point{X: math.Sin(tilt), Y: -math.Cos(tilt)}
	best := math.Inf(1)
	for _, u := range a.Us {
		q, tg, k := a.Leave(u)
		cv, err := fairStart(q, tg, k, b.Face, down)
		if err != nil || !oneWay(cv, -1) {
			continue
		}
		if e := curvatureRate(cv); e < best {
			best, b.Inner, b.TI = e, cv, u
		}
	}
	if math.IsInf(best, 1) {
		return b, fmt.Errorf("the inner edge cannot reach the face (%+v)", s)
	}
	if s.Ramp > 0 {
		b.ease(s.Ramp, down)
	}
	// A lift which no hollow can follow (a narrow bowl, a perturbed set) is
	// reduced in steps and at last given up.
	for _, k := range []float64{1, 0.75, 0.5, 0.25} {
		if s.Lift <= 0 {
			break
		}
		t := s
		t.Lift *= k
		if b.lift(t) == nil {
			break
		}
	}
	return b, nil
}

// ease replaces the last stretch of the bracket, over which it turns by ramp
// degrees, and the straight face below it by one curve (footCubic), which
// leaves the bracket with its direction and curvature and arrives at the foot
// of the face, TipIn, in the direction of the face without curvature: the
// bracket turns into the face without a jump of curvature, and the face is
// straight where it meets the flat end.  Face becomes TipIn.  Where no such
// curve turns one way, the bracket and the face stay as they are.
func (b *capBeakParts) ease(ramp float64, down outline.Point) {
	in := b.Inner
	u := tTurned(in, 1, ramp)
	c, ok := footCubic(in.at(u), in.tangent(u), in.curvature(u), b.TipIn, down)
	if !ok || !oneWay(c, -1) {
		return
	}
	b.Ease = c
	b.Inner = in.sub(0, u)
	b.Face = b.TipIn
}

// footCubic returns the cubic curve from p0, leaving in the direction t0 with
// the signed curvature k0, to p3, where it arrives in the direction t3 without
// curvature: its first handle ends on the line through p3 in the direction t3,
// and its second handle is the one which gives the curvature k0 at p0.  The
// result is false unless both handles point forward and the second lies
// between the first and p3.
func footCubic(p0, t0 outline.Point, k0 float64, p3, t3 outline.Point) (cubic, bool) {
	cr := func(u, v outline.Point) float64 { return u.X*v.Y - u.Y*v.X }
	s := cr(t0, t3)
	d := outline.Point{X: p3.X - p0.X, Y: p3.Y - p0.Y}
	a := cr(d, t3) / s
	b := (cr(t0, d) - 1.5*k0*a*a) / s
	c := cubic{p0,
		outline.Point{X: p0.X + a*t0.X, Y: p0.Y + a*t0.Y},
		outline.Point{X: p3.X - b*t3.X, Y: p3.Y - b*t3.Y},
		p3}
	reach := math.Hypot(p3.X-c.P1.X, p3.Y-c.P1.Y)
	return c, a > 0 && b > 0 && b < reach
}

// The tip of the serif, where the outer edge ends on the cut, stands Lift above
// the place where the round meets the cut.  The outer edge follows the round up
// to the point P, Run (arc length) before the cut, and from there leaves the
// continuation of the round in a short hollow: with the direction and the
// curvature of the round at P its curvature falls through zero (one inflection)
// to a small opposite value, and the edge reaches the cut at the tip,
// capBeakArriveMin degrees or more below the horizontal, in a true corner.  The
// inner edge and the face of the serif are those of the round (the arc is
// slightly thicker next to the tip, a small wedge).  Of the curves which keep
// within capBeakHollowDepth of the chord from P to the tip, the fairest (bending)
// is taken.
const (
	capBeakArriveMin   = 28.0 // degrees below the horizontal at the tip: the first angle at which a hollow exists
	capBeakArriveMax   = 45.0
	capBeakHollowDepth = 5.0   // units: the most the hollow departs from its chord
	capBeakHollowCurv  = 0.007 // 1/units: the largest curvature in the hollow
)

// lift replaces the end of the outer edge of b by a hollow (capBeakHollow) which
// ends in the tip s.Lift above the corner of the round.
func (b *capBeakParts) lift(s capBeakSpec) error {
	round := b.Outer // from the corner of the round to the top extreme
	// P, s.Run of arc length from the corner.
	length, prev, tP := 0.0, round.at(0), 1.0
	for i := 1; i <= 2000; i++ {
		t := float64(i) / 2000
		q := round.at(t)
		length += math.Hypot(q.X-prev.X, q.Y-prev.Y)
		prev = q
		if length >= s.Run {
			tP = t
			break
		}
	}
	P, tg, kP := round.at(tP), round.tangent(tP), round.curvature(tP)
	tip := outline.Point{X: b.Corner.X, Y: b.Corner.Y + s.Lift}
	chord := math.Hypot(P.X-tip.X, P.Y-tip.Y)
	best, bestE := cubic{}, math.Inf(1)
	for deg := capBeakArriveMin; deg <= capBeakArriveMax+1e-9 && math.IsInf(bestE, 1); deg += 0.5 {
		a := deg * math.Pi / 180
		dir := outline.Point{X: -math.Cos(a), Y: math.Sin(a)}
		for i := 0; i <= 190; i++ {
			c, err := g2End(tip, dir, (0.05+0.005*float64(i))*chord, P, tg, kP)
			if err != nil || !hollowOK(c, tip, P) {
				continue
			}
			if e := bending(c); e < bestE {
				best, bestE = c, e
			}
		}
	}
	if math.IsInf(bestE, 1) {
		return fmt.Errorf("no hollow from %v to the tip %v (tangent %v k %v)", P, tip, tg, kP)
	}
	b.Hollow = best
	b.Outer = round.sub(tP, 1)
	b.Corner = tip
	return nil
}

// hollowOK reports whether c, from the tip to P, has exactly one inflection
// with the curvature of the round at P and the opposite sign near the tip, no
// curvature beyond capBeakHollowCurv, and stays within capBeakHollowDepth of its
// chord.
func hollowOK(c cubic, tip, P outline.Point) bool {
	cx, cy := P.X-tip.X, P.Y-tip.Y
	n := math.Hypot(cx, cy)
	signs, last := 0, 0.0
	for i := 0; i <= 200; i++ {
		t := float64(i) / 200
		q, k := c.at(t), c.curvature(t)
		if math.Abs(k) > capBeakHollowCurv {
			return false
		}
		if d := math.Abs((q.X-tip.X)*cy-(q.Y-tip.Y)*cx) / n; d > capBeakHollowDepth {
			return false
		}
		if k != 0 && last != 0 && k*last < 0 {
			signs++
		}
		if k != 0 {
			last = k
		}
		if i == 0 && k > 0 {
			return false
		}
	}
	return signs == 1 && last > 0
}

// capUpperBeakEased returns the upper beak of C and G as a chain, see the
// comment above, with a bracket which eases into the face (beakRamp).
func capUpperBeakEased(p design.Params, f capRoundFrame) chain {
	return capBeak(p, f, capUpperBeakEasedSpec(p)).chain()
}

// capUpperBeakEasedSpec returns the size of the upper beak of C and G with the
// ramp of its bracket.
func capUpperBeakEasedSpec(p design.Params) capBeakSpec {
	s := capUpperBeakSpec(p)
	s.Ramp = beakRamp
	return s
}

// chain returns the beak as a chain, from the counter at the top extreme to the
// top extreme of the outer edge.
func (b capBeakParts) chain() chain {
	return join(
		chain{curveSeg(b.Counter.sub(0, b.TI))},
		splitAtExtrema(b.Inner),
		b.easeChain(),
		b.faceChain(),
		lineChain(b.TipIn, b.TipOut),
		lineChain(b.TipOut, b.Corner),
		b.hollowChain(),
		chain{curveSeg(b.Outer)},
	)
}

// capBeakOuter returns the outer edge of the arc of a beak of C and G, from the
// corner on the cut to the top extreme, which it meets running west with the
// curvature k of the round there: of the curves which leave the corner between
// capBeakArriveMin and capBeakArriveMax degrees below the horizontal and turn
// one way, the fairest (bending).
func capBeakOuter(corner, top outline.Point, k float64) (cubic, bool) {
	return capOuterArriving(corner, top, k, capBeakRoundArriveMin, capBeakRoundArriveMax)
}

// capOuterArriving returns the fairest curve (bending) from the corner to the
// top extreme, which it meets running west with the curvature k, among those
// which leave the corner between lo and hi degrees below the horizontal and
// turn one way: the outer edge of a beak of C and G and of the tail of C.
func capOuterArriving(corner, top outline.Point, k, lo, hi float64) (cubic, bool) {
	key := outerArrivingKey{corner, top, k, lo, hi}
	if v, ok := outerArrivings.Load(key); ok {
		res := v.(outerArrivingResult)
		return res.c, res.ok
	}
	c, ok := searchOuterArriving(corner, top, k, lo, hi)
	outerArrivings.Store(key, outerArrivingResult{c, ok})
	return c, ok
}

// outerArrivingKey holds every input of capOuterArriving, which is a pure
// function of them.
type outerArrivingKey struct {
	corner, top outline.Point
	k, lo, hi   float64
}

type outerArrivingResult struct {
	c  cubic
	ok bool
}

var outerArrivings sync.Map // outerArrivingKey -> outerArrivingResult

// searchOuterArriving is capOuterArriving without the cache.
func searchOuterArriving(corner, top outline.Point, k, lo, hi float64) (cubic, bool) {
	best, bestE := cubic{}, math.Inf(1)
	chord := math.Hypot(top.X-corner.X, top.Y-corner.Y)
	for deg := lo; deg <= hi; deg += 0.5 {
		a := deg * math.Pi / 180
		dir := outline.Point{X: -math.Cos(a), Y: math.Sin(a)}
		for i := 0; i <= 120; i++ {
			c, err := g2End(corner, dir, (0.05+0.005*float64(i))*chord, top, outline.Point{X: -1}, k)
			if err != nil || !oneWay(c, 1) {
				continue
			}
			if e := bending(c); e < bestE {
				best, bestE = c, e
			}
		}
	}
	return best, !math.IsInf(bestE, 1)
}

// easeChain returns the end of the bracket as a chain, empty if the beak has
// none.
func (b capBeakParts) easeChain() chain {
	if b.Ease == (cubic{}) {
		return nil
	}
	return chain{curveSeg(b.Ease)}
}

// faceChain returns the straight face of the serif, from Face to TipIn, empty
// where the bracket eases into the face (Ease).
func (b capBeakParts) faceChain() chain {
	if b.Ease != (cubic{}) {
		return nil
	}
	return lineChain(b.Face, b.TipIn)
}

// hollowChain returns the hollow of the outer edge as a chain, empty if the
// beak has none.
func (b capBeakParts) hollowChain() chain {
	if b.Hollow == (cubic{}) {
		return nil
	}
	return chain{curveSeg(b.Hollow)}
}
