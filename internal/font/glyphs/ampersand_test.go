// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sort"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// Tests of the ampersand.

// ampShape holds the measured parts of the ampersand: the outer widths of the
// loop and of the bowl through the middles of their counters, the widths of
// the counters, and the crotches of the upper and of the lower crossing.
type ampShape struct {
	loopW, bowlW               float64
	loopCounterW, bowlCounterW float64
	upper, lower               []outline.Point
	upperAngle, lowerAngle     []float64
}

// ampMeasure measures the ampersand; it reports false if the outline is not the
// outer contour with the counters of the loop and the bowl and the eight
// crotches of the two crossings.
func ampMeasure(path outline.Path) (ampShape, bool) {
	var m ampShape
	if len(path) != 3 {
		return m, false
	}
	// The counters: the upper one is the loop.
	type box struct{ x0, y0, x1, y1 float64 }
	var bs []box
	for _, ct := range path[1:] {
		x0, y0, x1, y1 := contourBox(ct)
		bs = append(bs, box{x0, y0, x1, y1})
	}
	sort.Slice(bs, func(i, j int) bool { return bs[i].y0 > bs[j].y0 })
	outer := func(b box) float64 {
		y, cx := (b.y0+b.y1)/2, (b.x0+b.x1)/2
		xs := path.CrossingsX(y)
		sort.Float64s(xs)
		for i := 1; i+2 < len(xs); i++ {
			if xs[i] < cx && cx < xs[i+1] {
				return xs[i+2] - xs[i-1]
			}
		}
		return 0
	}
	m.loopW, m.bowlW = outer(bs[0]), outer(bs[1])
	m.loopCounterW, m.bowlCounterW = bs[0].x1-bs[0].x0, bs[1].x1-bs[1].x0

	at, ang := crotches(path)
	if len(at) != 8 {
		return m, false
	}
	idx := make([]int, 8)
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool { return at[idx[i]].Y > at[idx[j]].Y })
	for k, i := range idx {
		if k < 4 {
			m.upper, m.upperAngle = append(m.upper, at[i]), append(m.upperAngle, ang[i])
		} else {
			m.lower, m.lowerAngle = append(m.lower, at[i]), append(m.lowerAngle, ang[i])
		}
	}
	return m, true
}

// meanY returns the mean height of the points.
func meanY(pts []outline.Point) float64 {
	var s float64
	for _, q := range pts {
		s += q.Y
	}
	return s / float64(len(pts))
}

// The lower bowl is at least ampBowlOverLoop times as wide as the upper loop;
// the counter of the loop is at least ampLoopCounterMin wide and that of the
// bowl at least ampBowlCounterMin.
const (
	ampBowlOverLoop   = 1.4
	ampLoopCounterMin = 150.0
	ampBowlCounterMin = 220.0
)

func TestAmpersandBowlAndLoop(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, _, _ := outlineOf("ampersand", np.params)
		m, ok := ampMeasure(path)
		if !ok {
			t.Errorf("%s: the outline is not the outer contour, two counters and eight crotches", np.name)
			continue
		}
		if m.bowlW < ampBowlOverLoop*m.loopW {
			t.Errorf("%s: the bowl is %.0f wide and the loop %.0f, want a ratio of at least %v", np.name, m.bowlW, m.loopW, ampBowlOverLoop)
		}
		if m.loopCounterW < ampLoopCounterMin || m.bowlCounterW < ampBowlCounterMin {
			t.Errorf("%s: the counters are %.0f and %.0f wide, want at least %v and %v", np.name, m.loopCounterW, m.bowlCounterW, ampLoopCounterMin, ampBowlCounterMin)
		}
	}
}

// The crotches of the upper crossing lie, on average, between ampUpperBand[0]
// and ampUpperBand[1] of the height from AmpBottom to AmpTop, and those of the
// lower crossing in ampLowerBand.  Every crotch opens at least ampCrotchMin
// degrees.
var (
	ampUpperBand = [2]float64{0.50, 0.62}
	ampLowerBand = [2]float64{0.12, 0.26}
)

const ampCrotchMin = 50.0

