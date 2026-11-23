// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"
)

// TestArrows asks, in every parameter set, that each arrow is one contour;
// that ← is the mirror image of → and ↓ is ↑ turned through half a turn about
// the middle of its ink; that the shaft of → has the stroke and the axis of
// the minus sign and runs unbroken from the tail to the tip; that the head
// is symmetric about the shaft and wider than three strokes; that the ends of
// the barbs keep at least a stroke of white from the shaft; that the level
// arrows have the advance of the em dash; and that the upright arrows are as
// long as the level ones, with the middle of their ink on the maths axis and
// in the middle of the advance of the figures, within a unit.
func TestArrows(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			q := np.params.Punct
			right, adv := cachedOutline(np, "arrowright")
			left, ladv := cachedOutline(np, "arrowleft")
			up, uadv := cachedOutline(np, "arrowup")
			down, dadv := cachedOutline(np, "arrowdown")
			for name, path := range map[string]int{"arrowright": len(right), "arrowleft": len(left), "arrowup": len(up), "arrowdown": len(down)} {
				if path != 1 {
					t.Errorf("%s has %d contours, want 1", name, path)
				}
			}
			if ladv != adv || !samePath(left, mirrorInk(right, adv/2)) {
				t.Error("the arrow to the left is not the mirror image of the arrow to the right")
			}
			_, em := cachedOutline(np, "emdash")
			if adv != em {
				t.Errorf("the level arrows have the advance %v, and the em dash %v", adv, em)
			}

			x0, y0, x1, y1 := right.BBox()
			if math.Abs((y0+y1)/2-q.MathAxis) > 0.01 {
				t.Errorf("the head stands from %v to %v, not about the axis %v", y0, y1, q.MathAxis)
			}
			if y1-y0 < 3*q.MathThick {
				t.Errorf("the head is %v high, less than three strokes of %v", y1-y0, q.MathThick)
			}
			minus, _ := cachedOutline(np, "minus")
			_, my0, _, my1 := minus.BBox()
			shaft := vruns(right, x0+(x1-x0)/4)
			if len(shaft) != 1 || math.Abs(shaft[0][0]-my0) > 0.01 || math.Abs(shaft[0][1]-my1) > 0.01 {
				t.Errorf("the shaft is %v, and the minus stands from %v to %v", shaft, my0, my1)
			}
			if axis := runs(right, q.MathAxis); len(axis) != 1 || axis[0][0] > x0+0.01 || axis[0][1] < x1-0.01 {
				t.Errorf("the ink on the axis is %v, want one run from %v to %v", axis, x0, x1)
			}
			// The ends of the barbs: the first vertical scan from the tail
			// which meets them.
			for x := x0; x < x1; x++ {
				rs := vruns(right, x+0.5)
				if len(rs) < 3 {
					continue
				}
				if len(rs) != 3 {
					t.Errorf("the scan at %v through the ends of the barbs has %d runs, want 3", x+0.5, len(rs))
				} else if lo, hi := rs[1][0]-rs[0][1], rs[2][0]-rs[1][1]; lo < q.MathThick || hi < q.MathThick || math.Abs(lo-hi) > 0.01 {
					t.Errorf("the ends of the barbs are %v and %v from the shaft, want equal and at least %v", lo, hi, q.MathThick)
				}
				break
			}

			_, zero := cachedOutline(np, "zero")
			ux0, uy0, ux1, uy1 := up.BBox()
			if uadv != zero || dadv != zero {
				t.Errorf("the upright arrows have the advances %v and %v, and the figures %v", uadv, dadv, zero)
			}
			if math.Abs((uy1-uy0)-(x1-x0)) > 0.01 {
				t.Errorf("the arrow up is %v long, and the arrow to the right %v", uy1-uy0, x1-x0)
			}
			if math.Abs((uy0+uy1)/2-q.MathAxis) > 1 || math.Abs((ux0+ux1)/2-uadv/2) > 1 {
				t.Errorf("the middle of the arrow up is (%v, %v), want (%v, %v)", (ux0+ux1)/2, (uy0+uy1)/2, uadv/2, q.MathAxis)
			}
			if !samePath(down, turned(up, (ux0+ux1)/2, (uy0+uy1)/2)) {
				t.Error("the arrow down is not the arrow up turned through half a turn")
			}
		})
	}
}

