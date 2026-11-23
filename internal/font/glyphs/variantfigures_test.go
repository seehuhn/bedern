// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"
)

// The old-style figures stand in the zones of the lowercase: 0, 1 and 2
// between the baseline and the x-height, 3, 4, 5, 7 and 9 from the x-height to
// the depth of the descenders, 6 and 8 as high as the lining figures.  The
// round tops reach the top of o and the flat tops of 5 and 7 the top of x; the
// top of 4 lies between the two and that of 1 above the top of x by at most
// twice the overshoot of o.  The round bottom of 0 is that of o, 1 and 2 stand
// on the baseline, the round bottoms of 3, 5 and 9 reach the foot of p, and
// the feet of 4 and 7 stand above it by the overshoot of o.
func TestOldStyleFigureZones(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		box := func(name string) (y0, y1 float64) {
			_, y0, _, y1 = capOutline(np, name).BBox()
			return
		}
		old := func(name string) (y0, y1 float64) { return box(name + ".taboldstyle") }
		oBottom, oTop := box("o")
		_, xTop := box("x")
		pBottom, _ := box("p")
		for _, n := range []string{"zero", "two", "three", "nine"} {
			if _, top := old(n); math.Abs(top-oTop) > 0.5 {
				t.Errorf("%s: the top of the old-style %s is at %.1f, that of o at %.1f", np.name, n, top, oTop)
			}
		}
		for _, n := range []string{"five", "seven"} {
			if _, top := old(n); math.Abs(top-xTop) > 0.5 {
				t.Errorf("%s: the top of the old-style %s is at %.1f, that of x at %.1f", np.name, n, top, xTop)
			}
		}
		if _, top := old("four"); top < xTop-0.5 || top > oTop+0.5 {
			t.Errorf("%s: the top of the old-style 4 is at %.1f, want %.1f to %.1f", np.name, top, xTop, oTop)
		}
		if _, top := old("one"); top < xTop+0.5 || top > xTop+2*(oTop-xTop)+0.5 {
			t.Errorf("%s: the top of the old-style 1 is %.1f above the top of x, want more than 0 and at most %.1f", np.name, top-xTop, 2*(oTop-xTop))
		}
		if bot, _ := old("zero"); math.Abs(bot-oBottom) > 0.5 {
			t.Errorf("%s: the bottom of the old-style 0 is at %.1f, that of o at %.1f", np.name, bot, oBottom)
		}
		for _, n := range []string{"one", "two"} {
			if bot, _ := old(n); math.Abs(bot) > 0.5 {
				t.Errorf("%s: the bottom of the old-style %s is at %.1f, want the baseline", np.name, n, bot)
			}
		}
		for _, n := range []string{"four", "seven"} {
			if bot, _ := old(n); math.Abs(bot-(pBottom-oBottom)) > 0.5 {
				t.Errorf("%s: the foot of the old-style %s is at %.1f, want %.1f", np.name, n, bot, pBottom-oBottom)
			}
		}
		for _, n := range []string{"three", "five", "nine"} {
			if bot, _ := old(n); math.Abs(bot-pBottom) > 1 {
				t.Errorf("%s: the bottom of the old-style %s is at %.1f, that of p at %.1f", np.name, n, bot, pBottom)
			}
		}
		for _, n := range []string{"six", "eight"} {
			y0, y1 := old(n)
			l0, l1 := box(n)
			if y0 != l0 || y1 != l1 {
				t.Errorf("%s: the old-style %s reaches from %.1f to %.1f, the lining one from %.1f to %.1f", np.name, n, y0, y1, l0, l1)
			}
		}
	})
}

// The bar of the old-style 4 stands on the baseline: two units above the
// baseline the ink reaches both ends of the figure, and two units below it
// there is the stem alone, less than half as wide as the figure.
func TestOldStyleFourBarOnBaseline(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		path := capOutline(np, "four.taboldstyle")
		x0, _, x1, _ := path.BBox()
		above, below := runs(path, 2), runs(path, -2)
		if len(above) == 0 || above[0][0] > x0+5 || above[len(above)-1][1] < x1-0.5 {
			t.Errorf("%s: two units above the baseline the ink of the old-style 4 is %v, the figure reaches from %.1f to %.1f", np.name, above, x0, x1)
		}
		if len(below) != 1 || below[0][1]-below[0][0] > (x1-x0)/2 {
			t.Errorf("%s: two units below the baseline the ink of the old-style 4 is %v, want the stem alone", np.name, below)
		}
	})
}

