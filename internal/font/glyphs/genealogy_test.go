// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"

	"seehuhn.de/go/bedern/internal/outline"
)

// inkMirrored says whether the ink of the path is symmetric about the vertical
// line at cx, within a unit, on the horizontal scans at n heights between y0
// and y1; it returns the first height at which it is not.
func inkMirrored(path outline.Path, cx, y0, y1 float64, n int) (float64, bool) {
	for i := range n {
		y := y0 + (y1-y0)*(float64(i)+0.5)/float64(n)
		rs := runs(path, y)
		for k := range rs {
			m := rs[len(rs)-1-k]
			if math.Abs(rs[k][0]+m[1]-2*cx) > 1 {
				return y, false
			}
		}
	}
	return 0, true
}

// contourKinds returns the numbers of counter-clockwise and of clockwise
// contours of the path.
func contourKinds(path outline.Path) (outer, inner int) {
	for _, ct := range path {
		if ct.SignedArea() > 0 {
			outer++
		} else {
			inner++
		}
	}
	return outer, inner
}

// TestBorn asks, in every parameter set, that the sign for a birth is one
// contour which is symmetric about the middle of its ink, has the advance of
// the figures and stands in its middle and on the maths axis within a unit;
// that it has five points, of which a scan meets one near the top, three at
// most at one height and two near the foot; and that the upper point ends
// in a flat between half a thin stroke and a stem wide.
func TestBorn(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, adv := cachedOutline(np, "born")
			_, zero := cachedOutline(np, "zero")
			x0, y0, x1, y1 := path.BBox()
			if len(path) != 1 || adv != zero {
				t.Errorf("the star has %d contours and the advance %v, want 1 and %v", len(path), adv, zero)
			}
			if math.Abs((x0+x1)/2-adv/2) > 1 || math.Abs((y0+y1)/2-p.Punct.MathAxis) > 1 {
				t.Errorf("the middle of the star is (%v, %v), want (%v, %v)", (x0+x1)/2, (y0+y1)/2, adv/2, p.Punct.MathAxis)
			}
			if y, ok := inkMirrored(path, (x0+x1)/2, y0, y1, 12); !ok {
				t.Errorf("the star is not symmetric at the height %.0f", y)
			}
			top, most, foot := 0, 0, 0
			for i := range 60 {
				f := (float64(i) + 0.5) / 60
				n := len(runs(path, y0+f*(y1-y0)))
				most = max(most, n)
				if f > 0.85 {
					top = max(top, n)
				}
				if f < 0.15 {
					foot = max(foot, n)
				}
			}
			if top != 1 || most != 3 || foot != 2 {
				t.Errorf("the scans meet %d points near the top, at most %d and %d near the foot, want 1, 3 and 2", top, most, foot)
			}
			if rs := runs(path, y1-0.5); len(rs) != 1 || rs[0][1]-rs[0][0] < 0.5*p.ThinLower || rs[0][1]-rs[0][0] > p.Stem {
				t.Errorf("the end of the upper point is %v, want one run between %v and %v wide", rs, 0.5*p.ThinLower, p.Stem)
			}
		})
	}
}

// TestDied asks, in every parameter set, that the sign for a death is one
// contour which is symmetric about the middle of its ink, stands on the
// baseline and reaches the cap height; that its arms lie between 0.55 and
// 0.75 of its height; and that the stem is wider at its foot and at its top,
// and the arms are thicker at their ends, than just beside the crossing.
func TestDied(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, _ := cachedOutline(np, "died")
			x0, y0, x1, y1 := path.BBox()
			if len(path) != 1 || math.Abs(y0) > 0.5 || math.Abs(y1-p.CapHeight) > 0.5 {
				t.Errorf("the cross has %d contours and stands from %v to %v, want 1 from 0 to %v", len(path), y0, y1, p.CapHeight)
			}
			cx := (x0 + x1) / 2
			if y, ok := inkMirrored(path, cx, y0, y1, 12); !ok {
				t.Errorf("the cross is not symmetric at the height %.0f", y)
			}
			arm := vruns(path, x0+0.5)
			if len(arm) != 1 || arm[0][0] < y0+0.55*(y1-y0) || arm[0][1] > y0+0.75*(y1-y0) {
				t.Fatalf("the end of an arm is %v, want one run between 0.55 and 0.75 of the height", arm)
			}
			width := func(y float64) float64 {
				for _, r := range runs(path, y) {
					if r[0] < cx && cx < r[1] {
						return r[1] - r[0]
					}
				}
				return 0
			}
			below, above := arm[0][0]-10, arm[0][1]+10
			if foot, waist := width(y0+0.5), width(below); foot < 1.2*waist {
				t.Errorf("the stem is %.1f wide at its foot and %.1f below the arms", foot, waist)
			}
			if top, waist := width(y1-0.5), width(above); top < 1.1*waist {
				t.Errorf("the stem is %.1f wide at its top and %.1f above the arms", top, waist)
			}
			inner := vruns(path, cx-width(below))
			if len(inner) != 1 || arm[0][1]-arm[0][0] < 1.05*(inner[0][1]-inner[0][0]) {
				t.Errorf("an arm is %v at its end and %v beside the stem", arm, inner)
			}
		})
	}
}

