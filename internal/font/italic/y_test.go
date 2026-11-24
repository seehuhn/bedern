// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import (
	"math"
	"testing"
)

// TestYIsUWithTail asks of y the tail of every tail (checkTail) under its
// right stem, and that above the baseline y is u with a narrower counter:
// its left stem is the stem of the italic down to the turn, and the turn
// overshoots the baseline as that of u does.
func TestYIsUWithTail(t *testing.T) {
	p := Default()
	e2 := p.SideStemLeft + p.Stem + p.YCounter
	checkTail(t, p, "y", e2)
	g := glyph(t, "y")
	for y := p.XHeight - p.YShoulder + 4.5; y < p.XHeight-p.EntryJoin; y += 8 {
		r := stemRun(t, p, g.Path, p.SideStemLeft+p.Stem/2, y)
		if w := r[1] - r[0]; math.Abs(w-p.Stem) > 1 {
			t.Errorf("the left stem of y is %.1f wide at %.0f, want %v", w, y, p.Stem)
		}
	}
	// The lowest ink of the turn, left of the right stem.
	up := unsheared(p, g.Path)
	low := math.Inf(1)
	for _, pts := range polys(up) {
		for _, q := range pts {
			if q.X < e2 && q.Y > -50 {
				low = math.Min(low, q.Y)
			}
		}
	}
	if math.Abs(low+p.Overshoot) > 2 {
		t.Errorf("the turn of y reaches %.1f, want about %v", low, -p.Overshoot)
	}
}
