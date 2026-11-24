// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestLigatureBridges asks of the drawn ligatures of f, upright: the crest
// stands at the crest of f, 704; the bridge is one run of ink at 15 below
// the crest from the stem of f to the stem after it, at least the thin of
// the pen thick on the vertical halfway between the stems; and the white
// under the bridge, between the stems, is at least 100 wide at 600.
func TestLigatureBridges(t *testing.T) {
	p := Default()
	top := p.XHeight + p.AscenderEntryRise + p.Overshoot
	for name := range ligatures {
		up := unsheared(p, glyph(t, name).Path)
		if _, _, _, y1 := bbox(up); math.Abs(y1-top) > 1 {
			t.Errorf("%s: the crest stands at %.1f, want %v", name, y1, top)
		}
		c1 := p.FSideLeft + p.Stem/2
		pitch := p.LigPitch
		switch name {
		case "f_b", "f_h", "f_k":
			pitch = p.LigBowlPitch
		}
		c2 := c1 + pitch
		found := false
		for _, r := range runs(up, top-15) {
			if r[0] <= c1+p.Stem && r[1] >= c2-p.Stem {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no bridge from %.0f to %.0f at %.0f: %v", name, c1, c2, top-15, runs(up, top-15))
			continue
		}
		mid := (c1 + c2) / 2
		ys := crossings(up, Point{X: mid, Y: p.XHeight + 40}, Point{Y: 1})
		if len(ys) == 0 || ys[len(ys)-1]-ys[len(ys)-2] < p.Thin {
			t.Errorf("%s: the bridge is thinner than %.1f halfway between the stems (%v)", name, p.Thin, ys)
		}
		rs := runs(up, 600)
		gap := 0.0
		for k := 1; k < len(rs); k++ {
			if rs[k-1][1] < mid && rs[k][0] > mid {
				gap = rs[k][0] - rs[k-1][1]
			}
		}
		if gap < 100 {
			t.Errorf("%s: the white under the bridge is %.0f wide at 600, want at least 100", name, gap)
		}
	}
}
