// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package greek

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// TestGlyphs checks that every letter builds, and that the glyphs hold the
// 25 code points U+03B1 to U+03C9 once each, with outlines and distinct names.
func TestGlyphs(t *testing.T) {
	gs, err := Glyphs(design.Regular())
	if err != nil {
		t.Fatal(err)
	}
	if len(gs) != 25 {
		t.Errorf("%d glyphs, want 25", len(gs))
	}
	runes := map[rune]bool{}
	names := map[string]bool{}
	for _, g := range gs {
		if runes[g.Rune] {
			t.Errorf("%U occurs twice", g.Rune)
		}
		if names[g.Name] {
			t.Errorf("the name %q occurs twice", g.Name)
		}
		runes[g.Rune], names[g.Name] = true, true
		if g.Name == "" || len(g.Path) == 0 || g.Advance <= 0 {
			t.Errorf("%U: name %q, %d segments, advance %g", g.Rune, g.Name, len(g.Path), g.Advance)
		}
	}
	for r := rune(0x3B1); r <= 0x3C9; r++ {
		if !runes[r] {
			t.Errorf("%U is missing", r)
		}
	}
}
