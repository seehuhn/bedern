// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"
	"sync"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// The old-style figures and the proportional figures.
//
// A figure has four forms.  The lining tabular figures are those of
// figures.go, at the code points.  The old-style tabular figures
// (.taboldstyle) stand on the same advance.  The proportional forms (.prop,
// lining, and .oldstyle, old-style) are the outlines of the tabular forms,
// moved by a whole number of units, between the side spaces of
// design.FigureSides.
//
// The old-style figures are drawn by the constructions of the lining figures
// with other heights and proportions (oldParams).  The figures 0, 1 and 2
// stand between the baseline and the x-height; 3, 4, 5, 7 and 9 have their
// tops at the x-height and descend to design.OldFigureDepth; 6 and 8 are the lining
// figures.  The flat tops are at the x-height, the round tops OvershootLower
// above it, and the round bottoms OvershootLower below the baseline or below
// the depth.

// liningDraw holds the constructions of the lining figures.
var liningDraw = [10]func(p design.Params) (outline.Path, float64){
	figZero, figOne, figTwo, figThree, figFour, figFive, figSix, figSeven, figEight, figNine,
}

// figureKey names a tabular figure in one set of parameters.
type figureKey struct {
	p   design.Params
	d   int
	old bool
}

// figureBuild is a tabular figure: its outline and its advance.
type figureBuild struct {
	path    outline.Path
	advance float64
}

// figureBuilds holds the tabular figures by the parameters, the figure and
// the form, so that the tabular and the proportional form of a figure in one
// set of parameters are drawn once.
var figureBuilds sync.Map // figureKey -> func() figureBuild

// tabularFigure returns the tabular figure d, old-style if old is set and
// lining otherwise.  An old-style figure is the figure drawn with oldParams,
// moved to its height.  The path is shared and must not be changed.
func tabularFigure(p design.Params, d int, old bool) figureBuild {
	f, _ := figureBuilds.LoadOrStore(figureKey{p, d, old}, sync.OnceValue(func() figureBuild {
		if !old {
			path, adv := liningDraw[d](p)
			return figureBuild{path, adv}
		}
		q, dy := oldParams(p, d)
		path, adv := liningDraw[d](q)
		return figureBuild{shiftInk(path, 0, dy), adv}
	}))
	return f.(func() figureBuild)()
}

// figure returns the function which draws the figure d in one of its four
// forms: old-style if old is set, proportional if prop is set.
func figure(d int, old, prop bool) func(p design.Params) (outline.Path, float64) {
	return func(p design.Params) (outline.Path, float64) {
		b := tabularFigure(p, d, old)
		if !prop {
			return shiftInk(b.path, 0, 0), b.advance
		}
		return placeInk(b.path, design.FigureSides(p, d, old))
	}
}

// oldParams returns the parameters with which the construction of the lining
// figure d draws the old-style figure, and the height by which the figure so
// drawn is moved.
//
// The figures 6 and 8 are the lining figures.  For the others the round tops
// and bottoms overshoot by OvershootLower.  The figures 3, 4, 5, 7 and 9 are
// drawn as high as they stand, from the feet of 4 and 7 to the x-height,
// with the heights of their parts in proportion, and moved down to those
// feet.  The figures 0, 1 and 2 are
// drawn as high as the x-height with the proportions of Old.  The optical
// shift of each is moved by less than a unit, so that the left side space
// which design.FigureLeft gives the figure is a whole number.
func oldParams(p design.Params, d int) (design.Params, float64) {
	if d == 6 || d == 8 {
		return p, 0
	}
	q, dy := oldHeights(p, d)
	left := design.FigureLeft(q, d)
	q.Figures.Shift[d] += math.Round(left) - left
	return q, dy
}

// oldHeights returns the parameters of oldParams before the figure is moved
// to a whole side space.
func oldHeights(p design.Params, d int) (design.Params, float64) {
	o := p.Figures.Old
	depth := design.OldFigureDepth(p) - p.OvershootLower
	q := p
	f := &q.Figures
	f.OverTop = p.OvershootLower
	q.OvershootUpper = p.OvershootLower
	switch d {
	case 0:
		f.Height = p.XHeight
		f.Round = o.ZeroRound
		f.Thin = o.ZeroThin
	case 1:
		f.Height = p.XHeight
		f.OneRise = o.OneRise
		f.OneFlag = o.OneFlag
		f.OneFlagDrop = o.OneFlagDrop
		f.OneFlagLean = o.OneFlagLean
		f.OneFoot = o.OneFoot
	case 2:
		f.Height = p.XHeight
		f.Ink[2] = o.TwoInk
		f.TwoBowlMid = o.TwoBowlMid
		f.TwoShoulder = o.TwoShoulder
		f.TwoOuterDrop = o.TwoOuterDrop
		f.TwoDiagFoot = o.TwoDiagFoot
		f.TwoDiagAngle = o.TwoDiagAngle
		f.TwoBallLeft = o.TwoBallLeft
		f.TwoBowlLeft = o.TwoBowlLeft
		f.Ball = o.TwoBall
		q.BallStalk = p.BallStalk * p.Figures.Ball / o.TwoBall
		f.TwoRise = o.TwoRise
		f.TwoBase = o.TwoBase
		f.TwoBaseEnd = o.TwoBaseEnd
		f.Heavy = o.TwoHeavy
	case 4:
		// The underside of the bar is on the baseline and the top as high
		// as the round tops; the diagonal leans so that the stem keeps its
		// place in the ink.
		g := p.Figures
		f.Height = p.XHeight + p.OvershootLower + depth
		f.FourBarLow = depth
		lean := math.Tan(degrees(g.FourAngle)) * (g.Height - g.FourBarLow) / (f.Height - f.FourBarLow)
		f.FourAngle = math.Atan(lean) * 180 / math.Pi
		return q, -depth
	default:
		f.Height = p.XHeight + depth
		k := f.Height / p.Figures.Height
		for _, v := range []*float64{&f.ThreeWaist, &f.FiveBowlTop, &f.SixBowlTop, &f.SixLeftY} {
			*v = math.Round(*v * k)
		}
		return q, -depth
	}
	return q, 0
}

// figureList lists the old-style figures and the proportional figures, in
// glyph order.
var figureList = func() []entry {
	names := [10]string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
	var es []entry
	for _, form := range []struct {
		suffix    string
		old, prop bool
	}{
		{".taboldstyle", true, false},
		{".oldstyle", true, true},
		{".prop", false, true},
	} {
		for d, name := range names {
			es = append(es, entry{name + form.suffix, 0, sideFigure, sideFigure, figure(d, form.old, form.prop)})
		}
	}
	return es
}()