// TestRadical asks, in every parameter set, that the radical is one contour
// which reaches below the baseline and above the cap height; that its bar
// lies along its top, ends at the advance and has the stroke of the minus
// sign; that the tick is a thin stroke, between 0.8 hairlines and the stem,
// which stands above the lower quarter of the sign; and that the thick
// stroke is heavier across than the long stroke, with at least half a stroke
// of white between the two at three quarters of the height of the peak.
func TestRadical(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			p := np.params
			path, adv := cachedOutline(np, "radical")
			if len(path) != 1 {
				t.Fatalf("the radical has %d contours, want 1", len(path))
			}
			x0, y0, x1, y1 := path.BBox()
			if y0 >= 0 || y1 <= p.CapHeight {
				t.Errorf("the radical stands from %v to %v, want from below the baseline to above the cap height %v", y0, y1, p.CapHeight)
			}
			if math.Abs(x1-adv) > 0.5 {
				t.Errorf("the bar ends at %v, and the advance is %v", x1, adv)
			}
			bar := vruns(path, x1-1)
			if len(bar) != 1 || math.Abs(bar[0][1]-y1) > 0.01 || math.Abs(bar[0][1]-bar[0][0]-p.Punct.MathThick) > 0.01 {
				t.Errorf("the bar is %v, want one run of %v below the top %v", bar, p.Punct.MathThick, y1)
			}
			tick := vruns(path, x0+20)
			if len(tick) != 1 {
				t.Fatalf("the scan through the tick has %d runs, want 1", len(tick))
			}
			if w := tick[0][1] - tick[0][0]; w < 0.8*p.Hair || w > p.Stem {
				t.Errorf("the tick is %v thick upright, want from 0.8 hairlines of %v to the stem %v", w, p.Hair, p.Stem)
			}
			if tick[0][0] < y0+0.25*(y1-y0) {
				t.Errorf("the tick reaches down to %v, into the lower quarter of the sign", tick[0][0])
			}
			// The strokes below the peak, the highest node to the left of the
			// foot: the thick one and the long one.
			foot, peak := x1, y0
			for _, sg := range path[0].Segs {
				if sg.P3.Y < y0+0.01 {
					foot = math.Min(foot, sg.P3.X)
				}
			}
			for _, sg := range path[0].Segs {
				if sg.P3.X < foot {
					peak = math.Max(peak, sg.P3.Y)
				}
			}
			at := func(y float64) (thick, long [2]float64) {
				rs := runs(path, y)
				if len(rs) != 2 {
					t.Fatalf("the scan at %v has %d runs, want 2", y, len(rs))
				}
				return rs[0], rs[1]
			}
			ya, yb := y0+0.65*(peak-y0), y0+0.75*(peak-y0)
			ta, la := at(ya)
			tb, lb := at(yb)
			across := func(a, b [2]float64) float64 {
				return (a[1] - a[0]) * math.Abs(math.Sin(math.Atan2(yb-ya, b[0]-a[0])))
			}
			if th, lg := across(ta, tb), across(la, lb); th < 1.2*lg {
				t.Errorf("the thick stroke is %.1f across, and the long stroke %.1f", th, lg)
			}
			if w := lb[0] - tb[1]; w < p.Punct.MathThick/2 {
				t.Errorf("the white between the thick and the long stroke is %.1f at %v", w, yb)
			}
		})
	}
}

// TestCentredAsterisk asks, in every parameter set, that the centred asterisk
// is the asterisk moved down, with its advance, and that the middle of its
// ink lies on the maths axis within a unit.
func TestCentredAsterisk(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		t.Run(np.name, func(t *testing.T) {
			t.Parallel()
			ast, adv := cachedOutline(np, "asterisk")
			path, madv := cachedOutline(np, "asteriskmath")
			_, y0, _, y1 := path.BBox()
			_, a0, _, _ := ast.BBox()
			if madv != adv || !samePath(path, shiftInk(ast, 0, y0-a0)) {
				t.Error("the centred asterisk is not the asterisk moved down")
			}
			if axis := np.params.Punct.MathAxis; math.Abs((y0+y1)/2-axis) > 1 {
				t.Errorf("the middle of the centred asterisk is at %v, and the axis at %v", (y0+y1)/2, axis)
			}
		})
	}
}
