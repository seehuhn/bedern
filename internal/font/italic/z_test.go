// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestZForm asks that z is the cursive z: its top stroke
// starts in an upturn below the x-height (the leftmost ink 0.65 to 0.82 of
// the x-height), its bottom stroke dips
// below the baseline no deeper than 0.05 of the x-height and its ink rises
// no higher than the top of o; the diagonal at ShearHeight is one run 0.5 to
// 1.0 stems wide; the bars are at least half a
// stem thick, measured vertically through the middle of the diagonal; and each bar is sharp at its corner with
// the diagonal: the outline turns there by more than 100 degrees within one
// unit.
func TestZForm(t *testing.T) {
	p := Default()
	g := glyph(t, "z")
	_, y0, _, y1 := bbox(g.Path)
	if y0 < -0.05*p.XHeight || y0 > -0.5 || y1 > p.XHeight+p.Overshoot+0.5 {
		t.Errorf("z: the ink runs from %.1f to %.1f, want from below 0 to no deeper than %.1f, and up to %v at most", y0, y1, -0.05*p.XHeight, p.XHeight+p.Overshoot)
	}
	// The upturn: the leftmost ink of the top half lies below the x-height
	// by at least 0.15 of it.
	lx, ly := math.Inf(1), 0.0
	for y := p.XHeight / 2; y <= p.XHeight; y++ {
		if rs := runs(g.Path, y); len(rs) > 0 && rs[0][0] < lx {
			lx, ly = rs[0][0], y
		}
	}
	if ly > 0.85*p.XHeight {
		t.Errorf("z: the leftmost ink of the top stroke lies at %.0f, want at most %.0f", ly, 0.85*p.XHeight)
	}
	rs := runs(g.Path, p.ShearHeight)
	if len(rs) != 1 {
		t.Fatalf("z: %d runs at %v, want 1", len(rs), p.ShearHeight)
	}
	if w := rs[0][1] - rs[0][0]; w < 0.5*p.Stem || w > 1.0*p.Stem {
		t.Errorf("z: the diagonal is %.1f wide at %v, want 0.5 to 1.0 stems", w, p.ShearHeight)
	}
	mid := (rs[0][0] + rs[0][1]) / 2
	ys := crossings(g.Path, Point{X: mid, Y: 0}, Point{Y: 1})
	if len(ys) != 6 {
		t.Fatalf("z: the vertical line at %.0f crosses the outline %d times, want 6", mid, len(ys))
	}
	for i, name := range []string{"bottom", "top"} {
		if th := ys[4*i+1] - ys[4*i]; th < 0.5*p.Stem {
			t.Errorf("z: the %s bar is %.1f thick at %.0f, want at least half a stem", name, th, mid)
		}
	}
	for _, c := range []Point{{X: p.ZTopRight, Y: p.XHeight}, {X: p.ZBotLeft, Y: 0}} {
		if turn := cornerTurn(g, c); turn < 100 {
			t.Errorf("z: the corner at %v turns by %.0f degrees, want more than 100", c, turn)
		}
	}
}

// cornerTurn returns the turn of the outline, in degrees, at its node
// nearest to c, between the segments before and after it.
func cornerTurn(g *Glyph, c Point) float64 {
	best, turn := math.Inf(1), 0.0
	for _, ct := range g.Path {
		n := len(ct.Segs)
		at := func(i int) Point {
			i = (i%n + n) % n
			if i == 0 {
				return ct.Start
			}
			return ct.Segs[i-1].P3
		}
		for i := 0; i < n; i++ {
			q := at(i)
			d := math.Hypot(q.X-c.X, q.Y-c.Y)
			if d >= best {
				continue
			}
			// The direction into q and out of it, from the handles.
			in := ct.Segs[(i-1+n)%n]
			from := at(i - 1)
			if in.IsCurve {
				from = in.P2
			}
			out := ct.Segs[i]
			to := out.P3
			if out.IsCurve {
				to = out.P1
			}
			a1 := math.Atan2(q.Y-from.Y, q.X-from.X)
			a2 := math.Atan2(to.Y-q.Y, to.X-q.X)
			da := math.Abs(math.Remainder(a2-a1, 2*math.Pi)) * 180 / math.Pi
			best, turn = d, da
		}
	}
	if best > 1.5 {
		return 0
	}
	return turn
}

// TestZLeans asks that z leans with the italic: its ink at 0.88 of the
// x-height lies on average at least 0.06 of the height to the right of its
// ink at 0.12 of it.
func TestZLeans(t *testing.T) {
	p := Default()
	g := glyph(t, "z")
	edges := func(y float64) (float64, float64) {
		rs := runs(g.Path, y)
		return rs[0][0], rs[len(rs)-1][1]
	}
	l0, r0 := edges(0.12 * p.XHeight)
	l1, r1 := edges(0.88 * p.XHeight)
	if lean := ((l1 - l0) + (r1 - r0)) / 2 / (0.76 * p.XHeight); lean < 0.06 {
		t.Errorf("z: the lean of the edges is %.2f, want at least 0.06", lean)
	}
}
