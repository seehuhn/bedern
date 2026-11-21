// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: OFL-1.1

package glyphs

import (
	"math"

	"seehuhn.de/go/bedern/internal/font/design"
	"seehuhn.de/go/bedern/internal/outline"
)

// figOne draws the figure 1: a stem Figures.Stem wide from the baseline to the
// height of the figures, with the bracketed foot serif of the capitals on both
// sides, OneFoot long, and a flag at the upper left.
//
// The flag leaves the top of the stem at its right-hand edge, which is the
// highest point of the figure, and falls to the left.  Its upper edge is one
// curve which leaves the top OneFlagLean below the horizontal and reaches the tip
// at OneFlagAngle, and is hollow; the tip is a vertical cut OneFlagThick
// long whose middle lies OneFlagDrop below the top and OneFlag left of the stem;
// the lower edge leaves the tip parallel to the upper edge and runs in one
// concave curve, a bracket, into the left-hand edge of the stem, which it meets
// vertically OneFlagJoin below the point where the line of its start meets the
// stem.
//
// The contour runs from the tip of the left foot serif along the baseline, round
// the right foot serif, up the right-hand edge to the top, along the upper edge
// of the flag, down the tip, along the lower edge to the stem and down the
// left-hand edge round the left foot serif.
func figOne(p design.Params) (outline.Path, float64) {
	f := p.Figures
	// The ink is the flag on the left and the foot serif on the right.
	x0 := design.FigureLeft(p, 1)
	xl := x0 + math.Max(f.OneFlag, f.OneFoot)
	xr := xl + f.Stem
	h := f.Height + f.OneRise
	cp := capSerifParams(p)

	right := capFootRight(p, xr, f.OneFoot)
	left := capFootLeft(p, xl, f.OneFoot)

	a := degrees(f.OneFlagAngle)
	tipX := xl - f.OneFlag
	tipMid := h - f.OneFlagDrop
	t1 := outline.Point{X: tipX, Y: tipMid + f.OneFlagThick/2}
	t2 := outline.Point{X: tipX, Y: tipMid - f.OneFlagThick/2}
	top := outline.Point{X: xr, Y: h}
	lean := degrees(f.OneFlagLean)
	// The upper edge, walked from the top to the tip: it leaves the top falling
	// at lean and arrives at the tip falling at a.
	up := quarter(top, t1, outline.Point{X: -math.Cos(lean), Y: -math.Sin(lean)}, outline.Point{X: -math.Cos(a), Y: -math.Sin(a)}, 0.5)
	// The lower edge, from the tip to the stem: the line of its start meets the
	// stem at the height m, and the edge runs into the stem OneFlagJoin below.
	m := t2.Y + (xl-tipX)*math.Tan(a)
	join0 := outline.Point{X: xl, Y: m - f.OneFlagJoin}
	low := quarter(t2, join0, outline.Point{X: math.Cos(a), Y: math.Sin(a)}, outline.Point{X: 0, Y: -1}, 0.55)

	c := join(
		lineChain(left.end(), right.start()),
		right,
		lineChain(right.end(), top),
		curveChain(up),
		lineChain(t1, t2),
		curveChain(low),
		lineChain(join0, outline.Point{X: xl, Y: cp.BracketHeight}),
		left,
	)
	return contourPath(mergeLines(c)), f.Advance
}
