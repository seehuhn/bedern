// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The numero sign, the estimated sign, the prescription sign and the degree
// Celsius.
//
// The numero sign is the N followed by the letter of the masculine ordinal
// over an underline on the baseline.  The degree Celsius is the degree sign followed by the
// C.  The prescription sign is the R with a stroke across its leg.  The
// estimated sign is drawn from two ellipses and a bar.

// Shape constants of these signs.
const (
	numeroGap   = 77.0  // the white between the right stem of N and the small o of the numero sign
	numeroBase  = 72.0  // the white between the underline and the small o of the numero sign
	numeroSide  = 46.0  // the right side space of the numero sign
	celsiusTuck = 15.0  // how far the C of the degree Celsius stands inside the advance of the degree sign
	estWide     = 1.0   // the width of the estimated sign, in its height
	estSide     = 0.146 // the sides of the estimated sign, in its width
	estOpen     = 37.0  // the end of the lower arc of the estimated sign, degrees below the middle
	estMargin   = 40.0  // the side spaces of the estimated sign
	recipeCross = 0.15  // the height at which the stroke of the prescription sign crosses the middle of the leg, in cap heights
	recipeLean  = 45.0  // the stroke of the prescription sign, degrees above the horizontal
	recipeUp    = 0.26  // how far the stroke of the prescription sign reaches from the crossing up to the right, in cap heights
	recipeBelow = 60.0  // how far the lower end of that stroke stands below the baseline
)

// numero draws the numero sign: the N; numeroGap to the right of its right
// stem the letter of the masculine ordinal, the o drawn heavier and scaled,
// with its lowest ink numeroBase above an underline as thick as the bar of
// e and as wide as the letter, which stands on the baseline; and numeroSide
// to the right of the letter.
func numero(p design.Params) (outline.Path, float64) {
	n, _, _ := Outline("N", p)
	o, _, _ := Outline("o", heavierLetters(p, ordWeight/ordScale))
	x0, y0, x1, _ := o.BBox()
	o = scaledPath(shiftInk(o, -x0, -y0), ordScale)
	w := ordScale * (x1 - x0)
	ox := math.Round(design.CapNStemRight(p) + numeroGap)
	path := append(n, shiftInk(o, ox, p.EBarThick+numeroBase)...)
	rect(&path, ox, 0, ox+w, p.EBarThick, true)
	return path, math.Ceil(ox + w + numeroSide)
}

// celsius draws the degree Celsius: the degree sign and, celsiusTuck inside
// its advance, the C.
func celsius(p design.Params) (outline.Path, float64) {
	d, dadv := degree(p)
	c, cadv := capC(p)
	dx := dadv - celsiusTuck
	return append(d, shiftInk(c, dx, 0)...), dx + cadv
}

// ellipseArc returns the arc of the ellipse about c with the half axes rx
// and ry from the parameter a0 to the parameter a1 (radians;
// counter-clockwise if a1 > a0), cut at the extremes of the ellipse.
func ellipseArc(c outline.Point, rx, ry, a0, a1 float64) chain {
	const eps = 1e-9
	cuts := []float64{a0}
	q := math.Pi / 2
	if a1 > a0 {
		for a := (math.Floor(a0/q+eps) + 1) * q; a < a1-eps; a += q {
			cuts = append(cuts, a)
		}
	} else {
		for a := (math.Ceil(a0/q-eps) - 1) * q; a > a1+eps; a -= q {
			cuts = append(cuts, a)
		}
	}
	cuts = append(cuts, a1)
	on := func(q outline.Point) outline.Point { return outline.Point{X: c.X + rx*q.X, Y: c.Y + ry*q.Y} }
	var ch chain
	for i := 0; i+1 < len(cuts); i++ {
		u := arcCubic(outline.Point{}, 1, cuts[i], cuts[i+1])
		ch = append(ch, curveSeg(cubic{on(u.P0), on(u.P1), on(u.P2), on(u.P3)}))
	}
	return ch
}

// estimated draws the estimated sign, from the overshoot below the baseline
// to the overshoot above the figure height: a round whose outer edge is an
// ellipse estWide times as wide as high and whose counter is an ellipse, so
// that its sides are estSide of its width thick and its top and bottom as
// thick as the thin strokes of the lowercase; a bar of that thickness across
// the middle; and below the bar no right side: the underside of the bar runs
// on to the outer edge, and the lower arc ends estOpen degrees below the
// middle in a cut across it.
func estimated(p design.Params) (outline.Path, float64) {
	top, bottom := p.Figures.Height+p.OvershootLower, -p.OvershootLower
	ry := (top - bottom) / 2
	rx := math.Round(estWide * ry)
	ctr := outline.Point{X: rx, Y: (top + bottom) / 2}
	thin := p.ThinLower
	ix, iy := rx-math.Round(estSide*2*rx), ry-thin
	yb, yt := ctr.Y-thin/2, ctr.Y+thin/2
	end := 2*math.Pi - estOpen*math.Pi/180
	outerA := math.Asin((yb - ctr.Y) / ry)
	innerB := math.Pi - math.Asin((yb-ctr.Y)/iy)
	outer := ellipseArc(ctr, rx, ry, outerA, end)
	inner := ellipseArc(ctr, ix, iy, end, innerB)
	round := join(outer, lineChain(outer.end(), inner.start()), inner, lineChain(inner.end(), outer.start()))
	eyeA := math.Asin((yt - ctr.Y) / iy)
	eye := ellipseArc(ctr, ix, iy, math.Pi-eyeA, eyeA)
	eye = join(eye, lineChain(eye.end(), eye.start()))
	return placeInk(chainsPath(round, eye), design.PunctSides{Left: estMargin, Right: estMargin})
}

// prescription draws the prescription sign: the R with a leg of its own, which
// ends in a flat cut on the baseline and has no foot serif, and a straight
// stroke as thick as the stroke of the dollar across the leg, which rises
// recipeLean degrees to the right, crosses the middle of the leg recipeCross of
// the cap height above the baseline, reaches recipeUp of the cap height from
// the crossing up to the right and ends recipeBelow below the baseline at the
// lower left.
func prescription(p design.Params) (outline.Path, float64) {
	path, adv := capRWith(p, true)
	leg := capRLeg(p)
	h := p.CapHeight
	t := p.Sym.DollarStroke
	l, r := leg.edgeAt(edgeLeft, recipeCross*h), leg.edgeAt(edgeRight, recipeCross*h)
	cx, cy := (l.X+r.X)/2, recipeCross*h
	a := recipeLean * math.Pi / 180
	up := recipeUp * h
	down := (cy + recipeBelow - t/2*math.Cos(a)) / math.Sin(a)
	stroke := taperBar(
		outline.Point{X: cx - down*math.Cos(a), Y: cy - down*math.Sin(a)},
		outline.Point{X: cx + up*math.Cos(a), Y: cy + up*math.Sin(a)}, t)
	return chainsPath(union(append(pathChains(path), stroke)...)...), adv
}
