// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"
)

// TestOldStyleSixAndEightOvershoot checks that the round bottoms of the
// old-style 6 and 8 overshoot the baseline as those of 0 and o do.
func TestOldStyleSixAndEightOvershoot(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		bottom := func(name string) float64 {
			_, y0, _, _ := capOutline(np, name).BBox()
			return y0
		}
		oBottom := bottom("o")
		for _, n := range []string{"zero", "six", "eight"} {
			if got := bottom(n + ".taboldstyle"); math.Abs(got-oBottom) > 0.5 {
				t.Errorf("%s: the bottom of the old-style %s is at %.1f, that of o at %.1f", np.name, n, got, oBottom)
			}
		}
	})
}