func TestAmpersandCrossings(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		s := np.params.Sym
		path, _, _ := outlineOf("ampersand", np.params)
		m, ok := ampMeasure(path)
		if !ok {
			t.Errorf("%s: the outline is not the outer contour, two counters and eight crotches", np.name)
			continue
		}
		h := s.AmpTop - s.AmpBottom
		up, low := (meanY(m.upper)-s.AmpBottom)/h, (meanY(m.lower)-s.AmpBottom)/h
		if up < ampUpperBand[0] || up > ampUpperBand[1] {
			t.Errorf("%s: the upper crossing lies at %.3f of the height, want %v", np.name, up, ampUpperBand)
		}
		if low < ampLowerBand[0] || low > ampLowerBand[1] {
			t.Errorf("%s: the lower crossing lies at %.3f of the height, want %v", np.name, low, ampLowerBand)
		}
		for i, a := range append(m.upperAngle, m.lowerAngle...) {
			if a < ampCrotchMin {
				q := append(m.upper, m.lower...)[i]
				t.Errorf("%s: the crotch at (%.0f, %.0f) opens %.1f degrees, want at least %v", np.name, q.X, q.Y, a, ampCrotchMin)
			}
		}
	}
}

// The top of the serif of the arm is a horizontal line at AmpArm at least
// ampSerifMinLen long, whose right end lies at most ampSerifMaxIn left of the
// right end of the ink.
const (
	ampSerifMinLen = 180.0
	ampSerifMaxIn  = 30.0
)

func TestAmpersandSerif(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		s := np.params.Sym
		path, _, _ := outlineOf("ampersand", np.params)
		_, _, inkR, _ := path.BBox()
		// The top of the serif: the longest straight piece with both ends at
		// AmpArm.
		var top [2]outline.Point
		for _, ct := range path {
			cur := ct.Start
			for _, sg := range ct.Segs {
				a, b := cur, sg.P3
				cur = sg.P3
				if sg.IsCurve || math.Abs(a.Y-s.AmpArm) > 1e-6 || math.Abs(b.Y-s.AmpArm) > 1e-6 {
					continue
				}
				if math.Abs(b.X-a.X) > math.Abs(top[1].X-top[0].X) {
					top = [2]outline.Point{a, b}
				}
			}
		}
		if l := math.Abs(top[1].X - top[0].X); l < ampSerifMinLen {
			t.Errorf("%s: the horizontal top of the serif at %v is %.0f long, want at least %v", np.name, s.AmpArm, l, ampSerifMinLen)
		}
		if in := inkR - math.Max(top[0].X, top[1].X); in > ampSerifMaxIn {
			t.Errorf("%s: the serif ends %.0f left of the right end of the ink, want at most %v", np.name, in, ampSerifMaxIn)
		}
	}
}

// Over its last ampTailRun units the tail grows thinner towards the cut: the
// height of the lowest ink at verticals ampTailStep apart, taken at right
// angles to the lower edge, falls at every step.  The lower corner of the cut lies between the baseline and
// ampTailMaxLift above it.
const (
	ampTailRun     = 120.0
	ampTailStep    = 10.0
	ampTailMaxLift = 40.0
)

func TestAmpersandTail(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		path, _, _ := outlineOf("ampersand", np.params)
		_, _, inkR, _ := path.BBox()
		prev := math.Inf(1)
		for x := inkR - ampTailRun; x < inkR-1; x += ampTailStep {
			ys := path.CrossingsY(x)
			sort.Float64s(ys)
			if len(ys) < 2 {
				t.Fatalf("%s: no ink at x = %.0f", np.name, x)
			}
			// The height of the lowest ink, at right angles to the lower edge.
			next := path.CrossingsY(x + 1)
			sort.Float64s(next)
			th := (ys[1] - ys[0]) * math.Cos(math.Atan(next[0]-ys[0]))
			if th >= prev {
				t.Errorf("%s: the tail is %.1f thick at x = %.0f and %.1f a step before, want it thinner", np.name, th, x, prev)
			}
			prev = th
		}
		ys := path.CrossingsY(inkR - 0.5)
		sort.Float64s(ys)
		if len(ys) == 0 || ys[0] < 0 || ys[0] > ampTailMaxLift {
			t.Errorf("%s: the lower corner of the cut lies at %v, want it between 0 and %v", np.name, ys, ampTailMaxLift)
		}
	}
}
