// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package upper

import (
	"seehuhn.de/go/bedern/internal/font/glyphs"
)

// The hyphens U+2010 and U+2011, the figure dash, the horizontal bar, the
// third prime, the per ten thousand sign with the zero of the per mille sign,
// the visible space and the capital schwa, made as the punctuation is: sloped
// from the roman with their rounds turned.  The bars of the hyphens and
// dashes are sloped like the hyphen of ASCII, whose ends the shear leans.
// The visible space is sloped too: its uprights lean with the text it
// stands in.  The zeros of the per mille signs turn as those of the percent
// sign do, and the eye of the capital schwa as the counter of O does.
//
// The sides of the hyphens are those of the hyphen and the figure dash and
// the bar are those of the en dash and the em dash, as in the roman
// (the sides are the fit table's).
func init() {
	for _, g := range []struct {
		name string
		r    rune
	}{
		{"uni2010", '‐'}, {"uni2011", '‑'},
		{"figuredash", '‒'}, {"uni2015", '―'},
	} {
		sources[g.name] = source{r: g.r, dots: punctDots}
	}
	sources["uni2034"] = source{r: '‴'}
	sources["uni2031"] = source{r: '‱', rounds: allCounters("uni2031")}
	sources["perthousandzero"] = source{r: privateUse("perthousandzero"), rounds: allCounters("perthousandzero")}
	sources["uni2423"] = source{r: '␣'}
	sources["uni018F"] = source{r: 'Ə', rounds: allCounters("uni018F")}
}

// privateUse returns the code point of the glyph of the roman in the Private
// Use Area.
func privateUse(name string) rune {
	for _, e := range glyphs.PrivateUseTable() {
		if e.Name == name {
			return e.Rune
		}
	}
	panic("upper: " + name + " is not in the table of the Private Use Area")
}
