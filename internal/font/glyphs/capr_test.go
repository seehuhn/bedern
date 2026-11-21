// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// The tests of the leg of R measure the outline in every variant and perturbed
// parameter set: the leg is straight and leans as the leg of K does, it is as
// heavy as the thick diagonals, and the right side of the letter is open, with
// one clear notch between the bowl and the leg and open whites on both sides of
// the leg.  The foot serif of the leg is checked with those of the other
// diagonals in TestDiagonalBracketsAreThoseOfTheStem.

// capRLegSpan returns the heights between which the edges of the leg of R are
// free of the serifs and their brackets and of the bowl: from 5 units above the
// longest bracket of a foot serif on a diagonal to 5 units below the underside of
// the bowl.
func capRLegSpan(p design.Params) (lo, hi float64) {
	c := p.Caps
	lo = c.SerifThick + design.DiagBracketObtuse*(c.BracketHeight-c.SerifThick) + 5
	hi = c.RBar - c.Bar/2 - 5
	return lo, hi
}

// legEdges returns the left and the right edge of the leg of R at the height y:
// the last run of the scan.
func legEdges(np namedParams, y float64) (l, r float64) {
	rs := runs(capOutline(np, "R"), y)
	last := rs[len(rs)-1]
	return last[0], last[1]
}

// capRLegAngleBand is the band of the angle of the leg of R from the vertical, in
// degrees: about that of the leg of K (34).  capRLegAngleToK is the largest
// difference from the angle of the leg of K.
const (
	capRLegAngleLo, capRLegAngleHi = 30.0, 40.0
	capRLegAngleToK                = 4.0
)

// Both edges of the leg of R are straight lines from the bracket of the foot
// serif up to the bowl: at five heights each edge lies within half a unit of the
// line through its ends.  The leg leans between capRLegAngleLo and
// capRLegAngleHi degrees from the vertical and within capRLegAngleToK of the leg
// of K.
func TestCapRLegIsStraight(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		lo, hi := capRLegSpan(p)
		l0, r0 := legEdges(np, lo)
		l1, r1 := legEdges(np, hi)
		for i := 1; i < 4; i++ {
			f := float64(i) / 4
			l, r := legEdges(np, lo+f*(hi-lo))
			if d := l - (l0 + f*(l1-l0)); math.Abs(d) > 0.5 {
				t.Errorf("the left edge of the leg of R is %.2f off a straight line at %.0f", d, lo+f*(hi-lo))
			}
			if d := r - (r0 + f*(r1-r0)); math.Abs(d) > 0.5 {
				t.Errorf("the right edge of the leg of R is %.2f off a straight line at %.0f", d, lo+f*(hi-lo))
			}
		}
		angle := math.Atan2(l0-l1, hi-lo) * 180 / math.Pi
		k := math.Atan(design.CapKLegLean(p)) * 180 / math.Pi
		if angle < capRLegAngleLo || angle > capRLegAngleHi || math.Abs(angle-k) > capRLegAngleToK {
			t.Errorf("the leg of R leans %.1f degrees, want %v to %v and within %v of the leg of K (%.1f)", angle, capRLegAngleLo, capRLegAngleHi, capRLegAngleToK, k)
		}
		if np.name == "regular" {
			t.Logf("the leg of R leans %.1f degrees, the leg of K %.1f", angle, k)
		}
	})
}

// capRLegWeight is the band of the thickness of the leg of R at right angles, as
// a multiple of the thick diagonals of the capitals (the leg of K is 0.97).
var capRLegWeight = [2]float64{0.93, 1.07}

// The leg of R is as heavy as the thick diagonals: its thickness at right angles,
// half way between the serif and the bowl, lies in capRLegWeight of
// CapDiag.Thick.
func TestCapRLegIsAsHeavyAsTheDiagonals(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		lo, hi := capRLegSpan(np.params)
		l0, _ := legEdges(np, lo)
		l1, _ := legEdges(np, hi)
		lean := math.Atan2(l0-l1, hi-lo)
		l, r := legEdges(np, (lo+hi)/2)
		w := (r - l) * math.Cos(lean)
		ref := np.params.CapDiag.Thick
		if q := w / ref; q < capRLegWeight[0] || q > capRLegWeight[1] {
			t.Errorf("the leg of R is %.1f thick, %.2f of the thick diagonals (%.0f), want %v to %v", w, q, ref, capRLegWeight[0], capRLegWeight[1])
		}
		if np.name == "regular" {
			t.Logf("the leg of R is %.1f thick at right angles, %.2f of the thick diagonals", w, w/ref)
		}
	})
}

// capRWhite is the least width of the two whites beside the top of the leg of R,
// as a multiple of the capital stem.
const capRWhite = 0.8

// The right side of R is open.  The rightmost ink, from the middle of the bowl
// down to the serif of the leg, moves in once and out once (within half a unit),
// with one notch between the bowl and the leg.  The
// white triangle between the stem and the leg, 20 units below the underside of
// the bowl, and the white between the right edge of the leg and the extreme of
// the bowl, 40 units below the notch, are each at least capRWhite of the capital
// stem wide.
func TestCapRRightSideIsOpen(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		c := p.Caps
		path := capOutline(np, "R")
		b := capRBowl(p)
		lo, _ := capRLegSpan(p)

		// The rightmost ink, from the middle of the bowl down.
		var xs, ys []float64
		for y := b.mid(); y >= lo; y-- {
			rs := runs(path, y)
			xs, ys = append(xs, rs[len(rs)-1][1]), append(ys, y)
		}
		low := 0
		for i := range xs {
			if xs[i] < xs[low] {
				low = i
			}
		}
		for i := 1; i < len(xs); i++ {
			if i <= low && xs[i] > xs[i-1]+0.5 || i > low && xs[i] < xs[i-1]-0.5 {
				t.Errorf("the right side of R turns back at %.0f: %.1f after %.1f, the notch is at %.0f", ys[i], xs[i], xs[i-1], ys[low])
				break
			}
		}

		rs := runs(path, b.Bottom-20)
		if len(rs) != 2 {
			t.Fatalf("R has %d runs at %.0f, want the stem and the leg", len(rs), b.Bottom-20)
		}
		inner := rs[1][0] - rs[0][1]
		_, q := legEdges(np, ys[low]-40)
		outer := b.Right - q
		if inner < capRWhite*c.Stem || outer < capRWhite*c.Stem {
			t.Errorf("the whites beside the leg of R are %.1f (stem to leg) and %.1f (leg to bowl), want at least %.1f", inner, outer, capRWhite*c.Stem)
		}
		if np.name == "regular" {
			t.Logf("the notch of R is at %.0f; the whites beside the leg are %.1f and %.1f", ys[low], inner, outer)
		}
	})
}
