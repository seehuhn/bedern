// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The dagger † and the double dagger ‡.
//
// Both signs are the union of limbs which leave a crossing: the head above it,
// the two arms beside it and, on †, the tail below it.  The arms and the head
// are drops: each edge leaves the crossing, where the limb is a hairline wide,
// as a straight line and runs into a round end on the spiral of the asterisk,
// which leaves the line with no curvature and joins the circle of the end with
// its curvature.  The tail swells from the crossing to its greatest width and
// runs on as a straight blade into a small round foot, joined in the same way.
// ‡ has the head of † at its top and, turned through half a turn, at its foot,
// the arms of † at both crossings, and between the crossings a stem which swells
// from a hairline at each crossing to its greatest width in the middle.  The
// limbs of a sign are joined in two contours, the stem and the arms, whose
// union has a corner in each crotch.

// Shape constants of the daggers.
const (
	daggerTop    = 696.0 // the top of the daggers
	daggerBottom = -150.0
	daggerCross  = 0.706 // the crossing of †, and the upper crossing of ‡, in heights from the bottom
	daggerReach  = 0.24  // how far an arm reaches from the middle of the stem, in heights
	daggerSide   = 50.0  // the side spaces of the daggers
	daggerDrop   = 0.8   // the diameter of the ends of the arms and the head, in ball diameters
	daggerFoot   = 0.27  // the diameter of the foot of †, in ball diameters
	daggerSwell  = 0.3   // where the straight edge of the tail begins, as a share of its length
	daggerTail   = 1.0   // the width of the tail there, in stems
	daggerMiddle = 0.8   // the greatest width of the stem of ‡ between its crossings, in stems
)

// A daggerLimbSpec gives a limb: its width at the crossing, the point of its
// upper edge where the straight edge towards the end begins, and its end.
type daggerLimbSpec struct {
	Waist float64 // the width at the crossing, in hairlines (ThinLower)
	At    float64 // the place where the straight edge begins, as a share of the length; 0 for a drop
	Wide  float64 // the width there, in stems
	End   float64 // the diameter of the round end, in ball diameters
}

var (
	daggerDropSpec = daggerLimbSpec{Waist: 1, End: daggerDrop}
	daggerTailSpec = daggerLimbSpec{Waist: 1, At: daggerSwell, Wide: daggerTail, End: daggerFoot}
)

// A daggerLimb is a limb in its own frame: it points along the positive x-axis
// from the crossing at the origin.  The upper edge passes through W, where its
// straight part begins; for a drop W is the edge point at the crossing.  R is
// the radius of the round end.
type daggerLimb struct {
	Len, Waist, R float64
	W             outline.Point
}

// newDaggerLimb returns the limb of the spec with the length l.
func newDaggerLimb(p design.Params, s daggerLimbSpec, l float64) daggerLimb {
	g := daggerLimb{Len: l, Waist: s.Waist * p.ThinLower, R: s.End * p.BallDiameter / 2}
	g.W = xy(s.At*l, s.Wide*p.Stem/2)
	if s.At == 0 {
		g.W = xy(0, g.Waist/2)
	}
	return g
}

// touching returns the signed curvature of the circle which touches the line
// through p in the direction t at p and passes through q (positive to the left).
func touching(p, t, q outline.Point) float64 {
	d := outline.Point{X: q.X - p.X, Y: q.Y - p.Y}
	return 2 * (t.X*d.Y - t.Y*d.X) / (d.X*d.X + d.Y*d.Y)
}

// daggerMinStraight is the shortest straight edge which a limb keeps as a line
// between its swelling and its end, and daggerLead the length of the straight
// before the spiral into the end which the curve into the end replaces, as on
// the asterisk.  A straight shorter than the sum of the two is replaced as a
// whole.
const (
	daggerMinStraight = 50.0
	daggerLead        = 30.0
)

