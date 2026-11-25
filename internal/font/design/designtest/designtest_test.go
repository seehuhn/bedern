// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package designtest

import (
	"slices"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// TestRoutineOrderNamesEverySet checks that routineOrder names every perturbed
// set exactly once and that Sets returns the first n of them in that order.
func TestRoutineOrderNamesEverySet(t *testing.T) {
	var names []string
	for _, s := range Perturbed() {
		names = append(names, s.Name)
	}
	order := slices.Clone(routineOrder)
	slices.Sort(names)
	slices.Sort(order)
	if !slices.Equal(names, order) {
		t.Errorf("the perturbed sets are\n%v\nbut routineOrder holds\n%v", names, order)
	}
	if got := Sets(true, 3); len(got) != 3 || got[0].Name != routineOrder[0] || got[2].Name != routineOrder[2] {
		t.Errorf("Sets(true, 3) = %d sets, want the first three of routineOrder", len(got))
	}
	t.Setenv(AllSetsVariable, "")
	if got := Sets(false, 3); len(got) != len(fullOrder) {
		t.Errorf("Sets(false, 3) has %d sets, want %d", len(got), len(fullOrder))
	}
	t.Setenv(AllSetsVariable, "1")
	if got := Sets(false, 3); len(got) != len(names) {
		t.Errorf("Sets(false, 3) with %s has %d sets, want all %d", AllSetsVariable, len(got), len(names))
	}
}

// Every perturbed set is valid, and differs from the regular parameters.
func TestPerturbedSetsAreValidAndDiffer(t *testing.T) {
	t.Parallel()
	reg := design.Regular()
	for _, s := range Perturbed() {
		if err := s.Params.Validate(); err != nil {
			t.Errorf("set %s: %v", s.Name, err)
		}
		if s.Params == reg {
			t.Errorf("set %s equals the regular parameters", s.Name)
		}
	}
}
