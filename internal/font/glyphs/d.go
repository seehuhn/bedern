// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// dBowlShape is the shape of the connection of the bowl of d with its stem:
// the lengths of the handles as fractions of the handles of the parabola.
var dBowlShape = bowlShape{
	OuterStemHandle:    0.9,
	OuterExtremeHandle: 1.2,
	InnerStemHandle:    0.9,
	InnerExtremeHandle: 1.2,
	Parabolic:          true,
}

// dBowl returns the bowl of d.  Its free side is the left half of the o: it
// has the left extreme, the half width, the thin parts, the thickest part,
// the overshoot and the squareness of the o.
func dBowl(p design.Params) bowlSpec {
	return bowlFrom(p, p.DBowl, design.DStemLeft(p)+design.AscenderInset(p), design.DBowlLeft(p))
}

// d draws the lowercase letter d as one outer contour and one counter: an
// ascender stem on the right with a head serif on the left at the ascender
// height and a foot serif on the right only, and on its left the bowl.  The
// stem has no serif on the left at the baseline, where the bowl is.  The outer
// contour goes up the right-hand edge of the stem, along the head serif and
// down the left-hand edge to the upper notch of the bowl, round the bowl and
// down the left-hand edge to the baseline.
func d(p design.Params) (outline.Path, float64) {
	stem := stemSpec{
		Left: design.DStemLeft(p) + design.AscenderInset(p), Width: p.AscenderStem, Top: p.Ascender,
		FootRight: p.SerifJut,
		HeadJut:   p.SerifJut, HeadDrop: p.HeadSerifDrop,
	}
	bowl := dBowl(p)

	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, stem.Top)
	stemCap(&path, p, stem)
	path.LineTo(stem.Left, bowl.NotchUpper)
	bowlOuter(dBowlShape, bowl).appendTo(&path)
	stemDown(&path, p, stem)
	bowlCounter(dBowlShape, bowl).addContour(&path)
	return path, design.DAdvance(p)
}