// TestMarriedAndDivorced asks, in every parameter set, that the sign for a
// marriage is one outer contour with three counters, of which the middle one
// is at least half a stroke wide, and that the sign for a divorce is two
// rings, each an outer contour with a counter, and a bar; that both are
// symmetric about the middle of their ink, have the middle of their ink on
// the maths axis within a unit and rings as high as the x-height within a
// tenth, whose stroke at the side is that of the minus sign; that the bar
// is at least one and a half times as high as the ovals and keeps at least
// three quarters of a stroke of white from each; and that the ovals are
// narrower than high.
func TestMarriedAndDivorced(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			rule, axis := p.Punct.MathThick, p.Punct.MathAxis

			m, _ := cachedOutline(np, "uni26AD")
			x0, y0, x1, y1 := m.BBox()
			if outer, inner := contourKinds(m); outer != 1 || inner != 3 {
				t.Fatalf("the rings have %d outer contours and %d counters, want 1 and 3", outer, inner)
			}
			if y, ok := inkMirrored(m, (x0+x1)/2, y0, y1, 12); !ok || math.Abs((y0+y1)/2-axis) > 1 {
				t.Errorf("the rings are not symmetric at %.0f, or their middle %v is not on the axis %v", y, (y0+y1)/2, axis)
			}
			if h := y1 - y0; h < 0.9*p.XHeight || h > 1.1*p.XHeight {
				t.Errorf("the rings are %v high, and the x-height is %v", h, p.XHeight)
			}
			mid := runs(m, (y0+y1)/2)
			if len(mid) != 4 || math.Abs(mid[0][1]-mid[0][0]-rule) > 0.5 {
				t.Fatalf("the rings at their middle are %v, want four runs, the first %v wide", mid, rule)
			}
			if lens := mid[2][0] - mid[1][1]; lens < rule/2 {
				t.Errorf("the middle counter of the rings is %.1f wide, want at least %v", lens, rule/2)
			}

			d, _ := cachedOutline(np, "uni26AE")
			dx0, dy0, dx1, dy1 := d.BBox()
			if outer, inner := contourKinds(d); outer != 3 || inner != 2 {
				t.Fatalf("the sign for a divorce has %d outer contours and %d counters, want 3 and 2", outer, inner)
			}
			if y, ok := inkMirrored(d, (dx0+dx1)/2, dy0, dy1, 12); !ok || math.Abs((dy0+dy1)/2-axis) > 1 {
				t.Errorf("the sign for a divorce is not symmetric at %.0f, or its middle %v is not on the axis %v", y, (dy0+dy1)/2, axis)
			}
			rs := runs(d, (dy0+dy1)/2)
			if len(rs) != 5 {
				t.Fatalf("the sign for a divorce at its middle is %v, want five runs", rs)
			}
			oval := vruns(d, (rs[0][0]+rs[1][1])/2)
			if len(oval) != 2 {
				t.Fatalf("the scan through an oval has %d runs, want 2", len(oval))
			}
			oh, ow := oval[1][1]-oval[0][0], rs[1][1]-rs[0][0]
			if math.Abs(oh-(y1-y0)) > 1 || ow >= oh || math.Abs(rs[0][1]-rs[0][0]-rule) > 0.5 {
				t.Errorf("an oval is %v by %v with a side of %v, want the height %v of the rings, less wide, with a side of %v", ow, oh, rs[0][1]-rs[0][0], y1-y0, rule)
			}
			if bar := dy1 - dy0; bar < 1.5*oh {
				t.Errorf("the bar is %v high, and the ovals %v", bar, oh)
			}
			if w := rs[2][0] - rs[1][1]; w < 0.75*rule {
				t.Errorf("the white between an oval and the bar is %v, want at least %v", w, 0.75*rule)
			}
		})
	}
}

