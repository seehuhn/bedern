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

// The helpers of this file check the parts which several letters share, in a
// path which holds other ink as well: the foot serifs and the head serif of a
// stem, and the arch between two stems.  They look for the crossing nearest to
// the expected position.

// nearest returns the value of vs which is nearest to want.  It fails the test
// if there is none within the tolerance.
func nearest(t *testing.T, what string, vs []float64, want float64) float64 {
	t.Helper()
	best := math.NaN()
	for _, v := range vs {
		if math.IsNaN(best) || math.Abs(v-want) < math.Abs(best-want) {
			best = v
		}
	}
	if math.IsNaN(best) || math.Abs(best-want) > tol {
		t.Errorf("%s: no crossing near %.1f in %v", what, want, vs)
		return want
	}
	return best
}

// lowest returns the two lowest values of vs, which are the bottom and the top
// of the ink at the bottom of the column.
func lowest(t *testing.T, what string, vs []float64) (bottom, top float64) {
	t.Helper()
	s := append([]float64(nil), vs...)
	sort.Float64s(s)
	if len(s) < 2 {
		t.Fatalf("%s: fewer than two crossings: %v", what, vs)
	}
	return s[0], s[1]
}

// checkFoot checks the foot serif at the bottom of a stem with the edge at x,
// which has the length jut and is rise thick at its tip: on the right of the
// stem if right is set, else on its left.  The path may hold other ink, and need
// hold none of the other foot of the stem.  It checks the length, the thickness
// at the tip, the flat top, and that the bracket is concave and leaves the stem
// at BracketHeight.
func checkFoot(t *testing.T, p design.Params, path outline.Path, x float64, right bool, jut, rise float64) {
	t.Helper()
	side := 1.0
	if !right {
		side = -1
	}
	// The serif, measured just above the baseline.
	tip := nearest(t, "foot tip", path.CrossingsX(footScanY), x+side*jut)
	near(t, "foot serif length", side*(tip-x), jut)

	// The thickness at the tip: the lowest ink of a vertical line just inside.
	bottom, top := lowest(t, "tip", path.CrossingsY(tip-side*tipInset))
	near(t, "foot serif bottom", bottom, 0)
	near(t, "foot serif thickness at the tip", top-bottom, rise)

	// Just above the bracket only the stem remains.
	nearest(t, "stem edge above the bracket", path.CrossingsX(p.BracketHeight+1), x)

	// The bracket is concave: halfway up it stands out from the stem by clearly
	// less than the straight chamfer's reach/2, but still visibly.
	midY := rise + (p.BracketHeight-rise)/2
	cs := path.CrossingsX(midY)
	// A crossing exactly on the edge (the stem itself) is not the bracket: the
	// bracket is the crossing between the edge and the tip.
	var offsets []float64
	for _, v := range cs {
		if o := side * (v - x); o > 0.5 && o < jut {
			offsets = append(offsets, o)
		}
	}
	if len(offsets) == 0 {
		t.Errorf("no bracket at mid height %.1f: %v", midY, cs)
	} else if o := offsets[0]; o < 0.1*p.BracketReach || o > 0.4*p.BracketReach {
		t.Errorf("bracket at half height stands out %.1f from the stem, want between %.1f and %.1f (concave)", o, 0.1*p.BracketReach, 0.4*p.BracketReach)
	}

	// The bracket meets the flat top of the serif at BracketReach.
	b, tp := lowest(t, "beyond the bracket", path.CrossingsY(x+side*(p.BracketReach+2)))
	near(t, "serif bottom", b, 0)
	near(t, "serif top", tp, rise)
}

// archSamples samples the arch between two stems: every half unit between the
// stems, the bottom and the top of the topmost run of the vertical scan, where
// that run lies above 0.3 of the x-height.  The stems are given by the
// right-hand edge l1 of the left one and the left-hand edge r0 of the right
// one.
func archSamples(path outline.Path, xh, l1, r0 float64) (top, bot []outline.Point) {
	for x := l1 + 0.5; x < r0-0.25; x += 0.5 {
		ys := path.CrossingsY(x)
		if len(ys) >= 2 && ys[len(ys)-2] > 0.3*xh {
			bot = append(bot, outline.Point{X: x, Y: ys[len(ys)-2]})
			top = append(top, outline.Point{X: x, Y: ys[len(ys)-1]})
		}
	}
	return top, bot
}

// edgeThickness returns, for each sample of one edge of a stroke, the distance
// to the nearest point of the polyline of the other edge: the thickness of the
// stroke at right angles.
func edgeThickness(a, b []outline.Point) []float64 {
	d := make([]float64, len(a))
	for i, q := range a {
		d[i] = math.Inf(1)
		for j := 0; j+1 < len(b); j++ {
			d[i] = math.Min(d[i], segDist(q, b[j], b[j+1]))
		}
	}
	return d
}

// notchHeight returns the height of the notch of an arch, where its outer edge
// meets the stem with the right-hand edge at x: the top of the outer edge 2 and
// 5 units to the right of the stem, extended linearly back to the stem, and at
// most the height cap.
func notchHeight(path outline.Path, x, cap float64) float64 {
	top := func(dx float64) float64 { c := path.CrossingsY(x + dx); return c[len(c)-1] }
	y2, y5 := top(2), top(5)
	return math.Min(cap, y2-(y5-y2)*2/3)
}

// joinHeight returns the height of the join of an arch, where its underside
// leaves the stem with the right-hand edge at x: the lowest height, going up
// from 0.3 of the x-height, at which the run of ink through the stem extends
// more than 4 units beyond the stem edge, that is at which the white to its
// right lies more than 4 units from the stem.
func joinHeight(path outline.Path, xh, x float64) float64 {
	for y := 0.3 * xh; y < xh; y += 0.25 {
		xs := path.CrossingsX(y)
		for i := 0; i+1 < len(xs); i += 2 {
			if xs[i] < x-1 && xs[i+1] > x-1 && xs[i+1] > x+4 {
				return y
			}
		}
	}
	return math.NaN()
}

// peakOf returns the position and the height of the highest point of the
// samples, the position as the centre of the part within half a unit of it.
func peakOf(top []outline.Point) (x, y float64) {
	for _, q := range top {
		y = math.Max(y, q.Y)
	}
	var sum, cnt float64
	for _, q := range top {
		if q.Y >= y-0.5 {
			sum += q.X
			cnt++
		}
	}
	return sum / cnt, y
}
