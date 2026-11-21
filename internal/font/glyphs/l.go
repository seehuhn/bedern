// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// l draws the lowercase letter l: a stem of ascender height with a foot
// serif on both sides and a sloped head serif on the left.
func l(p design.Params) (outline.Path, float64) {
	stem := stemSpec{
		Left: design.NarrowStemLeft(p) + design.AscenderInset(p), Width: p.AscenderStem, Top: p.Ascender,
		FootLeft: p.SerifJut, FootRight: p.SerifJut,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	var path outline.Path
	stemAlone(&path, p, stem)
	return path, design.NarrowAdvance(p)
}
