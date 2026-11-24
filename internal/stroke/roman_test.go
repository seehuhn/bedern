// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package stroke

import (
	"testing"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// reproTolerance is the largest distance, in font units, between a reproduced
// part and the outline of the font.
const reproTolerance = 3.0

// TestRomanPartsReproduce builds each part of a roman letter from its stroke
// description and asks that it lies within reproTolerance, or the tolerance of
// the part, of the rounded outline of the font everywhere.
func TestRomanPartsReproduce(t *testing.T) {
	p := design.Regular()
	for _, pt := range RomanParts(p) {
		t.Run(pt.Glyph, func(t *testing.T) {
			t.Parallel()
			font, _, ok := glyphs.Outline(pt.Glyph, p)
			if !ok {
				t.Fatalf("no glyph %s", pt.Glyph)
			}
			rs, err := pt.Build()
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range rs {
				for _, w := range r.Warnings {
					t.Logf("warning: %s", w)
				}
			}
			d, at := Distance(rs, font.Round(), 10)
			t.Logf("%s: largest distance %.2f at %.0f,%.0f", pt.Name, d, at.X, at.Y)
			tol := pt.Tolerance
			if tol == 0 {
				tol = reproTolerance
			}
			if d > tol {
				t.Errorf("%s: largest distance %.2f at %.0f,%.0f, more than %.1f", pt.Name, d, at.X, at.Y, tol)
			}
		})
	}
}
