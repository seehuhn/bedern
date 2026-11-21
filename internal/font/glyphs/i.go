// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// i draws the lowercase letter i: the stem of l at the x-height, with the head
// serif on the left and a foot serif on both sides, and above it the dot, a
// separate counter-clockwise contour.  The dot is a superellipse whose
// squareness is DotSuperness, which is that of a circle in the regular
// parameters, with the centre design.DotOffset to the right of the centre of the stem (a
// negative value: DotShift times the length of the head serif to the left).
func i(p design.Params) (outline.Path, float64) {
	stem := stemSpec{
		Left: design.NarrowStemLeft(p), Width: p.Stem, Top: p.XHeight,
		FootLeft: p.SerifJut, FootRight: p.SerifJut,
		HeadJut: p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	var path outline.Path
	stemAlone(&path, p, stem)
	cx := stem.Left + p.Stem/2 + design.DotOffset(p)
	r := p.DotDiameter / 2
	superEllipse(&path, cx-r, p.DotCentre-r, cx+r, p.DotCentre+r, p.DotSuperness, true)
	return path, design.NarrowAdvance(p)
}