// swell returns the lower edge from the crossing point n0 to the point w, where
// it arrives in the direction t with the curvature k: one cubic curve which
// leaves n0 in the direction of the circle which touches the line through w in
// the direction t at w and passes through n0, with the handle of that circular
// arc at n0 where k is not zero and of half the chord at w where it is.
func swell(n0, w, t outline.Point, k float64) cubic {
	t0 := dirOf(2*math.Atan2(w.Y-n0.Y, w.X-n0.X) - math.Atan2(t.Y, t.X))
	h := arcHandle(n0, t0, w, t)
	if math.Abs(k) < 1e-9 {
		// The handle at n0 which leaves no curvature at w, and at w half the
		// chord.
		d := xy(w.X-n0.X, w.Y-n0.Y)
		a := (d.X*t.Y - d.Y*t.X) / (t0.X*t.Y - t0.Y*t.X)
		b := math.Hypot(d.X, d.Y) / 2
		return cubic{n0, xy(n0.X+a*t0.X, n0.Y+a*t0.Y), xy(w.X-b*t.X, w.Y-b*t.Y), w}
	}
	c, err := g2End(n0, t0, h, w, t, k)
	if err != nil {
		panic(fmt.Sprintf("glyphs: the swelling of a dagger: %v", err))
	}
	return c
}

// open returns the limb as an open chain, counter-clockwise, from the lower
// edge at the crossing round the end to the upper edge at the crossing.
func (g daggerLimb) open() chain {
	n0 := xy(0, -g.Waist/2)
	wl := xy(g.W.X, -g.W.Y)
	mirror := func(c chain) chain {
		return c.mapPoints(func(q outline.Point) outline.Point { return xy(q.X, -q.Y) }).reversed()
	}
	// The straight edge from wl runs into the circle of the end on the spiral
	// of the asterisk.
	cp := newAsteriskCap(g.R, g.Len-g.R, g.W.X, g.W.Y)
	t := dirOf(cp.Alpha)
	straight := math.Hypot(cp.P0.X-wl.X, cp.P0.Y-wl.Y)
	lead := straight
	if straight >= daggerMinStraight+daggerLead {
		lead = daggerLead
	}
	q := xy(cp.P0.X-lead*t.X, cp.P0.Y-lead*t.Y)
	cuts := cp.cuts(0)
	n := len(cuts) - 1
	var cap chain
	for i := 1; i <= n; i++ {
		pa, ha, ka := cp.at(cuts[i-1])
		pb, hb, kb := cp.at(cuts[i])
		ta, tb := dirOf(ha), dirOf(hb)
		var c cubic
		var err error
		if kb != 0 {
			c, err = g2Cubic(pa, ta, ka, pb, tb, kb)
		} else {
			c, err = g2Cubic(pb, neg(tb), -kb, pa, neg(ta), -ka)
			c = cubic{c.P3, c.P2, c.P1, c.P0}
		}
		// The first and the last curve start and end at q and its mirror
		// image, with the handle there lengthened by the straight they
		// replace.
		if err == nil && i == 1 {
			h := math.Hypot(c.P1.X-pa.X, c.P1.Y-pa.Y)
			c, err = g2End(q, ta, h+lead, pb, tb, kb)
		}
		if err == nil && i == n {
			h := math.Hypot(c.P2.X-pb.X, c.P2.Y-pb.Y)
			c, err = g2Start(c.P0, ta, ka, xy(q.X, -q.Y), tb, h+lead)
		}
		if err != nil {
			panic(fmt.Sprintf("glyphs: the end of a dagger: %v", err))
		}
		cap = append(cap, curveSeg(c))
	}
	var low chain
	if g.W.X > 0 {
		k := 0.0
		if lead == straight {
			k = cap[0].cubic.curvature(0)
		}
		low = chain{curveSeg(swell(n0, wl, t, k))}
	}
	if lead < straight {
		low = append(low, lineSeg(wl, q))
	}
	if len(low) == 0 {
		return cap
	}
	return join(low, cap, mirror(low))
}

// turnedBy returns the chain turned counter-clockwise by the angle a about the
// origin and moved by (dx, dy).
func turnedBy(c chain, a, dx, dy float64) chain {
	cs, sn := math.Cos(a), math.Sin(a)
	return c.mapPoints(func(q outline.Point) outline.Point {
		return xy(q.X*cs-q.Y*sn+dx, q.X*sn+q.Y*cs+dy)
	})
}

