// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// TestAnalysisScale checks that only the four glyphs which are scaled copies
// of a construction have a scale other than 1, so that every other glyph is
// analysed at its own size, and that the scale of each of the four is a
// reduction.
func TestAnalysisScale(t *testing.T) {
	scaled := map[string]bool{"ordfeminine": true, "ordmasculine": true, "copyright": true, "registered": true}
	sets := []namedParams{{name: "regular", params: design.Regular()}}
	if !testing.Short() {
		sets = append(sets, namedParams{name: "bold", params: design.Bold()})
	}
	for _, np := range sets {
		for _, g := range setOf(np.params) {
			s := AnalysisScale(np.params, g.Name)
			switch {
			case scaled[g.Name] && (s <= 0.4 || s >= 1):
				t.Errorf("%s %s: the scale is %.3f, want a reduction", np.name, g.Name, s)
			case !scaled[g.Name] && s != 1:
				t.Errorf("%s %s: the scale is %.3f, want 1", np.name, g.Name, s)
			}
		}
	}
}
