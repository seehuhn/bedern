// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// figFour draws the closed figure 4: a heavy stem, Figures.Stem wide, from the
// baseline to the height of the figures with the bracketed foot serif of the
// capitals on both sides (FourFoot); a thin diagonal (FourDiag at right angles,
// FourAngle from the vertical) from the left end of the bar, which is the ink
// left, up to the top, where a short flat (FourTopFlat) joins its outer edge to
// the right-hand edge of the stem; and a thin bar (FourBar, its underside at
// FourBarLow) from the diagonal through the stem to the ink right, cut square.
// The counter is the triangle between the inner edge of the diagonal, the top of
// the bar and the left-hand edge of the stem.  The stem stands where the angle
// and the flat put it.
//
// The contour runs from the tip of the left foot serif round the right foot, up
// the right-hand edge to the bar, round the end of the bar, up to the top, along
// the flat, down the outer edge of the diagonal to the left end of the bar,
// along its underside to the stem and down the left-hand edge round the left
// foot; the counter is a second contour, clockwise.
func figFour(p design.Params) (outline.Path, float64) {
	f := p.Figures
	x0 := design.FigureLeft(p, 4)
	h := f.Height
	tan := math.Tan(degrees(f.FourAngle))
	sr := x0 + (h-f.FourBarLow)*tan + f.FourTopFlat // the right-hand edge of the stem
	sl := sr - f.Stem
	end := x0 + f.Ink[4]
	low, high := f.FourBarLow, f.FourBarLow+f.FourBar
	cp := capSerifParams(p)

	right := capFootRight(p, sr, f.FourFoot)
	left := capFootLeft(p, sl, f.FourFoot)
	flat := outline.Point{X: sr - f.FourTopFlat, Y: h}
	c := join(
		lineChain(left.end(), right.start()),
		right,
		lineChain(right.end(), outline.Point{X: sr, Y: low}),
		lineChain(outline.Point{X: sr, Y: low}, outline.Point{X: end, Y: low}),
		lineChain(outline.Point{X: end, Y: low}, outline.Point{X: end, Y: high}),
		lineChain(outline.Point{X: end, Y: high}, outline.Point{X: sr, Y: high}),
		lineChain(outline.Point{X: sr, Y: high}, outline.Point{X: sr, Y: h}),
		lineChain(outline.Point{X: sr, Y: h}, flat),
		lineChain(flat, outline.Point{X: x0, Y: low}),
		lineChain(outline.Point{X: x0, Y: low}, outline.Point{X: sl, Y: low}),
		lineChain(outline.Point{X: sl, Y: low}, outline.Point{X: sl, Y: cp.BracketHeight}),
		left,
	)
	// The inner edge of the diagonal lies FourDiag at right angles from the
	// outer one, which is the horizontal distance d.
	d := f.FourDiag / math.Cos(degrees(f.FourAngle))
	xin := func(y float64) float64 { return x0 + d + (y-low)*tan }
	apex := low + (sl-x0-d)/tan
	counter := chain{
		lineSeg(outline.Point{X: sl, Y: apex}, outline.Point{X: sl, Y: high}),
		lineSeg(outline.Point{X: sl, Y: high}, outline.Point{X: xin(high), Y: high}),
		lineSeg(outline.Point{X: xin(high), Y: high}, outline.Point{X: sl, Y: apex}),
	}
	return mergedPath(c, counter), f.Advance
}