// TestLeaf asks, in every parameter set, that the leaf is one outer contour
// with six counters, none smaller than three squares of the stroke of the
// leaf; that its leftmost ink lies in the lower third of its height and its
// rightmost ink in the upper third; and that its edge, measured where an
// upright scan through the middle of its width and a level scan through the
// middle of its height leave the ink, is between 0.9 and 1.8 strokes thick.
func TestLeaf(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			path, _ := cachedOutline(np, "leaf")
			stroke := math.Round(leafStroke * np.params.Punct.MathThick)
			if outer, inner := contourKinds(path); outer != 1 || inner != 6 {
				t.Fatalf("the leaf has %d outer contours and %d counters, want 1 and 6", outer, inner)
			}
			for i, ct := range path {
				if a := -ct.SignedArea(); a > 0 && a < 3*stroke*stroke {
					t.Errorf("the counter %d of the leaf has the area %.0f, less than %v", i, a, 3*stroke*stroke)
				}
			}
			x0, y0, x1, y1 := path.BBox()
			left, right := vruns(path, x0+1), vruns(path, x1-1)
			if len(left) != 1 || len(right) != 1 || left[0][1] > y0+(y1-y0)/3 || right[0][0] < y1-(y1-y0)/3 {
				t.Errorf("the tips of the leaf are %v and %v, want the left one in the lower and the right one in the upper third of %v to %v", left, right, y0, y1)
			}
			up, level := vruns(path, (x0+x1)/2), runs(path, (y0+y1)/2)
			for _, r := range [][2]float64{up[0], up[len(up)-1], level[0], level[len(level)-1]} {
				if w := r[1] - r[0]; w < 0.9*stroke || w > 1.8*stroke {
					t.Errorf("the edge of the leaf is %.1f thick along a scan, want from %v to %v", w, 0.9*stroke, 1.8*stroke)
				}
			}
		})
	}
}

// TestMusicalNote asks, in every parameter set, that the note is one
// contour from the overshoot below the baseline to the cap height; that its
// stem is a hairline wide and stands with its right edge on the rightmost
// ink of the head, within half a unit; that the head is wider than high; that
// the flag reaches at least half the width of the head to the right of the
// stem and ends in the upper six tenths of the sign; and that the flag,
// measured on level scans, grows narrower from six tenths to a quarter of
// the way up from its end and to its end, which is between the clearance
// floor and 1.2 hairlines wide.
func TestMusicalNote(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, _ := cachedOutline(np, "musicalnote")
			x0, y0, x1, y1 := path.BBox()
			if len(path) != 1 || y0 > 0 || y0 < -p.OvershootLower-1 || math.Abs(y1-p.CapHeight) > 0.5 {
				t.Errorf("the note has %d contours and stands from %v to %v", len(path), y0, y1)
			}
			h := y1 - y0
			stem := runs(path, y0+0.42*h)
			if len(stem) != 1 || math.Abs(stem[0][1]-stem[0][0]-p.Hair) > 0.5 {
				t.Fatalf("the stem is %v, want one run of %v", stem, p.Hair)
			}
			// The head: the ink below the free part of the stem.
			headRight, headTop := x0, y0
			for y := y0 + 1; y < y0+0.4*h; y++ {
				rs := runs(path, y)
				if len(rs) == 1 && rs[0][0] < stem[0][0]-1 {
					headRight, headTop = math.Max(headRight, rs[0][1]), y
				}
			}
			if math.Abs(headRight-stem[0][1]) > 0.5 {
				t.Errorf("the head ends at %v, and the right edge of the stem is at %v", headRight, stem[0][1])
			}
			hw, hh := headRight-x0, headTop-y0
			if hw <= hh {
				t.Errorf("the head is %v wide and %v high", hw, hh)
			}
			if x1-stem[0][1] < hw/2 {
				t.Errorf("the flag reaches %v beyond the stem, less than half the head of %v", x1-stem[0][1], hw)
			}
			end := vruns(path, x1-1)
			if len(end) != 1 || end[0][0] < y0+0.4*h {
				t.Fatalf("the end of the flag is %v, want one run in the upper six tenths", end)
			}
			// The flag on level scans between its end and the top.
			wide := func(f float64) float64 {
				rs := runs(path, end[0][0]+f*(y1-end[0][0]))
				return rs[len(rs)-1][1] - rs[len(rs)-1][0]
			}
			tip, low, high := wide(0.002), wide(0.25), wide(0.6)
			if !(tip < low && low < high) {
				t.Errorf("the flag is %.1f wide at its end, %.1f a quarter and %.1f six tenths of the way up", tip, low, high)
			}
			if tip < clearanceFloor(p) || tip > 1.2*p.Hair {
				t.Errorf("the end of the flag is %.1f wide, want from %v to %v", tip, clearanceFloor(p), 1.2*p.Hair)
			}
		})
	}
}
