// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The signs of ASCII mathematics: plus, minus, equals, less-than and
// greater-than.
//
// All five stand on one axis (MathAxis, 250 above the baseline), have strokes of
// one thickness (MathThick) with flat ends, and one advance
// (design.Params.MathAdvance).  The plus is a square of MathWidth, and the minus
// and the equals are as wide.  The less-than and greater-than are narrower
// (0.92 of the plus): two strokes at MathAngle degrees to the horizontal which
// meet at a vertex cut flat, with vertical cuts at the free ends.  The ink of a
// sign is centred in the advance.  The hyphen-minus is the hyphen; the minus
// sign U+2212 is the bar of the plus.

// placeCentred moves the path sideways so that its ink is centred in the advance
// adv, to a whole number of units.
func placeCentred(path outline.Path, adv float64) (outline.Path, float64) {
	x0, _, x1, _ := path.BBox()
	return shiftInk(path, math.Round((adv-(x0+x1))/2), 0), adv
}

// mathBar returns the horizontal bar of the minus: MathWidth long, MathThick
// thick, on the axis.
func mathBar(p design.Params, centre float64) outline.Path {
	q := p.Punct
	return horizontalBar(q.MathWidth, q.MathThick, centre)
}

// plus draws the plus sign.
func plus(p design.Params) (outline.Path, float64) {
	q := p.Punct
	w, t, ay := q.MathWidth/2, q.MathThick/2, q.MathAxis
	var path outline.Path
	addPolygon(&path,
		xy(-t, ay-w), xy(t, ay-w), xy(t, ay-t), xy(w, ay-t), xy(w, ay+t), xy(t, ay+t),
		xy(t, ay+w), xy(-t, ay+w), xy(-t, ay+t), xy(-w, ay+t), xy(-w, ay-t), xy(-t, ay-t))
	return placeCentred(path, p.MathAdvance())
}

// minus draws the minus sign.
func minus(p design.Params) (outline.Path, float64) {
	return placeCentred(mathBar(p, p.Punct.MathAxis), p.MathAdvance())
}

// equal draws the equals sign: two bars, MathGap apart, about the axis.
func equal(p design.Params) (outline.Path, float64) {
	q := p.Punct
	pitch := (q.EqualsGap + q.MathThick) / 2
	var path outline.Path
	addBar(&path, 0, q.MathAxis-pitch-q.MathThick/2, q.MathWidth, q.MathAxis-pitch+q.MathThick/2)
	addBar(&path, 0, q.MathAxis+pitch-q.MathThick/2, q.MathWidth, q.MathAxis+pitch+q.MathThick/2)
	return placeCentred(path, p.MathAdvance())
}

// lessPath returns the less-than sign with its vertex at the left.
func lessPath(p design.Params) outline.Path {
	q := p.Punct
	beta := degrees(q.MathAngle)
	tan, sin := math.Tan(beta), math.Sin(beta)
	v := q.MathThick / (2 * math.Cos(beta)) // the vertical half thickness of a stroke
	ay := q.MathAxis
	// The outer edges meet at the abscissa -MathThick/(2 sin beta); the vertex is cut
	// by a vertical line where they are MathVertexFlat apart.
	xl := -q.MathThick/(2*sin) + q.MathVertexFlat/(2*tan)
	xr := xl + q.MathAngleWidth
	xi := q.MathThick / (2 * sin) // the crotch of the inner edges
	var path outline.Path
	addPolygon(&path,
		xy(xl, ay+q.MathVertexFlat/2), xy(xl, ay-q.MathVertexFlat/2),
		xy(xr, ay-tan*xr-v), xy(xr, ay-tan*xr+v), xy(xi, ay),
		xy(xr, ay+tan*xr-v), xy(xr, ay+tan*xr+v))
	return path
}

// less draws the less-than sign.
func less(p design.Params) (outline.Path, float64) {
	return placeCentred(lessPath(p), p.MathAdvance())
}

// greater draws the greater-than sign: the mirror image of the less-than sign.
func greater(p design.Params) (outline.Path, float64) {
	return placeCentred(mirrorInk(lessPath(p), 0), p.MathAdvance())
}
