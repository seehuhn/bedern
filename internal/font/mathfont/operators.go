// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package mathfont

import "seehuhn.de/go/bedern/internal/font/design"

// operatorGlyphs returns the big operators (bigops.go), the integral
// (integral.go) and the radical (radical.go) with their sizes and parts, in
// the order of the font.
func operatorGlyphs(p design.Params) ([]Glyph, error) {
	out := bigOpGlyphs(p)
	ig, err := integralGlyphs(p)
	if err != nil {
		return nil, err
	}
	out = append(out, ig...)
	rg, err := radicalGlyphs(p)
	if err != nil {
		return nil, err
	}
	return append(out, rg...), nil
}
