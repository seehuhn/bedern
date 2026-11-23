// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The signs of genealogy: born, died, married and divorced.
//
// Born is a star of five points, one pointing up: each point is an arm with
// straight edges from the notches beside it to its end, which is cut flat at
// right angles to the arm.  The middle of its ink lies on the maths axis, and
// it stands in the advance of the figures.
//
// Died is a cross which stands on the baseline and reaches the cap height:
// an upright stem and two arms, each limb with straight edges which widen
// from the crossing to its end, which is cut flat; the arms are thicker at
// the crossing than the stem.
//
// Married is two rings of the stroke of the mathematical signs
// (Punct.MathThick), as high as the x-height, which overlap; the outline is
// their union.  Divorced is two upright ovals of that height and stroke with
// an upright bar between them, divorcedBar strokes thick and
// divorcedHeight ovals high.  The middle of both signs lies on the maths axis.

// Shape constants of the signs of genealogy.
const (
	bornSide  = 5.0  // the side spaces of the star
	bornNotch = 1.95 // the notches of the star from its middle, in thin strokes of the lowercase
	bornTip   = 0.85 // the flat at the end of a point of the star, in thin strokes of the lowercase

	diedWidth  = 0.70 // the ink width of the cross, in cap heights
	diedArms   = 0.65 // the middle of the arms of the cross, in cap heights
	diedWaist  = 1.05 // the stem of the cross at the crossing, in strokes of the mathematical signs
	diedArm    = 1.2  // the arms of the cross at the crossing, in strokes of the mathematical signs
	diedEnd    = 1.7  // the stem of the cross at its foot and its top, in strokes of the mathematical signs
	diedArmEnd = 1.5  // the arms of the cross at their ends, in strokes of the mathematical signs

	marriedOverlap = 0.44 // the overlap of the two rings, in their diameter
	divorcedOval   = 0.75 // the width of an oval, in its height
	divorcedGap    = 85.0 // the white between an oval and the bar
	divorcedBar    = 0.75 // the thickness of the bar, in strokes of the mathematical signs
	divorcedHeight = 1.7  // the height of the bar, in heights of an oval

	pictureSide = 40.0 // the side spaces of the cross, the rings and the ovals
)

// placeOnMathAxis returns the path moved up or down, by a whole number of
// units, so that the middle of its ink lies on the maths axis.
func placeOnMathAxis(p design.Params, path outline.Path) outline.Path {
	_, y0, _, y1 := path.BBox()
	return shiftInk(path, 0, math.Round(p.Punct.MathAxis-(y0+y1)/2))
}

// born draws the sign for the year of birth.
func born(p design.Params) (outline.Path, float64) {
	const points = 5
	w := p.Figures.Advance - 2*bornSide
	flat := bornTip * p.ThinLower
	notch := bornNotch * p.ThinLower
	// The ends of the two points beside the upper one are the leftmost and
	// the rightmost ink.
	side := math.Pi/2 - 2*math.Pi/points
	r := (w/2 - flat/2*math.Sin(side)) / math.Cos(side)
	var pts []outline.Point
	for i := range points {
		a := math.Pi/2 + float64(i)*2*math.Pi/points
		u, n := dirOf(a), dirOf(a+math.Pi/2)
		v := dirOf(a + math.Pi/points)
		pts = append(pts,
			xy(r*u.X-flat/2*n.X, r*u.Y-flat/2*n.Y),
			xy(r*u.X+flat/2*n.X, r*u.Y+flat/2*n.Y),
			xy(notch*v.X, notch*v.Y))
	}
	var path outline.Path
	addPolygon(&path, pts...)
	return placeCentred(placeOnMathAxis(p, path), p.Figures.Advance)
}

// died draws the sign for the year of death.
func died(p design.Params) (outline.Path, float64) {
	h := p.CapHeight
	w := math.Round(diedWidth*h) / 2
	ym := math.Round(diedArms * h)
	t := p.Punct.MathThick
	c, e := diedWaist*t/2, diedEnd*t/2
	b, a := diedArm*t/2, diedArmEnd*t/2
	var path outline.Path
	addPolygon(&path,
		xy(-e, 0), xy(e, 0), xy(c, ym-b), xy(w, ym-a), xy(w, ym+a), xy(c, ym+b),
		xy(e, h), xy(-e, h), xy(-c, ym+b), xy(-w, ym+a), xy(-w, ym-a), xy(-c, ym-b))
	return placeInk(path, design.PunctSides{Left: pictureSide, Right: pictureSide})
}

// married draws the sign for a marriage.
func married(p design.Params) (outline.Path, float64) {
	d := math.Round(p.XHeight)
	t := p.Punct.MathThick
	var parts []chain
	for _, x := range []float64{d / 2, d/2 + math.Round((1-marriedOverlap)*d)} {
		c := xy(x, 0)
		parts = append(parts, circleChain(c, d/2), circleChain(c, d/2-t).reversed())
	}
	path := placeOnMathAxis(p, chainsPath(union(parts...)...))
	return placeInk(path, design.PunctSides{Left: pictureSide, Right: pictureSide})
}

// divorced draws the sign for a divorce.
func divorced(p design.Params) (outline.Path, float64) {
	h := math.Round(p.XHeight)
	w := math.Round(divorcedOval * h)
	t := p.Punct.MathThick
	long := math.Round(divorcedHeight * h)
	bar := math.Round(divorcedBar * t)
	var parts []chain
	for _, x := range []float64{0, w + 2*divorcedGap + bar} {
		outer, counter := ringChains(x, -h/2, x+w, h/2, t, t, degSuperness, degSuperness)
		parts = append(parts, outer, counter)
	}
	parts = append(parts, rectChain(w+divorcedGap, -long/2, w+divorcedGap+bar, long/2))
	path := placeOnMathAxis(p, chainsPath(parts...))
	return placeInk(path, design.PunctSides{Left: pictureSide, Right: pictureSide})
}
