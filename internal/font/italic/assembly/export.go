// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package assembly

import (
	"seehuhn.de/go/bedern/internal/font/glyphs"
	"seehuhn.de/go/bedern/internal/outline"
)

// The fit of the italic for packages which take its letters: the sides of
// the fit table, the measure of the sides and the fit itself, and the dotless
// letters.

// FitSides returns the side spaces of the glyph of the given name in the fit
// table, and false for a glyph which the table does not list.
func FitSides(name string) (Sides, bool) {
	s, ok := fitTable[name]
	return s, ok
}

// SidesOf returns the sides of the glyph as the fit table measures them, by
// the band of its name, and false for a glyph without ink.
func SidesOf(g glyphs.Glyph) (Sides, bool) { return sidesOf(g) }

// Fitted returns the glyph moved and with the advance which give it the
// sides s, measured as SidesOf measures them.
func Fitted(g glyphs.Glyph, s Sides) glyphs.Glyph { return fitted(g, s) }

// WithoutDot returns the outline without the contours which lie wholly above
// the x-height xh and a little more: i and j without their dot.
func WithoutDot(path outline.Path, xh float64) outline.Path {
	keep, _ := withoutDot(path, xh)
	return keep
}
