// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sort"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// twoBowlSide is the stroke of the bowl of 2 from the top of the counter round
// its right side down to twoNeck times the height of the right extreme: the
// points of the inner edge in order and, for each, the distance to the outer
// edge and the middle of the stroke there.
type twoBowlSide struct {
	inner, mid []outline.Point
	thick      []float64
}

const (
	// twoNeck is the height, as a fraction of the height of the right
	// extreme, down to which the stroke of the bowl is followed.
	twoNeck = 0.6

	// twoStretch is the greatest height, as a fraction of the height of the
	// figure, over which the stroke of the bowl keeps 90 per cent of its
	// greatest thickness.
	twoStretch = 0.18

	// twoHandles is the range of the handles of the inner edge from the top
	// of the counter to the right extreme, each as a fraction of the width
	// and of the height of that quarter, and twoHandleRatio the greatest
	// ratio of the two fractions.
	twoHandleRatio = 1.25
)

var twoHandles = [2]float64{0.5, 0.8}

// bowlSide samples the stroke of the bowl of 2 in the set.
func bowlSide(np namedParams) twoBowlSide {
	p := np.params
	f := p.Figures
	path := capOutline(np, "two")
	x0 := design.FigureLeft(p, 2)
	top := figTop(p)

	// The inner edge over the top of the counter (vertical scans) and down
	// its right side (horizontal scans); the outer edge likewise.
	var in, out []outline.Point
	xt, yt := 0.0, math.Inf(-1)
	for x := x0 + 0.3*f.Ink[2]; x <= x0+f.Ink[2]; x++ {
		r := vruns(path, x)
		if len(r) < 2 {
			continue
		}
		q := r[len(r)-1]
		if q[1] < top-2*f.Thin {
			continue
		}
		out = append(out, outline.Point{X: x, Y: q[1]})
		if x <= x0+0.8*f.Ink[2] && q[0] > yt {
			xt, yt = x, q[0]
		}
		if r[len(r)-2][1] < q[0]-1 && q[0] > f.TwoShoulder {
			in = append(in, outline.Point{X: x, Y: q[0]})
		}
	}
	for y := twoNeck * f.TwoShoulder; y < top; y++ {
		r := runs(path, y)
		if len(r) == 0 {
			continue
		}
		q := r[len(r)-1]
		out = append(out, outline.Point{X: q[1], Y: y})
		if len(r) >= 2 && y < yt && q[0] > xt {
			in = append(in, outline.Point{X: q[0], Y: y})
		}
	}
	// Along the inner edge, clockwise round the centre of the quarter.
	var kept []outline.Point
	for _, q := range in {
		if q.X >= xt {
			kept = append(kept, q)
		}
	}
	c := outline.Point{X: xt, Y: f.TwoShoulder}
	ang := func(q outline.Point) float64 { return math.Atan2(q.X-c.X, q.Y-c.Y) }
	sort.Slice(kept, func(i, j int) bool { return ang(kept[i]) < ang(kept[j]) })

	var s twoBowlSide
	for _, q := range kept {
		best, at := math.Inf(1), outline.Point{}
		for _, o := range out {
			if d := math.Hypot(o.X-q.X, o.Y-q.Y); d < best {
				best, at = d, o
			}
		}
		s.inner = append(s.inner, q)
		s.thick = append(s.thick, best)
		s.mid = append(s.mid, outline.Point{X: (q.X + at.X) / 2, Y: (q.Y + at.Y) / 2})
	}
	return s
}

