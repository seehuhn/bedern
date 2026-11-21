// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The percent and the number sign.

// percent draws the percent sign: two small zeros and a thin diagonal between
// them, three contours which do not touch.  The zeros are ovals with vertical
// stress in the manner of the figure 0 (PctZeroThick at the sides, PctZeroThin
// at the top and the bottom, the squareness of o), the upper at the left with
// its top at the overshoot of the round figures above Figures.Height, the lower at
// the right with its bottom as far below the baseline.  The diagonal is
// PctSlash thick, rises from the baseline at the left to Figures.Height at the right
// and is cut horizontally at both ends.
//
// Side kinds: open, PctSide to the ink.
func percent(p design.Params) (outline.Path, float64) {
	s := p.Sym
	x0, x1 := s.PctSide, s.PctSide+s.PctInk
	ov := p.OvershootLower
	top := p.Figures.Height + ov
	zero := func(zx0, zy0 float64) (chain, chain) {
		return ringChains(zx0, zy0, zx0+s.PctZeroW, zy0+s.PctZeroH, s.PctZeroThick, s.PctZeroThin, p.SupernessOuter, p.SupernessInner)
	}
	uo, uc := zero(x0, top-s.PctZeroH)
	lo, lc := zero(x1-s.PctZeroW, -ov)

	// The diagonal runs from xb at the baseline to xt at Figures.Height, measured on
	// its centre line; its horizontal width follows from its thickness and its
	// slope.
	xb, xt := x0+0.20*s.PctInk, x1-0.20*s.PctInk
	h := p.Figures.Height
	hw := s.PctSlash / 2 * math.Hypot(xt-xb, h) / h
	slash := chain{
		lineSeg(outline.Point{X: xb - hw, Y: 0}, outline.Point{X: xb + hw, Y: 0}),
		lineSeg(outline.Point{X: xb + hw, Y: 0}, outline.Point{X: xt + hw, Y: h}),
		lineSeg(outline.Point{X: xt + hw, Y: h}, outline.Point{X: xt - hw, Y: h}),
		lineSeg(outline.Point{X: xt - hw, Y: h}, outline.Point{X: xb - hw, Y: 0}),
	}
	return chainsPath(uo, uc, slash, lo, lc), x1 + s.PctSide
}

// numbersign draws the number sign: two strokes NumStroke thick which lean
// NumLean degrees to the right, cut horizontally at the top and the bottom,
// and two horizontal bars NumBar thick across the whole ink, the union of the
// four (symunion.go), centred on NumCentre.
//
// Side kinds: open, NumSide to the ends of the bars.
func numbersign(p design.Params) (outline.Path, float64) {
	s := p.Sym
	x0, x1 := s.NumSide, s.NumSide+s.NumInk
	cx, cy := (x0+x1)/2, s.NumCentre
	y0, y1 := cy-s.NumHeight/2, cy+s.NumHeight/2
	lean := math.Tan(s.NumLean * math.Pi / 180)
	hw := s.NumStroke / 2 / math.Cos(s.NumLean*math.Pi/180)
	quad := func(a, b, c, d outline.Point) chain {
		return chain{lineSeg(a, b), lineSeg(b, c), lineSeg(c, d), lineSeg(d, a)}
	}
	stroke := func(xc float64) chain {
		at := func(y, dx float64) outline.Point { return outline.Point{X: xc + (y-cy)*lean + dx, Y: y} }
		return quad(at(y0, -hw), at(y0, hw), at(y1, hw), at(y1, -hw))
	}
	bar := func(yc float64) chain {
		b, t := yc-s.NumBar/2, yc+s.NumBar/2
		return quad(outline.Point{X: x0, Y: b}, outline.Point{X: x1, Y: b}, outline.Point{X: x1, Y: t}, outline.Point{X: x0, Y: t})
	}
	off := (s.NumGap + s.NumBar) / 2
	parts := union(stroke(cx-s.NumPitch/2), stroke(cx+s.NumPitch/2), bar(cy-off), bar(cy+off))
	return chainsPath(parts...), x1 + s.NumSide
}
