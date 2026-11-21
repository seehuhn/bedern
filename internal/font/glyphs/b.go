// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// bBowl returns the bowl of b: the bowl of d on the right of the stem, as the
// mirror image in the edge of the stem.  Its free side is the right half of the
// o, and its extreme lies where the cell of the stem and the width of the bowl
// put it (design.BBowlRight).
func bBowl(p design.Params) bowlSpec {
	b := bowlFrom(p, p.BBowl, design.BStemRight(p), design.BBowlRight(p))
	b.OnRight = true
	return b
}

// b draws the lowercase letter b as one outer contour and one counter: an
// ascender stem on the left with the head serif of l, h and d at the ascender
// height, and on its right the mirrored bowl of d.  The stem ends at the
// baseline in a plain corner, without a foot serif or a spur.  The outer
// contour goes up the right-hand edge of the stem to the lower notch, round the
// bowl to the upper notch, on up the stem, along the head serif and down the
// left-hand edge to the baseline.
func b(p design.Params) (outline.Path, float64) {
	stem := stemSpec{
		Left: design.BStemLeft(p), Width: p.AscenderStem, Top: p.Ascender,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	bowl := bBowl(p)

	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, bowl.NotchLower)
	bowlOuter(dBowlShape, bowl).appendTo(&path)
	stemCap(&path, p, stem)
	stemDown(&path, p, stem)
	bowlCounter(dBowlShape, bowl).addContour(&path)
	return path, design.BAdvance(p)
}
