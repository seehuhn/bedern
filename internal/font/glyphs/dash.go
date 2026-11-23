// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The plain bars: hyphen, figure dash, en dash, em dash, underscore, and the
// vertical bar.
// A bar has flat ends and straight edges.

// horizontalBar returns a bar of the given length and thickness whose left end
// is at the abscissa 0 and whose centre line is at the height centre.
func horizontalBar(length, thick, centre float64) outline.Path {
	var path outline.Path
	addBar(&path, 0, centre-thick/2, length, centre+thick/2)
	return path
}

// hyphen draws the hyphen-minus: a bar at half the x-height.
func hyphen(p design.Params) (outline.Path, float64) {
	q := p.Punct
	return placeInk(horizontalBar(q.HyphenLength, q.HyphenThick, q.HyphenCentre), q.Hyphen)
}

// endash draws the en dash: a bar on the axis of the mathematics which with its
// side space (Dash) is as wide as the advance of the figures.
func endash(p design.Params) (outline.Path, float64) {
	q := p.Punct
	return placeInk(horizontalBar(q.EnLength, q.DashThick, q.DashCentre), q.Dash)
}

// figuredash draws the figure dash: the bar of the en dash, as long as leaves
// the side space of the en dash (Dash) in the advance of the figures.
func figuredash(p design.Params) (outline.Path, float64) {
	q := p.Punct
	length := p.Figures.Advance - q.Dash.Left - q.Dash.Right
	return placeInk(horizontalBar(length, q.DashThick, q.DashCentre), q.Dash)
}

// emdash draws the em dash: a bar which with its side space (EmDash) is twice
// the advance of the figures.
func emdash(p design.Params) (outline.Path, float64) {
	q := p.Punct
	return placeInk(horizontalBar(q.EmLength, q.DashThick, q.DashCentre), q.EmDash)
}

// underscore draws the underscore: a bar as long as its advance, below the
// baseline.
func underscore(p design.Params) (outline.Path, float64) {
	q := p.Punct
	return placeInk(horizontalBar(q.UnderLength, q.UnderThick, -(q.UnderTop+q.UnderThick/2)), q.Underscore)
}

// verticalbar draws the vertical bar over the height of the brackets.
func verticalbar(p design.Params) (outline.Path, float64) {
	q := p.Punct
	var path outline.Path
	addBar(&path, 0, p.BracketBottom(), q.BarThick, p.BracketTop())
	return placeInk(path, q.Bar)
}
