// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import (
	"math"
	"sync"
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
)

// greekLowerOnce builds the italic Greek lowercase once for the tests.
var greekLowerOnce = sync.OnceValues(greekLower)

func greekLowerGlyphs(t *testing.T) []Glyph {
	t.Helper()
	gs, err := greekLowerOnce()
	if err != nil {
		t.Fatal(err)
	}
	return gs
}

// TestGreekLowerCodePoints asks that every italic Greek letter stands at the
// code point of the mathematical italic which unicode-math sets for its
// command, under the name of that code point.
func TestGreekLowerCodePoints(t *testing.T) {
	if testing.Short() {
		t.Skip("the Greek is tested in the full tier")
	}
	for _, g := range greekLowerGlyphs(t) {
		ok := g.Rune >= 0x1D6FC && g.Rune <= 0x1D714 || g.Rune >= 0x1D716 && g.Rune <= 0x1D71B
		if !ok {
			t.Errorf("%s stands at %U, outside the mathematical italic Greek", g.Name, g.Rune)
		}
	}
}

// TestGreekLowerOverhang asks that no italic Greek letter reaches more than
// maxOverhangLeft before its origin or maxOverhangRight beyond its advance
// (a symbol beside a letter in a formula is not kerned).
func TestGreekLowerOverhang(t *testing.T) {
	if testing.Short() {
		t.Skip("the Greek is tested in the full tier")
	}
	for _, g := range greekLowerGlyphs(t) {
		x0, _, x1, _ := g.Path.BBox()
		if x0 < -maxOverhangLeft-1 || x1 > g.Advance+maxOverhangRight+1 {
			t.Errorf("%s: ink %.0f..%.0f on an advance of %.0f, an overhang beyond %d and %d", g.Name, x0, x1, g.Advance, maxOverhangLeft, maxOverhangRight)
		}
	}
}

// TestGreekLowerCorrectionAndAccent asks that the italics correction covers
// the ink above half the height of a letter which reaches beyond its
// advance there, and that the top accent stands over the ink of the top of
// the letter, right of the middle of its advance on a letter whose top
// leans right.
func TestGreekLowerCorrectionAndAccent(t *testing.T) {
	if testing.Short() {
		t.Skip("the Greek is tested in the full tier")
	}
	p := design.Regular()
	for _, g := range greekLowerGlyphs(t) {
		_, _, _, y1 := g.Path.BBox()
		right := rightmostAbove(g.Path, y1/2)
		if right > g.Advance && g.Advance+g.Italic < right {
			t.Errorf("%s: the ink reaches %.0f above half its height, the advance and the correction %.0f", g.Name, right, g.Advance+g.Italic)
		}
		if !g.Attached {
			t.Errorf("%s has no top accent attachment", g.Name)
			continue
		}
		top := p.XHeight
		if y1 > p.XHeight+p.OvershootLower+2 {
			top = p.Ascender
		}
		x0, x1 := inkSpan(g.Path, 0.85*top, y1+1)
		if d := accentLower*p.XHeight + 1; g.Attach < x0-1 || g.Attach > x1+d {
			t.Errorf("%s: the accent attaches at %.0f, outside the ink of its top %.0f..%.0f", g.Name, g.Attach, x0, x1)
		}
		if math.IsNaN(g.Attach) {
			t.Errorf("%s: no ink at its top", g.Name)
		}
	}
}
