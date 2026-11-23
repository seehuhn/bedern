// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"testing"
)

// TestPrescriptionStrokeReachesBelowTheBaseline checks that the stroke of the
// prescription sign ends recipeBelow under the baseline, as far as the
// rounding of the outline allows, and that its leg ends in a flat cut on the
// baseline.
func TestPrescriptionStrokeReachesBelowTheBaseline(t *testing.T) {
	t.Parallel()
	forCaps(t, func(t *testing.T, np namedParams) {
		path := capOutline(np, "prescription")
		_, y0, _, _ := path.BBox()
		if math.Abs(y0+recipeBelow) > 2 {
			t.Errorf("%s: the lowest ink of the prescription sign is at %.1f, want %.0f", np.name, y0, -recipeBelow)
		}
		leg := capRLeg(np.params)
		// The plain leg of the sign has no foot serif: the ink on the baseline
		// right of the stem of R is the width of the leg and no wider.
		w := leg.edgeAt(edgeRight, 0).X - leg.edgeAt(edgeLeft, 0).X
		xs := path.CrossingsX(1)
		if len(xs) < 2 {
			t.Fatalf("%s: no crossings of the prescription sign at the baseline", np.name)
		}
		last := xs[len(xs)-1] - xs[len(xs)-2]
		if math.Abs(last-w) > 3 {
			t.Errorf("%s: the foot of the leg is %.1f wide at the baseline, want %.1f", np.name, last, w)
		}
	})
}
