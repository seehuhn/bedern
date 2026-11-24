// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestEszettForm asks of ß, upright: the tail is the tail of f (checkTail)
// and the crest stands at the crest of f, 704; at a tenth of the x-height
// the stem, the end of the lower bowl and the right side of the bowl are
// three runs of ink, the end apart from the stem by at least the clearance
// floor (25); the upper
// bowl reaches less far to the right than the lower, as a 3 does; the
// waist, the least reach of the right edge between the two bowls in the
// italic, lies at 0.68 to 0.80 of the ink and at 0.75 to 0.93 of
// the x-height; and both counters are at
// least 25 wide at their middle height.
func TestEszettForm(t *testing.T) {
	p := Default()
	e := p.FSideLeft
	checkTail(t, p, "germandbls", e)
	g := glyph(t, "germandbls")
	up := unsheared(p, g.Path)
	_, _, _, y1 := bbox(up)
	if top := p.XHeight + p.AscenderEntryRise + p.Overshoot; math.Abs(y1-top) > 1 {
		t.Errorf("the crest of ß stands at %.1f, want %v", y1, top)
	}
	rs := runs(up, 0.1*p.XHeight)
	if len(rs) != 3 {
		t.Fatalf("ß has %d runs of ink at a tenth of the x-height, want 3", len(rs))
	}
	if gap := rs[1][0] - rs[0][1]; gap < 25 {
		t.Errorf("the end of the lower bowl of ß is %.1f from the stem, want at least 25", gap)
	}
	wy := p.SzWaistY * p.XHeight
	_, lower := inkRange(up, 0, wy-40)
	_, upper := inkRange(up, wy+40, y1-20)
	if upper >= lower {
		t.Errorf("the upper bowl of ß reaches %.0f, the lower %.0f, want the upper less", upper, lower)
	}
	// The waist is measured in the italic, as a share
	// of the whole ink.
	x0, _, x1, _ := bbox(g.Path)
	waistX, waistY := math.Inf(1), 0.0
	for y := 0.45 * p.XHeight; y < 0.95*y1; y += 2 {
		r := runs(g.Path, y)
		if x := r[len(r)-1][1]; x < waistX {
			waistX, waistY = x, y
		}
	}
	if f := (waistX - x0) / (x1 - x0); f < 0.68 || f > 0.80 {
		t.Errorf("the waist of ß lies at %.2f of its ink, want 0.68 to 0.80", f)
	}
	if f := waistY / p.XHeight; f < 0.75 || f > 0.93 {
		t.Errorf("the waist of ß lies at %.2f of the x-height, want 0.75 to 0.93", f)
	}
	for _, y := range []float64{(wy + y1) / 2, wy / 2} {
		r := runs(up, y)
		if len(r) < 2 {
			t.Errorf("ß has no counter at %.0f", y)
			continue
		}
		if w := r[len(r)-1][0] - r[len(r)-2][1]; w < 25 {
			t.Errorf("the counter of ß is %.1f wide at %.0f, want at least 25", w, y)
		}
	}
}
