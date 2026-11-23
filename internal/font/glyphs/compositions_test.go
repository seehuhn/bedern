// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// TestCompositionsDescribeTheSet checks that every composed letter of
// Compositions is a glyph of the set with its base, and that its mark and
// every mark of MarkGlyphs has an outline with ink.
func TestCompositionsDescribeTheSet(t *testing.T) {
	p := design.Regular()
	ink := func(mark, form string) bool {
		m, ok := MarkOutline(p, mark, form)
		return ok && len(m) > 0
	}
	for _, c := range Compositions() {
		if _, ok := lookup(c.Name); !ok {
			t.Errorf("%s: not a glyph of the set", c.Name)
		}
		if _, ok := lookup(c.Base); !ok {
			t.Errorf("%s: no base %s", c.Name, c.Base)
		}
		if !ink(c.Mark, c.Form) {
			t.Errorf("%s: no outline of the mark %s in the form %s", c.Name, c.Mark, c.Form)
		}
	}
	for _, m := range MarkGlyphs() {
		if _, ok := lookup(m.Name); !ok {
			t.Errorf("%s: not a glyph of the set", m.Name)
		}
		if !ink(m.Mark, "lower") {
			t.Errorf("%s: no outline of the mark %s", m.Name, m.Mark)
		}
	}
	if _, ok := MarkOutline(p, "acute", "upright"); ok {
		t.Error("an unknown form gives an outline")
	}
}
