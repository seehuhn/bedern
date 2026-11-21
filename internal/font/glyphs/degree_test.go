// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import "testing"

// degreeNeighbourGap is the least distance between the degree sign and a
// following C or F or a preceding 0, 5 or 7 when the glyphs stand side by side.
const degreeNeighbourGap = 40.0

func TestDegreeKeepsDistanceFromNeighbours(t *testing.T) {
	t.Parallel()
	for _, np := range allParams() {
		deg, advDeg, _ := outlineOf("degree", np.params)
		for _, name := range []string{"C", "F"} {
			g, _, _ := outlineOf(name, np.params)
			d, _, _ := minDistanceBetween(deg, shiftedPath(g, advDeg))
			t.Logf("%s: degree then %s: %.1f", np.name, name, d)
			if d < degreeNeighbourGap {
				t.Errorf("%s: the degree sign is %.1f from a following %s, want at least %v", np.name, d, name, degreeNeighbourGap)
			}
		}
		for _, name := range []string{"zero", "five", "seven"} {
			g, adv, _ := outlineOf(name, np.params)
			d, _, _ := minDistanceBetween(g, shiftedPath(deg, adv))
			t.Logf("%s: %s then degree: %.1f", np.name, name, d)
			if d < degreeNeighbourGap {
				t.Errorf("%s: the degree sign is %.1f from a preceding %s, want at least %v", np.name, d, name, degreeNeighbourGap)
			}
		}
	}
}