// The thickness of the stroke of the bowl of 2 rises without falling back
// from the top of the counter to one greatest value, which lies at the height
// of the right extreme of the outer edge, of that of the inner edge or between
// them, and falls without rising below it down to the neck, each within half a
// unit.
func TestFigureTwoBowlSwells(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		s := bowlSide(np)
		if len(s.inner) < 50 {
			t.Fatalf("%s: %d points on the inner edge of the bowl of 2", np.name, len(s.inner))
		}
		peak := 0
		for i, d := range s.thick {
			if d > s.thick[peak] {
				peak = i
			}
		}
		f := np.params.Figures
		slack := 0.03 * f.TwoShoulder
		if y := s.inner[peak].Y; y < f.TwoShoulder-f.TwoOuterDrop-slack || y > f.TwoShoulder+slack {
			t.Errorf("%s: the stroke of the bowl of 2 is thickest at %.0f, the right extremes of its edges are at %.0f and %.0f", np.name, y, f.TwoShoulder-f.TwoOuterDrop, f.TwoShoulder)
		}
		for i := 1; i < len(s.thick); i++ {
			d := s.thick[i] - s.thick[i-1]
			if i <= peak && d < -0.5 || i > peak && d > 0.5 {
				t.Errorf("%s: the thickness of the bowl of 2 goes from %.1f to %.1f at (%.0f, %.0f)", np.name, s.thick[i-1], s.thick[i], s.inner[i].X, s.inner[i].Y)
				break
			}
		}
	})
}

// The stroke of the bowl of 2 keeps 90 per cent of its greatest thickness over
// a height of less than twoStretch times the height of the figure.
func TestFigureTwoBowlStretch(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		s := bowlSide(np)
		peak := 0
		for i, d := range s.thick {
			if d > s.thick[peak] {
				peak = i
			}
		}
		lo, hi := peak, peak
		for lo > 0 && s.thick[lo-1] >= 0.9*s.thick[peak] {
			lo--
		}
		for hi+1 < len(s.thick) && s.thick[hi+1] >= 0.9*s.thick[peak] {
			hi++
		}
		h := figTop(np.params)
		if d := math.Abs(s.inner[lo].Y - s.inner[hi].Y); d >= twoStretch*h {
			t.Errorf("%s: the bowl of 2 keeps 90 per cent of its thickness %.1f over %.0f units, want under %.0f", np.name, s.thick[peak], d, twoStretch*h)
		}
	})
}

// The middle of the stroke of the bowl of 2 moves only to the right and only
// down from the top of the counter to the height of the lower of the two right
// extremes, within a unit.
func TestFigureTwoBowlMiddle(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		s := bowlSide(np)
		f := np.params.Figures
		for i := 1; i < len(s.mid) && s.inner[i].Y >= f.TwoShoulder-f.TwoOuterDrop; i++ {
			a, b := s.mid[i-1], s.mid[i]
			if b.X < a.X-1 || b.Y > a.Y+1 {
				t.Errorf("%s: the middle of the bowl of 2 turns back from (%.1f, %.1f) to (%.1f, %.1f)", np.name, a.X, a.Y, b.X, b.Y)
				break
			}
		}
	})
}

// The inner edge of 2 from the top of the counter to its right extreme is one
// curve whose handles, as fractions of the width and of the height of that
// quarter, lie within twoHandles and differ by at most the factor
// twoHandleRatio.
func TestFigureTwoCounterRound(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		ct := capOutline(np, "two")[0]
		f := np.params.Figures
		found := false
		for i := range ct.Segs {
			if !ct.Segs[i].IsCurve {
				continue
			}
			c := segCubic(ct, i)
			t0, t3 := c.tangent(0), c.tangent(1)
			if !(t0.X > 0.999 && t3.Y < -0.999 && math.Abs(c.P3.Y-f.TwoShoulder) < 0.5 && c.P0.Y > c.P3.Y) {
				continue
			}
			found = true
			w, h := c.P3.X-c.P0.X, c.P0.Y-c.P3.Y
			a, b := (c.P1.X-c.P0.X)/w, (c.P2.Y-c.P3.Y)/h
			if a < twoHandles[0] || a > twoHandles[1] || b < twoHandles[0] || b > twoHandles[1] || math.Max(a, b) > twoHandleRatio*math.Min(a, b) {
				t.Errorf("%s: the handles of the counter of 2 from (%.0f, %.0f) are %.2f of its width and %.2f of its height", np.name, c.P0.X, c.P0.Y, a, b)
			}
		}
		if !found {
			t.Errorf("%s: no curve from the top of the counter of 2 to its right extreme", np.name)
		}
	})
}
