// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"fmt"
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// The tests of the forms of single figures: the balls, the corner of 2, the
// flat of 4 and the round tops.

// The terminals of 2, 3, 5, 6 and 9 are one ball in one manner of attachment.
// Each figure carries as many circles of the diameter Figures.Ball as it has
// terminals (the part is that of the lowercase, scaled with its stalk).  The ball
// lies close under the curl of its own stroke: on the ray from its centre
// figBallWedgeRay degrees past the tip of the wedge, away from the stalk, the
// white between the ball and the stroke is figBallCurl[0] to figBallCurl[1]
// wide in every figure.  The white between the ball and any other ink, on every
// ray from figBallClearFrom degrees past the tip, is at least figBallClear.
const (
	figBallWedgeRay  = 40
	figBallClearFrom = 75
	figBallClear     = 60.0
)

var figBallCurl = [2]float64{55, 85}

func TestFigureBalls(t *testing.T) {
	t.Parallel()
	// Each set whose figures differ is measured once.
	seen := map[string]bool{}
	for _, np := range allParams() {
		p := np.params
		key := fmt.Sprint(p.Figures, p.BallStalk, p.BallDiameter, p.SupernessOuter, p.SupernessInner, p.OvershootUpper)
		if seen[key] {
			continue
		}
		seen[key] = true
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			testFigureBalls(t, np)
		})
	}
}

func testFigureBalls(t *testing.T, np namedParams) {
	r := np.params.Figures.Ball / 2
	{
		for _, c := range []struct {
			name string
			n    int
		}{{"two", 1}, {"three", 2}, {"five", 1}, {"six", 1}, {"nine", 1}} {
			balls := figureBalls(capOutline(np, c.name), r)
			if len(balls) != c.n {
				t.Errorf("%s: %s: %d circles of the diameter %v, want %d", np.name, c.name, len(balls), 2*r, c.n)
				continue
			}
			for _, b := range balls {
				if b.curl < figBallCurl[0] || b.curl > figBallCurl[1] {
					t.Errorf("%s: %s: the ball at (%.0f, %.0f) has %.1f of white under its curl, want %v to %v", np.name, c.name, b.centre.X, b.centre.Y, b.curl, figBallCurl[0], figBallCurl[1])
				}
				if b.clear < figBallClear {
					t.Errorf("%s: %s: the ball at (%.0f, %.0f) has only %.1f of white beyond it, want %v", np.name, c.name, b.centre.X, b.centre.Y, b.clear, figBallClear)
				}
			}
		}
	}
}

// A figureBall is a ball terminal found on an outline with the white round it.
type figureBall struct {
	centre outline.Point
	curl   float64 // the white figBallWedgeRay degrees past the tip of the wedge
	clear  float64 // the least white beyond the ball from figBallClearFrom degrees past the tip
}

// whiteRun returns the length of the white on the ray from c in the direction th
// (radians) which starts at the circle of the radius r, or -1 where the ray
// leaves the circle into ink (the stalk).
func whiteRun(path outline.Path, c outline.Point, r, th float64) float64 {
	return whiteRunIn(pathIndex(path), c, r, th)
}

// whiteRunIn is whiteRun on the index of the path.
func whiteRunIn(ix *bandIndex, c outline.Point, r, th float64) float64 {
	q := func(d float64) outline.Point { return outline.Point{X: c.X + d*math.Cos(th), Y: c.Y + d*math.Sin(th)} }
	d := r + 1
	if ix.leftOdd(q(d)) {
		return -1
	}
	for d < r+400 && !ix.leftOdd(q(d)) {
		d++
	}
	return d - r
}

// figureBalls finds the balls of the radius r on the outline and measures the
// white round each.  The tip of the wedge is the end of the stalk, seen from the
// centre, on the side where the white is narrow.
func figureBalls(path outline.Path, r float64) []figureBall {
	var out []figureBall
	ix := pathIndex(path)
	for _, c := range findBalls(path, r) {
		if math.IsNaN(c.X) || math.IsNaN(c.Y) {
			continue
		}
		var white [360]float64
		for a := range white {
			white[a] = whiteRunIn(ix, c, r, float64(a)*math.Pi/180)
		}
		at := func(a int) float64 { return white[((a%360)+360)%360] }
		tip, side, narrow := 0, 1, math.Inf(1)
		for a := 0; a < 360; a++ {
			for _, s := range []int{1, -1} {
				if at(a) < 0 && at(a+s) >= 0 && at(a+2*s) >= 0 && at(a+2*s) < narrow {
					tip, side, narrow = a, s, at(a+2*s)
				}
			}
		}
		b := figureBall{centre: c, curl: at(tip + figBallWedgeRay*side), clear: math.Inf(1)}
		for k := figBallClearFrom; k < 360 && at(tip+k*side) >= 0; k++ {
			b.clear = math.Min(b.clear, at(tip+k*side))
		}
		out = append(out, b)
	}
	return out
}

