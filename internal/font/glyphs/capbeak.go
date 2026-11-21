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

// The beaks of s.
//
// A beak is a serif on the thin end of an arc.  The arc comes over the top of a round and ends in a
// corner on the vertical cut, at the abscissa X and Drop below the top extreme.  Its outer edge is
// one curve from the corner to the top extreme, which it meets running west with the curvature of
// the round there; of the curves which come down to the corner at capBeakRoundArriveMin to
// capBeakRoundArriveMax degrees below the horizontal, the fairest is taken (bending), which is
// nearly an arc of a circle; the last stretch of it is replaced by a short hollow to a tip a little
// higher (lift).  The arc ends where the round is still thin and bends at less than half a right
// angle.  The serif hangs from the corner: the cut runs down to End, and the inner face leaves the
// flat end at design.ZBeakTilt from the vertical and is thick from the cut halfway along the cut.
// The face runs straight up to where the bracket leaves it: (reach - Bar)/2 below the point V in
// which it meets the underside of an arc Arc thick (the line at right angles Arc in from the outer
// edge), reach being how far the serif reaches beyond V.  The inner edge of the arc is one curve
// from the counter of the round near its top, which it leaves with the direction and the curvature
// of the counter, to the top of the face, which it meets in the direction of the face; of the
// points at which it may leave the counter, the one whose curve changes its curvature least
// abruptly is taken (curvatureRate).  The underside of the arc runs thin from the top and turns,
// bending more and more and without a knee, into the face: the bracket.

// The outer edge of the arc of a beak comes down to the corner on the
// cut at capBeakRoundArriveMin to capBeakRoundArriveMax degrees below the horizontal.
const (
	capBeakRoundArriveMin = 30
	capBeakRoundArriveMax = 65
)

// A capBeakParts holds the curves and the points of a beak.
type capBeakParts struct {
	Inner         cubic         // the inner edge of the arc, from the counter to Face
	Face          outline.Point // the top of the straight inner face of the serif
	TipIn, TipOut outline.Point // the ends of the flat end, inside and on the cut
	Corner        outline.Point // the top of the cut, where the outer edge meets it
	Outer         cubic         // the outer edge of the round, from Hollow (or from Corner) to the top extreme
	Hollow        cubic         // the last stretch of the outer edge, from Corner to the start of Outer; empty without a lift
}

// A capBeakSpec gives the size of a beak at the top of a round: Drop is the
// depth below the top extreme at which the round meets the cut, Arc the thickness
// of the arc which the serif is placed against, and End the height of the end.
// The tip of the serif, the corner, stands Lift higher, where the outer edge
// lifts off the round Run (of arc length) before the cut, see capBeakParts.lift.
type capBeakSpec struct{ Drop, Arc, End, Lift, Run float64 }

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

// beakSized returns the beak of the size s on the arc a with a serif which is
// thick from the cut halfway along it and a bracket which leaves the face as
// under a bar of the thickness bar.
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
			best, b.Inner = e, cv
		}
	}
	if math.IsInf(best, 1) {
		return b, fmt.Errorf("the inner edge cannot reach the face (%+v)", s)
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

// capBeakOuter returns the outer edge of the arc of a beak, from the
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
// turn one way: the outer edge of a beak.
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
