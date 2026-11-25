// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import "testing"

// TestDoubleStruckLettersReview checks the runs of ink of the
// double-struck capitals: G has four runs of ink at a fifth of its height,
// the two lines of the doubled right side and the lens on the left, where C
// has three; Y has three at 0.7 of the cap height, the doubled left arm and
// the right arm.  The bend of U and of J is held by
// TestDoubleStruckBendsAreNotHeavy (dsrule_test.go).
func TestDoubleStruckLettersReview(t *testing.T) {
	s := derivedSets(0)[0]
	m := dsSet(s)
	cap := s.Params.CapHeight
	for _, c := range []struct {
		name string
		frac float64
		want int
	}{
		{"u1D53E", 0.2, 4}, {"uni2102", 0.2, 3}, {"u1D550", 0.7, 3},
	} {
		g, ok := m[c.name]
		if !ok {
			t.Fatalf("no %s", c.name)
		}
		if n := len(runsH(g.Path.Round(), c.frac*cap)); n != c.want {
			t.Errorf("%s has %d runs of ink at %.0f of the cap height, want %d", c.name, n, c.frac, c.want)
		}
	}
}