// The diagonal of 2 meets the top of the base in an inner corner which opens by
// at least twoCornerMin degrees; its lower edge comes down to the left end of
// the base, which stands above the baseline.
const twoCornerMin = 36.0

func TestFigureTwoCorner(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		f := np.params.Figures
		ct := capOutline(np, "two")[0]
		n := len(ct.Segs)
		found := false
		for i := 0; i < n; i++ {
			a, b := segCubic(ct, i), segCubic(ct, i+1)
			if ct.Segs[i].IsCurve || math.Abs(a.P0.Y-f.Heavy) > 0.5 || math.Abs(a.P3.Y-f.Heavy) > 0.5 || a.P3.X > a.P0.X {
				continue
			}
			// The top of the base, walked from the serif to the diagonal.
			found = true
			ta, tb := unit(a.P0, a.P3), b.tangent(0)
			turn := math.Atan2(ta.X*tb.Y-ta.Y*tb.X, ta.X*tb.X+ta.Y*tb.Y) * 180 / math.Pi
			if open := 180 - math.Abs(turn); open < twoCornerMin {
				t.Errorf("%s: the inner corner of 2 at (%.0f, %.0f) opens by %.1f degrees, want %v", np.name, a.P3.X, a.P3.Y, open, twoCornerMin)
			}
		}
		if !found {
			t.Errorf("%s: no inner corner on the top of the base of 2", np.name)
		}
		x0, _, _, _ := outline.Path{ct}.BBox()
		if !inkAt(outline.Path{ct}, outline.Point{X: x0 + 1, Y: f.TwoBaseEnd / 2}) {
			t.Errorf("%s: the base of 2 does not stand at its left end", np.name)
		}
	})
}

// The top of 4 is a flat, not a point: the ink at the top is fourFlat[0] to
// fourFlat[1] times the stem of 4 wide, 3 per cent below the top.
var fourFlat = [2]float64{0.6, 1.0}

func TestFigureFourFlat(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		f := np.params.Figures
		r := runs(capOutline(np, "four"), f.Height-0.5)
		if len(r) != 1 {
			t.Fatalf("%s: %d runs at the top of 4", np.name, len(r))
		}
		if w := (r[0][1] - r[0][0]) / f.Stem; w < fourFlat[0] || w > fourFlat[1] {
			t.Errorf("%s: the top of 4 is %.2f stems wide, want %v to %v", np.name, w, fourFlat[0], fourFlat[1])
		}
	})
}

// The round tops of 0, 8 and 9 and the round bottoms of 0, 6 and 8 show one
// roundness: the chord of the outer contour roundChordDepth below the top (above
// the bottom) is at most roundChordSpread times as long in one as in another, and
// the top of 9, the bottom of 6 turned, differs from the top of 8 by at most
// roundNineEight of it.
const (
	roundChordDepth  = 40.0
	roundChordSpread = 1.3
	roundNineEight   = 0.06
)

func TestFigureRoundTops(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		chord := func(name string, top bool) float64 {
			path := capOutline(np, name)
			_, y0, _, y1 := path.BBox()
			y := y0 + roundChordDepth
			if top {
				y = y1 - roundChordDepth
			}
			r := runs(path, y)
			return r[len(r)-1][1] - r[0][0]
		}
		c := map[string]float64{
			"top of 0": chord("zero", true), "top of 8": chord("eight", true), "top of 9": chord("nine", true),
			"bottom of 0": chord("zero", false), "bottom of 6": chord("six", false), "bottom of 8": chord("eight", false),
		}
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, v := range c {
			lo, hi = math.Min(lo, v), math.Max(hi, v)
		}
		if hi > roundChordSpread*lo {
			t.Errorf("%s: the chords of the round tops and bottoms are %.0f to %.0f: %v", np.name, lo, hi, c)
		}
		if d := math.Abs(c["top of 9"]-c["top of 8"]) / c["top of 8"]; d > roundNineEight {
			t.Errorf("%s: the top of 9 (%.0f) and of 8 (%.0f) differ by %.3f", np.name, c["top of 9"], c["top of 8"], d)
		}
	})
}
