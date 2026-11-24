// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// xRunAt returns the run of ink of x at the height y which contains the
// abscissa x, or false.
func xRunAt(g *Glyph, x, y float64) ([2]float64, bool) {
	for _, r := range runs(g.Path, y) {
		if r[0] <= x && x <= r[1] {
			return r, true
		}
	}
	return [2]float64{}, false
}

// TestXMeeting asks that the two arcs of x meet without a gap at
// ShearHeight, in one run of ink no wider than 1.4 stems, and that the run which they share there
// is at most 0.62 of the x-height long.
func TestXMeeting(t *testing.T) {
	p := Default()
	g := glyph(t, "x")
	mid := p.XWidth / 2
	r, ok := xRunAt(g, mid, p.ShearHeight+0.5)
	if !ok || len(runs(g.Path, p.ShearHeight+0.5)) != 1 {
		t.Fatalf("x: the arcs do not meet in one run at %v", p.ShearHeight)
	}
	if w := r[1] - r[0]; w > 1.4*p.Stem {
		t.Errorf("x: the meeting is %.1f wide, want at most %.1f", w, 1.4*p.Stem)
	}
	lo, hi := xSharedRun(g, p)
	if l := hi - lo; l > xSharedMax*p.XHeight {
		t.Errorf("x: the arcs share a run from %.0f to %.0f, want at most %.0f long", lo, hi, xSharedMax*p.XHeight)
	}
}

// xSharedMax is the longest run of ink which the two arcs of x may share, in
// x-heights.
const xSharedMax = 0.62

// xSharedRun returns the heights between which the two arcs of x are one run
// of ink, the run which reaches both sides of the vertical of the meeting,
// 20 units to either side.
func xSharedRun(g *Glyph, p Params) (lo, hi float64) {
	mid := p.XWidth / 2
	shared := func(y float64) bool {
		r, ok := xRunAt(g, mid, y)
		return ok && r[0] < mid-20 && r[1] > mid+20
	}
	lo, hi = p.ShearHeight, p.ShearHeight
	for shared(lo-1) && lo > 0 {
		lo--
	}
	for shared(hi+1) && hi < p.XHeight {
		hi++
	}
	return lo, hi
}

// TestXContrast asks that the stroke from the upper left to the lower right
// is the thick one: at 0.86 and 0.14 of the x-height, where the arcs have
// parted, the run of the thick arc next to the meeting is at least 1.4 times
// as wide as the run of the thin one.
func TestXContrast(t *testing.T) {
	p := Default()
	g := glyph(t, "x")
	for _, cs := range []struct {
		y          float64
		upperThick bool
	}{{0.86 * p.XHeight, true}, {0.14 * p.XHeight, false}} {
		// The two arcs are the runs on either side of the gap nearest to the
		// vertical of the meeting.
		xm := p.XWidth / 2
		rs := runs(g.Path, cs.y)
		best := -1
		for i := 0; i+1 < len(rs); i++ {
			d := math.Abs((rs[i][1]+rs[i+1][0])/2 - xm)
			if best < 0 || d < math.Abs((rs[best][1]+rs[best+1][0])/2-xm) {
				best = i
			}
		}
		if best < 0 {
			t.Fatalf("x: the arcs have not parted at %.0f", cs.y)
		}
		left, right := rs[best], rs[best+1]
		thick, thin := left, right
		if !cs.upperThick {
			thick, thin = right, left
		}
		if a, b := thick[1]-thick[0], thin[1]-thin[0]; a < 1.4*b {
			t.Errorf("x: at %.0f the thick arc is %.1f and the thin %.1f wide", cs.y, a, b)
		}
	}
}

// TestXEntry checks the entry of x as every entry at the x-height.
func TestXEntry(t *testing.T) {
	p := Default()
	xl := p.XWidth/2 - p.XMeetW/2 + p.XOverlap/2
	checkEntryOf(t, p, "x", xl-p.XTopIn-p.EntryTopShift, 0)
}

// TestXLeans asks that x leans as the italic does and is as wide as its
// neighbours: the column in which the two arcs meet, the run nearest the
// middle at 0.3 and 0.7 of the x-height, leans by at least 0.7 of the slant;
// the ink of x is at most 1.2 times as wide as that of y and at least as wide as that of v.
func TestXLeans(t *testing.T) {
	p := Default()
	g := glyph(t, "x")
	centre := func(y float64) float64 {
		best, bd := 0.0, math.Inf(1)
		for _, r := range runs(g.Path, y) {
			if d := math.Abs((r[0]+r[1])/2 - p.XWidth/2); d < bd {
				best, bd = (r[0]+r[1])/2, d
			}
		}
		return best
	}
	lean := (centre(0.7*p.XHeight) - centre(0.3*p.XHeight)) / (0.4 * p.XHeight)
	if lean < 0.7*p.slant() {
		t.Errorf("x: the meeting leans by %.2f, want at least %.2f", lean, 0.7*p.slant())
	}
	width := func(name string) float64 {
		x0, _, x1, _ := bbox(glyph(t, name).Path)
		return x1 - x0
	}
	if w, wv, wy := width("x"), width("v"), width("y"); w > 1.2*wy || w < wv {
		t.Errorf("x: the ink is %.0f wide, want from %.0f (v) to %.0f (1.2 y)", w, wv, 1.2*wy)
	}
}