// The tabular and the proportional form of a figure are one outline, moved
// sideways by a whole number of units, in the lining and in the old-style
// figures.  The tabular old-style figures have the advance of the lining
// ones, the proportional 1 is narrower than that, and every proportional
// figure has white on both sides.
func TestFigureFormsShareTheOutline(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		_, tabular := cachedOutline(np, "zero")
		for _, n := range figureNames {
			if _, adv := cachedOutline(np, n+".taboldstyle"); adv != tabular {
				t.Errorf("%s: the advance of %s.taboldstyle is %g, that of the lining figures %g", np.name, n, adv, tabular)
			}
			for _, pair := range [][2]string{{n, n + ".prop"}, {n + ".taboldstyle", n + ".oldstyle"}} {
				tab, prop := cachedRounded(np, pair[0]), cachedRounded(np, pair[1])
				a0, _, _, _ := tab.Path.BBox()
				b0, _, b1, _ := prop.Path.BBox()
				if !pathsEqual(shiftInk(tab.Path, b0-a0, 0), prop.Path) {
					t.Errorf("%s: %s is not %s moved sideways", np.name, pair[1], pair[0])
				}
				if prop.Advance != math.Round(prop.Advance) || b0 < 5 || prop.Advance-b1 < 5 {
					t.Errorf("%s: %s has the advance %g and the side spaces %.1f and %.1f, want a whole advance and at least 5 on each side", np.name, pair[1], prop.Advance, b0, prop.Advance-b1)
				}
			}
		}
		for _, n := range []string{"one.prop", "one.oldstyle"} {
			if _, adv := cachedOutline(np, n); adv >= tabular {
				t.Errorf("%s: the advance of %s is %g, want less than %g", np.name, n, adv, tabular)
			}
		}
	})
}

// figureWhite returns the white between the glyphs a and b set side by side,
// between the baseline and the height top: the mean, over ten slices, of the
// distance from the rightmost ink of a to the leftmost ink of b, capped at 250
// units.
func figureWhite(np namedParams, a, b string, top float64) float64 {
	pa, adv := cachedOutline(np, a)
	pb, _ := cachedOutline(np, b)
	const slices, limit = 10, 250.0
	sum := 0.0
	for i := 0; i < slices; i++ {
		y := top * (float64(i) + 0.5) / slices
		ra, rb := runs(pa, y), runs(pb, y)
		if len(ra) == 0 || len(rb) == 0 {
			sum += limit
			continue
		}
		sum += math.Min(limit, adv-ra[len(ra)-1][1]+rb[0][0])
	}
	return sum / slices
}

// The proportional figures fit the letters.  The white between n and an
// old-style figure, in the band of the x-height and relative to the white
// between two n, is 0.81 to 1.49 on the left of the figure and 0.74 to 1.35
// on its right; the white between H and a lining figure, in the band of the
// figures and relative to the white between two H, is 0.79 to 1.23 on the
// left and 0.71 to 1.23 on the right.  The 0 has less white beside it than the
// 1, on both sides.
func TestProportionalFiguresFitTheLetters(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		_, _, _, xh := capOutline(np, "x").BBox()
		_, _, _, fh := capOutline(np, "seven").BBox()
		for _, c := range []struct {
			suffix, stem   string
			top            float64
			left0, left1   float64
			right0, right1 float64
		}{
			{".oldstyle", "n", xh, 0.81, 1.49, 0.74, 1.35},
			{".prop", "H", fh, 0.79, 1.23, 0.71, 1.23},
		} {
			unit := figureWhite(np, c.stem, c.stem, c.top)
			var left, right [10]float64
			for d, n := range figureNames {
				left[d] = figureWhite(np, c.stem, n+c.suffix, c.top) / unit
				right[d] = figureWhite(np, n+c.suffix, c.stem, c.top) / unit
				if left[d] < c.left0 || left[d] > c.left1 {
					t.Errorf("%s: the white between %s and %s%s is %.2f of that between two %s, want %.2f to %.2f", np.name, c.stem, n, c.suffix, left[d], c.stem, c.left0, c.left1)
				}
				if right[d] < c.right0 || right[d] > c.right1 {
					t.Errorf("%s: the white between %s%s and %s is %.2f of that between two %s, want %.2f to %.2f", np.name, n, c.suffix, c.stem, right[d], c.stem, c.right0, c.right1)
				}
			}
			if left[0] >= left[1] || right[0] >= right[1] {
				t.Errorf("%s: the white beside zero%s is %.2f and %.2f, that beside one%s %.2f and %.2f; want less beside the 0", np.name, c.suffix, left[0], right[0], c.suffix, left[1], right[1])
			}
		}
	})
}

// The ball of the old-style 2 stands clear of the diagonal.  The vertical
// through the middle of the ball crosses the foot of the figure and, above
// it, the ball, and there are at least 60 units of white between the two.
func TestOldStyleTwoBallStandsClear(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		path := capOutline(np, "two.taboldstyle")
		f := np.params.Figures
		x0, _, _, _ := path.BBox()
		ys := path.CrossingsY(x0 + f.Old.TwoBallLeft + f.Old.TwoBall/2)
		sortFloats(ys)
		if len(ys) < 4 {
			t.Fatalf("%s: the vertical through the ball of the old-style 2 crosses the outline at %v, want the foot and the ball", np.name, ys)
		}
		if gap := ys[2] - ys[1]; gap < 60 {
			t.Errorf("%s: %.1f units of white below the ball of the old-style 2, want at least 60", np.name, gap)
		}
	})
}

// sortFloats sorts the values in increasing order.
func sortFloats(vs []float64) {
	for i := 1; i < len(vs); i++ {
		for j := i; j > 0 && vs[j] < vs[j-1]; j-- {
			vs[j], vs[j-1] = vs[j-1], vs[j]
		}
	}
}
