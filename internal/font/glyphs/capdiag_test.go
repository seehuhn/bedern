// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of the diagonal capitals A, V, W, X, Y, Z, K, M and N measure the
// outlines in every variant and perturbed parameter set: the overshoot of the
// pointed forms, the level tops, the weight of the strokes, the bar of A and the
// straight underside of the arm of K.

// The apex of A is a pointed top: it stands at or above the round top of O and at
// most twice the overshoot of O above the cap height.
func TestCapApexOvershootsMoreThanRound(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		_, _, _, ya := capOutline(np, "A").BBox()
		_, _, _, yo := capOutline(np, "O").BBox()
		if ya < yo-0.5 || ya-cap > 2*(yo-cap)+0.5 {
			t.Errorf("the apex of A is %.1f above the cap height, the top of O %.1f; want at least as much and at most twice as much", ya-cap, yo-cap)
		}
	})
}

// The apex of A overshoots the cap height and the vertices of V, W and N overshoot the
// baseline, each by at least 1 per cent of the cap height; the vertex of M stays
// above the baseline; the flat tops of the other letters stand on the cap height.
func TestCapDiagOvershoot(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		min := 0.01 * cap
		_, _, _, ya := capOutline(np, "A").BBox()
		if ya-cap < min {
			t.Errorf("the apex of A is %.1f above the cap height, want at least %.1f", ya-cap, min)
		}
		for _, g := range []string{"V", "W", "N"} {
			_, y0, _, y1 := capOutline(np, g).BBox()
			if -y0 < min {
				t.Errorf("%s reaches %.1f below the baseline, want at least %.1f", g, -y0, min)
			}
			if math.Abs(y1-cap) > 0.5 {
				t.Errorf("%s has its top at %.1f, want the cap height %.1f", g, y1, cap)
			}
		}
		_, y0, _, y1 := capOutline(np, "M").BBox()
		if y0 != 0 || math.Abs(y1-cap) > 0.5 {
			t.Errorf("M spans %.1f to %.1f, want 0 to %.1f", y0, y1, cap)
		}
		for _, g := range []string{"X", "Y", "Z", "K"} {
			_, y0, _, y1 := capOutline(np, g).BBox()
			if math.Abs(y0) > 0.5 || math.Abs(y1-cap) > 0.5 {
				t.Errorf("%s spans %.1f to %.1f, want 0 to %.1f", g, y0, y1, cap)
			}
		}
	})
}

// The thick diagonal of V is thicker than the thin one by at least a factor of 1.6
// and the thick diagonals of the letters are not heavier than the stem of H,
// measured horizontally at a quarter of the cap height above the vertex.
func TestCapDiagStrokeWeights(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		r := runs(capOutline(np, "V"), 0.75*cap)
		if len(r) != 2 {
			t.Fatalf("V has %d runs at 0.75 cap, want 2", len(r))
		}
		thick, thin := r[0][1]-r[0][0], r[1][1]-r[1][0]
		if thick < 1.6*thin {
			t.Errorf("the thick diagonal of V is %.1f wide, the thin one %.1f: want a ratio of at least 1.6", thick, thin)
		}
		lean := design.CapVLean(np.params)
		if thick/math.Sqrt(1+lean*lean) > np.params.Caps.Stem+1e-6 {
			t.Errorf("the thick diagonal of V is %.1f thick at right angles, more than the stem %.1f", thick/math.Sqrt(1+lean*lean), np.params.Caps.Stem)
		}
	})
}

// The crossbar of A is the thickness of the bar of H and stands below the middle
// of the cap height.
func TestCapABar(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		cap := capHeight(np)
		a := capOutline(np, "A")
		x0, _, x1, _ := a.BBox()
		ys := vruns(a, (x0+x1)/2)
		var bar [2]float64
		for _, r := range ys {
			if r[0] > 0.2*cap && r[1] < 0.5*cap {
				bar = r
			}
		}
		if bar == [2]float64{} {
			t.Fatalf("no crossbar found in %v", ys)
		}
		near(t, "bar thickness", bar[1]-bar[0], np.params.Caps.Bar)
		if c := (bar[0] + bar[1]) / 2 / cap; c < 0.30 || c > 0.45 {
			t.Errorf("the bar of A is centred at %.3f of the cap height, want 0.30 to 0.45", c)
		}
	})
}

// The underside of the arm of K runs from the stem to the leg as one straight line.
func TestCapKArmUndersideIsStraight(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		p := np.params
		k := capOutline(np, "K")
		arm, leg := capKDiagonals(p)
		right := design.CapKStemLeft(p) + p.Caps.Stem
		p2 := edgeMeet(arm, edgeRight, leg, edgeLeft)
		// The lowest ink above the white triangle at three places between the
		// stem and the leg.
		var pts [3][2]float64
		for i := range pts {
			x := right + 2 + (p2.X-right-4)*float64(i)/2
			ys := vruns(k, x)
			found := false
			for _, r := range ys {
				if r[0] > 0.2*p.CapHeight && r[0] < 0.6*p.CapHeight {
					pts[i] = [2]float64{x, r[0]}
					found = true
				}
			}
			if !found {
				t.Fatalf("no underside of the arm at x = %.1f: %v", x, ys)
			}
		}
		// The middle point lies on the line through the outer two.
		want := pts[0][1] + (pts[2][1]-pts[0][1])*(pts[1][0]-pts[0][0])/(pts[2][0]-pts[0][0])
		if math.Abs(pts[1][1]-want) > 0.5 {
			t.Errorf("the underside of the arm is %.2f off a straight line", pts[1][1]-want)
		}
	})
}

// footInk returns the ink of the rightmost foot of the outline below the height h, the
// area of the last run of each scan line, in square units.
func footInk(path outline.Path, h float64) float64 {
	var area float64
	for y := 0.5; y < h; y++ {
		xs := path.CrossingsX(y)
		if n := len(xs); n >= 2 {
			area += xs[n-1] - xs[n-2]
		}
	}
	return area
}

// TestCapKFootIsLighterThanTheFootOfX asks that the foot serif of the leg of K, with
// its bracket, carries no more ink than the foot of the thick diagonal of X, the
// leg being KLeg thick against XThick, and at least 0.9 of it.
func TestCapKFootIsLighterThanTheFootOfX(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		h := 3 * np.params.Caps.SerifThick
		k := footInk(capOutline(np, "K"), h)
		x := footInk(capOutline(np, "X"), h)
		if r := k / x; r > 1 || r < 0.9 {
			t.Errorf("the foot of the leg of K has %.3f of the ink of the foot of X below %.0f, want 0.9 to 1", r, h)
		}
	})
}
