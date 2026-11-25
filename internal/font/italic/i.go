// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package italic

import "seehuhn.de/go/bedern/internal/stroke"

// i draws the i: one stroke from the entry down the stem into the exit, and
// the dot, a circle whose centre lies on the vertical through the centre of
// the stem at the x-height, moved by DotOffset.
func i(p Params) letter {
	e := p.SideStemLeft
	return letter{
		strokes: []stroke.Stroke{p.stem(e, p.XHeight, true)},
		dots:    []dot{{C: p.dotCentre(e), D: p.Dot}},
		advance: e + p.Stem + p.SideStemRight,
	}
}

// dotCentre returns the centre of the dot of a stem with its left edge at e
// (upright, before the shear).  The shear moves a point to the right by the
// slant times its height, so the centre is set back by the slant times its
// distance below the dot, to stand on the vertical through the centre of the
// stem at the x-height, DotOffset to its right.  The long entry stroke weights
// the top of the stem to the left, and the dot moves with it, off the
// slanted axis.
func (p Params) dotCentre(e float64) Point {
	return pt(e+p.Stem/2+p.slant()*(p.XHeight-p.DotCentre)+p.DotOffset, p.DotCentre)
}
