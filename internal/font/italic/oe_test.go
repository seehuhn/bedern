// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "testing"

// TestSharedStroke asks of œ and æ, upright, at 0.6 and 0.7 of the
// x-height, in the eye of the e: three runs of ink, the left side of the first letter,
// the stroke which the two letters share and the right side of the e; the
// shared stroke 0.95 to 1.25 stems wide; and both counters at least 25
// wide.
func TestSharedStroke(t *testing.T) {
	p := Default()
	for _, name := range []string{"oe", "ae"} {
		up := unsheared(p, glyph(t, name).Path)
		for _, y := range []float64{0.6 * p.XHeight, 0.7 * p.XHeight} {
			rs := runs(up, y)
			if len(rs) != 3 {
				t.Errorf("%s: %d runs of ink at %.0f, want 3", name, len(rs), y)
				continue
			}
			if w := (rs[1][1] - rs[1][0]) / p.Stem; w < 0.95 || w > 1.25 {
				t.Errorf("%s: the shared stroke is %.2f stems wide at %.0f, want 0.95 to 1.25", name, w, y)
			}
			for k := 0; k < 2; k++ {
				if g := rs[k+1][0] - rs[k][1]; g < 25 {
					t.Errorf("%s: counter %d is %.1f wide at %.0f, want at least 25", name, k+1, g, y)
				}
			}
		}
	}
}
