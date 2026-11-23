// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The spaces of fixed width U+2000 to U+2008 and U+205F, the compound word
// mark and the visible space.
//
// A space has no outline.  The en quad and the en space are half an em wide,
// the em quad and the em space one em, and the three-per-em, four-per-em and
// six-per-em spaces a third, a quarter and a sixth of the em.  The figure
// space has the advance of the figures and the punctuation space that of the
// full stop.  The medium mathematical space is four eighteenths of the em.
// The compound word mark is empty and has no width.

// emSpace returns the drawing function of a space of the given fraction of
// the em.
func emSpace(frac float64) func(design.Params) (outline.Path, float64) {
	return func(p design.Params) (outline.Path, float64) {
		return nil, math.Round(frac * p.UnitsPerEm)
	}
}

func figureSpace(p design.Params) (outline.Path, float64) { return nil, p.Figures.Advance }

func punctuationSpace(p design.Params) (outline.Path, float64) {
	_, adv := period(p)
	return nil, adv
}

func compoundWordMark(design.Params) (outline.Path, float64) { return nil, 0 }

// Shape constants of the visible space.
const (
	visibleDepth  = 0.57 // the depth of the underside of the bar below the baseline, in descenders
	visibleHeight = 0.51 // the height of the uprights above the underside of the bar, in x-heights
)

// visiblespace draws the visible space: a bar below the baseline with an
// upright at each end, of the stroke of the underscore, one contour.  It has
// the advance of the figures and the side spaces of the en dash.
func visiblespace(p design.Params) (outline.Path, float64) {
	q := p.Punct
	t := q.UnderThick
	adv := p.Figures.Advance
	x0, x1 := q.Dash.Left, adv-q.Dash.Right
	y0 := -math.Round(visibleDepth * p.Descender)
	y1 := y0 + math.Round(visibleHeight*p.XHeight)
	var path outline.Path
	addPolygon(&path,
		xy(x0, y0), xy(x1, y0), xy(x1, y1), xy(x1-t, y1),
		xy(x1-t, y0+t), xy(x0+t, y0+t), xy(x0+t, y1), xy(x0, y1))
	return path, adv
}
