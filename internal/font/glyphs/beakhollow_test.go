// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// A beakPart is the end of the outer edge of a beak in the frame in which the
// beak hangs at the top on the right: the hollow from the tip to P, the round
// from P to the top extreme and the cut.
type beakPart struct {
	name         string
	hollow       cubic // from the tip to P
	round        cubic // from P towards the top extreme
	tip, tipFoot outline.Point
	lift         float64 // the height of the tip above the round alone
}

// beakParts returns the beaks of C and G and the two of s and S for the regular
// set.
func beakParts(p design.Params) []beakPart {
	var out []beakPart
	spec := capUpperBeakSpec(p)
	b := capBeak(p, capRound(p), spec)
	flat := spec
	flat.Lift = 0
	f0 := capBeak(p, capRound(p), flat)
	out = append(out, beakPart{"C and G", b.Hollow, b.Outer, b.Corner, b.TipOut, b.Corner.Y - f0.Corner.Y})
	turn := func(q outline.Point) outline.Point { return outline.Point{X: -q.X, Y: -q.Y} }
	turnC := func(c cubic) cubic { return cubic{turn(c.P0), turn(c.P1), turn(c.P2), turn(c.P3)} }
	for _, f := range []struct {
		name string
		fr   sFrame
		lift float64
	}{{"s", newSFrame(p), p.SBeakLift}, {"S", newCapSFrame(p), p.Caps.CGSJQ.BeakLift}} {
		for i, e := range []sEdges{f.fr.UpperE, f.fr.LowerE} {
			out = append(out, beakPart{f.name + " " + sBeakNames[i],
				turnC(e.Outer[3]).reversed(), turnC(e.Outer[2]).reversed(), turn(e.Corner), turn(e.TipOuter), f.lift})
		}
	}
	return out
}

// The tip of the serif of every beak of C, G, s and S stands 8 to 24 units
// above the place where the round of the arc, followed on without the hollow,
// meets the cut, and the outer edge keeps the round of the curve up to the
// hollow.  The angle between the outer edge and the cut at the tip is 115 to 128
// degrees on C and G and 118 to 136 on s and S.
func TestBeakTipAndAngle(t *testing.T) {
	t.Parallel()
	p := design.Regular()
	for _, b := range beakParts(p) {
		if b.lift < 8 || b.lift > 24 {
			t.Errorf("the tip of the %s beak stands %.1f above the round, want 8 to 24", b.name, b.lift)
		}
		d := outline.Point{X: b.hollow.P1.X - b.hollow.P0.X, Y: b.hollow.P1.Y - b.hollow.P0.Y}
		a := math.Acos(-d.Y/math.Hypot(d.X, d.Y)) * 180 / math.Pi
		lo, hi := 118.0, 136.0
		if b.name == "C and G" {
			lo, hi = 115, 128
		}
		t.Logf("%s: tip at height %.1f above the round, angle %.1f", b.name, b.lift, a)
		if a < lo || a > hi {
			t.Errorf("the corner of the %s beak is %.1f degrees, want %.0f to %.0f", b.name, a, lo, hi)
		}
	}
}

// The hollow from P to the tip is one shallow curve: the tangent and the
// curvature are those of the round at P, the curvature changes sign once and
// stays between -0.007 and 0.007 per unit, the curve departs from its chord by at
// most 5 units, and it is not straight anywhere (its curvature is above 0.0004 in
// size except within 8 units of the inflection).
func TestBeakHollow(t *testing.T) {
	t.Parallel()
	for _, b := range beakParts(design.Regular()) {
		h, r := b.hollow, b.round
		ta, tb := h.tangent(1), r.tangent(0)
		if j := math.Abs(math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y)); j > 0.002 {
			t.Errorf("%s: the tangent jumps by %.4f radians at P", b.name, j)
		}
		if ka, kb := h.curvature(1), r.curvature(0); math.Abs(ka-kb) > 1e-4 {
			t.Errorf("%s: the curvature jumps from %.5f to %.5f at P", b.name, ka, kb)
		}
		cx, cy := h.P3.X-h.P0.X, h.P3.Y-h.P0.Y
		n := math.Hypot(cx, cy)
		changes, last := 0, 0.0
		for i := 0; i <= 400; i++ {
			u := float64(i) / 400
			k := h.curvature(u)
			if k < -0.007 || k > 0.007 {
				t.Errorf("%s: the curvature of the hollow is %.5f", b.name, k)
				break
			}
			q := h.at(u)
			if d := math.Abs((q.X-h.P0.X)*cy-(q.Y-h.P0.Y)*cx) / n; d > 5 {
				t.Errorf("%s: the hollow departs from its chord by %.1f", b.name, d)
				break
			}
			if last != 0 && k*last < 0 {
				changes++
			}
			if k != 0 {
				last = k
			}
		}
		if changes != 1 {
			t.Errorf("%s: the curvature of the hollow changes sign %d times, want once", b.name, changes)
		}
	}
}
