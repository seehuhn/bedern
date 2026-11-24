// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "testing"

// TestWForm checks the round w as v, and that both of its counters are at
// least 110 wide at 0.65 of the x-height, just below the ball which hangs
// in the second (the counter of n is
// 185).
func TestWForm(t *testing.T) {
	p := Default()
	g := glyph(t, "w")
	checkRoundDiagonal(t, "w", g.Advance)
	y := 0.65 * p.XHeight
	rs := runs(g.Path, y)
	if len(rs) != 3 {
		t.Fatalf("w: %d runs of ink at %.0f, want 3", len(rs), y)
	}
	for i := 0; i < 2; i++ {
		if gap := rs[i+1][0] - rs[i][1]; gap < 110 {
			t.Errorf("w: counter %d is %.1f wide at %.0f, want at least 110", i+1, gap, y)
		}
	}
}
