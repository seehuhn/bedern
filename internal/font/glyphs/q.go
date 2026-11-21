// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// qBowl returns the bowl of q: the bowl of d, on the left of its stem.
func qBowl(p design.Params) bowlSpec {
	return bowlFrom(p, p.QBowl, design.QStemLeft(p), design.DBowlLeft(p))
}

// q draws the lowercase letter q as one outer contour and one counter: the bowl
// of d on the left of a stem which runs from the x-height down to the descender
// line, where it has a foot serif on both sides.  The right-hand edge of the
// stem rises QSpur above QStemTop to a slight point, the spur, and the top is cut
// by one straight line which falls to the left, to the upper notch of the bowl,
// where the upper stroke of the bowl leaves it in a corner.  Nothing points
// to the right beyond the stem and there is no serif.  The outer contour starts
// at the left tip of the left foot serif, goes up the right-hand edge of the
// stem, down the cut to the upper notch of the bowl, round the bowl and down
// the left-hand edge to the foot.
func q(p design.Params) (outline.Path, float64) {
	stem := descenderStem(p, design.QStemLeft(p), p.QStemTop+p.QSpur, false)
	bowl := qBowl(p)

	var path outline.Path
	path.MoveTo(stem.start())
	stemUp(&path, p, stem, stem.Top)
	path.LineTo(stem.Left, bowl.NotchUpper)
	bowlOuter(dBowlShape, bowl).appendTo(&path)
	stemDown(&path, p, stem)
	bowlCounter(dBowlShape, bowl).addContour(&path)
	return path, design.QAdvance(p)
}
