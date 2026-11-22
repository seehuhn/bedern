// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The signs made of the dot of i: full stop, colon, semicolon and exclamation
// mark, and the straight quotation marks, which are tapered bars of the same
// family as the stem of the exclamation mark.
//
// Every dot is the dot of i: a circle of DotDiameter across drawn by the code of
// i.  The full stop lies a little below the baseline (Punct.DotLow); the upper
// dot of the colon and the semicolon has its top a little above the x-height; the
// dot of the exclamation mark is the full stop.

// period draws the full stop.
func period(p design.Params) (outline.Path, float64) {
	var path outline.Path
	addDot(&path, p, 0, p.PeriodCentre())
	return placeInk(path, p.Punct.Period)
}

// colon draws the colon: the full stop under a dot whose top is level with the
// x-height.
func colon(p design.Params) (outline.Path, float64) {
	var path outline.Path
	addDot(&path, p, 0, p.PeriodCentre())
	addDot(&path, p, 0, p.ColonUpperCentre())
	return placeInk(path, p.Punct.Colon)
}

// semicolon draws the semicolon: the upper dot of the colon over the comma.
func semicolon(p design.Params) (outline.Path, float64) {
	var path outline.Path
	commaAt(p, 0, commaTop(p)).addContour(&path)
	addDot(&path, p, 0, p.ColonUpperCentre())
	return placeOnAxis(path, p.DotsAdvance())
}

// taperedBar adds a bar which is wide at the top and narrow at the foot, with flat
// ends, centred on the abscissa cx, from the height bottom to the height top, as a
// counter-clockwise contour.
func taperedBar(path *outline.Path, cx, bottom, top, topWidth, bottomWidth float64) {
	addPolygon(path,
		xy(cx-bottomWidth/2, bottom), xy(cx+bottomWidth/2, bottom),
		xy(cx+topWidth/2, top), xy(cx-topWidth/2, top))
}

// exclam draws the exclamation mark: a stem tapering downwards over the full stop.
func exclam(p design.Params) (outline.Path, float64) {
	q := p.Punct
	var path outline.Path
	addDot(&path, p, 0, p.PeriodCentre())
	taperedBar(&path, 0, p.ExclamFoot(), p.ExclamTop(), q.ExclamTopWidth, q.ExclamBottomWidth)
	return placeInk(path, q.Exclam)
}

// quotesingle draws the straight apostrophe: a tapered bar at the height of the
// ascender.
func quotesingle(p design.Params) (outline.Path, float64) {
	q := p.Punct
	var path outline.Path
	taperedBar(&path, 0, q.StraightTop-q.StraightHeight, q.StraightTop, q.StraightTopWidth, q.StraightBottomWidth)
	return placeInk(path, q.Straight)
}

// quotedbl draws the straight double quotation mark: two of the bars of the
// apostrophe.
func quotedbl(p design.Params) (outline.Path, float64) {
	q := p.Punct
	var path outline.Path
	for _, cx := range []float64{0, q.StraightPitch} {
		taperedBar(&path, cx, q.StraightTop-q.StraightHeight, q.StraightTop, q.StraightTopWidth, q.StraightBottomWidth)
	}
	return placeInk(path, q.StraightDouble)
}