// closed returns the chains one after the other as one closed chain, with
// straight lines where one does not start where the one before ends.
func closed(parts ...chain) chain {
	var all chain
	for _, c := range parts {
		if len(all) > 0 {
			if a, b := all.end(), c.start(); math.Hypot(a.X-b.X, a.Y-b.Y) > 1e-9 {
				all = append(all, lineSeg(a, b))
			}
		}
		all = append(all, c...)
	}
	if a, b := all.end(), all.start(); math.Hypot(a.X-b.X, a.Y-b.Y) > 1e-9 {
		all = append(all, lineSeg(a, b))
	}
	return all
}

// daggerArmsAt returns the two arms which cross the stem at the height y, as
// one contour.
func daggerArmsAt(p design.Params, y float64) chain {
	arm := newDaggerLimb(p, daggerDropSpec, daggerReach*(daggerTop-daggerBottom)).open()
	return closed(turnedBy(arm, 0, 0, y), turnedBy(arm, math.Pi, 0, y))
}

// daggerHead returns the head of a dagger whose crossing is at y.
func daggerHead(p design.Params, y float64) chain {
	return newDaggerLimb(p, daggerDropSpec, daggerTop-y).open()
}

// daggerPath returns the outline of †.
func daggerPath(p design.Params) outline.Path {
	y := daggerBottom + daggerCross*(daggerTop-daggerBottom)
	head := turnedBy(daggerHead(p, y), math.Pi/2, 0, y)
	tail := turnedBy(newDaggerLimb(p, daggerTailSpec, y-daggerBottom).open(), -math.Pi/2, 0, y)
	return chainsPath(union(closed(tail, head), daggerArmsAt(p, y))...)
}

// daggerdblPath returns the outline of ‡.
func daggerdblPath(p design.Params) outline.Path {
	y1 := daggerBottom + daggerCross*(daggerTop-daggerBottom)
	y0 := daggerTop + daggerBottom - y1
	head := daggerHead(p, y1)
	// The stem between the crossings: each edge is one arc from the crossing
	// to the middle.
	d := y1 - y0
	n0, nm := xy(0, p.ThinLower/2), xy(d/2, daggerMiddle*p.Stem/2)
	k := touching(nm, xy(1, 0), n0)
	t0 := dirOf(2 * math.Atan2(nm.Y-n0.Y, nm.X-n0.X))
	half := chain{curveSeg(arcThrough(n0, t0, nm, xy(1, 0), k))}
	upper := join(half, half.mapPoints(func(q outline.Point) outline.Point { return xy(d-q.X, q.Y) }).reversed())
	lower := upper.mapPoints(func(q outline.Point) outline.Point { return xy(q.X, -q.Y) })
	// In the frame of the stem the x-axis runs down from the upper crossing.
	stem := closed(lower, turnedBy(head, 0, d, 0), upper.reversed(), turnedBy(head, math.Pi, 0, 0))
	stem = turnedBy(stem, -math.Pi/2, 0, y1)
	return chainsPath(union(stem, daggerArmsAt(p, y1), daggerArmsAt(p, y0))...)
}

// dagger draws the dagger.
func dagger(p design.Params) (outline.Path, float64) {
	return placeInk(daggerPath(p), design.PunctSides{Left: daggerSide, Right: daggerSide})
}

// daggerdbl draws the double dagger.
func daggerdbl(p design.Params) (outline.Path, float64) {
	return placeInk(daggerdblPath(p), design.PunctSides{Left: daggerSide, Right: daggerSide})
}

// arcThrough returns the cubic curve from p0 in the direction t0 to p3 in the
// direction t3 with the curvature k at both ends, or the one whose curvatures
// come nearest to k.
func arcThrough(p0, t0, p3, t3 outline.Point, k float64) cubic {
	c, err := g2Cubic(p0, t0, k, p3, t3, k)
	if err != nil {
		return nearG2(p0, t0, k, p3, t3, k)
	}
	return c
}
